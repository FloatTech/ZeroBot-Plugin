# Windows OneBot 环境

## 当前状态

已安装并实际启动 NapCat，QQ 已完成登录，最近检查状态为 `ready`。WebUI 返回 HTTP 200，实际监听地址为 `127.0.0.1:6099`，未认证的管理请求被拒绝，本地凭据认证成功。

OneBot 已实际监听 `127.0.0.1:6700`，仅本机访问。带 Token 的 WebSocket 连接及 `get_login_info`、`get_status`、`get_version_info` 均通过实测，在线状态为 true。2026-10-01 已重新启动并再次扫码登录，上述检查再次通过。ZeroBot 的 `.local/config.json` 已使用匹配的 URL 和 Token，主人账号已填写，且与机器人账号不同。

ZeroBot 之前发现的公告及资源校验网络等待已增加超时，缺失启动素材已补齐；主程序已实际启动并与 NapCat 握手成功。私聊命令收发正在验证。NapCat 登录成功与 ZeroBot 插件程序正常启动是两个独立检查点。

## 安装来源和依赖

| 内容 | 版本或来源 | 验证 |
| --- | --- | --- |
| NapCat 官方包 | `NapCat.Shell.Windows.Node.zip`，v4.18.28 | SHA256 与官方 Release asset digest 一致 |
| SHA256 | `fb64fa3b036ad2df1a5d7c204c482694c20e4b763978c8a4968fd3474c05b4a8` | 本机已核对 |
| 包内 Node | v22.11.0 | 实际执行版本命令验证 |
| 包内 QQ native | 9.9.31-49738 | 来自包内元数据 |
| 本机已安装 QQ | 9.9.35-52892 | 从安装配置和 package.json 核对；未改动其文件或停止进程 |
| 补充原生库 | 本机 QQ `resources/app/crypto.dll`、`ssl.dll` | 首次启动明确缺失；复制到独立运行目录后启动成功 |

安装记录保存在 `.local/onebot/install-manifest.json`，包括官方包来源以及补充 DLL 的路径和 SHA256。不是修改 NapCat 业务代码，也不是给系统 QQ 安装注入插件。QQ native 的登录缓存可能仍由其自身写入用户数据目录；`NAPCAT_WORKDIR` 控制的是 NapCat 配置、日志、插件与缓存路径。

