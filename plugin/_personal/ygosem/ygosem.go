// 本插件查卡通过网页"https://www.ygo-sem.cn/"获取的
package ygosem

import (
	"encoding/base64"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/process"
	sql "github.com/FloatTech/sqlite"
)

type groupinfo struct {
	Groupid    string
	Switch     int
	Updatetime string
}

var (
	db      = &sql.Sqlite{}
	dbmu    sync.RWMutex
	reqconf = [...]string{"GET", "https://www.ygo-sem.cn/",
		"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"}
)

// 正则筛选数据
func regexpmatch(rule, str string) (regexpresult [][]string, ok bool) {
	regexp_rule := regexp.MustCompile(rule)
	regexpresult = regexp_rule.FindAllStringSubmatch(str, -1)
	if regexpresult == nil {
		ok = false
	} else {
		ok = true
	}
	return
}

func init() {
	en := control.Register("ygosem", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:           "游戏王进阶平台卡查",
		Help: "1.指令：/ys [卡名] [-(卡图|描述|调整)]\n" +
			"2.(开启|关闭)每日分享卡片",
		PrivateDataFolder: "ygosem",
	})

	go func() {
		db.DBPath = en.DataFolder() + "ygosem.db"
		db.Open(time.Hour * 24)
		db.Create("lookupgroupid", &groupinfo{})
	}()

	en.OnRegex(`^[(.|。|\/|\\)]ys((\s)?(.+))?`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		searchdata := strings.SplitN(ctx.State["regex_matched"].([]string)[3], " -", 2)
		searchName := searchdata[0]
		condition := ""
		if len(searchdata) > 1 {
			condition = searchdata[1]
		}
		if searchName == "" { // 如果是随机抽卡
			url := "https://www.ygo-sem.cn/Cards/Default.aspx"
			// 请求html页面
			list_body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
			if err != nil {
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("服务器读取错误：", err)))
				return
			}
			// 获取卡牌数量
			listmax, ok := regexpmatch("条 共:(?s:(.*?))条</span>", string(list_body))
			if !ok {
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("数据存在错误: 无法获取当前卡池数量")))
				return
			}
			maxnumber, _ := strconv.Atoi(listmax[0][1])
			searchName = strconv.Itoa(rand.Intn(maxnumber + 1))
		}
		url := "https://www.ygo-sem.cn/Cards/S.aspx?q=" + searchName
		// 请求html页面
		body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
		if err != nil {
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("服务器读取错误：", err)))
			return
		}
		// 获取卡牌信息
		listmax, ok := regexpmatch("找到(?s:(.*?))个卡片", string(body))
		if !ok { // 只存在一张卡时
			card_data, pic, ok := getYGOdata(string(body), condition)
			switch {
			case !ok:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("数据获取失败")))
			case card_data != "" && pic != "":
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image("base64://"+pic), message.Text(card_data)))
			case card_data != "":
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(card_data)))
			case pic != "":
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image("base64://"+pic)))
			}
			return
		}
		listmaxn := strings.ReplaceAll(listmax[0][1], " ", "")
		listmaxn = strings.ReplaceAll(listmaxn, "\r\n", "")
		if listmaxn == "0" {
			ctx.SendChain(message.Text("没找到相关卡片，请检查卡名是否正确"))
			return
		}
		// 获取查找的列表
		cardsname, cardshref := getYGolist(string(body))
		list_data := "找到" + listmaxn + "张相关卡片,当前显示以下卡名：\n" + strings.Join(cardsname, "\n") +
			"\n————————————————\n输入对应数字获取卡片信息，\n或回复“取消”、“下一页”指令"
		ctx.SendChain(message.Text(list_data))
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
					searchpage += 1
					if searchpage > maxpage {
						searchpage = 0
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("已是最后一页，返回到第一页"))
					}
					url := "https://www.ygo-sem.cn/Cards/S.aspx?dRace=&attr=&q=" + searchName + "&start=" + strconv.Itoa(searchpage*30)
					// 请求html页面
					body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
					if err != nil {
						ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("服务器读取错误：", err)))
						return
					}
					// 更新数据
					cardsname, cardshref = getYGolist(string(body))
					list_data := "找到" + listmaxn + "张相关卡片,当前显示以下卡名：\n" + strings.Join(cardsname, "\n") +
						"\n————————————————\n输入对应数字获取卡片信息，\n或回复“取消”、“下一页”指令"
					ctx.SendChain(message.Text(list_data))
				default:
					cardint, err := strconv.Atoi(nextcmd)
					switch {
					case err != nil:
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
					default:
						if cardint < len(cardsname) {
							cancel()
							url := "https://www.ygo-sem.cn/" + cardshref[cardint]
							// 请求html页面
							body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
							if err != nil {
								ctx.Send(message.Text("网页数据读取错误：", err))
								return
							}
							card_data, pic, ok := getYGOdata(string(body), condition)
							switch {
							case !ok:
								ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("数据获取失败")))
							case card_data != "" && pic != "":
								ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image("base64://"+pic), message.Text(card_data)))
							case card_data != "":
								ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(card_data)))
							case pic != "":
								ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Image("base64://"+pic)))
							}
							return
						} else {
							ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
						}
					}
				}
			}
		}
	})
	en.OnFullMatch("分享卡片", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		url := "https://www.ygo-sem.cn/Cards/Default.aspx"
		// 请求html页面
		list_body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
		if err != nil {
			ctx.SendChain(message.Text("今日分享卡片失败\n[error]:无法连接当前服务器"))
			return
		}
		// 获取卡牌数量
		listmax, ok := regexpmatch("条 共:(?s:(.*?))条</span>", string(list_body))
		if !ok {
			ctx.SendChain(message.Text("今日分享卡片失败\n[error]:无法获取当前卡池数量"))
			return
		}
		listmaxn := listmax[0][1]
		maxnumber, _ := strconv.Atoi(listmaxn)
		searchName := strconv.Itoa(rand.Intn(maxnumber + 1))
		url = "https://www.ygo-sem.cn/Cards/S.aspx?q=" + searchName
		// 请求html页面
		body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
		card_data, pic, ok := getYGOdata(string(body), "")
		if !ok {
			ctx.SendChain(message.Text("今日分享卡片失败\n[error]:无法获取卡片信息"))
			return
		}
		// 分享卡片
		ctx.SendChain(message.Text("当前游戏王卡池总数："+listmaxn+"\n\n今日分享卡牌：\n\n"), message.Image("base64://"+pic), message.Text(card_data))
	})
	// */每天12点随机分享一张卡
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		// 如果数据库不存在则下载
		db.DBPath = en.DataFolder() + "ygosem.db"
		// _, _ = engine.GetLazyData("SetuTime.db", false)
		err := db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		if err := db.Create("lookupgroupid", &groupinfo{}); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		return true
	})
	en.OnRegex(`^(开启|关闭|查询)每日分享卡片$`, zero.OnlyGroup, zero.AdminPermission, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cmd := ctx.State["regex_matched"].([]string)[1]
		gid := ctx.Event.GroupID
		dbmu.Lock()
		defer dbmu.Unlock()
		var ginfo groupinfo
		gidstr := strconv.FormatInt(gid, 10)
		err := db.Find("lookupgroupid", &ginfo, "where Groupid = "+gidstr)
		if cmd == "查询" {
			switch {
			case err != nil:
				ctx.SendChain(message.Text("服务未开启"))
			case ginfo.Switch == 1:
				ctx.SendChain(message.Text("服务已开启"))
			case ginfo.Switch == 0:
				ctx.SendChain(message.Text("服务未开启"))
			}
			return
		}
		if err != nil {
			ginfo.Groupid = gidstr
		}
		switch cmd {
		case "开启":
			ginfo.Switch = 1
		case "关闭":
			ginfo.Switch = 0
		}
		ginfo.Updatetime = time.Now().Format("2006-01-02 15:04:05")
		if err := db.Insert("lookupgroupid", &ginfo); err != nil {
			ctx.SendChain(message.Text("服务状态更改失败！\n[ERROER]", err))
			return
		}
		ctx.SendChain(message.Text("服务状态已更改"))
	})
	process.CronTab.AddFunc("00 12 * * *", func() {
		url := "https://www.ygo-sem.cn/Cards/Default.aspx"
		// 请求html页面
		list_body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
		if err != nil {
			return
		}
		// 获取卡牌数量
		listmax, ok := regexpmatch("条 共:(?s:(.*?))条</span>", string(list_body))
		if !ok {
			return
		}
		listmaxn := listmax[0][1]
		maxnumber, _ := strconv.Atoi(listmaxn)
		var ginfo groupinfo
		// 向各群分享
		zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
			for _, g := range ctx.GetGroupList().Array() {
				grp := g.Get("group_id").Int()
				gidstr := g.Get("group_id").String()
				err := db.Find("lookupgroupid", &ginfo, "where Groupid = "+gidstr)
				if err != nil {
					continue
				}
				if ginfo.Switch == 1 {
					searchName := strconv.Itoa(rand.Intn(maxnumber + 1))
					url = "https://www.ygo-sem.cn/Cards/S.aspx?q=" + searchName
					// 请求html页面
					body, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2])
					if err != nil {
						ctx.SendGroupMessage(grp, message.Message{message.Text("今日分享卡片失败")})
					}
					card_data, pic, ok := getYGOdata(string(body), "")
					if !ok {
						ctx.SendGroupMessage(grp, message.Message{message.Text("今日分享卡片失败")})
					}
					// 输出数据
					ctx.SendGroupMessage(
						grp,
						message.Message{
							message.Text("当前游戏王卡池总数：" + listmaxn + "\n\n今日分享卡牌：\n\n"),
							message.Image("base64://" + pic),
							message.Text(card_data),
						})
				}
				process.SleepAbout1sTo2s()
			}
			return true
		})
	}) // */
}

