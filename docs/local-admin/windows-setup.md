# Windows 开发环境与实测记录

## 结论

本项目支持 Windows 原生开发，不需要 WSL。本机已完成 Windows/amd64 编译、选定测试和依赖校验。2026-10-01 增加公告请求总超时及资源校验协议读写超时，补齐缺失启动素材后，帮助命令完成初始化并输出用法，实际 ZeroBot 进程已启动并连接到本机 NapCat。

当前具备本机编译、运行和真实 QQ 联调条件；帮助及服务列表私聊回复已由用户与日志确认通过。各业务插件及新管理后台尚未完整验收，见 [本次基本功能检查](basic-smoke-2026-10-01.md) 和 [OneBot 环境与自测](onebot-windows.md)。

## 已拉取的内容

| 内容 | 位置/版本 | 状态 |
| --- | --- | --- |
| 主项目完整 Git 历史 | 当前仓库；开发分支 `codex/local-admin` | 已拉取，非浅克隆 |
| 个人 Fork | `origin` → `https://github.com/qiye-yuri/ZeroBot-Plugin.git` | 已拉取远端分支和备份分支 |
| 上游 | `upstream` → `https://github.com/FloatTech/ZeroBot-Plugin.git` | 已拉取；本次上游与 Fork master 均为 `65619f0` |
| Go 工具链 | Go `1.26.1 windows/amd64` | 已下载并用于编译 |
| Go 模块 | `go.mod` / `go.sum` 锁定版本；本机 `<项目目录>` | 已下载，`go mod verify` 通过 |
| 旧 WebUI 源码与 dist | `.local/references/ZeroBot-Plugin-Webui`，`v1.0.1-0.20230412164529-cfe878675931` | 已从锁定模块版本复制供参考；未安装前端 npm 依赖，未启用旧后台 |
| data 资源子模块 | `6708169f6346d3b97998d4230e44b931617ac5fc` | 已接回子模块并核对版本；使用部分克隆和稀疏检出 |
| Windows 程序 | `.local/bin/zerobot.exe` | 已生成，42,877,440 字节；后续构建体积可能变化 |
| 启动词典 | `data/Chat/dict.txt` | 按资源子模块锁定提交补齐；约 5 MiB |
| 百人一首启动 CSV | `data/hyaku/小倉百人一首.csv` | 从作者 GitHub 原仓库补齐；19,579 字节；来源摘要见本次检查记录 |
| registry 超时补丁 | `third_party/go-registry` | 基于锁定 v0.2.7 源码，保留许可；通过 go.mod replace 使用 |

资源稀疏检出目前包括 `Control/kanban.png`、`Control/zbpuwu.png`、四个启动字体 GlowSans、Impact、Torus、Yuruka 和 `Chat/dict.txt`，以及资源仓库说明/许可文件。它们用于插件控制与分词初始化；运行期间另有插件按需下载素材，不代表全部资源已下载。

锁定资源快照包含约 733 MiB 文件，另外还有嵌套资源仓库；完整素材下载曾停滞。大型游戏、图片和其他插件素材尚未全部下载，后续根据保留插件逐项补齐。不要把部分克隆描述成完整素材包。若确实需要全部 data 素材，可在网络稳定时于项目根目录执行：

```powershell
git -C data -c http.sslBackend=openssl -c http.version=HTTP/1.1 sparse-checkout disable
git -C data submodule update --init --recursive
```

这会下载更多素材，且懒加载资源仍可能来自其他服务。无需为本地后台第一版预先准备全部娱乐插件素材。

## 修改后编译与测试

在项目根目录执行：

```powershell
# 下载缺失依赖、校验、生成代码并编译
.\scripts\build-windows.ps1

# 同时执行已选定的基线测试
.\scripts\build-windows.ps1 -RunTests
```

脚本不会调用 `go mod tidy` 或写入全局 Go 配置，避免仅为启动而改动依赖；`CGO_ENABLED=0` 只作用于本次脚本并在结束后恢复。

当前选定测试为 `kanban`、`plugin/chess` 和 `plugin/tarot`，以及单独执行本地 `third_party/go-registry` 模块的超时测试。包初始化可能访问公告服务或下载字体。其他包中存在外部服务测试和显式 `t.Fail()` / `t.Fatal()` 的调试测试，尚未作为可靠的全仓验收门槛。

## 真实机器人运行准备

