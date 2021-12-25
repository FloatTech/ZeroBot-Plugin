package juejuezi

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	juejueziURL = "https://www.offjuan.com/api/juejuezi/text"
	prio        = 15
	referer     = "https://juejuezi.offjuan.com/"
	ua          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
)

var (
	engine = control.Register("juejuezi", &control.Options{
		DisableOnDefault: false,
		Help: "绝绝子生成器\n" +
			"- 喝奶茶绝绝子|吃饭绝绝子",
	})
	limit = rate.NewManager(time.Minute, 20)
)

func init() {
	engine.OnRegex("[\u4E00-\u9FA5]{0,10}绝绝子[\u4E00-\u9FA5]{0,10}").SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		if !limit.Load(ctx.Event.GroupID).Acquire() {
			return
		}
		toDealStr := []rune(strings.Replace(ctx.ExtractPlainText(), "绝绝子", "", -1))
		if len(toDealStr) < 2 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("不要只输入绝绝子"))
		} else if len(toDealStr) == 2 {
			data, err := juejuezi(string(toDealStr[0]), string(toDealStr[1]))
			if err != nil {
				ctx.SendChain(message.Text(err))
			}
			ctx.SendChain(message.Text(gjson.Get(helper.BytesToString(data), "text").String()))
		} else {
			params := ctx.GetWordSlices(string(toDealStr)).Get("slices").Array()
			data, err := juejuezi(params[0].String(), params[1].String())
			if err != nil {
				ctx.SendChain(message.Text(err))
			}
			ctx.SendChain(message.Text(gjson.Get(helper.BytesToString(data), "text").String()))
		}
	})
}

func juejuezi(verb, noun string) (data []byte, err error) {
	juejueziStr := fmt.Sprintf("{\"verb\":\"%s\",\"noun\":\"%s\"}", verb, noun)
	client := &http.Client{}
	// 提交请求
	request, err := http.NewRequest("POST", juejueziURL, strings.NewReader(juejueziStr))
	if err != nil {
		log.Errorln("[juejuezi]:", err)
	}
	request.Header.Add("Referer", referer)
	request.Header.Add("User-Agent", ua)
	response, err := client.Do(request)
	if err != nil {
		log.Errorln("[juejuezi]:", err)
	}
	data, err = io.ReadAll(response.Body)
	response.Body.Close()
	return
}