// 获取字段列表
func getYGolist(body string) (cardsname []string, cardshref []string) {
	// 获取卡名列表
	regsult_name := regexp.MustCompile("height=\"144px\"  alt=\"(?s:(.*?))\" src=\"")
	namelist := regsult_name.FindAllStringSubmatch(body, -1)
	for i, names := range namelist {
		cardsname = append(cardsname, strconv.Itoa(i)+"."+names[1])
	}
	// 获取链接列表
	regsult_href := regexp.MustCompile("<a href=\"..(.*?)\" target=\"_blank\">")
	hreflist := regsult_href.FindAllStringSubmatch(body, -1)
	for i, frefs := range hreflist {
		if i > 1 && i%2 == 0 {
			cardshref = append(cardshref, frefs[1])
		}
	}
	return cardsname, cardshref
}

func getYGOdata(body, condition string) (card_data, pic string, ok bool) {
	switch condition {
	case "卡图":
		pic, ok = getPic(body, true)
	case "描述":
		card_data, ok = getDescribe(body)
	case "调整":
		card_data, ok = getAdjustment(body)
	default:
		card_data, ok = getCarddata(body)
		pic, ok = getPic(body, false)
	}
	return
}

// 获取卡面信息
func getCarddata(body string) (cardata string, ok bool) {
	var cardinfo []string
	// 获取卡名*/
	card_name, ok := regexpmatch("<b>中文名</b> </span>&nbsp;<span class=\"item_box_value\">(?s:(.*?))</span>", body)
	if !ok {
		return
	}
	cardname := strings.ReplaceAll(card_name[0][1], " ", "") // 去掉空格
	cardname = strings.ReplaceAll(cardname, "\r\n", "")      // 去掉空格
	cardinfo = append(cardinfo, cardname)
	/*种类*/
	card_type, ok := regexpmatch("<span class=\"item_box_value\" id=\"dCnType\">(?s:(.*?))</span>", body)
	if !ok {
		return
	}
	cardtype := strings.ReplaceAll(card_type[0][1], " ", "")
	cardtype = strings.ReplaceAll(cardtype, "\r\n", "")
	if strings.Contains(cardtype, "魔法") || strings.Contains(cardtype, "陷阱") {
		cardinfo = append(cardinfo, cardtype)
	} else {
		/*种族*/
		card_race, ok := regexpmatch("<span id=\"dCnRace\" class=\"item_box_value\">(?s:(.*?))</span>", body)
		if !ok {
			return "", ok
		}
		cardrace := strings.ReplaceAll(card_race[0][1], " ", "")
		cardrace = strings.ReplaceAll(cardrace, "\r\n", "")
		/*属性*/
		card_attr, ok := regexpmatch("<span class=\"item_box_value\" id=\"attr\">(?s:(.*?))</span>", body)
		if !ok {
			return "", ok
		}
		cardattr := strings.ReplaceAll(card_attr[0][1], " ", "")
		cardattr = strings.ReplaceAll(cardattr, "\r\n", "")
		cardinfo = append(cardinfo, cardtype+"  "+cardrace+"族/"+cardattr)
		/*星数*/
		level := ""
		switch {
		case strings.Contains(cardtype, "连接"):
			levelinfo, ok := regexpmatch("LINK(?s:(.*?))</span>", body)
			if !ok {
				levelinfo, ok = regexpmatch("\"item_box_value\">Link(?s:(.*?))</span>", body)
			}
			level = "LINK-" + strings.ReplaceAll(levelinfo[0][1], " ", "")
			level = strings.ReplaceAll(level, "\r\n", "")
		default:
			levelinfo, ok := regexpmatch("<b>星数/阶级</b> </span><span class=\"item_box_value\">(?s:(.*?))</span>", body)
			if !ok {
				return "", ok
			}
			if strings.Contains(cardtype, "超量") || strings.Contains(cardtype, "XYZ") {
				level = "阶级"
			} else {
				level = "等级"
			}
			level += strings.ReplaceAll(levelinfo[0][1], " ", "")
			level = strings.ReplaceAll(level, "\r\n", "")
		}
		// 攻击力
		card_atk, ok := regexpmatch("<b>ATK</b> </span>&nbsp;<span class=\"item_box_value\">(?s:(.*?))</span>", body)
		if !ok {
			return "", ok
		}
		cardatk := strings.ReplaceAll(card_atk[0][1], " ", "")
		cardatk = strings.ReplaceAll(cardatk, "\r\n", "")
		/*守备力*/
		card_def, ok := regexpmatch("<b>DEF</b></span>(?s:(.*?))</span>", body)
		if ok {
			card_def2 := strings.ReplaceAll(card_def[0][0], "\r\n", "") // 去掉空格
			card_def2 = strings.ReplaceAll(card_def2, " ", "")          // 去掉空格
			carddef, ok := regexpmatch("\"item_box_value\">(.*?)</span>", card_def2)
			if !ok {
				return "", ok
			}
			cardinfo = append(cardinfo, level+"  atk:"+cardatk+"  def:"+carddef[0][1])
		} else {
			cardinfo = append(cardinfo, level+"  atk:"+cardatk)
		}
	}
	cardata = strings.Join(cardinfo, "\n")
	/*效果*/
	result_depict, ok := regexpmatch("<div class=\"item_box_text\" id=\"cardDepict\">(?s:(.*?))</div>", body)
	card_depict := strings.ReplaceAll(result_depict[0][1], " ", "") // 去掉空格
	cardata += card_depict
	return
}

