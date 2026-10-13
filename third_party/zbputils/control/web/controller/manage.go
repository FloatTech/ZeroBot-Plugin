// Package controller 主要处理逻辑
package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/RomiChan/syncx"
	"github.com/RomiChan/websocket"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/binary"
	ctrl "github.com/FloatTech/zbpctrl"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/web/middleware"
	"github.com/FloatTech/zbputils/control/web/model"
	"github.com/FloatTech/zbputils/control/web/types"
)

var (
	// MsgConn 向前端推送消息的ws链接
	MsgConn *websocket.Conn
	// LogConn 向前端推送日志的ws链接
	LogConn *websocket.Conn
	// 实现Write接口
	l logWriter
	// 存储请求事件，flag作为键，一个request对象作为值
	requestData syncx.Map[string, *zero.Event]
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}
	validate = validator.New()
)

func init() {
	// 日志设置
	writer := io.MultiWriter(l, os.Stdout)
	log.SetOutput(writer)

	zero.OnMessage().SetBlock(false).FirstPriority().Handle(func(ctx *zero.Ctx) {
		if MsgConn != nil {
			mi := types.MessageInfo{
				MessageType: ctx.Event.MessageType,
				MessageID:   ctx.Event.MessageID,
				GroupID:     ctx.Event.GroupID,
				GroupName:   ctx.GetGroupInfo(ctx.Event.GroupID, false).Name,
				UserID:      ctx.Event.UserID,
				Nickname:    ctx.GetStrangerInfo(ctx.Event.UserID, false).Get("nickname").String(),
				RawMessage:  ctx.Event.RawMessage,
			}
			err := MsgConn.WriteJSON(mi)
			if err != nil {
				log.Errorln("[gui] 连接前端ws失败:", err)
				return
			}
		}
	})
	// 直接注册一个request请求监听器，优先级设置为最高，设置不阻断事件传播
	zero.OnRequest().SetBlock(false).FirstPriority().Handle(func(ctx *zero.Ctx) {
		requestData.Store(ctx.Event.Flag, ctx.Event)
	})
}

// logWriter
type logWriter struct{}

