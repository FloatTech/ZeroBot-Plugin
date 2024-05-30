// Package types 结构体
package types

// Response 包装返回体
//
//	@Description	包装返回体
type Response struct {
	Code         int         `json:"code"`    // 错误码
	Message      string      `json:"message"` // 错误信息
	Result       interface{} `json:"result"`  // 数据
	ResponseType string      `json:"type"`    // 待定
}

// BotParams GetGroupList,GetFriendList的入参
//
//	@Description	GetGroupList,GetFriendList的入参
type BotParams struct {
	SelfID int64 `json:"selfId" form:"selfId"` // 机器人qq
}

// GetGroupMemberListReq 获得群成员的入参
//
//	@Description	获得群成员的入参
type GetGroupMemberListReq struct {
	SelfID  int64 `json:"selfId" form:"selfId"`   // 机器人qq
	GroupID int64 `json:"groupId" form:"groupId"` // 群id
}

// AllPluginParams GetAllPlugin的入参
//
//	@Description	GetAllPlugin的入参
type AllPluginParams struct {
	GroupID int64 `json:"groupId" form:"groupId"` // 群id, gid>0为群聊,gid<0为私聊,gid=0为全部群聊
}

// DeleteGroupParams 退群或删除好友的入参
//
//	@Description	退群或删除好友的入参
type DeleteGroupParams struct {
	SelfID  int64 `json:"selfId" form:"selfId"`   // 机器人qq
	GroupID int64 `json:"groupId" form:"groupId"` // 群id, gid>0为群聊,gid<0为私聊,gid=0为全部群聊
}

// PluginParams GetPlugin的入参
//
//	@Description	GetPlugin的入参
type PluginParams struct {
	GroupID int64  `json:"groupId" form:"groupId"` // 群id, gid>0为群聊,gid<0为私聊,gid=0为全部群聊
	Name    string `json:"name" form:"name"`       // 插件名
}

// PluginStatusParams UpdatePluginStatus的入参
//
//	@Description	UpdatePluginStatus的入参
type PluginStatusParams struct {
	GroupID int64  `json:"groupId" form:"groupId"`               // 群id, gid>0为群聊,gid<0为私聊,gid=0为全部群聊
	Name    string `json:"name" form:"name" validate:"required"` // 插件名
	Status  int    `json:"status" form:"status"`                 // 插件状态,0=禁用,1=启用,2=还原
}

// ResponseStatusParams UpdateResponseStatus的入参
//
//	@Description	UpdateResponseStatus的入参
type ResponseStatusParams struct {
	GroupID int64 `json:"groupId" form:"groupId"` // 群id, gid>0为群聊,gid<0为私聊,gid=0为全部群聊
	Status  int   `json:"status" form:"status"`   // 响应状态,0=沉默,1=响应
}

// AllPluginStatusParams UpdateAllPluginStatus的入参
//
//	@Description	UpdateAllPluginStatus的入参
type AllPluginStatusParams struct {
	GroupID int64 `json:"groupId" form:"groupId"` // 群id, gid>0为群聊,gid<0为私聊,gid=0为全部群聊
	Status  int   `json:"status" form:"status"`   // 插件状态,0=禁用,1=启用,2=还原
}

// HandleRequestParams 处理事件的入参
//
//	@Description	处理事件的入参
type HandleRequestParams struct {
	Flag    string `json:"flag" form:"flag"`       // 事件的flag
	Reason  string `json:"reason" form:"reason"`   // 事件的原因, 拒绝的时候需要填
	Approve bool   `json:"approve" form:"approve"` // 是否同意, true=同意,false=拒绝
}

// SendMsgParams 发送消息的入参
//
//	@Description	处理事件的入参
type SendMsgParams struct {
	SelfID  int64   `json:"selfId" form:"selfId"`   // 机器人qq
	GIDList []int64 `json:"gidList" form:"gidList"` // 群聊数组
	Message string  `json:"message" form:"message"` // CQ码格式的消息
}

// LoginParams 登录参数
//
//	@Description	登录参数
type LoginParams struct {
	Username string `json:"username" form:"username"` // 用户名
	Password string `json:"password" form:"password"` // 密码
}

// LoginResultVo 登录返回参数
//
//	@Description	登录返回参数
type LoginResultVo struct {
	UserID   int        `json:"userId"`   // 用户id
	Username string     `json:"username"` // 用户名
	RealName string     `json:"realName"` // 实际名
	Desc     string     `json:"desc"`     // 描述
	Token    string     `json:"token"`    // token
	Roles    []RoleInfo `json:"roles"`    // 角色
}

// RoleInfo 角色参数
//
//	@Description	角色参数
type RoleInfo struct {
	RoleName string `json:"roleName"` // 角色名
	Value    string `json:"value"`    // 角色值
}

// UserInfoVo 用户信息
//
//	@Description	用户信息
type UserInfoVo struct {
	UserID   int        `json:"userId"`   // 用户id
	Username string     `json:"username"` // 用户名
	RealName string     `json:"realName"` // 实际名
	Desc     string     `json:"desc"`     // 描述
	Token    string     `json:"token"`    // token
	Roles    []RoleInfo `json:"roles"`    // 角色
	Avatar   string     `json:"avatar"`   // 头像
	HomePath string     `json:"homePath"` // 主页路径
	Password string     `json:"password"` // 密码
}

// MessageInfo 消息信息
//
//	@Description	消息信息
type MessageInfo struct {
	MessageType string      `json:"message_type"` // 消息类型, group为群聊,private为私聊
	MessageID   interface{} `json:"message_id"`   // 消息id
	GroupID     int64       `json:"group_id"`     // 群id
	GroupName   string      `json:"group_name"`   // 群名
	UserID      int64       `json:"user_id"`      // 用户名
	Nickname    string      `json:"nickname"`     // 昵称
	RawMessage  string      `json:"raw_message"`  // 初始消息
}

// PluginVo 全部插件的返回
//
//	@Description	全部插件的返回
type PluginVo struct {
	ID             int    `json:"id"`             // 插件序号
	Name           string `json:"name"`           // 插件名
	Brief          string `json:"brief"`          // 简述
	Usage          string `json:"usage"`          // 用法
	Banner         string `json:"banner"`         // 头像
	PluginStatus   bool   `json:"pluginStatus"`   // 插件状态,false=禁用,true=启用
	ResponseStatus bool   `json:"responseStatus"` // 响应状态, false=沉默,true=响应
}

// RequestVo 请求返回
//
//	@Description	请求返回
type RequestVo struct {
	Flag        string `json:"flag"`        // 请求flag
	RequestType string `json:"requestType"` // 请求类型
	SubType     string `json:"subType"`     // 请求子类型
	Comment     string `json:"comment"`     // 注释
	GroupID     int64  `json:"groupId"`     // 群id
	GroupName   string `json:"groupName"`   // 群名
	UserID      int64  `json:"userId"`      // 用户id
	Nickname    string `json:"nickname"`    // 昵称
	SelfID      int64  `json:"selfId"`      // 机器人qq
}
