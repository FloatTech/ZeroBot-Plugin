// Package magicprompt MagicPrompt-Stable-Diffusion吟唱提示
package magicprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	hf "github.com/FloatTech/AnimeAPI/huggingface"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/RomiChan/websocket"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	magicpromptRepo = "Gustavosta/MagicPrompt-Stable-Diffusion"
)

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "MagicPrompt-Stable-Diffusion吟唱提示",
		Help:              "- 吟唱提示 xxx",
		PrivateDataFolder: "magicprompt",
	})

	// 开启
	engine.OnPrefixGroup([]string{`吟唱提示`, "吟唱补全"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			_ctx, _cancel := context.WithTimeout(context.Background(), hf.TimeoutMax*time.Second)
			defer _cancel()
			ctx.SendChain(message.Text("少女祈祷中..."))

			magicpromptURL := fmt.Sprintf(hf.WssJoinPath, magicpromptRepo)
			args := ctx.State["args"].(string)
			c, _, err := websocket.DefaultDialer.Dial(magicpromptURL, nil)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			defer c.Close()

			r := hf.PushRequest{
				FnIndex: 0,
				Data:    []interface{}{args},
			}
			b, err := json.Marshal(r)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}

			err = c.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			t := time.NewTicker(time.Second * 1)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					_, data, err := c.ReadMessage()
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					j := gjson.ParseBytes(data)
					if j.Get("msg").String() == hf.WssCompleteStatus {
						m := message.Message{}
						for _, v := range strings.Split(j.Get("output.data.0").String(), "\n\n") {
							m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(v)))
						}
						if id := ctx.Send(m).ID(); id == 0 {
							ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
						}
						return
					}
				case <-_ctx.Done():
					ctx.SendChain(message.Text("ERROR: 吟唱提示指令超时"))
					return
				}
			}
		})
}
