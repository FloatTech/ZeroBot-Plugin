# 第一版验证摘要与复验入口

更新日期：2026-10-13。公开文档记录可重复执行的方法与验证范围；真实QQ明细、私有日志、截图、凭据和原始结果不随源码上传。

| 类别 | 已执行范围 | 边界 |
| --- | --- | --- |
| 后端与配置 | 认证/来源/CSRF、共享持久化、冲突/写失败保护、备份/恢复、统计及脱敏 | Go包测试与隔离程序验证；不改日常配置演练恢复 |
| 页面与布局 | 现有242项界面回归、35项相关概览回归、64项离线布局；实际Chrome备份/AI/日志/连接/主题等流程 | 实际浏览器资料保留原执行范围，离线检查不替代所有浏览器行为 |
| 观察工具 | 210项模型/回复/资源联合工具回归，45次模拟GPT与15分钟模拟资源观察 | 模型HTTP、任务和发送回执各有口径，不强求计数相等 |
| 当前日常资源 | 20分钟、240样本、CPU平均约0.002018%、私有内存最大增长72KiB | 资源窗口，不是配套业务负载，也不能证明长期无泄漏 |
| 真实QQ | 指定群/主人私聊文字、实际引用、识图、热词出图、群单条语音与私聊限制 | 仅记录已验证业务，不扩大到所有消息类型 |
| 真实角色 | 普通成员/非主人私聊/管理员/真人群主的本轮会话参数允许与拒绝 | 临时身份已恢复；初次管理员无响应，刷新消息成员缓存后正向测试成功 |
| 主人真实Agent | 查询本群、保存记忆、4条群回执、4个完成任务、6次成功动作响应，单次重启后代号延续 | 最初查询收到两条事件但只有一个完成任务；事件不能当作任务数 |
| 取消项目 | 非主人Agent真人测试、客户端体验、真实持续业务负载 | 使用者取消，不记作实测通过，后续出现问题再优化 |

最终交付已从显式文件清单生成独立源码副本并完成构建、相关Go检查、界面及工具回归和Windows隔离整机检查，文档链接和私有资料排除核对通过。结果汇总见[第一版说明](release-v0.1.md)。

## 开发者复验

在仓库根目录执行，先初始化data子模块并按[运行维护](operations.md)准备本地环境。测试使用合成账号与独立目录；不要把日常凭据传入夹具。

```powershell
.\scripts\build-windows.ps1 -OutputName zerobot.exe
go test ./internal/... ./kanban ./plugin/llm ./plugin/aichat ./plugin/aichatcfg ./plugin/airecord ./plugin/wordcount ./plugin/tarot -count=1 -timeout=120s
go run scripts/smoke-local-admin.go -binary-name zerobot.exe
go run scripts/test-admin-system.go -mode restart -binary-name zerobot.exe
.\scripts\test-service-paths.ps1
```

前端测试按实际文件列表执行scripts/test-*-ui.mjs；离线布局使用scripts/test-appearance-layout.mjs。模型/回复/资源工具回归入口为scripts/test-observation-evidence.mjs、scripts/test-reply-evidence.mjs与scripts/test-combined-observation.mjs，其他用例按scripts/test-*.mjs的用途运行。保留原始结果，失败不能改写成通过。

本地依赖补丁可在各自目录单独开发：控制模块和工具模块的go.mod已显式引用相邻补丁，主程序根go.mod也保留同一组替换。隔离/观察启动器默认读取.local/bin/zerobot.exe，可通过binary-name或BinaryName参数指定其他已构建副本；不依赖开发者的内部构建文件名。

依赖模块原有TestMaru只是无条件失败的演示，没有业务断言，现保留名称并明确跳过，不计功能测试通过。最终整机夹具使用data子模块资源及经典诗歌资源缓存，不复制日常控制库或私有模型配置。
