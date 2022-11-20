package ygoscore

import (
	"archive/zip"
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	//获取卡片信息
	"github.com/FloatTech/floatbox/web"

	// 图片输出
	"github.com/FloatTech/floatbox/file"
)

type cardInfo struct {
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
}

var (
	//游戏列表
	gamesInfo = map[string]string{
		"福悲喜": "双方玩家从ygo全卡池中随机抽取一张。\n" +
			"把那些卡攻击力确认(攻击力?的怪兽或者魔法·陷阱卡的场合攻击力当作0使用)。\n" +
			"展示攻击力高的玩家获得 10 枚ATRI币。",
		"救金鱼": "双方玩家从ygo全卡池中随机抽取一张。\n" +
			"把那些卡属性确认(魔法·陷阱卡的场合属性当作魔使用)。\n" +
			"如果属性相同则双方均获得 16 枚ATRI币。",
		"打赌胜负": "双方玩家从ygo全卡池中随机抽取一张。\n" +
			"把那些卡等级/阶级/连接值确认(魔法·陷阱卡的场合等级当作0使用)。\n" +
			"展示数值高的玩家获得 10 枚ATRI币。",
		"骰子壶": "双方玩家各自投掷1个骰子。\n" +
			"投掷出来的数目低的玩家将另一方投掷出的数目x2的ATRI币交给对方。\n" +
			"如果输给投掷出来的数目为6的场合,移交的ATRI币变成20。\n平局的场合再掷一次直到数目不一样。",
		"大金星!?": "双方玩家宣言1到12的任意数值,并各自进行一次投掷硬币。\n" +
			"如果都是表的场合,双方各自获得对方宣言的数值ATRI币。\n" +
			"如果都是里的场合,双方各自失去对方宣言的数值ATRI币。\n" +
			"否则投掷出里的玩家失去对方宣言的数值ATRI币;\n投掷出表的玩家获得自己宣言的数值ATRI币。",
		"通贩卖员": "双方玩家从ygo全卡池中随机抽取一张。\n" +
			"把那些卡种类确认。\n" +
			"同为怪兽时,各自的签到天数+2。\n" +
			"同为魔法时,各自的ATRI币+10。\n" +
			"同为陷阱时,各自的ATRI币-2。\n",
	}
	cards     = make(map[int]*cardInfo)
	cardsList []int
)

