// Package ygo 一些关于ygo的插件
package ygo

// 本插件查卡通过网页"https://www.ygo-sem.cn/"获取的

import (
	"bytes"
	"errors"
	"image"
	"math/rand"
	"regexp"
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
)

var (
	reqconf = [...]string{"GET", "https://www.ygo-sem.cn/",
		"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"}
)

func init() {
	en := control.Register("ygosem", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "游戏王进阶平台卡查",
		Help: "1.指令：/ys [卡名] [-(卡图|描述|调整)]\n" +
			"2.(开启|关闭)每日分享卡片",
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
			listBody, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
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
		url := "https://www.ygo-sem.cn/Cards/S.aspx?q=" + searchName
		// 请求html页面
		body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
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
				describePic, err := text.RenderToBase64(cardText, text.BoldFontFile, 600, 50)
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
				describePic, err := text.RenderToBase64(cardText, text.BoldFontFile, 600, 50)
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
		listData := "找到" + listmaxn + "张相关卡片,当前显示以下卡名：\n" + strings.Join(cardsnames, "\n") +
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
					body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
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
					listData = "找到" + listmaxn + "张相关卡片,当前显示以下卡名：\n" + strings.Join(cardsnames, "\n") +
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
							body, err = web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
							if err != nil {
								ctx.Send(message.Text("网页数据读取错误：", err))
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
								describePic, err := text.RenderToBase64(cardText, text.BoldFontFile, 600, 50)
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
								describePic, err := text.RenderToBase64(cardText, text.BoldFontFile, 600, 50)
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
		listBody, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 获取卡牌数量
		listmax := regexpmatch(`找到\s*(.*)\s*个卡片`, helper.BytesToString(listBody))
		if len(listmax) == 0 {
			ctx.SendChain(message.Text("今日分享卡片失败\n[error]:无法获取当前卡池数量"))
			return
		}
		listmaxn := listmax[0][1]
		maxnumber, _ := strconv.Atoi(listmaxn)
		searchName := strconv.Itoa(rand.Intn(maxnumber + 1))
		url = "https://www.ygo-sem.cn/Cards/S.aspx?q=" + searchName
		// 请求html页面
		body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		cardInfo := helper.BytesToString(body)
		cardData := getCarddata(cardInfo)
		pictrue, err := getPic(cardInfo, false)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 分享卡片
		ctx.SendChain(message.Text("当前游戏王卡池总数："+listmaxn+"\n\n今日分享卡牌:\n\n"), message.ImageBytes(pictrue), message.Text(cardData))
	})
}

// 正则筛选数据
func regexpmatch(rule, str string) [][]string {
	return regexp.MustCompile(rule).FindAllStringSubmatch(str, -1)
}

// 获取卡名列表
func getYGolist(body string) (cardsname map[string]string) {
	nameList := regexpmatch(`<div class="icon_size" style="white-space: nowrap;">\s*<a href="..(.*)" target="_blank">(.*)</a>\s*</div>`, body)
	if len(nameList) != 0 {
		cardsname = make(map[string]string, len(nameList)*2)
		for _, names := range nameList {
			cardsname[names[2]] = names[1]
		}
	}
	return
}

// 获取卡面信息
func getCarddata(body string) (cardata map[string]string) {
	cardata = make(map[string]string, 20)
	// 获取卡名
	cardName := regexpmatch(`<b>中文名</b> </span>&nbsp;<span class="item_box_value">\s*(.*)</span>\s*</div>`, body)
	if len(cardName) == 0 {
		return
	}
	cardata["卡名"] = cardName[0][1]
	// 获取卡密
	cardID := regexpmatch(`<b>卡片密码</b> </span>&nbsp;<span class="item_box_value">\s*(.*)\s*</span>`, body)
	cardata["卡密"] = cardID[0][1]
	// 种类
	cardType := regexpmatch(`<b>卡片种类</b> </span>&nbsp;<span class="item_box_value" id="dCnType">\s*(.*?)\s*</span>\s*<span`, body)
	cardata["种类"] = cardType[0][1]
	if strings.Contains(cardType[0][1], "怪兽") {
		// 种族
		cardRace := regexpmatch(`<span id="dCnRace" class="item_box_value">\s*(.*)\s*</span>\s*<span id="dEnRace"`, body)
		cardata["种族"] = cardRace[0][1]
		// 属性
		cardAttr := regexpmatch(`<b>属性</b> </span>&nbsp;<span class="item_box_value" id="attr">\s*(.*)\s*</span>`, body)
		cardata["属性"] = cardAttr[0][1]
		/*星数*/
		switch {
		case strings.Contains(cardType[0][1], "连接"):
			cardLevel := regexpmatch(`<span class="item_box_value">(LINK.*)</span>`, body)
			cardata["等级"] = cardLevel[0][1]
		default:
			cardLevel := regexpmatch(`<b>星数/阶级</b> </span><span class=\"item_box_value\">\s*(.*)\s*</span>`, body)
			cardata["等级"] = cardLevel[0][1]
			// 守备力
			cardDef := regexpmatch(`<b>DEF</b></span>\s*&nbsp;<span class="item_box_value">\s*(\d+|\?|？)\s*</span>\s*</div>`, body)
			cardata["守备力"] = cardDef[0][1]
		}
		// 攻击力
		cardAtk := regexpmatch(`<b>ATK</b> </span>&nbsp;<span class=\"item_box_value\">\s*(\d+|\?|？)\s*</span>`, body)
		cardata["攻击力"] = cardAtk[0][1]
	}
	/*效果*/
	cardDepict := regexpmatch(`<div class="item_box_text" id="cardDepict">\s*(?s:(.*?))\s*</div>`, body)
	cardata["效果"] = cardDepict[0][1]
	//cardata["效果"] = strings.ReplaceAll(cardDepict[0][1], " ", "")
	return
}

