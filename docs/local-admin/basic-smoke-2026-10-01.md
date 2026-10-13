# Windows 基本功能验证：2026-10-01

代码基线 `65619f08b09da44d57f44beacb3ce261130c5457`，包含尚未提交的本地修复。环境：Windows/amd64、Go 1.26.1、NapCat v4.18.28、OneBot 11 正向 WebSocket。本文验证已有机器人基本链路，不代替新后台 TC-01～25 验收。

## 当前结果

| 编号 | 检查 | 结果 | 证据与限制 |
| --- | --- | --- | --- |
| BASIC-01 | 实际 JSON、主人账号类型、URL/Token 配对 | 通过 | 主人已填写，正整数数组；Token 非空且匹配，不输出明文 |
| BASIC-02 | NapCat 重新启动、再次扫码登录 | 通过 | 新进程及 WebUI HTTP 200；用户确认扫码，API 登录状态 ready |
| BASIC-03 | 未认证管理请求拒绝、三个 OneBot 只读 API | 通过 | `test-onebot.ps1`；登录信息、在线状态及版本接口成功 |
| BASIC-04 | WebUI/OneBot 实际回环监听 | 通过 | `127.0.0.1:6099`、`127.0.0.1:6700` |
| BASIC-05 | 公告正常响应、错误状态及无响应读超时 | 通过 | `go test ./kanban -count=1 -timeout=30s` |
| BASIC-06 | registry Get/Cat/MD5 比较无响应时结束 | 通过 | `third_party/go-registry` 模块内 `go test ./... -count=1 -timeout=30s` |
| BASIC-07 | Windows 重新编译、帮助初始化 | 通过 | `CGO_ENABLED=0 go build -trimpath`；帮助日志完成 Usage 输出，未保存该次最终退出码 |
| BASIC-08 | 实际配置启动 ZeroBot、WS 握手 | 通过 | `.local/zerobot/process.json` 和日志；进程存活，握手成功，无启动 panic；主人与机器人不是同一账号 |
| BASIC-09 | 主人私聊 `/help` | 通过 | 原 `/help\n` 不匹配已修复；编译及重启后，用户确认正常显示，ZeroBot 新日志记录收到帮助命令并发送帮助文字 |
| BASIC-10 | 主人私聊 `/服务列表` | 通过 | 用户确认有反应；ZeroBot 与 NapCat 日志均记录向主人发送图片，消息 ID 非零 |
| BASIC-11 | chess/tarot 基线回归 | 通过 | 修改公共网络依赖后补跑，两包分别约 7.240 秒、7.227 秒 |
| BASIC-12 | 隔离目录内完整程序处理帮助与服务列表 | 通过 | 新 Windows 程序处理普通帮助、末尾换行帮助、首尾空白帮助及服务列表均收到预期回复；`.local/smoke-basic/result.json`；未向真实 QQ 发消息 |
| BASIC-13 | 独立宿主恢复、正确账号校验 | 通过 | 两个手动任务 Running，宿主由 Windows svchost 启动；配置指定账号登录及新 WS 握手通过；真实基础命令回复已确认 |
| BASIC-14 | 帮助空白及消息路由回归 | 通过 | `go test ./internal/basiccommands -count=1 -timeout=20s`；11 个场景，原规则在换行与首尾空白场景失败，修复后全部通过；群聊仍需 @bot，较长命令及带参数命令不误匹配 |
| BASIC-15 | ZeroBot 停止、替换程序、重启 | 通过 | 发现 Stop-ScheduledTask 只结束宿主而留下子进程；新增路径及启动时间校验后结束本项目进程树，旧 PID 已消失，新程序连接成功；OneBot 停止/移除未回归 |

NapCat 停止及任务移除尚未单独实测；ZeroBot 已实际停止、替换及重启。当前两个服务保持运行，NapCat 登录会话持续保留。

后续用户报告发送命令没有回复：复查时两个旧进程均已退出，旧 ZeroBot 日志没有收到用户测试命令。因此前面的“进程在线”是当次检查时的状态，不代表跨工具会话持续运行已通过。已新增独立计划任务宿主；用户手动登录正确账号后，OneBot 只读 API、账号校验与新 ZeroBot 握手均通过。两个任务都在运行，父进程核实为 Windows `svchost.exe`。修复帮助末尾换行问题后，用户与日志已确认帮助及服务列表真实回复通过；浏览器安全检查仍阻止自动页面操作。详见 [恢复与登录](operations.md)。

隔离验证使用合成账号与独立配置/数据目录，不复用真实 Token。首次隔离启动缺少 `McFish/articlesInfo.json` 而退出，补齐已在本机取得的启动素材后通过。模拟进程已结束，真实 ZeroBot 保持运行。模拟验证能检查命令匹配、文字构造及图片生成；实际 QQ 投递仍以 BASIC-09、10 为准。

## 修复与资源

- 原公告请求没有连接及读写总时限；改为每次独立连接，整个请求最多五秒，失败只显示公告暂不可用。
- 资源校验使用的锁定 registry 库只有连接超时，响应读取可能无限等待；首次实测控制层初始化约 106 秒。保留 v0.2.7 源码及 GPL-3.0 许可，通过本地 replace 为协议请求加五秒读写时限，连接也有限时。详情见 [补丁说明](../../third_party/go-registry/LOCAL_PATCH.md)。
- 百人一首插件因镜像返回 HTTP 530 而 panic；从 [作者原仓库](https://github.com/fumiama/OguraHyakuninIsshu) 补齐 CSV。Git blob 为 `7fd14b91eb84c02ed00ad92b9e31b2a35a070f57`，SHA256 为 `6f48eabde117d41b05e1afadc68657f30467797903bccfd4aa38a1b956aeee09`，大小 19,579 字节。
- 从锁定 data 子模块补齐 `Chat/dict.txt`，Git blob 为 `fc6075f64943e1861c420db4da38063de9d8afc5`。其他插件也可能按需下载素材；当前仍不是全部素材包。
- 帮助命令原本要求正文完全相等；实际 QQ 事件中的 `/help\n` 没有匹配。将注册抽到 `internal/basiccommands`，只对帮助匹配去除首尾空白；保留原命令别名、优先级及私聊/群聊 @bot 限制。消息 `raw_message` 会省略换行，不能仅凭普通日志判定正文完全一致。

日志位于 `.local/onebot/runtime/launcher-logs/`、`.local/smoke-restart/`、`.local/zerobot/logs/`，由 Git 忽略。日志可能含账号、聊天内容、WebUI 密钥或二维码链接，仅在本地保存。
