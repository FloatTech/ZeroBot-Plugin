package control

// GroupConfig holds the group config for the Manager.
type GroupConfig struct {
	GroupID int64 `db:"gid"`     // GroupID 群号
	Disable int64 `db:"disable"` // Disable 默认启用该插件
}

// BanStatus 在某群封禁某人的状态
type BanStatus struct {
	ID      int64 `db:"id"`
	UserID  int64 `db:"uid"`
	GroupID int64 `db:"gid"`
}

// BlockStatus 全局 ban 某人
type BlockStatus struct {
	UserID int64 `db:"uid"`
}

// ResponseGroup 响应的群
type ResponseGroup struct {
	GroupID int64  `db:"gid"` // GroupID 群号, 个人为负
	Extra   string `db:"ext"` // Extra 该群的扩展数据
}

// Options holds the optional parameters for the Manager.
type Options[CTX any] struct {
	DisableOnDefault  bool
	Extra             int16     // 插件申请的 Extra 记录号, 可为 -32768~32767, 0 不可用
	Brief             string    // 简介
	Help              string    // 帮助文本信息
	Banner            string    // 背景图路径, 可为 http 或 本地 路径
	PrivateDataFolder string    // 全部小写的数据文件夹名, 不出现在 zbpdata
	PublicDataFolder  string    // 驼峰的数据文件夹名, 出现在 zbpdata
	OnEnable          func(CTX) // 启用插件后执行的命令, 为空则打印 “已启用服务: xxx”
	OnDisable         func(CTX) // 禁用插件后执行的命令, 为空则打印 “已禁用服务: xxx”
}
