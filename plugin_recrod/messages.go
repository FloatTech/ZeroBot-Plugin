package record

import (
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"os"
	"strconv"
)

const (
	dbpath = "data/record/"
	dbfile = dbpath + "messages.db"
	hint   = "====消息记录====\n"
)

var engine = control.Register("record", order.AcquirePrio(), &control.Options{
	DisableOnDefault: false,
	Help:             hint,
})

type messages struct {
	gorm.Model
	Mid interface{} `gorm:"column:message_id;type:bigint"`
	QQ  int64       `gorm:"column:qq"`
	Un  string      `gorm:"column:username"`
	Gn  string      `gorm:"column:groupname"`
	Msg string      `gorm:"column:messages;type:varchar(1024)"`
	Ts  int64       `gorm:"column:timestamp;type:timestamp"`
}

func init() {
	engine.OnMessage(zero.OnlyGroup).
		Handle(func(ctx *zero.Ctx) {
			var err error
			if _, err = os.Stat(dbfile); err != nil || os.IsNotExist(err) {
				// 生成文件
				_ = os.MkdirAll(dbpath, 0755)
				f, err := os.Create(dbfile)
				if err != nil {
					logrus.Errorln(err)
				}
				defer f.Close()
			}

			db, err := gorm.Open("sqlite3", dbfile)
			if err != nil {
				logrus.Errorln("打开数据库失败：", err)
			}
			// 制表
			tableName := "groupMessages_" + strconv.FormatInt(ctx.Event.GroupID, 10)
			db.Table(tableName).AutoMigrate(messages{})
			//插入数据
			db.Table(tableName).Create(&messages{
				Mid: ctx.Event.MessageID,
				QQ:  ctx.Event.UserID,
				Un:  ctx.Event.Sender.NickName,
				Gn:  ctx.Event.Sender.Card,
				Msg: ctx.Event.Message.String(),
				Ts:  ctx.Event.Time,
			})
			logrus.Infof("[recrod]消息（%v）插入数据库成功", ctx.Event.GroupID)
			defer db.Close()
		})
}
