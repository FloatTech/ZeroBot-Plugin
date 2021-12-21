package control

// grpcfg holds the group config for the Manager.
type grpcfg struct {
	GroupID int64 `db:"gid"`     // GroupID 群号
	Disable int64 `db:"disable"` // Disable 默认启用该插件
}

type ban struct {
	ID      int64 `db:"id"`
	UserID  int64 `db:"uid"`
	GroupID int64 `db:"gid"`
}

// Options holds the optional parameters for the Manager.
type Options struct {
	DisableOnDefault bool
	Help             string // 帮助文本信息
}
