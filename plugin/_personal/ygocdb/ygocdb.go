// Package ygocdb 本插件查卡通过网页"https://www.ygo-sem.cn/"获取的
package ygocdb

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	api     = "https://ygocdb.com/api/v0/?search="
	picherf = "https://cdn.233.momobako.com/ygopro/pics/"
)

type searchResult struct {
	Result []struct {
		Cid    int    `json:"cid"`
		ID     int    `json:"id"`
		CnName string `json:"cn_name"`
		CnocgN string `json:"cnocg_n"`
		JpRuby string `json:"jp_ruby"`
		JpName string `json:"jp_name"`
		EnName string `json:"en_name"`
		Text   struct {
			Types string `json:"types"`
			Pdesc string `json:"pdesc"`
			Desc  string `json:"desc"`
		} `json:"text"`
		Data struct {
			Ot        int `json:"ot"`
			Setcode   int `json:"setcode"`
			Type      int `json:"type"`
			Atk       int `json:"atk"`
			Def       int `json:"def"`
			Level     int `json:"level"`
			Race      int `json:"race"`
			Attribute int `json:"attribute"`
		} `json:"data"`
		Weight int           `json:"weight"`
		Faqs   []interface{} `json:"faqs"`
		Artid  int           `json:"artid"`
	} `json:"result"`
	Next int `json:"next"`
}

func init() {
	en := control.Register("ygocdb", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:"游戏王百鸽卡查API卡查",
		Help:
			"- /ydp [xxx]\n" +
			"- /yds [xxx]\n" +
			"- /ydb [xxx]\n" +
			"[xxx]为搜索内容\np:返回一张图片\ns:返回一张效果描述\nb:高级搜索",
	})

	en.OnRegex(`^[/,.\\!。，！]ydp\s?(.*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctxtext := ctx.State["regex_matched"].([]string)[1]
		if ctxtext == "" {
			ctx.SendChain(message.Text("你是想查询「空手假象」吗？"))
			return
		}
		data, err := web.GetData(api + url.QueryEscape(ctxtext))
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		var result searchResult
		err = json.Unmarshal(data, &result)
		if err != nil {
			ctx.SendChain(message.Text("json ERROR:", err))
			return
		}
		switch len(result.Result) {
		case 0:
			ctx.SendChain(message.Text("没有找到相关的卡片额"))
			return
		default:
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image(picherf+strconv.Itoa(result.Result[0].ID)+".jpg")))
		}
	})

	en.OnRegex(`^[/,.\\!。，！]yds\s?(.*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctxtext := ctx.State["regex_matched"].([]string)[1]
		if ctxtext == "" {
			ctx.SendChain(message.Text("你是想查询「空手假象」吗？"))
			return
		}
		data, err := web.GetData(api + url.QueryEscape(ctxtext))
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		var result searchResult
		err = json.Unmarshal(data, &result)
		if err != nil {
			ctx.SendChain(message.Text("json ERROR:", err))
			return
		}
		switch len(result.Result) {
		case 0:
			ctx.SendChain(message.Text("没有找到相关的卡片额"))
			return
		default:
			cardtextout := cardtext(result, 0)
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(cardtextout)))
		}
	})

	en.OnRegex(`^[/,.\\!。，！]ydb\s?(.*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctxtext := ctx.State["regex_matched"].([]string)[1]
		if ctxtext == "" {
			ctx.SendChain(message.Text("你是想查询「空手假象」吗？"))
			return
		}
		data, err := web.GetData(api + url.QueryEscape(ctxtext))
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		var result searchResult
		err = json.Unmarshal(data, &result)
		if err != nil {
			ctx.SendChain(message.Text("json ERROR:", err))
			return
		}
		maxpage := len(result.Result)
		switch maxpage {
		case 0:
			ctx.SendChain(message.Text("没有找到相关的卡片额"))
			return
		case 1:
			cardtextout := cardtext(result, 0)
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image(picherf+strconv.Itoa(result.Result[0].ID)+".jpg"), message.Text(cardtextout)))
			return
		default:
			break
		}
		var listName []string
		var listid []int
		for _, v := range result.Result {
			listName = append(listName, strconv.Itoa(len(listName))+"."+v.CnName)
			listid = append(listid, v.ID)
		}
		var (
			currentPage = 10
			nextpage    = 0
		)
		if maxpage < 10 {
			currentPage = maxpage
		}
		ctx.SendChain(message.Text("找到", strconv.Itoa(maxpage), "张相关卡片,当前显示以下卡名：\n",
			strings.Join(listName[:currentPage], "\n"),
			"\n————————————————\n输入对应数字获取卡片信息，",
			"\n或回复“取消”、“下一页”指令"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(取消)|(下一页)|\d+`), zero.OnlyGroup, zero.CheckUser(ctx.Event.UserID)).Repeat()
		after := time.NewTimer(20 * time.Second)
		for {
			select {
			case <-after.C:
				cancel()
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("等待超时,搜索结束"),
					),
				)
				return
			case e := <-recv:
				nextcmd := e.Event.Message.String()
				switch nextcmd {
				case "取消":
					cancel()
					after.Stop()
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("用户取消,搜索结束"),
						),
					)
					return
				case "下一页":
					after.Reset(20 * time.Second)
					if maxpage < 11 {
						continue
					}
					nextpage += 1
					if nextpage*10 >= maxpage {
						nextpage = 0
						currentPage = 10
						ctx.SendChain(message.Text("已是最后一页，返回到第一页"))
					} else if nextpage == maxpage/10 {
						currentPage = maxpage % 10
					}
					ctx.SendChain(message.Text("找到", strconv.Itoa(maxpage), "张相关卡片,当前显示以下卡名：\n",
						strings.Join(listName[nextpage*10:nextpage*10+currentPage], "\n"),
						"\n————————————————\n输入对应数字获取卡片信息,",
						"\n或回复“取消”、“下一页”指令"))
				default:
					cardint, err := strconv.Atoi(nextcmd)
					switch {
					case err != nil:
						after.Reset(20 * time.Second)
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
					default:
						if cardint < nextpage*10+currentPage {
							cancel()
							after.Stop()
							cardtextout := cardtext(result, cardint)
							ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image(picherf+strconv.Itoa(listid[cardint])+".jpg"), message.Text(cardtextout)))
							return
						} else {
							after.Reset(20 * time.Second)
							ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
						}
					}
				}
			}
		}
	})
}

func cardtext(List searchResult, cardid int) (textout string) {
	var cardtext []string
	cardtext = append(cardtext, "中文卡名：\n    "+List.Result[cardid].CnName)
	if List.Result[cardid].JpName == "" {
		cardtext = append(cardtext, "英文卡名：\n    "+List.Result[cardid].EnName)
	} else {
		cardtext = append(cardtext, "日文卡名：\n    "+List.Result[cardid].JpName)
	}
	cardtext = append(cardtext, "卡片密码："+strconv.Itoa(List.Result[cardid].ID))
	cardtext = append(cardtext, List.Result[cardid].Text.Types)
	cardtext = append(cardtext, List.Result[cardid].Text.Desc)
	textout = strings.Join(cardtext, "\n")
	return textout
}
