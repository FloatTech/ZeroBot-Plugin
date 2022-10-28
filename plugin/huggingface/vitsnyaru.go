package huggingface

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	vitsnyaruRepo = "/innnky/vits-nyaru"
)

func init() { // 插件主体
	engine := control.Register("vitsnyaru", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "vits猫雷\n" +
			"- 让猫雷说 xxx",
		PrivateDataFolder: "vitsnyaru",
	})

	// 开启
	engine.OnPrefix(`让猫雷说`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			_ctx, _cancel := context.WithTimeout(context.Background(), timeoutMax*time.Second)
			defer _cancel()
			ch := make(chan []byte, 1)

			args := ctx.State["args"].(string)
			pushURL := embed + vitsnyaruRepo + pushPath
			statusURL := embed + vitsnyaruRepo + statusPath
			ctx.SendChain(message.Text("少女祈祷中..."))
			var (
				pushReq   pushRequest
				pushRes   pushResponse
				statusReq statusRequest
				statusRes statusResponse
				data      []byte
			)

			// 获取clean后的文本
			pushReq = pushRequest{
				Action:      defaultAction,
				Data:        []interface{}{args},
				FnIndex:     1,
				SessionHash: defaultSessionHash,
			}
			pushRes, err := push(pushURL, pushReq)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			statusReq = statusRequest{
				Hash: pushRes.Hash,
			}
			go func(c context.Context) {
				t := time.NewTicker(time.Second * 1)
				defer t.Stop()
			LOOP:
				for {
					select {
					case <-t.C:
						data, err = status(statusURL, statusReq)
						if err != nil {
							ch <- data
							break LOOP
						}
						if gjson.ParseBytes(data).Get("status").String() == completeStatus {
							ch <- data
							break LOOP
						}
					case <-c.Done():
						ch <- data
						break LOOP
					}
				}
			}(_ctx)
			data = <-ch
			err = json.Unmarshal(data, &statusRes)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}

			// 用clean的文本预测语音
			pushReq = pushRequest{
				Action:      defaultAction,
				Data:        statusRes.Data.Data,
				FnIndex:     2,
				SessionHash: defaultSessionHash,
			}
			pushRes, err = push(pushURL, pushReq)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			statusReq = statusRequest{
				Hash: pushRes.Hash,
			}
			go func(c context.Context) {
				t := time.NewTicker(time.Second * 1)
				defer t.Stop()
			LOOP:
				for {
					select {
					case <-t.C:
						data, err = status(statusURL, statusReq)
						if err != nil {
							ch <- data
							break LOOP
						}
						if gjson.ParseBytes(data).Get("status").String() == completeStatus {
							ch <- data
							break LOOP
						}
					case <-c.Done():
						ch <- data
						break LOOP
					}
				}
			}(_ctx)
			data = <-ch
			err = json.Unmarshal(data, &statusRes)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}

			// 发送语音
			if len(statusRes.Data.Data) < 2 {
				ctx.SendChain(message.Text("ERROR: 未能获取语音"))
				return
			}
			ctx.SendChain(message.Record("base64://" + strings.TrimPrefix(statusRes.Data.Data[1].(string), "data:audio/wav;base64,")))
		})
}
