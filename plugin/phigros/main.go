package phigros

import (
	"strconv"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	//"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	dbnum int
	db    = &sql.Sqlite{}
	en    = control.Register("phigros", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "",
		PublicDataFolder: "Phigros",
	}) //.ApplySingle(ctxext.DefaultSingle)
	filepath = en.DataFolder()
)

func init() {
	db.DBPath = en.DataFolder() + "data.db"
	go func() {

		err := db.Open(time.Hour * 24)
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
		err = db.Create("songdata", &songdata{})
		if err != nil {
			panic(err)
		}
	}()
	en.OnFullMatch("/phi init").SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
		uid := ctx.Event.UserID
		struid := strconv.FormatInt(uid, 10)
		var d data
		err := db.Find("gamedata", &d, "WHERE UID = "+struid)
		if err != nil {
			ctx.SendChain(message.Text("看来你还没有绑定过呢"))
			return
		}
		plname := d.Gamename
		var c challen
		var chal, chalnum string
		err = db.Find("challen", &c, "WHERE UID = "+struid)
		if err != nil {
			chal, chalnum = "无", "0"
		}
		chal, chalnum = c.Chall, strconv.FormatInt(c.Challnum, 10)
		var list = make([]result, 0, 40)
		var r result
		err = db.Find(struid, &r, "WHERE Rank = 'phi'")
		if err != nil {
			list = append(list, result{Songname: "",
				Diff:    "",
				Diffnum: 0,
				Score:   0,
				Acc:     0,
				Rank:    "",
				Rksm:    0})
		} else {
			list = append(list, r)
		}

		dbnum, err = db.Count(struid)
		if err != nil || dbnum == 0 {
			ctx.SendChain(message.Text("emm...看起来你好像还没添加过数据?"))
			return
		}

		err = db.FindFor(struid, &r, "ORDER BY Rksm DESC", func() error {
			list = append(list, r)
			return nil
		})
		for i := len(list); i < 20; i++ {
			list = append(list, result{Songname: "",
				Diff:    "",
				Diffnum: 0,
				Score:   0,
				Acc:     0,
				Rank:    "",
				Rksm:    0})
		}
		var arks float64
		for i := 0; i < 20; i++ {
			arks = +list[i].Rksm
		}
		err = renderb19(plname, strconv.FormatFloat(arks, 'f', 3, 64), chal, chalnum, list)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + filepath + "output.png"))
	})
	en.OnRegex(`^/phi set (.*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^/phi add (.*) ([a-z|A-Z]{2}) ([0-9]{2,3}\.?([0-9]{2})?) ([0-9]{6,7}) ?([0,1])?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		songname := ctx.State["regex_matched"].([]string)[1]
		var sd songdata
		err := db.Find("songdata", &sd, "WHERE Name  LIKE '%"+songname+"%'")
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
		if tac {
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
		ctx.SendChain(message.Text("存储成功!"))
	})

}
