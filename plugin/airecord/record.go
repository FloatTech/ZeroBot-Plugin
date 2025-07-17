// Package airecord 群应用：AI声聊
package airecord

import (
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

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

			// 转换为字符串数组
			var names []string
			// 初始化两个映射表
			nameToID := make(map[string]string)
			nameToURL := make(map[string]string)
			characters := jsonData.Get("#.characters")

			// 遍历每个角色对象
			characters.ForEach(func(_, group gjson.Result) bool {
				group.ForEach(func(_, character gjson.Result) bool {
					// 提取当前角色的三个字段
					name := character.Get("character_name").String()
					names = append(names, name)
					// 存入映射表（重复名称会覆盖，保留最后出现的条目）
					nameToID[name] = character.Get("character_id").String()
					nameToURL[name] = character.Get("preview_url").String()
					return true // 继续遍历
				})
				return true // 继续遍历
			})
			var builder strings.Builder
			// 写入开头文本
			builder.WriteString("请选择语音模型序号：\n")

			// 遍历names数组，拼接序号和名称
			for i, v := range names {
				// 将数字转换为字符串（不依赖fmt）
				numStr := strconv.Itoa(i)
				// 拼接格式："序号. 名称\n"
				builder.WriteString(numStr)
				builder.WriteString(". ")
				builder.WriteString(v)
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
					if num < 0 || num >= len(names) {
						ctx.SendChain(message.Text("序号非法!"))
						continue
					}
					err = airecord.SetRecordModel(names[num], nameToID[names[num]])
					if err != nil {
						ctx.SendChain(message.Text("ERROR: set model err: ", err))
						continue
					}
					ctx.SendChain(message.Text("已选择语音模型: ", names[num]))
					ctx.SendChain(message.Record(nameToURL[names[num]]))
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
			record := ctx.GetAIRecord(recCfg.ModelID, recCfg.Customgid, u)
			if record == "" {
				id := ctx.SendGroupAIRecord(recCfg.ModelID, ctx.Event.GroupID, u)
				if id == "" {
					ctx.SendChain(message.Text("ERROR: get record err: empty record"))
					return
				}
			}
			ctx.SendChain(message.Record(record))
		})
}
