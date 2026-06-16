# GoTutor

> [English](README.md) | 中文

一款交互式桌面应用，通过在真实迷你项目里填补 `// TODO` 空缺来学习
Go。应用会用 `go test` 真正编译并运行你的代码——不靠字符串匹配。

## 状态

已端到端发布 **11 个章节**：
- **两个基础章节** —— `calc`（命令行计算器：`os.Args`、`strconv`、switch
  分发、除零处理）和 `urlcheck`（并发 URL 检查器：goroutine、channel、
  `sync.WaitGroup`、`net/http` 客户端）。
- **九个进阶章节** —— 取材自一个真实的 LLM API 网关（AiDeptus）所用
  Go 模式：错误处理、接口与策略模式、并发求和、channel 与 select、
  context 取消、令牌桶限流、熔断器、HTTP 重试与退避、SSE 流式转发。

全部章节都走完整的 Monaco 编辑器 + 沙箱化 `go test` 流程。

## 技术栈

| 层级         | 选型                                                |
|-------------|-----------------------------------------------------|
| 前端         | Vue 3 + Vite + Electron + Monaco + Pinia + vue-i18n |
| 后端         | Go + Gin HTTP 旁路进程                              |
| 存储         | SQLite（`modernc.org/sqlite`，纯 Go——无 CGO）       |
| 校验器       | 临时目录内沙箱化 `go test` + AST 标准库白名单         |
| 打包         | electron-builder；Go 二进制作为 `extraResources` 打入 |

界面**双语**（zh-CN + en），侧边栏可切换；明暗主题跟随系统或手动选择。

## 前置条件

- **Go 1.26+** 在 PATH 中——校验器会调用它。缺失时应用会显示"安装 Go"页面。
- **Node.js 22+** 与 **pnpm 8+**，用于前端/Electron 开发。

## 开发环境

```bash
git clone <repo> GoTutor && cd GoTutor

# 1. 后端（终端 1）
make backend-dev                          # 服务于 :8081

# 2. 前端（终端 2）
make frontend-install                     # 一次性
make frontend-dev                         # 服务于 :5173，/api 代理到 :8081

# 3. Electron 外壳（终端 3，可选——封装 1 + 2）
make electron-install                     # 一次性，Electron 下载约 5–25 分钟
GOTUTOR_DEV=1 make electron-dev           # 打开桌面窗口
```

不想起 Electron、只想快速冒烟测试的话，在完成第 1、2 步后用浏览器打开
`http://localhost:5173`。

## 构建安装包

```bash
# 交叉编译 Go 后端 + 构建前端 + 打包 Electron
make package-darwin        # → release/GoTutor-<ver>-arm64.dmg
make package-linux         # → release/GoTutor-<ver>.AppImage + .deb
make package-win           # → release/GoTutor-Setup-<ver>-x64.exe
```

纯 Go 的 SQLite 意味着没有 CGO 的麻烦；同一份 Go 源码无需为每个 OS 准备
C 工具链即可交叉编译到全部四个目标。

> **macOS Gatekeeper**：v1 未签名。运行 dmg 时，右键应用 → 打开 → 确认。
> 后续版本会用 Apple Developer ID 签名。

## API

| 端点                                   | 方法   | 请求体                        | 返回                                  |
|---------------------------------------|--------|-------------------------------|--------------------------------------|
| `/api/health`                         | GET    | —                             | `{ok, port, goFound, goVersion}`     |
| `/api/chapters`                       | GET    | —                             | `Chapter[]`（按语言返回）              |
| `/api/chapters/:id/template`          | GET    | —                             | `{code, todos: [{line, hint}]}`      |
| `/api/chapters/:id/hint?line=N`       | GET    | —                             | `{text}`                             |
| `/api/chapters/:id/submit`            | POST   | `{userCode: string}`          | `{passed, output, durationMs, ...}`  |
| `/api/reset`                          | POST   | —                             | 204                                  |

`Accept-Language` 在每个对语言敏感的端点上切换 zh-CN 与 en。

## 文档

- [打包](docs/PACKAGING-zh.md) —— 构建 `.dmg` / `.exe` / `.AppImage`
  安装包、Gatekeeper 绕过、排错。
- [架构](docs/ARCHITECTURE-zh.md) —— 系统分层、端到端提交流程、存储布局。
- [安全模型](docs/SECURITY-zh.md) —— 沙箱防御什么、残余风险、未来加固
  （RLIMIT、Docker）。
- [添加章节](docs/ADDING_A_CHAPTER-zh.md) —— 编写你自己的练习
  （9 个进阶章节已覆盖网关模式）。

## 许可证

MIT，Copyright © 2026 TeaHouse。详见 [LICENSE](LICENSE)。
