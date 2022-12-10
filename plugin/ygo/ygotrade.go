// Package ygo 一些关于ygo的插件
package ygo

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	rarityTrade = "https://api.jihuanshe.com/api/market/search/match-product?game_key=ygo&game_sub_key=ocg&page=1&keyword="
	storeTrade  = "https://api.jihuanshe.com/api/market/card-versions/products?game_key=ygo&game_sub_key=ocg&page=1&condition=1&card_version_id="
)

type apiInfo struct {
	Data []tradeInfo `json:"data"`
}

type tradeInfo struct {
	// 卡片信息
	ID       int    `json:"id"`
	NameCn   string `json:"name_cn"`
	CardID   int    `json:"card_id"`
	Number   string `json:"number"`
	Rarity   string `json:"rarity"`
	ImageURL string `json:"image_url"`
	MinPrice string `json:"min_price"`
	// 卡店信息
	SellerUserID     int    `json:"seller_user_id"`
	SellerUsername   string `json:"seller_username"`
	SellerUserAvatar string `json:"seller_user_avatar"`
	SellerProvince   string `json:"seller_province"`
	SellerCity       string `json:"seller_city"`
	SellerCreditRank string `json:"seller_credit_rank"`
	Quantity         string `json:"quantity"`
	CardVersionImage string `json:"card_version_image"`
}

func init() {
	engine := control.Register("ygotrade", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "游戏王卡价查询", // 本插件基于集换社API
		Help:             "- 查卡价 [卡名]\n- 查卡价 [卡名] -r [稀有度 稀有度 ...]\n- 查卡店  [卡名]\n- 查卡店  [卡名] -r [稀有度]",
	}).ApplySingle(ctxext.DefaultSingle)
	engine.OnPrefix("查卡价", func(ctx *zero.Ctx) bool {
		ctx.State["args"] = strings.TrimSpace(ctx.State["args"].(string))
		return ctx.State["args"].(string) != ""
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cardName, rarity, _ := strings.Cut(ctx.State["args"].(string), " -r ")
		listOfTrace, err := getRarityTrade(cardName, rarity)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		msg := make(message.Message, len(listOfTrace))
		for i := 0; i < len(listOfTrace); i++ {
			msg[i] = ctxext.FakeSenderForwardNode(ctx, message.Text(
				"卡名:", listOfTrace[i].NameCn,
				"\nID:", listOfTrace[i].ID,
				"\n卡序:", listOfTrace[i].Number,
				"\n罕贵度:", listOfTrace[i].Rarity,
				"\n当前最低价:", listOfTrace[i].MinPrice),
				message.Image(listOfTrace[i].ImageURL))
		}
		if id := ctx.Send(msg); id.ID() == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}
	})
	engine.OnPrefix("查卡店", func(ctx *zero.Ctx) bool {
		ctx.State["args"] = strings.TrimSpace(ctx.State["args"].(string))
		return ctx.State["args"].(string) != ""
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cardName, rarity, _ := strings.Cut(ctx.State["args"].(string), " -r ")
		if strings.Count(rarity, " ") > 0 {
			ctx.SendChain(message.Text("ERROR: ", "卡店查询不支持查找多个罕贵度"))
			return
		}
		listOfTrace, err := getRarityTrade(cardName, rarity)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		listStroe, err := getStoreTrade(listOfTrace[0].ID)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		msg := make(message.Message, len(listStroe))
		for i := 0; i < len(listStroe); i++ {
			msg[i] = ctxext.FakeSenderForwardNode(ctx, message.Text(
				"卖家名:", listStroe[i].SellerUsername,
				"\nID:", listStroe[i].SellerUserID,
				"\n地区:", listStroe[i].SellerCity,
				"\n信誉度:", listStroe[i].SellerCreditRank,
				"\n数量:", listStroe[i].Quantity,
				"\n当前最低价:", listStroe[i].MinPrice),
				message.Image(listStroe[i].CardVersionImage))
		}
		if id := ctx.Send(msg); id.ID() == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}
	})
}

// 获取卡名该罕贵度卡片数据
func getRarityTrade(key, rarity string) (tradeInfo []tradeInfo, err error) {
	listOfTrace, err := web.GetData(rarityTrade + url.QueryEscape(key) + "&rarity=" + url.QueryEscape(rarity))
	if err != nil {
		return
	}
	var apiInfo apiInfo
	err = json.Unmarshal(listOfTrace, &apiInfo)
	if len(apiInfo.Data) == 0 {
		err = errors.New("没有找到相关卡片或输入参数错误")
		return
	}
	tradeInfo = apiInfo.Data
	return
}

// 获取卡店卡片数据
func getStoreTrade(cardID int) (stroeInfo []tradeInfo, err error) {
	listOfTrace, err := web.GetData(storeTrade + url.QueryEscape(strconv.Itoa(cardID)))
	if err != nil {
		return
	}
	var apiInfo apiInfo
	err = json.Unmarshal(listOfTrace, &apiInfo)
	if len(apiInfo.Data) == 0 {
		err = errors.New("没有找到相关卡片或输入参数错误")
		return
	}
	stroeInfo = apiInfo.Data
	return
}
