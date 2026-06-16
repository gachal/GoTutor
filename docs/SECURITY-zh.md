# GoTutor 安全模型

> [English](SECURITY.md) | 中文

GoTutor 通过 `go test` 运行用户提交的 Go 代码。这些代码按定义是不可信
的——学习者可能粘贴任何东西。本文档说明 v1 防御什么、不防御什么，以及
未来的加固路径。

## v1 沙箱

每次 `POST /api/chapters/:id/submit` 在返回判定前都会做以下事情：

1. **请求体上限**。请求体限制为 256 KiB。防止超大 payload 在校验前把
   handler 撑爆 OOM。
2. **AST 导入检查**。`internal/verifier/astcheck.go` 用 `go/parser`
   解析用户源码，拒绝任何不在本章白名单里的导入。默认值（在
   `policy.go`）：
   - **允许**：`fmt`、`os`、`strconv`、`strings`、`math`、`time`、
     `errors`、`log`、`sort`、`path`、`path/filepath`、`io`、`io/fs`、
     `bytes`、`unicode`、`regexp`、`context`、`sync`。
   - **始终拒绝**（`DeniedStdlibImports`）：`os/exec`、`unsafe`、
     `syscall`、`reflect`、`net`、`C`、`cgo`。无论章节策略如何，这些都
     优先生效。
   - 章节通过 `Chapter.AllowImports` 扩展白名单（urlcheck 放开
     `net/http`、`sync`、`time`）。
   - 除非 `AllowListen` 为 true（v1 中从不为 true），`net.Listen` 与
     `http.ListenAndServe` 调用点也会被拒绝。
3. **临时目录隔离**。每次提交运行在
   `os.TempDir()/gotutor-<chapter>-<uuid>/` 下，有自己的 `go.mod`
   （`module gotutoruser`）。提交之间无共享状态。
4. **分层超时**。
   - handler 外层 ctx：15 秒。
   - `go test -timeout`：10 秒。
   - `cmd.WaitDelay`：2 秒（回收用户代码产生的孤儿子进程）。
   - 超时时 `cmd.Cancel` 触发 `SIGKILL`，用户代码无法靠信号处理器逃逸。
5. **输出上限**。合并的 stdout+stderr 在 64 KiB 处截断，并加
   `[output truncated at 65536 bytes]` 标记。防止日志炸弹 DoS。
6. **路径守卫**。`pathguard.go` 在任何写入前解析符号链接，并断言解析后
   的路径仍在沙箱根目录下。
7. **并发限制**。一个 `NumCPU` 大小的 `chan struct{}` 约束并行编译，使
   得提交洪峰无法占满每个核心。

## v1 防御的内容

- **进程派生**：`os/exec` 在 AST 阶段被拒，学习者无法 shell 出去跑
  `rm`、`curl` 等。
- **unsafe 内存把戏**：`unsafe` 与 `syscall` 被拒。
- **基于反射的绕过**：`reflect` 被拒，AST 无法被反射式导入查询骗过。
- **网络监听**：`net.Listen` / `http.ListenAndServe` 被阻。出站 HTTP
  仅 `urlcheck`（`net/http` 客户端）允许。
- **fork 炸弹**：进程数受 NumCPU 信号量 + 10 秒测试超时约束。
- **死循环**：`go test -timeout 10s` + 外层 15 秒 ctx。
- **日志炸弹 DoS**：64 KiB 输出上限。

## v1 不防御的内容

这些是已记录的残余风险；有决心的攻击者在 v1 安装上很可能能做到：

- **用循环里的 os.WriteFile 填满磁盘**。10 秒超时限制时长但不限总字节数。
  缓解：未来的 `RLIMIT_FSIZE`。
- **CPU 烧**。10 秒超时有界，但紧凑循环跑满所有核心仍会压垮机器。缓解：
  未来的 `RLIMIT_CPU` + cgroups。
- **`net/http` 客户端滥用**（仅 urlcheck 章）。学习者可以从自己的机器
  对第三方 URL 发起 DoS。在 v1 的"单一可信用户"模型下可接受；多租户则
  不可接受。
- **Windows 专属的资源限制**。`RLIMIT_*` 仅 Linux/BSD 有。v1 Windows
  只靠超时 + AST。

## 未来加固路径

当 GoTutor 进入多租户场景时，依次叠加：

1. **`RLIMIT_FSIZE`**（10 MiB），Linux/macOS——约束磁盘写入。
2. **`RLIMIT_NPROC`**（1），Linux——防 fork 炸弹。
3. **`RLIMIT_CPU`**（10 秒），Linux/macOS——内核级 CPU 上限。
4. **Docker 后端**。配置开关 `verifier.backend = "docker"` 在
   `--network=none --memory=256m --pids-limit=64 --read-only --tmpfs /tmp`
   内跑 `go test`。黄金标准沙箱。macOS/Windows 上需 Docker Desktop。
5. **Linux 网络命名空间**（`unshare -n`），无需 Docker 即可实现真正的
   网络隔离。
6. **Windows Job Objects**，约束 CPU/内存/pids。

## 报告安全问题

发邮件到 security@teahouse.dev 说明详情。请不要为安全敏感的 bug 开公开
issue。
