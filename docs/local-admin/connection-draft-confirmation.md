# 连接与主人账号草稿的页面确认

入口为本机后台登录后的“连接设置”。已有连接地址或 Token 有未保存更改时，切换连接先确认；主人账号草稿保留。点击“放弃当前连接更改”或“放弃主人账号更改”只撤销对应部分。重新加载配置和退出登录会先提醒丢弃连接设置草稿，包括尚未保存的新增连接。

确认在页面内显示，默认焦点为“取消”。取消或按 Esc 保留草稿，不发送该操作的请求；确认后按原有规则切换、重新读取或退出。确认说明只显示操作与连接序号，不显示地址、Token 或主人账号值。登录失效、配置重新读取或待确认草稿变化后，旧确认不得继续操作新页面。

局部放弃不修改已保存配置，也不影响运行中的连接。重新读取成功后采用已保存配置；失败保留原草稿。连接保存仍需备份与版本核对，保存后按提示重启生效。本轮不改变插件或 AI 表单的确认方式；真实 QQ 角色和持续业务负载按各自待办验证。

## 自动验证（DEV-71 / REQ-68 / TC-91）

212项完整界面回归、40项连接相关回归通过，其中新增8项覆盖默认取消焦点、Esc/外部关闭、三部分草稿保留、目标草稿变化与另一部分保留、旧会话/设置替换、并发与新增弹窗互斥、旧关闭事件及迟到退出响应。既有64项页面布局与两类弹窗16项离线布局通过；相关Go模块、vet、构建通过。00:25:42完整Windows隔离回归通过，目录`.local/smoke-admin/run-4265110890/`。证据为`.local/draft-confirm-{ui-tests,focused-ui,layout,dialog-layout,go-tests,vet,build,smoke}.txt`、`-layout-result.json`及`.local/draft-confirm-dialog-layout/result.json`。

实际Chrome独立目录`.local/connection-browser-acceptance/run-2686500409/`的20项核对通过，涵盖已有连接/主人局部放弃、切换、重新加载、退出的取消和确认、三部分草稿范围、固定说明与默认焦点、原密码重新登录、保存备份与重启、旧登录拒绝及双模拟账号。8项实际弹窗几何/主题/焦点检查通过，视口与原主题恢复，自建测试标签关闭。控制器两次启动、三轮双账号检查共6条模拟回复通过，已停止；没有调用真实QQ或模型。控制器`result.json`与浏览器`browser-result.json`分别记录，互不代替。

桌面`draft-confirmation.png`和小屏`draft-mobile-confirmation-verified.png`已查看且包含弹窗。循环截图保留但不作为视觉通过证据；实际Chrome Esc输入受工具能力限制，不计通过，离线Chrome键盘Esc已验证。首轮整机资源检查仍查找旧反向条件字符串，脚本标记已同步为当前会话/配置保护及新弹窗入口后整机通过；原失败保存在`draft-confirm-smoke-initial.txt`。首轮完整界面中概览测试仍按同步退出立即完成旧请求，更新为等待实际退出请求后212项通过；原211/212记录保存在`draft-confirm-ui-tests-initial.txt`，不计通过。

新构建SHA-256：`0CF3DE9A2ACCE07772025577D0BBE59A76958443EECEBA9AB73B8CE57FC4E6F0`。本轮没有重复历史五分钟性能测量，不把上述模拟验证扩大到真实QQ其他角色或长期负载。

## 部署与回退

刷新后台重新登录即可使用。旧程序保存在`.local/bin/zerobot-before-draft-confirm-20261013-003334.exe`，SHA-256为`8C0204076E0E354758D4F9A0BC342BE6B864E399F6726CC3A3E924A7E8641288`；回退信息`.local/draft-confirm-rollback.json`，控制快照`20261012T163006.174452400Z-7c3d1f548e4aa48f-control`。只回退程序时不要恢复配置或停止NapCat，启停步骤见[运行维护](operations.md)。

日常后台已用Chrome正常自动填充登录，核对账号87654321、108插件和近期警告/错误均为0。未保存的主人草稿页面确认已自动执行取消与局部放弃，草稿清空、自建标签关闭，没有提交日常配置写入；五项配置摘要再次保持。证据`.local/draft-confirm-daily-browser.json`、`-daily-preview-check.json`、`-daily-config-check.json`；实际弹窗截图`-daily-confirmation.png`已查看。最终文档相对链接、资源入口、构建/回归/部署证据核对通过，见`.local/draft-confirm-evidence-check.json`。
