# 旧 WebUI 与数据接口复用结论

## 结论

新建独立、内嵌的本地 HTTP 服务与轻量网页；复用当前进程的 ZeroBot 账号注册表、插件注册列表和控制数据。旧 `control/web` 的认证、路由、日志与网页不直接启用。界面使用原生 HTML/CSS/JavaScript，由 `go:embed` 打包，不增加 Node 前端构建步骤或 CDN 依赖。

## 版本与证据

| 锁定模块 | 检查文件 | 观察与影响 |
| --- | --- | --- |
| `zbputils v1.7.2-0.20260510095247-3e9f8cf88d2d` | `control/web/router/router.go` | `/api/getLog`、`/api/data`、`/api/getBotList` 注册在不验证登录的路由组，不能直接用于新后台 |
| 同上 | `control/web/middleware/middleware.go`、`controller/manage.go` | CORS 反射任意 Origin 并允许凭据，WebSocket `CheckOrigin` 恒为 true；新入口采用固定 Host/Origin、回环来源和会话校验 |
| 同上 | `control/web/model/webui.go` | 明文密码保存在 `data/webui/user.db`，查询拼接用户名及密码；新认证不使用旧数据库 |
| 同上 | `control/web/controller/manage.go`、`gui.go` | 全局单个日志/消息连接、直接向连接写入并替换日志输出；导入时安装消息监听器。新后台不导入此包，不改变现有日志去向 |
| `ZeroBot-Plugin-Webui v1.0.1-0.20230412164529-cfe878675931` | `package.json`、嵌入产物 | 基于 Vben/Vue 的较大页面与旧接口绑定，头像/插件横幅取外部 URL；重建小型离线界面更适合当前范围 |
| `ZeroBot v1.8.3-0.20260407110317-d207de3b6c79` | `callermap.go`、`driver/wsclient.go`、`driver/wsserver.go` | 连接时写入 `APICallers`，断线时删除；新概览每次读取该表，不把已有账号列表永久缓存为在线 |
| `zbputils` | `control/managers.go`、`control/register.go` | `ForEachByPrio` 来自本次实际注册列表；可用于展示已编译插件及元数据，不扫描目录猜测可启用插件 |
| `zbpctrl v1.7.1` | `control.go`、`data.go` | 开关、群配置与扩展配置使用同一 Manager/SQLite；网页不得建立另一套插件状态数据库 |

## 状态与配置边界

| 数据 | 来源 | 第一阶段实现/后续处理 |
| --- | --- | --- |
| 后台认证 | `.local/admin/credential.json`；首用设置码 | 新建独立认证，仅保存带随机盐的密码派生值；不复用 QQ/OneBot 密码或 Token |
| 后台启用、地址 | `.local/admin/options.json` 或启动参数 | 默认不启用；仅 `127.0.0.1`；改端口/启用状态需要重启 ZeroBot |
| 机器人账号 | 当前 `zero.APICallers` | 每次查询；只标明 OneBot 连接，不能据此宣称 QQ 平台在线 |
| 插件元数据 | 当前 `control.ForEachByPrio` | 启动时快照；默认状态可能已包含已有持久化的默认翻转，页面标记“启动时默认状态” |
| 插件开关/群参数/AI 扩展 | `data/control/plugins.db` | 复用同一 Manager；后续写入口必须先解决失败回报、并发及备份 |
| OneBot 地址、Token、主人等 | 启动使用的 JSON，例如 `.local/config.json` | 需重启；DEV-05 以原始 JSON 字段树保存未知字段及未编辑连接，不能重新编码一个裁剪后的结构覆盖原文件 |
| AI 运行配置 | 原 `chat.AC` 已移除，使用控制库与 `CurrentConfig/ConfigOf` 快照 | DEV-08 网页和聊天复用 ConfigRepository，原磁盘字段兼容 |
| 日志 | 现有 logrus 输出 | DEV-09 已实现 1000 条/1 MiB 脱敏缓冲及有认证短请求读取；切页/暂停/退出取消轮询，日志钩子不访问控制库 |

`zbpctrl.IsEnabledIn` 的实际实现为：存在显式全局开关时全局状态优先；没有显式全局开关时使用群级状态，最后使用默认值。源码旁边的部分文字说明与实际逻辑不一致，以锁定实现及测试矩阵为准。未配置全局值时直接调用 `IsEnabledIn(0)` 还会得到特殊缓存标记导致的禁用值；不能用这个调用表达“没有全局覆盖”。后续页面需明确“继承默认”与“显式全局开关”。

上游锁定版本的 `Enable/Disable/Reset` 不返回写库错误，并在持久化完成前更新缓存；已在本地补丁中统一开关事务/锁、错误回报及提交后缓存失效。显式继承不删除插件参数；默认翻转标志正确持久化。DEV-08 已将 `chat.NewExtraSet*` 改为共享字段提交，取消可变 AC；扩展/群参数写入使用同一锁，提交成功后才影响后续读取。不认为整个依赖均已修复。

## 备份与恢复方案

JSON 使用版本校验、写前备份和同目录临时文件替换，已验证。插件开关和 AI 群参数采用共享 Manager 锁及 SQLite 事务；扩展更新在同一锁内合并并提交，写前以限时 VACUUM INTO 建立含 WAL 已提交数据的一致性快照；对应失败/并发测试已通过。DEV-10 实现三个独立范围的手动备份和恢复，逐项校验、绑定当前状态的预览与确认、恢复前备份、失败回滚及重启提示；不提供跨文件联合恢复。控制库使用同一锁和事务恢复匹配表，提交后清缓存，详见 [备份恢复记录](backup-restore-2026-10-02.md)。响应/沉默、用户封禁/解封及插件禁止/允许已接入同一写前备份，持久化查找失败不当成缺失状态，备份或写入失败不改缓存、不误报成功；验证见 [控制策略备份记录](control-policy-backups-2026-10-02.md)。业务计数及外部编辑不逐次备份。

## 性能验收口径

本机无后台基线采样（2026-10-01 16:36，现有 ZeroBot 进程）：Working Set 49,209,344 字节，Private Bytes 130,613,248 字节；累计 CPU 4.59375 秒。该采样不是峰值或稳定区间统计，不能单独作为完整性能验收。

概览不调用远程 OneBot API；插件列表现从同一控制数据库读取指定范围的实际状态。浏览器刷新及网页开关操作不会额外发送 QQ 消息，实测与限制见后台基础和插件管理记录。