官方参考：[发行包](https://github.com/NapNeko/NapCatQQ/releases/tag/v4.18.28)、[Windows 安装说明](https://napneko.github.io/guide/boot/Shell)、[OneBot 配置说明](https://napneko.github.io/config/basic)。

## 文件与端口

| 用途 | 位置/地址 |
| --- | --- |
| 独立程序 | `.local/onebot/napcat-v4.18.28/` |
| NapCat 配置和缓存 | `.local/onebot/runtime/` |
| WebUI 配置和登录密钥 | `.local/onebot/runtime/config/webui.json` |
| 默认 OneBot 配置 | `.local/onebot/runtime/config/onebot11.json` |
| QQ 登录后账户配置 | `config/onebot11_<QQ号>.json`，如果生成则以该文件为准 |
| ZeroBot 连接配置 | `.local/config.json` |
| 启动日志 | `.local/onebot/runtime/launcher-logs/` |
| WebUI | `http://127.0.0.1:6099/webui/` |
| OneBot 正向 WS | `ws://127.0.0.1:6700` |

WebUI 密钥与 OneBot Token 分别生成，二者用途不同。配置和日志已由 `.local/` 忽略规则排除，不提交至 Git。NapCat 的启动日志可能含 WebUI 密钥和二维码链接，应仅在本机查看。

## 常用操作

在项目根目录执行：

```powershell
# 重复执行会验证安装和配对，保留已有配置及密钥
.\scripts\install-onebot.ps1

# 在当前终端运行；已有记录且进程仍运行时不重复启动
.\scripts\start-onebot.ps1

# 在默认浏览器中打开 WebUI，自动读取本地登录密钥
.\scripts\open-onebot-webui.ps1

# 检查页面、认证和 QQ 登录状态；登录后检查三个只读 OneBot API
.\scripts\test-onebot.ps1

# 从另一终端停止本项目记录的 NapCat 启动器及其 Node 子进程
.\scripts\stop-onebot.ps1
```

首次安装需要 GitHub CLI 和下载网络。当前目录已有验证过的包；重装脚本不会再次下载。如果 QQ 安装在其他路径，可传入 `-QQAppPath` 指向含 `crypto.dll` 和 `ssl.dll` 的 `resources/app` 目录。

当前实例已改为 Windows 计划任务的独立宿主运行，输出重定向到 `.local/services/`。之前临时工具启动的两个进程已退出，需要跨会话运行时使用 [持续运行入口](operations.md) 中的 `start-local-services.ps1`；前台 `start-onebot.ps1` 仍可用于手工调试，终端需保持打开。用户提供的杀毒记录确认：原隐藏启动脚本在 2026-09-30 两次触发 `HEUR:TrojanDownloader/PS.NetLoader.ae` 并被删除；当前没有修改杀毒软件配置或添加排除项。

## 用户需要完成的登录步骤

1. 打开 WebUI，选择二维码登录；使用准备作为机器人的 QQ 扫码，并在手机上确认。
2. 若 WebUI 要求更改登录密码，按页面完成。该密码与 OneBot Token 无关；以后打开页面的脚本会读取更新后的配置。
3. 确认 QQ 状态变为在线，网络配置中 `zerobot-local` 已启用，地址为 `127.0.0.1`，端口为 `6700`。
4. 在 `.local/config.json` 的 `zero.super_users` 填入自己的主人 QQ 号。本机已由用户填写并通过格式检查；不要将机器人 QQ 号自动当作主人账号。
5. 执行 `test-onebot.ps1`，应显示 `qqLoggedIn: true` 和 `onebotReadOnlyAPIsPassed: true`。

登录二维码会过期，以 WebUI 中当前显示的二维码为准。无需把 QQ 密码发送到聊天中。

当前 ZeroBot 主程序已通过独立宿主启动并成功连接。后台入口使用 `start-local-services.ps1 -Service ZeroBot`，会先校验 `.local/onebot/account.json` 指定的机器人账号；前台调试可用 `run-windows.ps1 -ConfigPath .local/config.json -Headless`。PID 和日志位置记录于 `.local/zerobot/process.json`，不要重复启动第二个实例。默认 WS 配置使用消息数组格式、30 秒心跳、关闭自身消息上报和调试事件输出；未启用 HTTP API 或反向连接。

## 环境自测

| 编号 | 检查 | 本次结果 |
| --- | --- | --- |
| OB-01 | 官方包 SHA256、包内 Node 版本 | 通过 |
| OB-02 | 缺失 DLL 定位、补齐后启动 | 通过 |
| OB-03 | WebUI HTTP 200 和实际回环监听 | 通过，`127.0.0.1:6099` |
| OB-04 | 未认证请求拒绝、有效认证成功及会话退出 | 通过 |
| OB-05 | ZeroBot 与 OneBot URL/Token 一致 | 通过 |
| OB-06 | 重复安装/配置不修改既有密钥 | 通过，配置文件摘要保持一致 |
| OB-07 | 已运行时重复启动 | 通过，复用记录的 PID |
| OB-08 | QQ 登录并完成三个只读 API 检查 | 通过，`qqLoggedIn: true`、`loginPhase: ready`、`onebotReadOnlyAPIsPassed: true` |
| OB-09 | 停止与重新启动回归 | 重新启动、扫码及接口回归通过；停止脚本尚未实测 |
| OB-10 | ZeroBot 实际消息收发 | 通过：初始化、实际 WS 握手、隔离模拟及帮助/服务列表真实 QQ 回复均通过，用户及收发日志已确认 |

健康检查不会发送消息、添加好友、处理群邀请或更改 QQ 群设置。环境就绪不等于所有业务插件通过验收。
