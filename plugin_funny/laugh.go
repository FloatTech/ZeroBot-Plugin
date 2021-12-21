package funny

import (
	"io/ioutil"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	qqReg  = `\d+`
	dbpath = "data/funny/"
	dbfile = dbpath + "laugh.txt"
)

var (
	engine = control.Register("curse", &control.Options{
		DisableOnDefault: false,
		Help: "讲个笑话\n" +
			"- 讲个笑话[@xxx]|讲个笑话[qq号]\n",
	})
	limit = rate.NewManager(time.Minute, 20)
)

func init() {
	engine.OnPrefix("讲个笑话").SetBlock(true).FirstPriority().Handle(func(ctx *zero.Ctx) {
		if !limit.Load(ctx.Event.GroupID).Acquire() {
			ctx.SendChain(message.Text("请稍后重试0x0..."))
			return
		}
		var uid int64
		var text string
		reg := regexp.MustCompile(qqReg)
		if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
			uid, _ = strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		} else if reg.MatchString(ctx.Event.RawMessage) {
			result := reg.FindAllString(ctx.Event.RawMessage, -1)
			uid, _ = strconv.ParseInt(result[0], 10, 64)
		} else if uid == 0 {
			uid = ctx.Event.UserID
		}
		si := ctx.GetStrangerInfo(uid, false)
		nickname := si.Get("nickname").String()
		laugh, err := ioutil.ReadFile(dbfile)
		if err != nil {
			log.Println("err:", err)
		}
		laughList := strings.Split(helper.BytesToString(laugh), "\n")
		rand.Seed(time.Now().Unix())
		text = laughList[rand.Intn(len(laughList))]
		text = strings.Replace(text, "%name", nickname, -1)
		ctx.SendChain(message.Text(text))
	})
}