func init() {
	var gamesList []string
	helper := ""
	i := 0
	for key, value := range gamesInfo {
		i++
		gamesList = append(gamesList, key)
		helper += strconv.Itoa(i) + key + ":\n" + value + "\n"
	}
	engine := control.Register("ygogames", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:           "游戏王小游戏插件",
		Help:
			"-玩小游戏 @群友\n" +
			"======小游戏内容======\n" +
			"每个小游戏游玩均需交 6 枚ATRI币\n" +
			helper,
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("已经有正在进行的游戏..."),
				),
			)
		}),
	))

	zipfile := file.BOTPATH + "/data/ygoscore/ygocdb.com.cards.zip"
	err := parsezip(zipfile)
	if err != nil {
		panic(err)
	}

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		scoredata.db.DBPath = file.BOTPATH + "/data/ygoscore/score.db"
		err := scoredata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		return true
	})

	engine.OnRegex(`^玩小游戏\s*?\[CQ:at,qq=(\d+)\].*?`, zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		duelUser, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		uid := ctx.Event.UserID
		if duelUser == uid {
			ctx.SendChain(message.Text("禁止左右手互博"))
			return
		}
		userinfo, err := scoredata.checkuser(uid)
		if err != nil {
			ctx.SendChain(message.Text(serviceErr, err))
			return
		}
		if userinfo.UserName == "" {
			ctx.SendChain(message.Text("决斗者未注册!\n请输入“注册决斗者 xxx”进行登记(xxx为决斗者昵称)。"))
			return
		}
		challenginfo, err := scoredata.checkuser(duelUser)
		if err != nil {
			ctx.SendChain(message.Text(serviceErr, err))
			return
		}
		if challenginfo.UserName == "" {
			ctx.SendChain(message.Text("决斗者未注册!\n请输入“注册决斗者 xxx”进行登记(xxx为决斗者昵称)。"))
			return
		}
		userScore := wallet.GetWalletOf(uid)
		if userScore <= 6 {
			ctx.SendChain(message.Text("你的ATRI币不足以满足该游戏\n可通过签到获取,连续签到获得的ATRI币越高哟"))
			return
		}
		challengScore := wallet.GetWalletOf(duelUser)
		if challengScore <= 6 {
			ctx.SendChain(message.Text("他的ATRI币不足以满足该游戏\n可通过签到获取,连续签到获得的ATRI币越高哟"))
			return
		}
		// 等待对方响应
		ctx.SendChain(message.Text("等待对方发送“duel|决斗|拒绝”进行回复"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.FullMatchRule("duel", "决斗", "拒绝"), zero.CheckUser(duelUser), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()
		answer := ""
		wait := time.NewTimer(120 * time.Second)
		for {
			select {
			case <-wait.C:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("时间超时,游戏取消")))
				return
			case c := <-recv:
				answer = c.Event.Message.String()
				if answer == "拒绝" {
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("对方拒绝了你的邀请,游戏结束")))
					return
				}
			}
			if answer == "duel" || answer == "决斗" {
				break
			}
		}
		gamesListName := ""
		for i, name := range gamesList {
			gamesListName += strconv.Itoa(i+1) + "." + name + "\n"
		}
		ctx.SendChain(message.Text("请选择游戏模式：\n", gamesListName))
		recv, cancel = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^[1-`+strconv.Itoa(len(gamesList))+`]$`), zero.CheckUser(uid, duelUser), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()
		modeNum := 0
		wait = time.NewTimer(120 * time.Second)
		for {
			select {
			case <-wait.C:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("选择超时,游戏取消")))
				return
			case c := <-recv:
				modeNum, _ = strconv.Atoi(c.Event.Message.String())
			}
			if modeNum != 0 {
				break
			}
		}
		compare := gamesList[modeNum-1]
		ctx.SendChain(message.Text("游戏规则:\n", gamesInfo[compare]))
		time.Sleep(3 * time.Second)
		var duel = make(map[int64]int, 2)
		wait = time.NewTimer(120 * time.Second)
		var next *zero.FutureEvent
		switch {
		case compare == "骰子壶":
			ctx.SendChain(message.Text("游戏开始,请说出你的带“投”或“骰”的话语进行投掷骰子"))
			next = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.KeywordRule("投", "骰"), zero.CheckUser(duelUser, uid), zero.CheckGroup(ctx.Event.GroupID))
		case compare == "大金星!?":
			ctx.SendChain(message.Text("请各自宣言1~12的数值"))
			next = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^([1-9]|[1][0-2])$`), zero.CheckUser(uid, duelUser), zero.CheckGroup(ctx.Event.GroupID))
		default:
			ctx.SendChain(message.Text("游戏开始,请说出你的带“抽卡”的话语进行抽卡"))
			next = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.KeywordRule("抽卡"), zero.CheckUser(duelUser, uid), zero.CheckGroup(ctx.Event.GroupID))
		}
		recv, cancel = next.Repeat()
		defer cancel()
		for {
			select {
			case <-wait.C:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("等待超时,游戏取消\n(咕之人的ATRI币并不返回)")))
				_, ok := duel[uid]
				if !ok {
					err := wallet.InsertWalletOf(uid, -6)
					if err != nil {
						ctx.SendChain(message.Text(serviceErr, err))
					}
				}
				_, ok = duel[duelUser]
				if !ok {
					err := wallet.InsertWalletOf(duelUser, -6)
					if err != nil {
						ctx.SendChain(message.Text(serviceErr, err))
					}
				}
				return
			case c := <-recv:
				eventID := c.Event.UserID
				_, ok := duel[eventID]
				if !ok {
					switch {
					case compare == "骰子壶":
						dice := rand.Intn(6) + 1
						duel[eventID] = dice
						ctx.SendChain(message.At(eventID), message.Text("\n你投掷出了数目:", dice))
					case compare == "大金星!?":
						Num, _ := strconv.Atoi(c.Event.Message.String())
						duel[eventID] = Num
					default:
						cardID := cardsList[rand.Intn(len(cardsList))]
						duel[eventID] = cardID
						picID := cards[cardID].ID
						ctx.SendChain(message.At(eventID),
							message.Text("\n你抽到了:\n"),
							message.Image("https://cdn.233.momobako.com/ygopro/pics/"+
								strconv.Itoa(picID)+".jpg"))
					}
				}
			}
			dice1, ok1 := duel[uid]
			dice2, ok2 := duel[duelUser]
			if !ok1 || !ok2 {
				continue
			}
			if compare == "骰子壶" {
				if dice1 != dice2 {
					break
				}
				//如果点数一样就清空
				ctx.SendChain(message.Text("你们投掷的数目相同,请重新投掷"))
				duel = make(map[int64]int, 2)
			} else if ok1 && ok2 {
				break
			}
		}
		result := 0
		var winID int64
		switch compare {
		case "福悲喜":
			result = cards[duel[uid]].Data.Atk - cards[duel[duelUser]].Data.Atk
			switch {
			case result > 0:
				err = wallet.InsertWalletOf(uid, 10)
				winID = uid
			case result < 0:
				err = wallet.InsertWalletOf(duelUser, 10)
				winID = duelUser
			default:
				err = wallet.InsertWalletOf(uid, -1)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, -1)
				}
			}
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			if winID != 0 {
				ctx.SendChain(message.At(winID), message.Text("恭喜你赢得了「", compare, "」游戏,ATRI币+10\n可发送“/钱包”查看"))
			} else {
				ctx.SendChain(message.Text("打平局还行,你们双打组队一定很强吧。\n扣除游玩费 1 枚ATRI币,返回你们 5 枚ATRI币"))
			}
		case "通贩卖员":
			uType := cards[duel[uid]].Text.Types
			dType := cards[duel[duelUser]].Text.Types
			if strings.Contains(uType, "怪兽") && strings.Contains(dType, "怪兽") {
				userinfo.Continuous += 2
				challenginfo.Continuous += 2
				if err = scoredata.db.Insert("score", &userinfo); err == nil {
					err = scoredata.db.Insert("score", &challenginfo)
				}
				result = 1
			} else if strings.Contains(uType, "魔法") && strings.Contains(dType, "魔法") {
				err = wallet.InsertWalletOf(uid, 10)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, 10)
				}
				result = 2
			} else if strings.Contains(uType, "陷阱") && strings.Contains(dType, "陷阱") {
				err = wallet.InsertWalletOf(uid, -10)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, -10)
				}
				result = 3
			}
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			switch result {
			case 0:
				ctx.SendChain(message.Text("很遗憾,你们之间的默契太低了。下次再来玩~"))
			case 1:
				ctx.SendChain(message.Text("你们完成了「", compare, "」游戏,你们签到天数+2\n可发送“/钱包”查看"))
			case 2:
				ctx.SendChain(message.Text("你们完成了「", compare, "」游戏,你们ATRI币+10\n可发送“/钱包”查看"))
			case 3:
				ctx.SendChain(message.Text("你们完成了「", compare, "」游戏,你们ATRI币-10\n可发送“/钱包”查看"))
			}
		case "救金鱼":
			result = cards[duel[uid]].Data.Attribute - cards[duel[duelUser]].Data.Attribute
			if result == 0 {
				err = wallet.InsertWalletOf(uid, 16)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, 16)
				}
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
				ctx.SendChain(message.Text("你们真默契！你们完成了「", compare, "」游戏,你们的ATRI币+16\n可发送“/钱包”查看"))
			} else {
				err = wallet.InsertWalletOf(uid, -6)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, -6)
				}
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
				ctx.SendChain(message.Text("很遗憾,你们没有抽到相同的属性。游玩费 6 枚ATRI币我拿走了"))
			}
		case "打赌胜负":
			result = cards[duel[uid]].Data.Level - cards[duel[duelUser]].Data.Level
			switch {
			case result > 0:
				err = wallet.InsertWalletOf(uid, 10)
				winID = uid
			case result < 0:
				err = wallet.InsertWalletOf(duelUser, 10)
				winID = duelUser
			default:
				err = wallet.InsertWalletOf(uid, -5)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, -5)
				}
			}
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			if winID != 0 {
				ctx.SendChain(message.At(winID), message.Text("恭喜你赢得了「", compare, "」游戏,ATRI币+10\n可发送“/钱包”查看"))
			} else {
				ctx.SendChain(message.Text("打平局还行,你们双打组队一定很强吧。\n扣除游玩费 1 枚ATRI币,返回你们 5 枚ATRI币"))
			}
		case "骰子壶":
			uDice := duel[uid]
			cDice := duel[duelUser]
			result = uDice - cDice
			points := 20
			var lostID int64
			switch {
			case result > 0:
				if duel[uid] != 6 {
					points = uDice * 2
				}
				winID = uid
				lostID = duelUser
			case result < 0:
				if duel[duelUser] != 6 {
					points = uDice * 2
				}
				winID = duelUser
				lostID = uid
			}
			//数据结算
			err = wallet.InsertWalletOf(winID, points)
			if err == nil {
				err = wallet.InsertWalletOf(lostID, -points)
			}
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			ctx.SendChain(message.At(winID), message.Text("恭喜你赢得了「", compare, "」游戏,ATRI币+", points, "\n可发送“/钱包”查看"))
		case "大金星!?":
			result := []string{"里", "表"}
			uPoints := duel[uid]
			cPoints := duel[duelUser]
			ctx.SendChain(message.Text("游戏开始,双方同时扔出了银币..."))
			uDice := rand.Intn(2)
			time.Sleep(3 * time.Second)
			cDice := rand.Intn(2)
			ctx.SendChain(message.Text("结果出来了！\n",
				userinfo.UserName, "投出的硬币是", result[uDice], "\n",
				challenginfo.UserName, "投出的硬币是", result[cDice]))
			resultPoints := 0
			switch {
			case uDice == 0 && cDice == 0:
				err = wallet.InsertWalletOf(uid, -cPoints)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, -uPoints)
				}
			case uDice == 1 && cDice == 1:
				err = wallet.InsertWalletOf(uid, cPoints)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, uPoints)
				}
				resultPoints = 1
			case uDice == 0 && cDice == 1:
				err = wallet.InsertWalletOf(uid, -cPoints)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, cPoints)
				}
				resultPoints = 2
				winID = duelUser
			case uDice == 1 && cDice == 0:
				err = wallet.InsertWalletOf(uid, uPoints)
				if err == nil {
					err = wallet.InsertWalletOf(duelUser, -uPoints)
				}
				resultPoints = 3
				winID = uid
			}
			//数据结算
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			switch resultPoints {
			case 0:
				ctx.SendChain(message.Text("真是糟糕,你们各自失去了对方宣言数值的ATRI币\n可发送“/钱包”查看"))
			case 1:
				ctx.SendChain(message.Text("恭喜你们获得了对方宣言数值的ATRI币~\n可发送“/钱包”查看"))
			case 2:
				ctx.SendChain(message.At(winID), message.Text("恭喜你赢得了「", compare, "」游戏,ATRI币+", uPoints, "\n可发送“/钱包”查看"))
			case 3:
				ctx.SendChain(message.At(winID), message.Text("恭喜你赢得了「", compare, "」游戏,TRI币+", cPoints, "\n可发送“/钱包”查看"))
			}
		}
	})
}

// 获取卡表
func parsezip(zipFile string) error {
	if file.IsNotExist(zipFile) {
		cardsURL := "https://ygocdb.com/api/v0/cards.zip"
		data, err := web.GetData(cardsURL)
		if err != nil {
			return err
		}
		err = os.WriteFile(zipFile, data, 0666)
		if err != nil {
			return err
		}
	}
	zipReader, err := zip.OpenReader(zipFile) // will not close
	if err != nil {
		return err
	}
	for _, f := range zipReader.File {
		if f.FileInfo().Name() == "cards.json" {
			reader, err := f.Open()
			if err != nil {
				return err
			}
			err = json.NewDecoder(reader).Decode(&cards)
			if err != nil {
				return err
			}
			err = reader.Close()
			if err != nil {
				return err
			}
			break
		}
	}
	for key := range cards {
		cardsList = append(cardsList, key)
	}
	return nil
}
