// Package airecord 群应用：AI声聊
package airecord

import (
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/internal/qqvoice"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/airecord"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Extra:            control.ExtraFromString("airecord"),
		Brief:            "群应用：AI声聊",
		Help: "- 设置AI语音群号1048452984(tips：机器人任意所在群聊即可)\n" +
			"- 设置AI语音模型\n" +
			"- 查看AI语音配置\n" +
			"- 发送AI语音xxx",
		PrivateDataFolder: "airecord",
	})

	en.OnPrefix("设置AI语音群号", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			u := strings.TrimSpace(ctx.State["args"].(string))
			num, err := strconv.ParseInt(u, 10, 64)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: parse gid err: ", err))
				return
			}
			if num <= 0 {
				ctx.SendChain(message.Text("请填写正数群号。"))
				return
			}
			err = airecord.SetCustomGID(num)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: set gid err: ", err))
				return
			}
			ctx.SendChain(message.Text("设置AI语音群号为", num))
		})
	en.OnFullMatch("设置AI语音模型", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession())
			recv, cancel := next.Repeat()
			defer cancel()
			jsonData := ctx.GetAICharacters(0, 1)

			models := availableVoiceModels(jsonData)
			if len(models) == 0 {
				ctx.SendChain(message.Text("暂时无法获取AI语音声音，请稍后重试。"))
				return
			}
			var builder strings.Builder
			// 写入开头文本
			builder.WriteString("请选择语音模型序号：\n")

			// 遍历names数组，拼接序号和名称
			for i, v := range models {
				// 将数字转换为字符串（不依赖fmt）
				numStr := strconv.Itoa(i)
				// 拼接格式："序号. 名称\n"
				builder.WriteString(numStr)
				builder.WriteString(". ")
				builder.WriteString(v.name)
				builder.WriteString("\n")
			}
			// 获取最终字符串
			ctx.SendChain(message.Text(builder.String()))
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.SendChain(message.Text("设置AI语音模型指令过期"))
					return
				case ct := <-recv:
					msg := ct.Event.Message.ExtractPlainText()
					num, err := strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Text("请输入数字!"))
						continue
					}
					if num < 0 || num >= len(models) {
						ctx.SendChain(message.Text("序号非法!"))
						continue
					}
					err = airecord.SetRecordModel(models[num].name, models[num].id)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: set model err: ", err))
						continue
					}
					ctx.SendChain(message.Text("已选择语音模型: ", models[num].name))
					if models[num].preview != "" {
						ctx.SendChain(message.Record(models[num].preview))
					}
					return
				}
			}
		})
	en.OnFullMatch("查看AI语音配置", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(airecord.PrintRecordConfig()))
		})
	en.OnPrefix("发送AI语音", zero.UserOrGrpAdmin).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			u := strings.TrimSpace(ctx.State["args"].(string))
			recCfg := airecord.GetConfig()
			failure := qqvoice.Deliver(u, recCfg.ModelID, ctx.Event.GroupID, ctx.CallAction)
			if failure != "" {
				ctx.SendChain(message.Text(failure))
			}
		})
}
