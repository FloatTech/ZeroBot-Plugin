# 主题与小屏验证

用户已确认日志可以正常下载，TC-28 的实际下载操作记为通过；其他导出人工矩阵保留原边界。本次继续实现外观和窄窗口布局。

## 使用与行为

登录页与后台顶栏新增“外观”，支持跟随系统、浅色、深色。默认跟随系统；明确选择浅色/深色后，系统变化不会覆盖选择。初始主题在页面内容加载前应用，两个选择器保持同步。

只将选择写入当前浏览器、当前站点的 `zerobot.localadmin.theme.v1`，不写机器人配置或凭据。重开/退出后保留，同站点其他标签页通过存储事件同步。存储不可用仍可在当前页切换；刷新恢复跟随系统。未知存储值按跟随系统处理；不支持系统配色查询时回退浅色。

所有页面的表单、插件卡片、日志、错误/状态徽标及用法弹窗使用同一套明暗配色。正文、辅助文字、按钮、错误和警告的配色对比度检查均达到 4.5:1。保留键盘焦点提示。

640px 及以下，导航移动到顶部并保留名称，移除计数徽标以留出空间；表单、页标题和按钮按需换行。输入控件使用 16px 字体，常用按钮至少 44px 高；长插件名、日志和备份编号可折行。后台仍仅本机访问，不提供手机远程访问。

## 验证

| 项目 | 结果 |
| --- | --- |
| `CGO_ENABLED=0 go test ./internal/localadmin -count=1 -timeout=120s` | 通过，含新增离线资源及原后台接口回归 |
| `go vet ./internal/localadmin`、`go vet scripts/smoke-local-admin.go` | 通过 |
| `node --check internal/localadmin/assets/theme.js`、渲染脚本语法 | 通过 |
| `node --test scripts/test-theme-ui.mjs scripts/test-log-ui.mjs scripts/test-backup-ui.mjs` | 13 项全部通过；新增 4 项验证初始化/系统切换、选择保留/跨标签、异常偏好/存储禁用及配色对比度 |
| `node scripts/test-appearance-layout.mjs` | 56 组通过：320/375/640/1280px × 浅色/深色 × 登录及六个后台页面；没有横向溢出，主题控件与具名导航可见 |
| 截图核对 | 保存 28 张 375/1280px 截图，核对明暗、登录、概览、连接和插件页面；发现并修正小屏计数徽标选择器优先级，再次渲染通过 |
| `go build -trimpath -o .local/bin/zerobot.exe .` | Windows CGO=0 最终构建通过 |
| `go run scripts/smoke-local-admin.go -binary-name zerobot.exe` | 00:34:12 最终整机回归通过，`appearanceAssetsPassed` 及全部原标志为 true；日志导出、原指令、插件开关及备份恢复正常 |

独立渲染证据 `.local/appearance-layout/result.json` 及同目录 PNG。渲染脚本使用新的无登录 Chrome 上下文，将静态 HTML/CSS/主题脚本和合成数据直接载入，阻止全部网络请求；没有访问实际后台、实际账户或操作用户已登录的 Chrome。此项验证布局与静态外观，不代表日常数据或完整业务交互。

整机证据 `.local/smoke-admin/run-670955380/result.json`。使用独立配置、模拟 OneBot 和本地模拟模型，不发送真实 QQ 消息或调用真实模型。早期构建结果不作为最终部署证据。

## 本机更新

更新前控制库快照 `20261002T163519.698259200Z-5fcc5e440b8b20b3-control`；旧程序 `.local/bin/zerobot-before-appearance-20261003-003513.exe`。仅重启 ZeroBot，NapCat 登录保留；预期账号、OneBot 只读 API 与握手通过。启动 JSON 与控制库摘要未变，启动 fatal/panic 未发现，所有原未认证接口拒绝检查通过。证据 `.local/appearance-before.json`、`.local/appearance-deployment.json`。

既有外部 MD5 校验服务器不可达警告仍存在；本次没有将跳过远端校验记作资源完整性通过。主题不触发后台 API 写入，不改变日志导出范围、密钥及响应的 @ 规则。

## 剩余验收

刷新后台重新登录后可使用新控件。日常页面、真实数据和完整手工矩阵继续记为 TODO-UI-01；上述独立浏览器渲染不能代替业务验收。真实 AI 仍为 TODO-AI-01，备份清理策略未实施；没有重复上一构建的持续性能测量。

201项界面行为、离线64项布局、Go后台测试/vet/Windows构建通过，19:06:39完整独立Windows回归全部标志通过（.local/smoke-admin/run-1797868651/result.json）。修复版真实Chrome的320/375/640/1280px×深浅×登录/七后台页面64项矩阵通过；375px实际6个控件为51px高/16px字体/全宽/背景匹配，主题刷新与跨标签同步正确。run-4024527018已停止，视口及测试主题恢复。两类64项分别记录，不计作新增128项用例。原日志控件另完成12组核对，实际下载文件内容和原生日期输入明确保留待办；首轮下载等待超时/正文为空、日期工具报错、测试实例正常到期及视口定位/只读计算观察问题均保留，不改写为产品失败或全部通过。详见[日志与主题自动验收](logs-appearance-browser-acceptance.md)。
