// Package ygosem 基于ygosem的插件功能
package ygosem

// 本插件查卡通过网页"https://www.ygo-sem.cn/"获取的

import (
	"bytes"
	"image"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	// 图片输出
	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/zbputils/img"
)

var (
	reqconf = [...]string{"GET", "https://www.ygo-sem.cn/",
		"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"}
)

func init() {
	en := control.Register("ygosem", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "游戏王进阶平台卡查",
		Help:              "- /ys [卡名] [-(卡图|描述|调整)]\n- 分享卡片",
		PrivateDataFolder: "ygosem",
	})
	en.OnRegex(`^/ys\s*(.*)?`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		searchdata := strings.SplitN(ctx.State["regex_matched"].([]string)[1], " -", 2)
		searchName := searchdata[0]
		condition := ""
		if len(searchdata) > 1 {
			condition = searchdata[1]
		}
		if searchName == "" { // 如果是随机抽卡
			url := "https://www.ygo-sem.cn/Cards/Default.aspx"
			// 请求html页面
			listBody, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2], nil)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]", err))
				return
			}
			// 获取卡牌数量
			listmax := regexpmatch("条 共:(?s:(.*?))条</span>", string(listBody))
			if len(listmax) == 0 {
				ctx.SendChain(message.Text("数据存在错误: 无法获取当前卡池数量"))
				return
			}
			maxnumber, _ := strconv.Atoi(listmax[0][1])
			searchName = strconv.Itoa(rand.Intn(maxnumber + 1))
		}
		url := "https://www.ygo-sem.cn/Cards/S.aspx?q=" + url.QueryEscape(searchName)
		// 请求html页面
		body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2], nil)
		if err != nil {
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("[ERROR]", err)))
			return
		}
		cardInfo := helper.BytesToString(body)
		cardText := ""
		var pictrue []byte
		// 获取卡牌信息
		listmax := regexpmatch(`找到\s*(\d+)\s*个卡片`, cardInfo)
		if len(listmax) == 0 { // 只存在一张卡时
			switch condition {
			case "卡图":
				pictrue, err = getPic(cardInfo, true)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]", err))
					return
				}
				ctx.SendChain(message.ImageBytes(pictrue))
			case "描述":
				cardText = getDescribe(cardInfo)
				_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]", err))
					return
				}
				describePic, err := text.RenderToBase64(cardText, text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + helper.BytesToString(describePic)))
			case "调整":
				cardText = getAdjustment(cardInfo)
				_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]", err))
					return
				}
				describePic, err := text.RenderToBase64(cardText, text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + helper.BytesToString(describePic)))
			default:
				cardData := getCarddata(cardInfo)
				pictrue, err = getPic(cardInfo, false)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]", err))
					return
				}
				img, cl, err := drawimage(cardData, pictrue)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]", err))
					return
				}
				ctx.SendChain(message.ImageBytes(img))
				defer cl()
			}
			return
		}
		listmaxn := listmax[0][1]
		if listmaxn == "0" {
			ctx.SendChain(message.Text("没找到相关卡片，请检查卡名是否正确"))
			return
		}
		// 获取查找的列表
		cardList := getYGolist(string(body))
		cardsnames := make([]string, 0, len(cardList))
		i := 0
		for name := range cardList {
			cardsnames = append(cardsnames, strconv.Itoa(i)+"."+name)
			i++
		}
		listData := "找到" + listmaxn + "张相关卡片,当前显示以下卡名:\n" + strings.Join(cardsnames, "\n") +
			"\n————————————————\n输入对应数字获取卡片信息，\n或回复“取消”、“下一页”指令"
		ctx.SendChain(message.Text(listData))
		maxpage, _ := strconv.Atoi(listmaxn)
		var searchpage = 0 // 初始当前页面
		// 等待用户下一步选择
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(取消)|(下一页)|^\d+$`), zero.OnlyGroup, zero.CheckUser(ctx.Event.UserID)).Repeat()
		for {
			select {
			case <-time.After(time.Second * 40): // 40s等待
				cancel()
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("等待超时,搜索结束"),
					),
				)
				return
			case e := <-recv:
				nextcmd := e.Event.Message.String() // 获取下一个指令
				switch nextcmd {
				case "取消":
					cancel()
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("用户取消,搜索结束"),
						),
					)
					return
				case "下一页":
					searchpage++
					if searchpage > maxpage {
						searchpage = 0
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("已是最后一页，返回到第一页"))
					}
					url := "https://www.ygo-sem.cn/Cards/S.aspx?dRace=&attr=&q=" + searchName + "&start=" + strconv.Itoa(searchpage*30)
					// 请求html页面
					body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2], nil)
					if err != nil {
						ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("[ERROR]", err)))
						return
					}
					// 更新数据
					cardList = getYGolist(string(body))
					cardsnames = make([]string, 0, len(cardList))
					i := 0
					for name := range cardList {
						cardsnames = append(cardsnames, strconv.Itoa(i)+"."+name)
						i++
					}
					listData = "找到" + listmaxn + "张相关卡片,当前显示以下卡名:\n" + strings.Join(cardsnames, "\n") +
						"\n————————————————\n输入对应数字获取卡片信息，\n或回复“取消”、“下一页”指令"
					ctx.SendChain(message.Text(listData))
				default:
					cardint, err := strconv.Atoi(nextcmd)
					switch {
					case err != nil:
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
					default:
						if cardint < len(cardsnames) {
							cancel()
							cradsreach := strings.Split(cardsnames[cardint], ".")[1]
							url := "https://www.ygo-sem.cn/" + cardList[cradsreach]
							// 请求html页面
							body, err = web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2], nil)
							if err != nil {
								ctx.Send(message.Text("网页数据读取错误:", err))
								return
							}
							cardInfo = helper.BytesToString(body)
							switch condition {
							case "卡图":
								pictrue, err = getPic(cardInfo, true)
								if err != nil {
									ctx.SendChain(message.Text("[ERROR]", err))
									return
								}
								ctx.SendChain(message.ImageBytes(pictrue))
							case "描述":
								cardText = getDescribe(cardInfo)
								_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
								if err != nil {
									ctx.SendChain(message.Text("[ERROR]", err))
									return
								}
								describePic, err := text.RenderToBase64(cardText, text.BoldFontFile, 1280, 50)
								if err != nil {
									ctx.SendChain(message.Text("[ERROR]", err))
									return
								}
								ctx.SendChain(message.Image("base64://" + helper.BytesToString(describePic)))
							case "调整":
								cardText = getAdjustment(cardInfo)
								_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
								if err != nil {
									ctx.SendChain(message.Text("[ERROR]", err))
									return
								}
								describePic, err := text.RenderToBase64(cardText, text.BoldFontFile, 1280, 50)
								if err != nil {
									ctx.SendChain(message.Text("[ERROR]", err))
									return
								}
								ctx.SendChain(message.Image("base64://" + helper.BytesToString(describePic)))
							default:
								cardData := getCarddata(cardInfo)
								pictrue, err = getPic(cardInfo, false)
								if err != nil {
									ctx.SendChain(message.Text("[ERROR]", err))
									return
								}
								img, cl, err := drawimage(cardData, pictrue)
								if err != nil {
									ctx.SendChain(message.Text("[ERROR]", err))
									return
								}
								ctx.SendChain(message.ImageBytes(img))
								defer cl()
							}
							return
						}
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
					}
				}
			}
		}
	})
	en.OnFullMatch("分享卡片", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		url := "https://www.ygo-sem.cn/Cards/Default.aspx"
		// 请求html页面
		listBody, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2], nil)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 获取卡牌数量
		listmax := regexpmatch(`条 共:\s*(?s:(.*?))\s*条</span>`, string(listBody))
		if len(listmax) == 0 {
			ctx.SendChain(message.Text("数据存在错误: 无法获取当前卡池数量"))
			return
		}
		listnumber := listmax[0][1]
		maxnumber, _ := strconv.Atoi(listnumber)
		searchName := strconv.Itoa(rand.Intn(maxnumber + 1))
		url = "https://www.ygo-sem.cn/Cards/S.aspx?q=" + searchName
		// 请求html页面
		body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2], nil)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		cardInfo := helper.BytesToString(body)
		cardData := getCarddata(cardInfo)
		cardData.Maxcard = listnumber
		pictrue, err := getPic(cardInfo, false)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 分享卡片
		img, cl, err := drawimage(cardData, pictrue)
		defer cl()
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		ctx.SendChain(message.ImageBytes(img))
	})
}

// 绘制图片
func drawimage(cardInfo gameCardInfo, pictrue []byte) (data []byte, cl func(), err error) {
	// 卡图大小
	cardPic, _, err := image.Decode(bytes.NewReader(pictrue))
	if err != nil {
		return
	}
	cardPic = img.Size(cardPic, 400, 580).Im
	picx := cardPic.Bounds().Dx()
	picy := cardPic.Bounds().Dy()
	_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	textWidth := 1200
	if !strings.Contains(cardInfo.Type, "怪兽") {
		textWidth = 1150 - picx
	}
	textPic, err := text.Render(cardInfo.Depict, text.BoldFontFile, textWidth, 50)
	if err != nil {
		return
	}
	picHigh := picy + 30
	if strings.Contains(cardInfo.Type, "怪兽") {
		picHigh += textPic.Bounds().Dy() + 30
	} else {
		picHigh = math.Max(picHigh, 50+(20+30)*4+10+textPic.Bounds().Dy())
	}
	/***********设置图片的大小和底色***********/
	canvas := gg.NewContext(1300, picHigh)
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()
	// 放置卡图
	canvas.DrawImage(cardPic, 1270-picx, 10)
	// 写内容
	if err = canvas.LoadFontFace(text.BoldFontFile, 50); err != nil {
		return
	}
	canvas.SetRGB(0, 0, 0)
	_, h := canvas.MeasureString("游戏王")
	listnumber := cardInfo.Maxcard
	textHigh := 50.0
	if listnumber != "" {
		canvas.DrawString("当前卡池总数:"+listnumber, 10, 50)
		canvas.DrawString("今日分享卡片:", 10, textHigh+h+30)
		textHigh += (h + 30) * 2
	}
	textPicy := textHigh + h*3 + 30*2
	canvas.DrawString("卡名:    "+cardInfo.Name, 10, textHigh)
	canvas.DrawString("卡密:    "+cardInfo.ID, 10, textHigh+h+30)
	canvas.DrawString("种类:    "+cardInfo.Type, 10, textHigh+(h+30)*2)
	if strings.Contains(cardInfo.Type, "怪兽") {
		canvas.DrawString("种族:    "+cardInfo.Race, 10, textHigh+(h+30)*3+10)
		canvas.DrawString("属性:    "+cardInfo.Attr, 10, textHigh+(h+30)*4+10)
		if strings.Contains(cardInfo.Type, "连接") {
			canvas.DrawString(cardInfo.Level, 10, textHigh+(h+30)*5+10)
			canvas.DrawString("ATK:"+cardInfo.Atk, 10, textHigh+(h+30)*6+10)
		} else {
			canvas.DrawString("星数/阶级:"+cardInfo.Level, 10, textHigh+(h+30)*5+10)
			canvas.DrawString("ATK:"+cardInfo.Atk+"/def:"+cardInfo.Def, 10, textHigh+(h+30)*6+10)
		}
		textPicy = textHigh + (h+30)*6 + 30
	}
	// 放置效果
	canvas.DrawImage(textPic, 10, int(textPicy)+10)
	// 生成图片
	data, cl = writer.ToBytes(canvas.Image())
	return
}
