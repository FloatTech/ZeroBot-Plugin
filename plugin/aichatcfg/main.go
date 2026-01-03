// Package aichatcfg aichat 的配置, 优先级要比 aichat 高
package aichatcfg

import (
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

var (
	// en data [8 temp] [8 rate] LSB
	en = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Extra:            control.ExtraFromString("aichat"),
		Brief:            "aichat 的配置",
		Help: "- 设置AI聊天触发概率10\n" +
			"- 设置AI聊天温度80\n" +
			"- 设置AI聊天(|识图|Agent)接口类型[OpenAI|OLLaMA|GenAI]\n" +
			"- 设置AI聊天(不)使用Agent模式\n" +
			"- 设置AI聊天(不)支持系统提示词\n" +
			"- 设置AI聊天(|识图|Agent)接口地址https://api.siliconflow.cn/v1/chat/completions\n" +
			"- 设置AI聊天(|识图|Agent)密钥xxx\n" +
			"- 设置AI聊天(|识图|Agent)模型名Qwen/Qwen3-8B\n" +
			"- 查看AI聊天系统提示词\n" +
			"- 重置AI聊天系统提示词\n" +
			"- 设置AI聊天系统提示词xxx\n" +
			"- 设置AI聊天分隔符</think>(留空则清除)\n" +
			"- 设置AI聊天(不)响应AT\n" +
			"- 设置AI聊天最大长度4096\n" +
			"- 设置AI聊天TopP 0.9\n" +
			"- 设置AI聊天(不)以AI语音输出\n" +
			"- 查看AI聊天配置\n" +
			"- 重置AI聊天\n",
	})
)

func init() {
	en.UsePreHandler(func(ctx *zero.Ctx) bool {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		stor, err := chat.NewStorage(ctx, gid)
		if err != nil {
			logrus.Warnln("ERROR: ", err)
			return false
		}
		ctx.State[zero.StateKeyPrefixKeep+"aichatcfg_stor__"] = stor
		return true
	})
	en.OnPrefix("设置AI聊天触发概率", zero.AdminPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBitmapHandler(chat.BitmapRate, 0, 100))
	en.OnPrefix("设置AI聊天温度", zero.AdminPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBitmapHandler(chat.BitmapTemp, 0, 100))
	en.OnPrefix("设置AI聊天接口类型", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetModelType(&chat.AC.Type))
	en.OnPrefix("设置AI聊天识图接口类型", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetModelType(&chat.AC.ImageType))
	en.OnPrefix("设置AI聊天Agent接口类型", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetModelType(&chat.AC.AgentType))
	en.OnPrefix("设置AI聊天接口地址", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.API))
	en.OnPrefix("设置AI聊天识图接口地址", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.ImageAPI))
	en.OnPrefix("设置AI聊天Agent接口地址", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.AgentAPI))
	en.OnPrefix("设置AI聊天密钥", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.Key))
	en.OnPrefix("设置AI聊天识图密钥", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.ImageKey))
	en.OnPrefix("设置AI聊天Agent密钥", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.AgentKey))
	en.OnPrefix("设置AI聊天模型名", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.ModelName))
	en.OnPrefix("设置AI聊天识图模型名", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.ImageModelName))
	en.OnPrefix("设置AI聊天Agent模型名", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.AgentModelName))
	en.OnPrefix("设置AI聊天系统提示词", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.SystemP))
	en.OnFullMatch("查看AI聊天系统提示词", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text(chat.AC.SystemP))
	})
	en.OnFullMatch("重置AI聊天系统提示词", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: no such plugin"))
			return
		}
		chat.AC.SystemP = chat.SystemPrompt
		err := c.SetExtra(&chat.AC)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
	en.OnPrefix("设置AI聊天分隔符", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetStr(&chat.AC.Separator))
	en.OnRegex("^设置AI聊天(不)?响应AT$", chat.EnsureConfig, zero.SuperUserPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBoolHandler(chat.BitmapNrat))
	en.OnRegex("^设置AI聊天(不)?支持系统提示词$", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetBool(&chat.AC.NoSystemP))
	en.OnRegex("^设置AI聊天(不)?使用Agent模式$", chat.EnsureConfig, zero.SuperUserPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBoolHandler(chat.BitmapNagt))
	en.OnPrefix("设置AI聊天最大长度", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetUint(&chat.AC.MaxN))
	en.OnPrefix("设置AI聊天TopP", chat.EnsureConfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(chat.NewExtraSetFloat32(&chat.AC.TopP))
	en.OnRegex("^设置AI聊天(不)?以AI语音输出$", chat.EnsureConfig, zero.AdminPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBoolHandler(chat.BitmapNrec))
	en.OnFullMatch("查看AI聊天配置", chat.EnsureConfig, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			stor, err := chat.NewStorage(ctx, gid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(
				message.Text(
					"【当前AI聊天本群配置】\n",
					"• 触发概率：", int(stor.Rate()), "\n",
					"• 温度：", stor.Temp(), "\n",
					"• 以AI语音输出：", chat.ModelBool(!stor.NoRecord()), "\n",
					"• 使用Agent：", chat.ModelBool(!stor.NoAgent()), "\n",
					"• 响应@：", chat.ModelBool(!stor.NoReplyAt()), "\n",
				),
				message.Text("【当前AI聊天全局配置】\n", &chat.AC),
			)
		})
	en.OnFullMatch("重置AI聊天", chat.EnsureConfig, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		chat.ResetChat()
		ctx.SendChain(message.Text("成功"))
	})
}
