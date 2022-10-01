// Package yiyan 每日一言
package yiyan

import (
	"encoding/json"

	ctrl "github.com/FloatTech/zbpctrl"

	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const url = "https://v1.hitokoto.cn/?c=a&c=b&c=c&c=d&c=h&c=i" // 动漫 漫画 游戏 文学 影视 诗词

type RspData struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	Hitokoto   string `json:"hitokoto"`
	Type       string `json:"type"`
	From       string `json:"from"`
	FromWho    string `json:"from_who"`
	Creator    string `json:"creator"`
	CreatorUid int    `json:"creator_uid"`
	Reviewer   int    `json:"reviewer"`
	CommitFrom string `json:"commit_from"`
	CreatedAt  string `json:"created_at"`
	Length     int    `json:"length"`
}

func init() {
	engine := control.Register("yiyan", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "- 每日一言(建议配合job插件净化群友心灵)\n",
	})
	engine.OnFullMatch("每日一言").SetBlock(true).Handle(handle)
}
func handle(ctx *zero.Ctx) {
	var rsp RspData
	data, err := web.GetData(url)
	if err != nil {
		ctx.SendChain(message.Text("Err:", err))
	}
	json.Unmarshal(data, &rsp)
	msg := ""
	msg += rsp.Hitokoto + "\n出自：" + rsp.From + "\n"
	if len(rsp.FromWho) != 0 {
		msg += "作者：" + rsp.FromWho
	}
	ctx.SendChain(message.Text(msg))
}