// 获取卡图
func getPic(body string, choosepic bool) (imageBase64 string, ok bool) {
	// 获取卡图连接
	cardpic, ok := regexpmatch("picsCN(?s:(.*?)).jpg", string(body))
	if !ok {
		return
	}
	choose := "larg/"
	if !choosepic {
		choose = "picsCN/"
	}
	pic_href := "https://www.ygo-sem.cn/yugioh/" + choose + cardpic[0][1] + ".jpg"
	// 读取获取的[]byte数据
	data, err := web.RequestDataWith(web.NewDefaultClient(), pic_href, reqconf[0], reqconf[1], reqconf[2])
	if err != nil {
		ok = false
		return
	}
	imageBase64 = base64.StdEncoding.EncodeToString(data)
	return
}

// 获取描述
func getDescribe(body string) (describe string, ok bool) {
	card_name, ok := regexpmatch("<b>中文名</b> </span>&nbsp;<span class=\"item_box_value\">(?s:(.*?))</span>", body)
	if !ok {
		return
	}
	cardname := strings.ReplaceAll(card_name[0][1], " ", "") // 去掉空格
	cardname = strings.ReplaceAll(cardname, "\r\n", "")      // 去掉空格
	describeinfo, ok := regexpmatch("<span class=\"cont-list\">(?s:(.*?))<span style=\"display:block;", body)
	if !ok {
		return "无相关描述,请期待更新", true
	}
	getdescribe := strings.ReplaceAll(describeinfo[0][1], "\r\n", "")
	getdescribe = strings.ReplaceAll(getdescribe, " ", "")
	href1, _ := regexpmatch("<span(.*?)data-content=(.*?)'>(.*?)</span>", getdescribe)
	if href1 != nil {
		for _, hrefv := range href1 {
			getdescribe = strings.ReplaceAll(getdescribe, hrefv[0], "「"+hrefv[3]+"」")
		}
	}
	href2, _ := regexpmatch("<ahref='(.*?)'target='_blank'>(.*?)</a>", getdescribe)
	if href2 != nil {
		for _, hrefv := range href2 {
			getdescribe = strings.ReplaceAll(getdescribe, hrefv[0], hrefv[2])
		}
	}
	getdescribe = strings.ReplaceAll(getdescribe, "</span>", "")
	getdescribe = strings.ReplaceAll(getdescribe, "<br/>", "\r\n")
	describe = "卡名：" + cardname + "\n\n描述：\n" + getdescribe
	return
}

// 获取调整
func getAdjustment(body string) (adjustment string, ok bool) {
	adjustmentinfo, ok := regexpmatch(`<div class="accordion-inner" id="adjust">(?s:(.*?))</div>`, body)
	if !ok {
		return "无相关调整，可以尝试搜索相关效果的旧卡", true
	}
	getadjustmentinfo := strings.ReplaceAll(adjustmentinfo[0][1], "\r\n", "")
	getadjustmentinfo = strings.ReplaceAll(getadjustmentinfo, " ", "")
	adjustmentinfo, ok = regexpmatch("<tableclass=\"table\"><tbody><tr><td>(.*?)</td></tbody></table>", getadjustmentinfo)
	if !ok {
		return
	}
	href, _ := regexpmatch("<a href='(.*?)' target='_blank'>(.*?)</a>", getadjustmentinfo)
	if href != nil {
		for _, hrefv := range href {
			getadjustmentinfo = strings.ReplaceAll(getadjustmentinfo, hrefv[0], hrefv[2])
		}
	}
	getadjustmentinfo = strings.ReplaceAll(adjustmentinfo[0][1], "<br/>", "\r\n")
	adjustment = getadjustmentinfo
	return
}
