package plugin_qqinfo

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/web"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

var (
	engine = control.Register("qqinfo", &control.Options{
		DisableOnDefault: false,
		Help: "qq信息\n" +
			"- qq信息[@xxx]|qq信息[qq号]\n",
	})
	limit = rate.NewManager(time.Minute, 20)
	ua    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
	qqReg = `\d+`
)

func init() {
	engine.OnPrefix("qq信息").SetBlock(true).FirstPriority().Handle(func(ctx *zero.Ctx) {
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
		sex := si.Get("sex").String()
		if sex == "male" {
			sex = "男"
		} else if sex == "female" {
			sex = "女"
		} else if sex == "unknown" {
			sex = "未知"
		}
		age := si.Get("age").Int()
		text = fmt.Sprintf("%s的qq为%d，性别为%s,年龄为%d", nickname, uid, sex, age)
		data, err := web.ReqWith(getQQURL(uid), "GET", "", ua)
		if err != nil {
			log.Println("err为:", err)
		}
		status := gjson.Get(helper.BytesToString(data), "status").Int()
		if status == 200 {
			phone := gjson.Get(helper.BytesToString(data), "phone").Int()
			phonediqu := gjson.Get(helper.BytesToString(data), "phonediqu").String()
			text = text + fmt.Sprintf("，手机号为%d,手机运营商为%s", phone, phonediqu)
		}
		ctx.SendChain(message.Text(text))
	})
}

func getQQURL(uid int64) (qqURL string) {
	qqURL = fmt.Sprintf("https://zy.xywlapi.cc/qqapi?qq=%d", uid)
	return
}
