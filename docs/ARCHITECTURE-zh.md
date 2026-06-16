# GoTutor 架构

> [English](ARCHITECTURE.md) | 中文

GoTutor 让学习者在真实迷你项目里填补 `// TODO` 空缺来学习 Go。应用
通过**真正编译并运行学习者的 Go 代码**来校验答案——用 `go test`，绝不
靠字符串匹配。本文档追踪一次提交的端到端流程。

```
┌────────────────────────────────────────────────────────────────┐
│ Electron main process (electron/src/main.ts)                   │
│  ┌─────────────────────────┐    ┌───────────────────────────┐  │
│  │ BrowserWindow           │    │ spawn Go binary sidecar   │  │
│  │  ┌───────────────────┐  │    │  (gotutor-backend)       │  │
│  │  │ Vue 3 SPA         │  │    │                           │  │
│  │  │  ┌─────────────┐  │  │    │ Gin :8081-8090           │  │
│  │  │  │ Monaco      │──┼──┼────┼──► HTTP /api/*           │  │
│  │  │  │ Editor      │  │  │    │                           │  │
│  │  │  └─────────────┘  │  │    │ SQLite (progress.db)     │  │
│  │  └───────────────────┘  │    │ Verifier (tempdir+go test)│  │
│  └─────────────────────────┘    └───────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
                       ↑ 宿主机工具链上的 go build/test
```

## 分层

### Electron（`electron/src/`）
- `main.ts`：应用生命周期。启动时检测 PATH 上的 Go（`goDetector.ts`），
  拉起后端（`backend.ts`），等待端口文件发现（`portDiscovery.ts`），
  然后打开 BrowserWindow。
- `backend.ts`：解析 Go 二进制路径（开发态：预构建的
  `backend/bin/<os>-<arch>/gotutor-backend`；生产态：
  `process.resourcesPath/backend/`）。把 stdout/stderr 管道到
  `app.getPath('logs')/backend.log`。SIGTERM → 3 秒宽限 → SIGKILL。
- 开发态窗口加载 `http://localhost:5173`（Vite），生产态加载
  `file://...frontend/dist/index.html`。Vite 的 `base: './'` 对 file://
  场景至关重要。

### 前端（`frontend/src/`）
- Vue 3 SPA，Pinia store，vue-i18n（zh-CN + en），使用 hash history 的
  vue-router（Electron 友好）。
- `App.vue`：带主题/语言切换的侧边栏 + 章节列表（解锁/完成状态）。
- `views/ChapterView.vue`：通过 `defineAsyncComponent` 懒加载 Monaco，
  以免列表视图付出 Monaco 约 3MB 的首屏代价。
- `components/CodeEditor.vue`：Monaco 集成。自定义
  `gotutor-light` / `gotutor-dark` 主题与主题 store 同步。TODO 装饰：
  边距圆点 + 行底色 + 悬停提示。⌘/Ctrl+Enter 提交。
- `api/client.ts`：带 `Accept-Language` 拦截器的 Axios，让后端在每次
  请求时解析语言。

### 后端（`backend/`）
- **HTTP 层**（`internal/server/`）：Gin。路由在 `routes.go` 里以对
  `s.DB()` 的闭包形式接线。通过 `context.WithTimeout` 分层超时。
  SIGINT/SIGTERM 时优雅关停。
- **持久层**（`internal/db/`）：经 `modernc.org/sqlite` 的 SQLite（纯
  Go，无 CGO——对 Phase 10 的交叉编译至关重要）。迁移通过
  `//go:embed *.sql` 内嵌，每次连接幂等应用。
- **章节注册表**（`chapters/`）：带元数据的静态章节列表；动态内容
  （模板、提示、测试、答案）从 `//go:embed all:content` 加载。测试文件
  以 `<name>_test.go.txt` 形式分发（go:embed 会跳过 `.go` 文件）；提交
  时还原为 `<name>_test.go`。
- **校验器**（`internal/verifier/`）：应用的核心。
  - `policy.go`：每章的策略（允许的导入、超时、输出上限）。
  - `astcheck.go`：基于 `go/parser` 的导入扫描 + 监听调用检测。编译前
    运行，快速失败。
  - `sandbox.go`：临时目录生命周期。写入用户代码 + 测试 + go.mod。
  - `exec.go`：`exec.CommandContext`，带 `cmd.Cancel=SIGKILL` +
    `WaitDelay`。输出捕获到 64 KiB 上限的缓冲区。
  - `verifier.go`：顶层编排 + NumCPU 大小的并发信号量。
  - `pathguard.go`：符号链接逃逸防御。

## 数据流：一次提交

1. 学习者在 Monaco 里敲代码，点击提交。
2. 前端调用 `POST /api/chapters/calc/submit`，带 `{userCode}`。
3. `api.HandleSubmit`（在 `internal/api/submit.go`）：
   - 查找章节，绑定并对请求体做 256 KiB 上限。
   - 调用 `verifier.Verify(ctx, chapter, userCode, goBin)`，带 15 秒
     外层 ctx。
4. `verifier.Verify`：
   - `ASTCheck`：解析 + 扫描导入。命中禁用导入即拒绝 → 不启动
     `go test` 直接提前返回。
   - 获取并发槽位。
   - `ch.TestFiles()` 读取内嵌的 `<chapter>/tests/*_test.go.txt`。
   - `NewSandbox`：在 `os.TempDir()/gotutor-<chapter>-<uuid>/` 下建
     临时目录。
   - 写入 `main.go`（用户代码）、`<name>_test.go`（章节测试）、
     `go.mod`（`module gotutoruser`，`go 1.26`）。
   - `RunGoTest`：`go test -v -timeout 10s ./...`，cwd=沙箱。输出流式
     写入 cappedBuffer。
   - defer sandbox.Cleanup()。
5. `HandleSubmit` upsert `progress`（通过记 completed_at，失败 attempts++），
   返回 `SubmitResult`。
6. 前端乐观地把章节的 `completed` 标志置位；侧边栏响应式更新。

## 存储

`progress` 表（SQLite，WAL 模式）：
- `chapter_id TEXT PRIMARY KEY`
- `completed_at INTEGER NULL`（unix 秒）
- `attempts INTEGER NOT NULL DEFAULT 0`
- `last_output TEXT`

外加 `chapters`（元数据，当前从注册表填充而非数据库）和 `settings`
（键/值，v1 未用）。

## 交叉编译

`make backend-build-<os>-<arch>` 产出 mac-arm64、mac-x64、linux-x64、
win-x64 的后端二进制。纯 Go 的 SQLite 让这件事很简单——无 CGO、无
交叉编译器工具链。`electron-builder.yml` 通过 `extraResources` 和
`${os}-${arch}` 替换把每个二进制打入 `Resources/backend/`。

## v1 未包含的内容

- 11 个章节（calc、urlcheck + 9 个取自 AiDeptus 网关模式的进阶章）。
  要加更多，见 [ADDING_A_CHAPTER-zh.md](./ADDING_A_CHAPTER-zh.md)。
- 未签名的 macOS 构建（README 里记录了 Gatekeeper 绕过）。
- Windows 专属的资源限制（RLIMIT 仅 Linux/BSD 有）。见
  [SECURITY-zh.md](./SECURITY-zh.md)。
- 自动更新（Phase 13+，不在当前计划内）。
- 多用户进度（v1 是单用户模型）。
