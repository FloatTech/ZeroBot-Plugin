package ygotrade

import (
	"encoding/json"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	api     = "https://api.jihuanshe.com/api/market/search/match-product?game_key=ygo&game_sub_key=ocg&page=1&keyword="
	method  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"
	referer = "https://api.jihuanshe.com/"
)

type tradeInfo struct {
	Total       int         `json:"total"`
	PerPage     int         `json:"per_page"`
	CurrentPage int         `json:"current_page"`
	LastPage    int         `json:"last_page"`
	NextPageURL string      `json:"next_page_url"`
	PrevPageURL interface{} `json:"prev_page_url"`
	From        int         `json:"from"`
	To          int         `json:"to"`
	Data        []struct {
		Type       string      `json:"type"`
		GameKey    string      `json:"game_key"`
		GameSubKey string      `json:"game_sub_key"`
		ID         int         `json:"id"`
		NameCn     string      `json:"name_cn"`
		NameOrigin string      `json:"name_origin"`
		CardID     int         `json:"card_id"`
		Number     string      `json:"number"`
		Rarity     string      `json:"rarity"`
		ImageURL   string      `json:"image_url"`
		MinPrice   string      `json:"min_price"`
		Grade      interface{} `json:"grade"`
	} `json:"data"`
}

var ()

func init() {
	engine := control.Register("ygotrade", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "集换社游戏王的卡价查询",
		Help:             "- 查卡价 [卡名]\n- 查卡价 [卡名] [稀有度]",
	})
	engine.OnPrefix("查卡价").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		//args := strings.Split(model.Args, " ")
		// 请求html页面
		list_body, err := web.RequestDataWith(web.NewDefaultClient(), api+model.Args, "GET", method, referer)
		if err != nil {
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("服务器读取错误：", err)))
			return
		}
		var parsed tradeInfo
		err = json.Unmarshal(list_body, &parsed)
		if err != nil {
			ctx.SendChain(message.Text("无法解析歌单ID内容,", err))
			return
		}
		switch len(parsed.Data) {
		case 0:
			ctx.SendChain(message.Text("没有找到此关键字的卡片"))
		case 1:
			ctx.SendChain(
				message.Text(
					"卡名:", parsed.Data[0].NameCn,
					"卡密:", parsed.Data[0].ID,
					"卡序:", parsed.Data[0].Number,
					"罕贵度:", parsed.Data[0].Rarity,
					"当前最低价:", parsed.Data[0].MinPrice,
				),
				message.Image(parsed.Data[0].ImageURL),
			)
		default:
			msg := make(message.Message, len(parsed.Data))
			for i := 0; i < len(parsed.Data); i++ {
				msg[i] = ctxext.FakeSenderForwardNode(ctx, message.Text(
					"卡名:", parsed.Data[i].NameCn,
					"卡密:", parsed.Data[i].ID,
					"卡序:", parsed.Data[i].Number,
					"罕贵度:", parsed.Data[i].Rarity,
					"当前最低价:", parsed.Data[i].MinPrice),
					message.Image(parsed.Data[i].ImageURL))
			}
			if id := ctx.Send(msg); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		}
	})
}
