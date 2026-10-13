# 本地插件控制补丁

日期：2026-10-03。复制锁定模块源码并保留各目录 LICENSE；不修改全局 Go 模块缓存。

| 目录 / 锁定版本 | 修改 |
| --- | --- |
| zbpctrl / v1.7.1 | 单锁读取/写入、提交后清缓存、开关错误返回及读版本；全局/群继承元数据；默认翻转持久化；保留群参数和默认标志；SetData 读改写锁 |
| sqlite / v1.7.2 | 新增 Transaction、TransactionContext 和 ExecContext，原 API 保留；供控制写入事务、限时恢复及 VACUUM INTO 使用 |
| go-onebot-agent / v0.0.0-20260425093914-9c92a03776e3 | 可选 HistoryStore 接口、记录错误返回、恢复事件的图片预处理；原 JSON 解析/权限保持，详见 go-onebot-agent/LOCAL_PATCH.md |
| zbputils / v1.7.2-0.20260510095247-3e9f8cf88d2d | 聊天开关/还原/默认翻转不再忽略保存错误；批量命令报告部分失败；默认值读取使用锁保护；响应/沉默保留原 OnlyToMe 和管理员权限，响应类命令拒绝相似前缀与额外参数 |

根 go.mod 的 replace 指向上述目录。现有其他补丁仍见 go-registry/LOCAL_PATCH.md。

2026-10-03 Agent 扩展：zbputils/chat 实现私有 SQLite 历史、完整记录与有界请求输入，设定重载保留历史、显式重置留档；aichat 移除固定 8 轮循环并阻止重复动作周期。记录方法和 ResetAgents 新增 error 返回，应用端检查错误；外部插件若依赖旧的方法值/接口签名须迁移。行为与实测见 [Agent 记录](../docs/local-admin/agent-history-2026-10-03.md)。

响应命令保留原 OnlyToMe、私聊和全局权限与全局覆盖规则。AI 阶段将响应缓存移至各 Manager，响应读改写在同一锁内、成功提交后更新缓存；扩展读取独立于响应门控。完整程序验证与部署见 [群响应修复](../docs/local-admin/response-command-fix-2026-10-02.md)。

新 __switch_inherit 表仅保存显式继承标记，原插件表的 gid/disable 数据保持可读。继承不删除 62 位参数和默认翻转标志。无显式全局时，全局查询表示默认状态，不再把缺失缓存标记误当成禁用。对群/私聊仍保持显式全局优先的上游实际规则。

Enable/Disable/Reset 新增 error 返回，原直接调用可忽略返回值，已使用的聊天指令改为检查错误。Reset(0) 现在可恢复全局继承；Reset(群) 保留参数，和上游原先删整行的行为不同。旧二进制不识别继承标记，单独回退旧程序不能保证新的继承语义；更新前的完整数据库快照已保留。

后台安装 BeforeSwitch 后，网页和聊天开关及默认翻转都先建立同一控制数据库快照；关闭后台时不安装此钩子，兼容无后台运行。AI 阶段新增 Manager.BeforeExtra，覆盖共享扩展写入口；aichatcfg/llm 的 BeforeData 覆盖网页与聊天群参数写入，其他业务计数不安装此钩子。控制策略阶段新增 Manager.BeforePolicy，覆盖响应/沉默、用户封禁/解封及插件按群/全局禁止/允许，备份失败拒绝写入。批量聊天命令逐项执行，不构成全批事务。后台操作不构造 QQ 事件，也不调用需要聊天上下文的 OnEnable/OnDisable 回复回调。

AI 阶段新增 ExtraJSON/UpdateExtraJSON、DataVersion/UpdateData，统一版本比较、字段合并、备份和提交；SetData 保留开关、默认标志及继承语义。群参数读改写避免聊天和网页丢失不同字段更新。

zbputils/chat 移除可变全局 AC，改为 CurrentConfig/ConfigOf 配置快照及共享 ConfigRepository；NewExtraSet* 参数由字段指针改为字段名。根仓库所有使用者已迁移；自定义外部插件若直接使用 AC 或旧 setter，需要相同迁移。这是源码 API 变化，不能声称与所有上游扩展无改动兼容。磁盘字段名与原数据保持兼容，未编辑未知字段保留。

验证使用根模块的依赖版本：go test ./internal/localadmin ./internal/basiccommands ./plugin/llm ./plugin/tarot -count=1 -timeout=60s；相关包 go vet；完整 Windows 程序由 scripts/smoke-local-admin.go 验证。功能并发测试不能替代全仓 race 检测；未声称修复整个控制依赖中所有缓存。AI 细节见 [AI 设置记录](../docs/local-admin/ai-settings-2026-10-02.md)。

控制策略写入在 Manager 锁内读取持久化状态，区分缺失行与数据库读取失败；实际修改前快照，提交后才更新缓存。重复封禁/解封/禁止/允许不备份；重复响应/沉默返回可识别的 ErrAlreadyResponding/ErrAlreadySilent。无本地响应行时，沉默操作不删除全局响应，不建立无变化快照。按群允许不删除全局禁止。Ban/Permit 新增 error 返回；直接忽略返回值的调用仍可编译，外部插件若以旧无返回方法签名存储方法值或实现接口需迁移。聊天指令检查返回值，逐项报告成功或失败；失败回复不包含底层 SQL、路径或配置。详细验证见 [控制策略备份记录](../docs/local-admin/control-policy-backups-2026-10-02.md)。

恢复阶段新增 Manager.RestoredLocked：数据库事务成功后刷新默认翻转并清空开关、响应、黑名单和封禁缓存。IsBlocked/IsBannedIn 在同一锁内查库与缓存写回，避免恢复后的过期读者；Ban/Permit 查询与删除使用与存储一致的 int64 ID，群级未命中继续检查全局封禁，持久化失败不提前修改缓存。Ban 不再因群写入失败而意外扩大到全局封禁。隔离恢复、失败回滚、并发写入、符号位 ID 与全局封禁回归见 [备份恢复记录](../docs/local-admin/backup-restore-2026-10-02.md)。
