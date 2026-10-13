# 新增连接的页面确认

进入“连接设置”填写新增连接，点击“备份并新增连接”后，在页面弹窗中核对新增方向与保存说明，点击“确认新增”才提交；“取消”或Esc保留新增草稿且不提交请求。点击“放弃新增”同样先弹出页面确认，“确认放弃”只清空新增表单，取消保留原值。默认焦点放在取消按钮，弹窗期间不能操作背景表单。

弹窗只显示固定操作说明，不显示Token、地址或主人字段值。确认前后核对原会话、配置版本、表单内容与生命周期；发生退出、过期或重置时关闭确认并取消旧操作，不能用旧确认提交新会话。保存仍按原接口执行版本检查、写前备份和待重启规则；不会自动连接、监听、重启或修改NapCat。

本说明覆盖新增保存及放弃新增；2026-10-13已有连接/主人草稿的确认也已改为[页面弹窗](connection-draft-confirmation.md)。两类确认不会相互嵌套或替换。弹窗重置或会话过期会取消等待中的操作，后台请求完成前原忙时保护保持。以下2026-10-12记录保留当时验证范围。

## 2026-10-12 验证与部署（DEV-70 / REQ-67 / TC-90）

204项完整界面回归及其中21项相关流程通过，新增3项覆盖取消/Esc/默认焦点/不使用原生确认、旧草稿或设置替换、过期重置/重复提交与新会话隔离。既有64项布局及新增弹窗8项（1280/768/375/320px×深浅主题）通过；相关Go模块、vet、独立构建和22:18:36完整Windows回归通过。整机目录`.local/smoke-admin/run-3327827690/`；证据`.local/connection-confirm-{focused-ui,ui-tests,layout,dialog-layout,go-tests,vet,build,smoke}.txt`、`-layout-result.json`和`.local/connection-confirm-layout/result.json`。

双模拟账号独立`run-1231404508`于22:14:58～22:29:12运行，当前程序SHA-256 `8C0204076E0E354758D4F9A0BC342BE6B864E399F6726CC3A3E924A7E8641288`。20项实际Chrome通过：页面取消与确认放弃、默认取消焦点/固定文本/无原生JS确认、重复地址拒绝并保留草稿、新增选择及其他连接/Token/未知大整数保持、删除预览确认/取消/换目标失效、删除未加载项恢复启动值、删除运行项仍保持运行通信、最后一条保护、重新新增/保存后实际重启与会话/持久状态。另8项实际弹窗主题/小屏布局通过，视口和原主题恢复；新旧测试标签均已关闭。

控制器2次启动、4轮双账号往返共8条模拟回复通过，原测试密码及凭据保持、旧会话401、保存配置加载后提示一致，正常stop退出。记录各目录`result.json`、`browser-cases.json`、`browser-result.json`、`browser-dialog-layout.json`、`browser-confirmation-mobile.png`、`browser-after-restart.png`及汇总`.local/connection-confirm-browser-result.json`。控制器的`browserAcceptancePassed=false`仍只表达它本身不验证页面；独立浏览器结果按其已列范围判定。

原始观察保留3条错误测试预期，不计通过数：旧夹具使用`Url`，重新新增采用规范`URL`，恢复相同URL/Token仍存在字段表示差异，页面正确报告同数量内容不同并待重启；另一次观察把重登录后保留的连接页当成概览页，不能据此判断账号离线。实际配置字段、页面结果与修正后的用例均保留，不隐藏原探测失败。实际Chrome的Esc输入尝试被原始CDP接口拒绝，该Chrome绑定没有可用的对应按键工具，因此未计实际Esc通过；离线真实Chrome键盘Esc用例通过，实际按钮取消/接受均成功。见`actual-escape-tool-limit.json`与`observer-expectation-correction.json`。

既有连接/主人草稿的原生确认仍可能遇到浏览器工具焦点限制，单独保留未完成范围；此轮新增保存与放弃新增已可自动操作，不以本轮完成扩大到全部真实角色/业务/P0验收。

日常后台正常自动填充登录、正确账号/108插件/无近期异常及未保存新增草稿的取消/确认放弃已自动核对，未提交真实连接写入；预览后配置摘要再次保持，见`.local/connection-confirm-daily-browser.json`与`-daily-preview-check.json`。最终日常弹窗截图`.local/connection-confirm-daily-confirmation.png`已视觉核对，截图草稿取消并清空、自建标签关闭。独立实例最初的手机截图未完整包含弹窗，不作为视觉通过证据；原图保留，另有8项离线弹窗PNG和8项实际几何/主题指标。最终跨文件与690个相对文档链接核对通过，见`.local/connection-confirm-evidence-check.json`。
