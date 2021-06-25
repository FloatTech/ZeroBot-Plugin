package manager

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	zero.OnFullMatchGroup([]string{"/help", "/帮助", "/指令菜单"}).SetBlock(true).SetPriority(999).
			Handle(func(ctx *zero.Ctx) {
				ctx.SendChain(message.Text(
					"* 可交互指令菜单，被动技能不在以下 ", "\n",
					"* 说怪话：@我说话有几率触发哦，怪话都是跟你们学的捏", "\n",
					"* 发大病：随机触发", "\n",
					"* /info name： 查询粉丝实时数据（例：/info 嘉然）", "\n",
					"* /粉丝 name： 查询粉丝实时数据（例：/粉丝 嘉然）", "\n",
					"* /查 uid： 通过uid查询成分（与关注了2000uvp小号的共同关注）", "\n",
					"* /查 name： 通过名字查询成分（与关注了2000uvp小号的共同关注）", "\n",
          "* 来张涩图、风景、二次、车万： 随机从P站获取一张对应类型的图片并发送", "\n",
          "* 来张萝莉、来张萝莉r18：对应lolicon接口的涩图", "\n",
					"* /空调开or关：开关群里的空调", "\n",
					"* /小作文：从剪贴板随机发一篇小作文", "\n",
					"* /网抑云：随机获取一句网抑云伤痛文学", "\n",
					"* /网易点歌 歌名：分享歌名的歌曲到群", "\n",
					"* /戒断x分钟（x=int）：获得x分钟禁言", "\n",
					"* /全员戒断：全体禁言（仅管理员有权限）", "\n",
					"* /全员开溜：解除全体禁言（仅管理员有权限）", "\n",
					"* 有什么其他想要的功能可以私聊和我说", "\n",
					))
	})
}