// GetBotList 获取机器人qq号
//
//	@Tags			通用
//	@Summary		获取机器人qq号
//	@Description	获取机器人qq号
//	@Router			/api/getBotList [get]
//	@Success		200	{object}	types.Response	"成功"
func GetBotList(context *gin.Context) {
	var bots []int64
	zero.RangeBot(func(id int64, _ *zero.Ctx) bool {
		bots = append(bots, id)
		return true
	})
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       bots,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetFriendList 获取好友列表
//
//	@Tags			通用
//	@Summary		获取好友列表
//	@Description	获取好友列表
//	@Router			/api/getFriendList [get]
//	@Param			selfId	query		int				false	"机器人qq"
//	@Success		200		{object}	types.Response	"成功"
func GetFriendList(context *gin.Context) {
	_, bot, err := getBot(context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	var resp []any
	list := bot.GetFriendList().String()
	err = json.Unmarshal(binary.StringToBytes(list), &resp)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       resp,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetGroupMemberList 获取群成员列表
//
//	@Tags			通用
//	@Summary		获取群成员列表
//	@Description	获取群成员列表,groupId为0的时候拉取好友信息
//	@Router			/api/getGroupMemberList [get]
//	@Param			selfId	query		int				false	"机器人qq"
//	@Param			groupId	query		int				false	"群聊id"
//	@Success		200		{object}	types.Response	"成功"
func GetGroupMemberList(context *gin.Context) {
	var (
		d   types.GetGroupMemberListReq
		bot *zero.Ctx
	)
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	if d.SelfID == 0 {
		zero.RangeBot(func(_ int64, ctx *zero.Ctx) bool {
			bot = ctx
			return false
		})
	} else {
		bot = zero.GetBot(d.SelfID)
	}
	var resp []any
	list := bot.GetGroupMemberList(d.GroupID).String()
	_ = json.Unmarshal(binary.StringToBytes(list), &resp)
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       resp,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetGroupList 获取群列表
//
//	@Tags			通用
//	@Summary		获取群列表
//	@Description	获取群列表
//	@Router			/api/getGroupList [get]
//	@Param			selfId	query		int				false	"机器人qq"
//	@Success		200		{object}	types.Response	"成功"
func GetGroupList(context *gin.Context) {
	_, bot, err := getBot(context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	var resp []any
	list := bot.GetGroupList().String()
	err = json.Unmarshal(binary.StringToBytes(list), &resp)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       resp,
		Message:      "",
		ResponseType: "ok",
	})
}

// DeleteGroup 删除群聊或好友
//
//	@Tags			通用
//	@Summary		删除群聊或好友
//	@Description	删除群聊或好友
//	@Router			/api/deleteGroup [get]
//	@Param			object	body		types.DeleteGroupParams	false	"删除群聊或好友入参"
//	@Success		200		{object}	types.Response			"成功"
func DeleteGroup(context *gin.Context) {
	var (
		d   types.DeleteGroupParams
		bot *zero.Ctx
		gid int64
	)
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	if d.SelfID == 0 {
		zero.RangeBot(func(_ int64, ctx *zero.Ctx) bool {
			bot = ctx
			return false
		})
	} else {
		bot = zero.GetBot(d.SelfID)
	}
	if d.GroupID >= 0 {
		gid = d.GroupID
		bot.SetGroupLeave(gid, false)
	} else {
		gid = -d.GroupID
		bot.CallAction("delete_friend", zero.Params{
			"user_id": gid,
		})
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       nil,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetAllPlugin 获取所有插件的状态
//
//	@Tags			插件
//	@Summary		获取所有插件的状态
//	@Description	获取所有插件的状态
//	@Router			/api/manage/getAllPlugin [get]
//	@Param			object	body		types.AllPluginParams					false	"获取所有插件的状态入参"
//	@Success		200		{object}	types.Response{result=[]types.PluginVo}	"成功"
func GetAllPlugin(context *gin.Context) {
	var d types.AllPluginParams
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	var pluginVoList []types.PluginVo
	control.ForEachByPrio(func(i int, manager *ctrl.Control[*zero.Ctx]) bool {
		p := types.PluginVo{
			ID:             i,
			Name:           manager.Service,
			Brief:          manager.Options.Brief,
			Usage:          manager.String(),
			Banner:         "https://gitcode.net/u011570312/zbpbanner/-/raw/main/" + manager.Service + ".png",
			PluginStatus:   manager.IsEnabledIn(d.GroupID),
			ResponseStatus: control.CanResponse(d.GroupID),
		}
		pluginVoList = append(pluginVoList, p)
		return true
	})
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       pluginVoList,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetPlugin 获取某个插件的状态
//
//	@Tags			插件
//	@Summary		获取某个插件的状态
//	@Description	获取某个插件的状态
//	@Router			/api/manage/getPlugin [get]
//	@Param			object	body		types.PluginParams						false	"获取某个插件的状态入参"
//	@Success		200		{object}	types.Response{result=types.PluginVo}	"成功"
func GetPlugin(context *gin.Context) {
	var d types.PluginParams
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	con, b := control.Lookup(d.Name)
	if !b {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      d.Name + "服务不存在",
			ResponseType: "error",
		})
		return
	}
	p := types.PluginVo{
		Name:           con.Service,
		Brief:          con.Options.Brief,
		Usage:          con.String(),
		Banner:         "https://gitcode.net/u011570312/zbpbanner/-/raw/main/" + con.Service + ".png",
		PluginStatus:   con.IsEnabledIn(d.GroupID),
		ResponseStatus: control.CanResponse(d.GroupID),
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       p,
		Message:      "",
		ResponseType: "ok",
	})
}

// UpdatePluginStatus 更改某一个插件状态
//
//	@Tags			插件
//	@Summary		更改某一个插件状态
//	@Description	更改某一个插件状态
//	@Router			/api/manage/updatePluginStatus [post]
//	@Param			object	body		types.PluginStatusParams	false	"修改插件状态入参"
//	@Success		200		{object}	types.Response				"成功"
func UpdatePluginStatus(context *gin.Context) {
	var d types.PluginStatusParams
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	con, b := control.Lookup(d.Name)
	if !b {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      d.Name + "服务不存在",
			ResponseType: "error",
		})
		return
	}
	switch d.Status {
	case 0:
		con.Disable(d.GroupID)
	case 1:
		con.Enable(d.GroupID)
	case 2:
		con.Reset(d.GroupID)
	default:
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       nil,
		Message:      "",
		ResponseType: "ok",
	})
}

// UpdateResponseStatus 更改某一个群响应
//
//	@Tags			响应
//	@Summary		更改某一个群响应
//	@Description	更改某一个群响应
//	@Router			/api/manage/updateResponseStatus [post]
//	@Param			object	body		types.ResponseStatusParams	false	"修改群响应入参"
//	@Success		200		{object}	types.Response				"成功"
func UpdateResponseStatus(context *gin.Context) {
	var d types.ResponseStatusParams
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	if d.Status == 1 {
		err = control.Response(d.GroupID)
	} else {
		err = control.Silence(d.GroupID)
	}
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       nil,
		Message:      "",
		ResponseType: "ok",
	})
}