// 获取卡图
func getPic(body string, choosepic bool) (imageBytes []byte, err error) {
	// 获取卡图连接
	cardpic := regexpmatch(`picsCN(/\d+/\d+).jpg`, body)
	if len(cardpic) == 0 {
		return nil, errors.New("getPic正则匹配失败")
	}
	choose := "larg/"
	if !choosepic {
		choose = "picsCN/"
	}
	picHref := "https://www.ygo-sem.cn/yugioh/" + choose + cardpic[0][1] + ".jpg"
	// 读取获取的[]byte数据
	return web.RequestDataWith(web.NewDefaultClient(), picHref, reqconf[0], reqconf[1], reqconf[2])
}

// 获取描述
func getDescribe(body string) string {
	cardName := regexpmatch(`<b>中文名</b> </span>&nbsp;<span class="item_box_value">\s*(?s:(.*?))\s*</span>\s*</div>`, body)
	if len(cardName) == 0 {
		return "查无此卡"
	}
	describeinfo := regexpmatch(`<span class="cont-list" style="background-color: rgba(0, 0, 0, 0.7); padding: 4px; line-height: 24px; z-index: 2; color: rgb(255, 255, 255);">\s*(.*)\s*</span>`, body)
	if len(describeinfo) == 0 {
		return "无相关描述,请期待更新"
	}
	getdescribe := ""
	href1 := regexpmatch("<span(.*?)data-content=(.*?)'>(.*?)</span>", describeinfo[0][1])
	if href1 != nil {
		getdescribe = describeinfo[0][1]
		for _, hrefv := range href1 {
			getdescribe = strings.ReplaceAll(getdescribe, hrefv[0], "「"+hrefv[3]+"」")
		}
	}
	href2 := regexpmatch("<ahref='(.*?)'target='_blank'>(.*?)</a>", getdescribe)
	if len(href2) != 0 {
		for _, hrefv := range href2 {
			getdescribe = strings.ReplaceAll(getdescribe, hrefv[0], hrefv[2])
		}
	}
	getdescribe = strings.ReplaceAll(getdescribe, "</span>", "")
	getdescribe = strings.ReplaceAll(getdescribe, "<br/>", "\n")
	return "卡名：" + cardName[0][1] + "\n\n描述:\n" + getdescribe
}

// 获取调整
func getAdjustment(body string) string {
	adjustment := regexpmatch(`<div class="accordion-inner" id="adjust">.*<td>\s*(?s:(.*?))\s*</td>\s*</tr></tbody>\s*</table>\s*</div>`, body)
	if len(adjustment) == 0 {
		return "无相关调整，可以尝试搜索相关效果的旧卡"
	}
	return strings.ReplaceAll(adjustment[0][1], "<br/>", "\n")
}

// 绘制图片
func drawimage(cardInfo map[string]string, pictrue []byte) (data []byte, cl func(), err error) {
	// 卡图大小
	cardPic, _, err := image.Decode(bytes.NewReader(pictrue))
	if err != nil {
		return
	}
	picx := cardPic.Bounds().Dx()
	picy := cardPic.Bounds().Dy()
	_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	textWidth := 1200
	if !strings.Contains(cardInfo["种类"], "怪兽") {
		textWidth = 1250 - picx
	}
	dtext, err := text.Render(cardInfo["效果"], text.BoldFontFile, textWidth, 50)
	if err != nil {
		return
	}
	textPic := dtext.Image()
	/***********设置图片的大小和底色***********/
	canvas := gg.NewContext(1300, math.Max(500+textPic.Bounds().Dy()+30, picy+30))
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
	textPicy := 50 + h*3 + 30*3
	canvas.DrawString("卡名:    "+cardInfo["卡名"], 10, 50)
	canvas.DrawString("卡密:    "+cardInfo["卡密"], 10, 50+h+30)
	canvas.DrawString("种类:    "+cardInfo["种类"], 10, 50+h*2+30*2)
	if strings.Contains(cardInfo["种类"], "怪兽") {
		canvas.DrawString(cardInfo["种族"]+"族    "+cardInfo["属性"], 10, 50+h*3+30*3)
		if strings.Contains(cardInfo["种类"], "连接") {
			canvas.DrawString(cardInfo["等级"], 10, 20+h*5+30*4)
			canvas.DrawString("ATK:"+cardInfo["攻击力"], 10, 50+h*5+30*5)
		} else {
			canvas.DrawString("星数/阶级:"+cardInfo["等级"], 10, 50+h*4+30*4)
			canvas.DrawString("ATK:"+cardInfo["攻击力"]+"/def:"+cardInfo["守备力"], 10, 50+h*5+30*5)
		}
		textPicy = 50 + h*7 + 30*7
	}
	// 放置卡图
	canvas.DrawString("效果:", 10, textPicy)
	canvas.DrawImage(textPic, 10, int(textPicy))
	// 生成图片
	data, cl = writer.ToBytes(canvas.Image())
	return
}