本机已部署 NapCat v4.18.28，QQ 已登录。登录页监听 `127.0.0.1:6099`，OneBot 正向 WebSocket 实际监听 `127.0.0.1:6700`，三个只读 API 已通过检查。安装脚本已生成本地 Token 并与 `.local/config.json` 配对，重复执行保留现有配置。详见 [OneBot 环境与自测](onebot-windows.md)。后台本身不负责 QQ 登录，也尚未实现。

复制示例到已被 Git 忽略的位置；如果文件已存在，不覆盖原配置：

```powershell
New-Item -ItemType Directory -Path .local -Force | Out-Null
if (-not (Test-Path .local/config.json)) {
    Copy-Item docs/local-admin/config.windows.example.json .local/config.json
}
```

编辑 `.local/config.json`，填入自己的主人 QQ 号到 `zero.super_users`。本机 `Url` 和 `AccessToken` 已配对；更改 OneBot 服务配置时需同步更新。示例不包含真实凭据；其中时长数字遵循上游 JSON 格式，单位为纳秒。

正常手动启动使用以下命令。本次后台实例已运行，PID、配置和日志位置记录于 `.local/zerobot/process.json`；不要重复启动第二个实例。

```powershell
.\scripts\run-windows.ps1 -ConfigPath .local/config.json

# 需要将输出重定向到文件时使用现有 Windows 兼容开关
.\scripts\run-windows.ps1 -ConfigPath .local/config.json -Headless
```

`-Headless` 设置上游已有的 `DEBUG_MODE=1`，只用于容忍重定向时无有效控制台句柄，不代表后台网页已启用。脚本使用仓库根目录运行，插件数据写入 `data` 对应目录；修改资源或配置后提交前仍需检查主仓库和子模块状态。

FFmpeg、TiMidity 尚未在当前命令搜索路径中发现，使用视频解析、猜歌、MIDI 转换等相关插件时需另行准备；它们不是本次 Go 编译的前置条件。QQ 登录及 OneBot 三个只读 API 已通过验证，完整消息联调仍待完成。

## 实际验证结果

| 编号 | 验证 | 结果 | 证据/限制 |
| --- | --- | --- | --- |
| WIN-01 | `go generate main.go` 与 `CGO_ENABLED=0 go build -trimpath` | 通过 | 本机生成 Windows `.exe` |
| WIN-02 | `go test ./plugin/chess ./plugin/tarot -count=1 -timeout=60s` | 通过 | 两包分别约 18.1 秒、20.4 秒；不代表全仓测试通过 |
| WIN-03 | `go mod verify` | 通过 | `all modules verified` |
| WIN-04 | 新建构建脚本实际执行 | 通过 | 工具链、依赖校验、生成和编译全过程退出码为 0 |
| WIN-05 | PowerShell 脚本语法及 JSON 示例解析 | 通过 | 两个脚本无语法错误；JSON 可解析 |
| WIN-06 | 执行 `zerobot.exe -h`，日志重定向 | 新构建完成 | 旧构建超时；2026-10-01 补丁与素材就绪后输出 Usage，未单独保存该次最终退出码 |
| WIN-07 | 初始化跟踪 | 已记录 | 旧记录 `.local/smoke-help/trace-stderr.log`；新记录 `.local/smoke-restart/help2-stderr.log` 完成全部初始化 |
| WIN-08 | 真实 QQ 收发与新后台页面 | 进行中 | QQ/OneBot 已在线，ZeroBot 握手及帮助/服务列表真实回复通过；新后台尚未实现 |
| WIN-09 | 实际配置启动 ZeroBot 并连接 NapCat | 通过 | `.local/zerobot/process.json` 及日志；进程存活、WS 握手成功，无启动 panic |
| WIN-10 | 公告协议与 registry 无响应回归 | 通过 | `go test ./kanban` 及本地 registry 模块测试；正常公告、错误状态与超时均覆盖 |

`.local/` 与选定测试生成的 `plugin/*/data/` 已加入忽略规则。代码及文档修改尚未提交或推送。

## 后续必须处理的问题

1. 继续排查其他插件外部服务与资源下载边界；已修复公告及 registry 读写等待，不代表所有第三方接口都具备超时及降级。
2. 确定保留插件并补齐其资源；对其他外部接口逐项验证。
3. 旧 WebUI 的 `model/webui.go` 将密码作为字符串保存并拼接查询；路由中日志等接口未挂认证中间件，CORS 反射请求来源。当前仅作源码参考，不直接启用；复用结论需结合现有控制层重新设计认证及接口边界。
4. 完成基础私聊收发检查后，再逐项进行权限、保留插件和后台功能验收。
