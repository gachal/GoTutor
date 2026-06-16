# 把 GoTutor 打包为 macOS

> [English](PACKAGING.md) | 中文

从源码构建可运行 `.dmg` 的端到端指南。

## 一行命令构建（推荐）

```bash
make package-darwin
```

该目标串联：
1. **后端交叉编译** —— `make backend-build-darwin-arm64` 产出
   `backend/bin/mac-arm64/gotutor-backend`（约 27 MB）。
2. **前端构建** —— `pnpm --dir frontend build` 产出 `frontend/dist/`
   （初始包 gzip 约 76 KB + 30 KB Monaco chunk）。
3. **Electron TS 构建** —— `pnpm --dir electron build` 产出
   `electron/dist/`。
4. **electron-builder** —— 把上述三者加 Electron 运行时打包成 `.app`，
   再做成 DMG。

产物：**`release/GoTutor-<version>-arm64.dmg`**

## 前置条件

一次性安装（前端 + electron 依赖）：

```bash
make frontend-install    # Vue/Vite/Monaco 等
make electron-install    # Electron + electron-builder（约 200 MB，5–25 分钟）
```

构建机上的工具链：

```bash
go version       # 1.26+
node --version   # 22+
pnpm --version   # 8+
```

安装 `.dmg` 的最终用户只需 PATH 上有 **Go 1.26+**——后端校验器调用
`go test`。首次启动若 `go` 缺失，应用会显示"安装 Go"页面。

## Intel Mac

x64 构建（较老的 Intel Mac）：

```bash
make package-darwin-x64
```

产物：`release/GoTutor-<version>-x64.dmg`

## 通用二进制？

v1 不支持——我们按架构分发 dmg。通过 `lipo` 做通用版是未来选项
（`electron-builder` 支持 `--arch universal`）。

## 验证构建

```bash
ls -lh release/
# 期望：
# GoTutor-0.1.0-arm64.dmg
# GoTutor-0.1.0-arm64-mac.zip
```

双击 `.dmg`，把 GoTutor 拖到 Applications。

## Gatekeeper（v1 未签名）

首次启动会显示*"GoTutor 无法打开，因为它来自身份不明的开发者。"* 变通：

**方案 A —— Finder 右键**（推荐）：
1. 在 Finder 里找到 GoTutor.app。
2. **右键** → **打开**。
3. 在对话框里确认。

**方案 B —— 命令行**：
```bash
xattr -cr /Applications/GoTutor.app
open /Applications/GoTutor.app
```

未来的 v2 会用 Apple Developer ID 签名。见 CLAUDE.md →"已知缺口"。

## 开发模式（不打包）

做本地测试而不产出 `.dmg`：

```bash
# 终端 1 —— 后端
make backend-dev                  # 服务于 :8081

# 终端 2 —— 前端
make frontend-dev                 # Vite 于 :5173，/api 代理到 :8081

# 终端 3 —— Electron 外壳
GOTUTOR_DEV=1 make electron-dev   # 打开桌面窗口
```

每次保存热重载；无需重新构建。

## 交叉编译到其他 OS

同一个 Makefile，不同目标：

```bash
make package-linux                # release/GoTutor-<ver>.AppImage + .deb
make package-win                  # release/GoTutor-Setup-<ver>-x64.exe
```

得益于纯 Go 的 SQLite（`modernc.org/sqlite`），Go 后端的交叉编译很
简单——无 CGO、无需为每个 OS 准备 C 工具链。electron-builder 会自动
下载目标 OS 的 Electron 二进制。

## 排错

- **`tsc: command not found`** —— `make electron-install` 没跑完。重跑；
  若卡住，用 `pnpm install --dir electron` 重试。
- **`[prebuild] backend binary not found`** —— 目标平台/架构的 Go 二进制
  没构建。先跑 `make backend-build-darwin-arm64`（或对应目标），或直接用
  会串联两者的 `make package-darwin`。
- **应用启动但立即报错 `spawn .../Resources/backend/gotutor-backend
  ENOENT`** —— .app 打包时没把 Go 二进制放进去（路径修复前的旧构建）。
  从干净态重建：`rm -rf backend/bin release && make package-darwin`。
- **electron-builder 报 `extraResources` 错** —— 你目标架构的后端二进制
  不存在。prebuild 检查本该先拦住；若没有，跑
  `make backend-build-darwin-arm64`（或对应目标）。
- **应用打开但显示"无法连接后端"** —— 后端启动崩溃。查
  `~/Library/Logs/gotutor/backend.log`（macOS 路径；Electron 主进程经
  `app.getPath('logs')` 写到这里）。
- **应用显示"安装 Go"页面** —— `go` 不在 GUI 应用的 PATH 上。GUI 启动器
  不继承你 shell 的 PATH。用 go.dev 官方 `.pkg` 安装 Go（放到
  `/usr/local/go/bin`，应用能看到），或软链：
  `sudo ln -s $(which go) /usr/local/bin/go`。

## 打包后的应用文件布局

```
GoTutor.app/
├── Contents/
│   ├── Info.plist
│   ├── MacOS/
│   │   └── GoTutor              # Electron 启动器二进制
│   ├── Resources/
│   │   ├── backend/
│   │   │   └── gotutor-backend  # Go 旁路进程（交叉编译）
│   │   ├── frontend/dist/       # Vue SPA
│   │   ├── electron/dist/       # 编译后的 TS main + preload
│   │   └── app.asar             # Electron 外壳代码
│   └── Frameworks/              # Chromium、Node 等
└── ...
```

运行时，`electron/src/backend.ts` 经
`process.resourcesPath/backend/gotutor-backend` 解析旁路进程。
