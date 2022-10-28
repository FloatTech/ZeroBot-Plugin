package huggingface

import (
	"encoding/json"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	magicpromptRepo = "/Gustavosta/MagicPrompt-Stable-Diffusion"
)

func init() { // 插件主体
	engine := control.Register("magicprompt", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "MagicPrompt-Stable-Diffusion吟唱提示\n" +
			"- 吟唱提示 xxx",
		PrivateDataFolder: "magicprompt",
	})

	// 开启
	engine.OnPrefixGroup([]string{`吟唱提示`, "吟唱补全"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			magicpromptURL := huggingfaceSpaceWss + magicpromptRepo + joinPath
			args := ctx.State["args"].(string)
			c, _, err := websocket.DefaultDialer.Dial(magicpromptURL, nil)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			defer c.Close()

			done := make(chan struct{})

			go func() {
				defer close(done)
				for {
					_, data, err := c.ReadMessage()
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					j := gjson.ParseBytes(data)
					if j.Get("msg").String() == wssCompleteStatus {
						m := message.Message{}
						for _, v := range strings.Split(j.Get("output.data.0").String(), "\n\n") {
							m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(v)))
						}
						if id := ctx.Send(m).ID(); id == 0 {
							ctx.SendChain(message.Text("ERROR:  可能被风控或下载图片用时过长，请耐心等待"))
						}
						return
					}
				}
			}()

			r := pushRequest{
				FnIndex:     0,
				Data:        []interface{}{args},
				SessionHash: defaultSessionHash,
			}
			b, err := json.Marshal(r)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					err := c.WriteMessage(websocket.TextMessage, b)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
				}
			}
		})
}
