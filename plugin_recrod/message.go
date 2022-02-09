package plugin_record

import (
	"github.com/FloatTech/sqlite"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/process"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"os"
	"strconv"
	"time"
)

const (
	datapath = "data/record/"
	confile  = datapath + "messages.db"
	hint     = "====消息记录====\n"
)

type msg struct {
	MsgID interface{} `db:"messages_id"`
	QQ    int64       `db:"qq"`
	Un    string      `db:"username"`
	Msg   string      `db:"messages"`
	Ts    int64       `db:"timestamp"`
	Tm    string      `db:"datetime"`
}

var (
	db = &sql.Sqlite{DBPath: confile}
)

var engine = control.Register("record", order.AcquirePrio(), &control.Options{
	DisableOnDefault: false,
	Help:             hint,
})

func init() { // 插件主体
	go func() {
		engine.OnMessage(zero.OnlyGroup).
			Handle(func(ctx *zero.Ctx) {
				process.SleepAbout1sTo2s()
				_ = os.MkdirAll(datapath, 0755)
				tablename := "groupMessages_" + strconv.FormatInt(ctx.Event.GroupID, 10)
				err := db.Create(tablename, &msg{})
				if err != nil {
					log.Errorln("数据库创建失败", err)
				}

				db.Insert(tablename, &msg{
					MsgID: ctx.Event.MessageID,
					QQ:    ctx.Event.UserID,
					Un:    ctx.Event.Sender.NickName,
					Msg:   ctx.Event.Message.String(),
					Ts:    ctx.Event.Time,
					Tm:    time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05"),
				})

				log.Infof("消息ID(%v)插入数据库成功", ctx.Event.MessageID)
			})
	}()
}