// UpdateAllPluginStatus 更改某群所有插件状态
//
//	@Tags			响应
//	@Summary		更改某群所有插件状态
//	@Description	更改某群所有插件状态
//	@Router			/api/manage/updateAllPluginStatus [post]
//	@Param			object	body		types.AllPluginStatusParams	false	"修改插件状态入参"
//	@Success		200		{object}	types.Response				"成功"
func UpdateAllPluginStatus(context *gin.Context) {
	var d types.AllPluginStatusParams
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	switch d.Status {
	case 0:
		control.ForEachByPrio(func(_ int, manager *ctrl.Control[*zero.Ctx]) bool {
			manager.Disable(d.GroupID)
			return true
		})
	case 1:
		control.ForEachByPrio(func(_ int, manager *ctrl.Control[*zero.Ctx]) bool {
			manager.Enable(d.GroupID)
			return true
		})
	case 2:
		control.ForEachByPrio(func(_ int, manager *ctrl.Control[*zero.Ctx]) bool {
			manager.Reset(d.GroupID)
			return true
		})
	default:
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       nil,
		Message:      "",
		ResponseType: "ok",
	})
}

// HandleRequest 处理一个请求
//
//	@Tags			通用
//	@Summary		处理一个请求
//	@Description	处理一个请求
//	@Router			/api/handleRequest [post]
//	@Param			object	body		types.HandleRequestParams	false	"处理一个请求入参"
//	@Success		200		{object}	types.Response				"成功"
func HandleRequest(context *gin.Context) {
	var (
		d        types.HandleRequestParams
		typeName string
	)
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	r, ok := requestData.LoadAndDelete(d.Flag)
	if !ok {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      "flag not found",
			ResponseType: "error",
		})
		return
	}
	bot := zero.GetBot(r.SelfID)
	if r.RequestType == "friend" {
		bot.SetFriendAddRequest(r.Flag, d.Approve, "")
		typeName = "好友添加"
	} else {
		bot.SetGroupAddRequest(r.Flag, r.SubType, d.Approve, d.Reason)
		if r.SubType == "add" {
			typeName = "加群请求"
		} else {
			typeName = "群邀请"
		}
	}
	log.Debugln("[gui] 已处理", r.UserID, "的"+typeName)
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       nil,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetRequestList 获取所有的请求
//
//	@Tags			通用
//	@Summary		获取所有的请求
//	@Description	获取所有的请求
//	@Router			/api/getRequestList [get]
//	@Param			object	body		types.BotParams								false	"获取所有的请求入参"
//	@Success		200		{object}	types.Response{result=[]types.RequestVo}	"成功"
func GetRequestList(context *gin.Context) {
	d, bot, err := getBot(context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	var data []types.RequestVo
	requestData.Range(func(_ string, value *zero.Event) bool {
		if d.SelfID != 0 && value.SelfID != d.SelfID {
			return true
		}
		data = append(data, types.RequestVo{
			Flag:        value.Flag,
			RequestType: value.RequestType,
			SubType:     value.SubType,
			Comment:     value.Comment,
			GroupID:     value.GroupID,
			GroupName:   bot.GetGroupInfo(value.GroupID, false).Name,
			UserID:      value.UserID,
			Nickname:    bot.GetStrangerInfo(value.UserID, false).Get("nickname").String(),
			SelfID:      value.SelfID,
		})
		return true
	})
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       data,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetLog 连接日志
func GetLog(context *gin.Context) {
	conn, err := upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		return
	}
	LogConn = conn
}

// Upgrade 连接ws，向前端推送message
func Upgrade(context *gin.Context) {
	conn, err := upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		return
	}
	MsgConn = conn
}

