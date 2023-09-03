// Package juejuezi 绝绝子
package juejuezi

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	juejueziURL = "https://www.offjuan.com/api/juejuezi/text"
	referer     = "https://juejuezi.offjuan.com/"
	ua          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "绝绝子生成器",
		Help:             "例: 喝奶茶绝绝子\n绝绝子吃饭",
	}).OnRegex("[\u4E00-\u9FA5]{0,10}绝绝子[\u4E00-\u9FA5]{0,10}").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		toDealStr := []rune(strings.ReplaceAll(ctx.ExtractPlainText(), "绝绝子", ""))
		switch len(toDealStr) {
		case 0, 1:
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("不要只输入绝绝子"))
		case 2:
			data, err := juejuezi(string(toDealStr[0]), string(toDealStr[1]))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text(gjson.Get(helper.BytesToString(data), "text").String()))
		default:
			params := ctx.GetWordSlices(string(toDealStr)).Get("slices").Array()
			data, err := juejuezi(params[0].String(), params[1].String())
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
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
		return nil, err
	}
	request.Header.Add("Referer", referer)
	request.Header.Add("User-Agent", ua)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	data, err = io.ReadAll(response.Body)
	response.Body.Close()
	return
}
