// Package nbnhhsh 能不能好好说话
package nbnhhsh

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "拼音首字母释义工具",
		Help:             "- ?? [缩写]",
	}).OnRegex(`^[?？]{1,2} ?([a-z0-9]+)$`).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			ctx.SendChain(message.Text(keyword + ": " + strings.Join(getValue(keyword), ", ")))
		})
}

func getValue(text string) []string {
	urlValues := url.Values{}
	urlValues.Add("text", text)
	resp, err := http.PostForm("https://lab.magiconch.com/api/nbnhhsh/guess", urlValues)
	if err == nil {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body.Close()
			json := gjson.ParseBytes(body)
			res := make([]string, 0)
			var jsonPath string
			if json.Get("0.trans").Exists() {
				jsonPath = "0.trans"
			} else {
				jsonPath = "0.inputting"
			}
			for _, value := range json.Get(jsonPath).Array() {
				res = append(res, value.String())
			}
			return res
		}
		return []string{err.Error()}
	}
	return []string{err.Error()}
}
