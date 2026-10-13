# 本地后台基础：启动与验证

## 本机使用

后台地址默认是 `http://127.0.0.1:3010/`。首次启动生成 `.local/admin/setup-code.txt`，在页面选择该文件，设置至少 12 个字符的后台密码即可进入。设置码成功使用后失效并删除；此密码与 QQ 登录分开，不能用 OneBot Token 登录。

后台凭据仅保存随机盐与 PBKDF2-HMAC-SHA256 派生值（600,000 次），保存在 `.local/admin/credential.json`。会话存在内存中，有效期 8 小时；退出或重启后原会话失效。Windows 数据目录及凭据文件使用受保护的 DACL，仅当前用户与 SYSTEM 拥有访问权限。

用独立计划任务宿主运行时，在项目根目录执行：

```powershell
# 开启本机后台；本操作不改机器人账号、昵称或 OneBot 配置
.\scripts\configure-local-admin.ps1
.\scripts\stop-local-services.ps1 -Service ZeroBot
.\scripts\start-local-services.ps1 -Service ZeroBot

# 关闭后台，保留原机器人启动方式；同样需重启 ZeroBot
.\scripts\configure-local-admin.ps1 -Disable
.\scripts\stop-local-services.ps1 -Service ZeroBot
.\scripts\start-local-services.ps1 -Service ZeroBot
```

手动启动也可使用：

```powershell
.\.local\bin\zerobot.exe -c .local/config.json -admin -admin-addr 127.0.0.1:3010 -admin-data .local/admin
```

不传 `-admin` 时不会创建后台监听或读取后台凭据。地址只接受字面 IPv4 `127.0.0.1`；端口占用时明确报错并停止启动，不静默换端口。机器人无法连接 OneBot 时后台仍可加载并显示等待连接；概览标记的是 OneBot 传输连接，不替代 QQ 登录状态检查。

## 接口约定

| 接口 | 权限/作用 |
| --- | --- |
| `GET /api/session` | 只返回首次设置/登录状态；已登录时提供当前会话 CSRF 校验值 |
| `POST /api/setup` | 本地设置码与新密码；只能在尚未初始化时使用 |
| `POST /api/login` | 后台密码登录；每分钟最多五次失败尝试 |
| `POST /api/logout` | 已登录并持有 CSRF 校验值，注销当前会话 |
| `GET /api/overview` | 已登录；运行时间、构建信息、当前 OneBot 账号、连接及插件数量 |
| `GET /api/plugins` | 已登录；已编译插件元数据及启动时默认状态 |

全部接口仅接受回环来源及正确 Host，跨站 Origin/Fetch Metadata 被拒绝；写请求必须为同源 JSON 并带 `X-Local-Request: 1`，已登录写操作另需 `X-CSRF-Token`。Cookie 为 HttpOnly、SameSite=Strict；服务仅本机 HTTP，故不设置要求 HTTPS 的 Secure 标志。响应不缓存，不记录请求体或凭据。网页不把会话放入 URL/localStorage，不加载外部字体、头像、图标或脚本。

静态服务只暴露内嵌的三个页面文件，不映射本地数据目录或备份。插件用法和名称作为文本节点展示。

## 可重复执行的自测

在项目根目录执行；运行中的程序需使用另一个输出文件名，避免覆盖被 Windows 锁定的可执行文件：

```powershell
.\scripts\build-windows.ps1 -RunTests -OutputName zerobot-next.exe
go run scripts/smoke-local-admin.go -binary-name zerobot-next.exe
```

集成脚本创建独立 `.local/smoke-admin/run-*` 目录，只复制必要启动素材，使用假账号、本地模拟 OneBot 和测试密码；不读取实际机器人配置、控制数据库或后台凭据，不向真实 QQ 发消息。运行结果写入该目录的 `result.json` 和 `.local/smoke-admin/latest.json`，退出时停止自己的测试进程。未指定 `-binary-name` 时测试 `.local/bin/zerobot.exe`。