// SendMsg 前端调用发送信息
//
//	@Tags			通用
//	@Summary		前端调用发送信息
//	@Description	前端调用发送信息
//	@Router			/api/sendMsg [post]
//	@Param			object	body		types.SendMsgParams			false	"发消息参数"
//	@Success		200		{object}	types.Response{result=int}	"成功"
func SendMsg(context *gin.Context) {
	var (
		d     types.SendMsgParams
		msgID int64
	)
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	bot := zero.GetBot(d.SelfID)
	all := false
	for _, gid := range d.GIDList {
		if gid == 0 {
			all = true
			break
		}
	}
	// 避免机器人之间发消息
	botMap := make(map[int64]struct{}, 4)
	zero.RangeBot(func(id int64, _ *zero.Ctx) bool {
		botMap[id] = struct{}{}
		return true
	})
	// 检查是否有全部群聊
	if all {
		// 添加私聊
		r := make([]int64, 0, 16)
		for _, v := range d.GIDList {
			if v < 0 {
				r = append(r, v)
			}
		}
		d.GIDList = r
		// 添加全部群聊
		for _, v := range bot.GetGroupList().Array() {
			d.GIDList = append(d.GIDList, v.Get("group_id").Int())
		}
	}
	for _, gid := range d.GIDList {
		if gid > 0 {
			msgID = bot.SendGroupMessage(gid, message.UnescapeCQCodeText(d.Message))
		} else if gid < 0 {
			_, ok := botMap[-gid]
			if !ok {
				msgID = bot.SendPrivateMessage(-gid, message.UnescapeCQCodeText(d.Message))
			}
		}
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       msgID,
		Message:      "",
		ResponseType: "ok",
	})
}

// Write 写入日志
func (l logWriter) Write(p []byte) (n int, err error) {
	if LogConn != nil {
		err := LogConn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			return len(p), nil
		}
	}
	return len(p), nil
}

// Login 前端登录
//
//	@Tags			用户
//	@Summary		前端登录
//	@Description	前端登录
//	@Router			/api/login [post]
//	@Param			object	body		types.LoginParams							false	"前端登录"
//	@Success		200		{object}	types.Response{result=types.LoginResultVo}	"成功"
func Login(context *gin.Context) {
	var d types.LoginParams
	err := bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	user, err := model.FindUser(d.Username, d.Password)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	token := uuid.NewString()
	middleware.LoginCache.Set(token, &user, cache.DefaultExpiration)
	r := types.LoginResultVo{
		Desc:     "manager",
		RealName: user.Username,
		Roles: []types.RoleInfo{
			{
				RoleName: "Super Admin",
				Value:    "super",
			},
		},
		Token:    token,
		UserID:   int(user.ID),
		Username: user.Username,
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       r,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetUserInfo 获得用户信息
//
//	@Tags			用户
//	@Summary		获得用户信息
//	@Description	获得用户信息
//	@Router			/api/getUserInfo [get]
//	@Success		200	{object}	types.Response{result=types.UserInfoVo}	"成功"
func GetUserInfo(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")
	i, _ := middleware.LoginCache.Get(token)
	user := i.(*model.User)
	var qq int64
	if len(zero.BotConfig.SuperUsers) != 0 {
		qq = zero.BotConfig.SuperUsers[0]
	}
	r := types.UserInfoVo{
		Desc:     "manager",
		RealName: user.Username,
		Roles: []types.RoleInfo{
			{
				RoleName: "Super Admin",
				Value:    "super",
			},
		},
		UserID:   int(user.ID),
		Username: user.Username,
		Avatar:   "https://q1.qlogo.cn/g?b=qq&nk=" + strconv.FormatInt(qq, 10) + "&s=640",
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       r,
		Message:      "",
		ResponseType: "ok",
	})
}

// Logout 登出
//
//	@Tags			用户
//	@Summary		登出
//	@Description	登出
//	@Router			/api/logout [get]
//	@Success		200	{object}	types.Response	"成功"
func Logout(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")
	middleware.LoginCache.Delete(token)
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       nil,
		Message:      "",
		ResponseType: "ok",
	})
}

// GetPermCode 授权码
//
//	@Tags			用户
//	@Summary		授权码
//	@Description	授权码
//	@Router			/api/getPermCode [get]
//	@Success		200	{object}	types.Response{result=[]string}	"成功"
func GetPermCode(context *gin.Context) {
	r := []string{"1000", "3000", "5000"}
	// 先写死接口
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       r,
		Message:      "",
		ResponseType: "ok",
	})
}

func getBot(context *gin.Context) (d types.BotParams, bot *zero.Ctx, err error) {
	err = bind(&d, context)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	if d.SelfID == 0 {
		zero.RangeBot(func(_ int64, ctx *zero.Ctx) bool {
			bot = ctx
			return false
		})
	} else {
		bot = zero.GetBot(d.SelfID)
	}
	return
}

// bind 绑定结构体, 并校验
func bind(obj any, ctx *gin.Context) (err error) {
	err = ctx.ShouldBind(obj)
	if err != nil {
		return
	}
	return validate.Struct(obj)
}
