// Package kokomi 原神面板查询
package kokomi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	api     = "http://8.134.179.136/"
	genshin = "genshin/"
)

func init() {
	en := control.Register("kokomi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "原神面板查询",
		Help:             "- 绑定xxx\n",
	})
	en.OnRegex(`^(?:#|＃)?\s*绑定+?\s*(?:uid|UID|Uid)?\s*(\d+)?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		suid := ctx.State["regex_matched"].([]string)[1] // 获取uid
		body, err := getData(api + genshin + "bound?qq=" + strconv.Itoa(int(ctx.Event.UserID)) + "&uid=" + suid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", string(body), err))
			return
		}
		ctx.SendChain(message.Text(string(body)))
	})
	en.OnRegex(`^(?:#|＃)?(.*)面板\s*(?:(?:\[CQ:at,qq=)(\d+))?(\d+)?(.*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var i string
		str := ctx.State["regex_matched"].([]string)[1] // 获取key
		if str == "" {
			str = ctx.State["regex_matched"].([]string)[4]
		}
		if ctx.State["regex_matched"].([]string)[3] == "" {
			if i = ctx.State["regex_matched"].([]string)[2]; i == "" {
				i = strconv.FormatInt(ctx.Event.UserID, 10)
			}
			if str == "更新" {
				body, err := getData(api + genshin + "find?qq=" + i)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", string(body), err))
					return
				}
				ctx.SendChain(message.Text(string(body)))
			} else {
				body, err := getData(api + genshin + "qtop?qq=" + i + "&role=" + str)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", string(body), err))
					return
				}
				ctx.SendChain(message.ImageBytes(body))
			}
			return
		}
		i = ctx.State["regex_matched"].([]string)[3]
		if str == "更新" {
			body, err := getData(api + genshin + "find?uid=" + i)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", string(body), err))
				return
			}
			ctx.SendChain(message.Text(string(body)))
			return
		}
		body, err := getData(api + genshin + "utop?uid=" + i + "&role=" + str)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", string(body), err))
			return
		}
		ctx.SendChain(message.ImageBytes(body))
	})
	en.OnRegex(`^(?:#|＃)?\s*更新+?\s*(?:uid|UID|Uid)?\s*(\d+)?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		suid := ctx.State["regex_matched"].([]string)[1] // 获取uid
		body, err := getData(api + genshin + "find?uid=" + suid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", string(body), err))
			return
		}
		ctx.SendChain(message.Text(string(body)))
	})
}

// GetData 获取数据
func getData(url string) (data []byte, err error) {
	var response *http.Response
	response, err = http.Get(url)
	if err == nil {
		if response.StatusCode != http.StatusOK {
			s := fmt.Sprintf("\ncode: %d", response.StatusCode)
			err = errors.New(s)
		}
		data, _ = io.ReadAll(response.Body)
		response.Body.Close()
	}
	return
}
