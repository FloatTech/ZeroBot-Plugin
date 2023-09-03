// Package driftbottle 漂流瓶
package driftbottle

import (
	"fmt"
	"hash/crc64"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/FloatTech/floatbox/binary"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type sea struct {
	ID   int64  `db:"id"`   // ID qq_grp_name_msg 的 crc64 hashCheck.
	QQ   int64  `db:"qq"`   // Get current user(Who sends this)
	Name string `db:"Name"` //  his or her name at that time:P
	Msg  string `db:"msg"`  // What he or she sent to bot?
	Grp  int64  `db:"grp"`  // which group sends this msg?
	Time string `db:"time"` // we need to know the current time,master>
}

var seaSide = &sql.Sqlite{}
var seaLocker sync.RWMutex

// We need a container to inject what we need :(

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "漂流瓶",
		Help:              "- @bot pick" + "- @bot throw xxx (xxx为投递内容)",
		PrivateDataFolder: "driftbottle",
	})
	seaSide.DBPath = en.DataFolder() + "sea.db"
	err := seaSide.Open(time.Hour)
	if err != nil {
		panic(err)
	}

	_ = createChannel(seaSide)
	en.OnFullMatch("pick", zero.OnlyToMe, zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		be, err := fetchBottle(seaSide)
		if err != nil {
			ctx.SendChain(message.Text("ERR:", err))
		}
		idstr := strconv.Itoa(int(be.ID))
		qqstr := strconv.Itoa(int(be.QQ))
		grpstr := strconv.Itoa(int(be.Grp))
		botname := zero.BotConfig.NickName[0]
		msg := message.Message{message.CustomNode(botname, ctx.Event.SelfID, botname+"试着帮你捞出来了这个~\nID:"+idstr+"\n投递人: "+be.Name+"("+qqstr+")"+"\n群号: "+grpstr+"\n时间: "+be.Time+"\n内容: \n"+be.Msg)}
		ctx.Send(msg)
	})

	en.OnRegex(`throw.*?(.*)`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		senderFormatTime := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
		rawSenderMessage := ctx.State["regex_matched"].([]string)[1]
		rawMessageCallBack := message.UnescapeCQCodeText(rawSenderMessage)
		keyWordsNum := utf8.RuneCountInString(rawMessageCallBack)
		if keyWordsNum < 10 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("需要投递的内容过少( "))
			return
		}
		// check current needs and prepare to throw drift_bottle.
		err = globalbottle(
			ctx.Event.UserID,
			ctx.Event.GroupID,
			senderFormatTime,
			ctx.CardOrNickName(ctx.Event.UserID),
			rawMessageCallBack,
		).throw(seaSide)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("已经帮你丢出去了哦~")))
	})
}

func globalbottle(qq, grp int64, time, name, msg string) *sea { // Check as if the User is available and collect information to store.
	id := int64(crc64.Checksum(binary.StringToBytes(fmt.Sprintf("%d_%d_%s_%s_%s", grp, qq, time, name, msg)), crc64.MakeTable(crc64.ISO)))
	return &sea{ID: id, Grp: grp, Time: time, QQ: qq, Name: name, Msg: msg}
}

func (be *sea) throw(db *sql.Sqlite) error {
	seaLocker.Lock()
	defer seaLocker.Unlock()
	return db.Insert("global", be)
}

func fetchBottle(db *sql.Sqlite) (*sea, error) {
	seaLocker.Lock()
	defer seaLocker.Unlock()
	be := new(sea)
	return be, db.Pick("global", be)
}

func createChannel(db *sql.Sqlite) error {
	seaLocker.Lock()
	defer seaLocker.Unlock()
	return db.Create("global", &sea{})
}
