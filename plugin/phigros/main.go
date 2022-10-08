// Package phigros ...
package phigros

import (
	"strconv"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	db  = &sql.Sqlite{}
	sdb = &sql.Sqlite{}
	mu  sync.RWMutex
	en  = control.Register("phigros", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "Phigros查分, 初次使用请先/phi init初始化存档\n" +
			"/phi init 初始化存档\n" +
			"/phi 查询成绩\n" +
			"/phi set <name>\n" +
			"/phi chall <等级> <数字>\n" +
			"/phi add <曲名> <难度> <acc> <分数> <0|1>?\n" +
			"示例: /phi chall rainbow 48\n" +
			"/phi add iga AT 100.0 1000000\n" +
			"Tips: /phi add中最后的<0|1>, 1代表全连",
		PublicDataFolder: "Phigros",
	})
	filepath = en.DataFolder()
)

func init() {
	db.DBPath = en.DataFolder() + "data.db"
	sdb.DBPath = en.DataFolder() + "songdata.db"
	go func() {
		err := db.Open(time.Hour * 24)
		if err != nil {
			panic(err)
		}
		err = sdb.Open(time.Hour * 24)
		if err != nil {
			panic(err)
		}
		err = db.Create("gamedata", &data{})
		if err != nil {
			panic(err)
		}
		err = db.Create("challen", &challen{})
		if err != nil {
			panic(err)
		}
		err = sdb.Create("songdata", &songdata{})
		if err != nil {
			panic(err)
		}
	}()
	en.OnFullMatch("/phi init").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()
		uid := ctx.Event.UserID
		struid := strconv.FormatInt(uid, 10)
		_ = db.Drop(struid)
		err := db.Create(struid, &result{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("初始化成功!"))
	})

	en.OnFullMatch("/phi").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()
		uid := ctx.Event.UserID
		struid := strconv.FormatInt(uid, 10)
		var d data
		err := db.Find("gamedata", &d, "WHERE UID = "+struid)
		if err != nil {
			ctx.SendChain(message.Text("看来你还没有绑定过呢"))
			return
		}
		ctx.SendChain(message.Text("正在查询..."))
		plname := d.Gamename
		var c challen
		var chal, chalnum string
		err = db.Find("challen", &c, "WHERE UID = "+struid)
		if err != nil {
			chal, chalnum = "", ""
		} else {
			chal, chalnum = c.Chall, strconv.FormatInt(c.Challnum, 10)
		}
		var dbnum int
		dbnum, err = db.Count(struid)
		if err != nil || dbnum == 0 {
			ctx.SendChain(message.Text("emm...看起来你好像还没添加过数据?"))
			return
		}
		var list = make([]result, 0, 22)
		var r result
		var m max
		err = db.Query("SELECT *, max(rksm) FROM ["+struid+"] WHERE rank='phi';", &m)
		if err != nil {
			list = append(list, result{})
		} else {
			list = append(list, result{Songname: m.Songname,
				ID:      m.ID,
				Diff:    m.Diff,
				Diffnum: m.Diffnum,
				Score:   m.Score,
				Acc:     m.Acc,
				Rank:    m.Rank,
				Rksm:    m.Rksm})
		}

		err = db.FindFor(struid, &r, "ORDER BY Rksm DESC", func() error {
			if len(list) < 22 {
				list = append(list, r)
				return nil
			}
			return nil
		})
		if err != nil {
			ctx.SendChain(message.Text("emm...看起来你好像还没添加过数据?"))
			return
		}
		for i := len(list); i < 22; i++ {
			list = append(list, result{})
		}
		var allrks float64
		for i := 0; i < 20; i++ {
			allrks += list[i].Rksm
		}
		err = renderb19(plname, strconv.FormatFloat(allrks/20, 'f', 3, 64), chal, chalnum, struid, list)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + filepath + struid + "/output.png"))
	})
	en.OnRegex(`^/phi set (.*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()
		name := ctx.State["regex_matched"].([]string)[1]
		uid := ctx.Event.UserID
		struid := strconv.FormatInt(uid, 10)
		d := &data{
			UID:      uid,
			Gamename: name,
		}
		err := db.Insert("gamedata", d)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已成功绑定", struid, "的游戏名字为:", name))
	})
	en.OnRegex(`^/phi chall (.*) ([0-9]{1,2})`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()
		uid := ctx.Event.UserID
		chall := ctx.State["regex_matched"].([]string)[1]
		challnum := math.Str2Int64(ctx.State["regex_matched"].([]string)[2])
		_, ok := checkchall[chall]
		if !ok {
			ctx.SendChain(message.Text("输入的等级有误, 请重新输入"))
			return
		}
		if challnum > 48 {
			ctx.SendChain(message.Text("你是什么yyw"))
			return
		}
		if challnum < 3 {
			ctx.SendChain(message.Text("最低都有3"))
			return
		}
		c := &challen{
			UID:      uid,
			Chall:    chall,
			Challnum: challnum,
		}
		err := db.Insert("challen", c)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已设置课题模式等级为", chall, " ", challnum))
	})
	en.OnRegex(`^/phi add (.*) ([a-z|A-Z]{2}) ([0-9]{2,3}\.?([0-9]{1,2})?) ([0-9]{6,7}) ?([0,1])?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()
		songname := ctx.State["regex_matched"].([]string)[1]
		var sd songdata
		err := sdb.Find("songdata", &sd, "WHERE Name LIKE '"+songname+"%' OR ATName LIKE '"+songname+"%'")
		if err != nil {
			ctx.SendChain(message.Text("未找到该歌曲\nERROR: ", err))
			return
		}
		diff := ctx.State["regex_matched"].([]string)[2]
		_, ok := checkdiff[diff]
		var tdiff float64
		if ok {
			switch diff {
			case "AT":
				tdiff = sd.AT
			case "IN":
				tdiff = sd.IN
			case "HD":
				tdiff = sd.HD
			case "EZ":
				tdiff = sd.EZ
			}
		}
		if tdiff == 0 {
			ctx.SendChain(message.Text("未找到该歌曲所对应的等级"))
			return
		}
		acc := ctx.State["regex_matched"].([]string)[3]
		score := ctx.State["regex_matched"].([]string)[5]
		ac := ctx.State["regex_matched"].([]string)[6]
		scoreint := math.Str2Int64(score)
		if scoreint > 1000000 {
			ctx.SendChain(message.Text("这是什么分数啊, 理论值是吧"))
			return
		}
		accfloat, _ := strconv.ParseFloat(acc, 64)
		rksm := rksc(accfloat, tdiff)
		tac, _ := strconv.ParseBool(ac)
		var rank string
		if tac && score != "1000000" {
			rank = "fc"
		} else {
			rank = checkrank(scoreint)
		}
		r := &result{
			ID:       idof(sd.Name, strconv.FormatFloat(tdiff, 'f', 1, 64)),
			Songname: sd.Name,
			Diff:     diff,
			Diffnum:  tdiff,
			Score:    scoreint,
			Acc:      accfloat,
			Rank:     rank,
			Rksm:     rksm,
		}
		uid := ctx.Event.UserID
		struid := strconv.FormatInt(uid, 10)
		err = db.Insert(struid, r)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已设置", songname, "为", score, " ", acc))
	})

}
