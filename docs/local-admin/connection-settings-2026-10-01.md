# 连接设置与启动配置备份验证

## 实现范围

- 登录后读取、编辑 `-c` 指定文件中的已有 `ws`/`wss` 连接与 `zero.super_users`。不使用 `-c` 时显示不可编辑。
- Token 只返回“已配置/未配置”；URL 的用户凭据、查询和片段脱敏。保留、替换和清除分别提交，清除页面有确认勾选；占位符不能替代真实 Token。
- 校验 ws/wss 地址与端口、主人正整数/重复/数量、Token 长度/控制字符；主人 ID 以字符串在 API 中传递，避免 JavaScript 大整数精度丢失。读取与保存后的配置均限 1 MiB，格式化导致超限时拒绝保存并保留原文件。
- 保留其他连接、其他字段和未知字段数值精度。重复 JSON 字段和字段别名歧义会拒绝；仅改动指定字段。
- HMAC 版本标记阻止旧页面覆盖；服务内串行写入，写前重复检查外部变化。
- 每次实际修改先生成私有旧配置备份，再写入同目录临时文件、同步并通过 Windows MoveFileEx 替换。备份失败不覆盖；替换失败保留原文件和已完成备份；无修改不写盘。
- 保存值与本次启动快照分开，改变连接/主人配置后提示待重启。认证、来源与 CSRF 检查沿用已有后台；响应禁止缓存。

备份封装版本为 1、范围为 `startup-config`，包含 `createdAt` 和旧配置 `config` 字段。本机保存到 `.local/admin/backups/`；含 Token，限制为当前 Windows 用户及 SYSTEM。用法见 [运行操作](operations.md)。

## 可重复验证

在项目根目录执行，环境为 Go 1.26.1、Windows amd64、CGO=0：

```powershell
$env:CGO_ENABLED = '0'
go test ./internal/localadmin ./internal/basiccommands -count=1 -timeout=60s
go vet ./internal/localadmin
node --check internal/localadmin/assets/app.js
go build -trimpath -o .local/bin/zerobot-next.exe .
go run scripts/smoke-local-admin.go -binary-name zerobot-next.exe
```

完整程序隔离模拟结果，完成时间 22:56:05（Asia/Shanghai，包含格式化大小保护修复）：

| 检查 | 结果 |
| --- | --- |
| 后台认证与真实编译插件注册表 | 通过 |
| 连接读取无 Token、保存前备份、非法 CSRF 和旧版本拒绝 | 通过 |
| 保存后 help、带换行 help、服务列表图片回复 | 通过 |
| 真实 QQ 消息发送 | 未发送；使用模拟 OneBot 与独立配置 |

## 本机部署

22:52 已备份旧程序为 `.local/bin/zerobot-before-connections-20261001-225200.exe`，停止本项目 ZeroBot 进程树、替换程序并通过已有 Windows 计划任务重新启动。NapCat 保持登录；部署前后日常配置 SHA256 一致，新程序与测试构建 SHA256 一致。本次没有通过管理接口修改日常机器人配置。

| 部署检查 | 结果 |
| --- | --- |
| 后台 `/` | HTTP 200，包含连接设置入口 |
| 内嵌 `app.js` | HTTP 200，包含连接 API |
| 未认证 `/api/connections` | HTTP 401 |
| ZeroBot 实际进程 | 正常运行，程序路径符合本项目 |
| NapCat 登录与三个只读 API | 通过 |
| ZeroBot 与配置指定机器人账号的实际握手 | 新日志确认成功，无启动 fatal/panic |

重启使原后台会话失效；刷新页面并重新登录即可查看新入口。API/静态资源检查不代表浏览器交互或视觉验收。

## 待完成与边界

- DEV-05/06 仍标记进行中。页面校验、清除确认及配置修改后的真实连接重启联调待验收；本次只在模块测试中重新加载修改后的配置，完整程序隔离测试验证保存后继续处理消息。
- 尚无连接新增/删除、NapCat 配置联动、插件/AI 写接口、SQLite 备份、手动备份、备份列表与恢复。JSON 写前备份不能替代完整恢复演练。
- 当前原子替换及故障注入已验证；真实断电、磁盘硬件故障及多文件事务未验证。不承诺任意外部编辑器与后台共享锁，保存期间不要同时编辑配置。
- 不把此阶段视为全部插件 Windows 兼容、完整资源包或全部 P0 验收通过。进度及用例分别见 `progress.md`、`self-test.md`。
