// Package vitsnyaru vits猫雷
package vitsnyaru

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	hf "github.com/FloatTech/AnimeAPI/huggingface"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	vitsnyaruRepo = "innnky/vits-nyaru"
)

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "vits猫雷",
		Help:              "- 让猫雷说 xxx",
		PrivateDataFolder: "vitsnyaru",
	})

	// 开启
	engine.OnPrefix(`让猫雷说`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			_ctx, _cancel := context.WithTimeout(context.Background(), hf.TimeoutMax*time.Second)
			defer _cancel()
			ch := make(chan []byte, 1)

			args := ctx.State["args"].(string)
			pushURL := fmt.Sprintf(hf.HTTPSPushPath, vitsnyaruRepo)
			statusURL := fmt.Sprintf(hf.HTTPSStatusPath, vitsnyaruRepo)
			ctx.SendChain(message.Text("少女祈祷中..."))
			var (
				pushReq   hf.PushRequest
				pushRes   hf.PushResponse
				statusReq hf.StatusRequest
				statusRes hf.StatusResponse
				data      []byte
			)

			// 获取clean后的文本
			pushReq = hf.PushRequest{
				Action:  hf.DefaultAction,
				Data:    []interface{}{args},
				FnIndex: 1,
			}
			pushRes, err := hf.Push(pushURL, pushReq)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			statusReq = hf.StatusRequest{
				Hash: pushRes.Hash,
			}

			t := time.NewTicker(time.Second * 1)
			defer t.Stop()
		LOOP:
			for {
				select {
				case <-t.C:
					data, err = hf.Status(statusURL, statusReq)
					if err != nil {
						ch <- data
						break LOOP
					}
					if gjson.ParseBytes(data).Get("status").String() == hf.CompleteStatus {
						ch <- data
						break LOOP
					}
				case <-_ctx.Done():
					ch <- data
					break LOOP
				}
			}

			data = <-ch
			err = json.Unmarshal(data, &statusRes)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}

			// 用clean的文本预测语音
			pushReq = hf.PushRequest{
				Action:  hf.DefaultAction,
				Data:    statusRes.Data.Data,
				FnIndex: 2,
			}
			pushRes, err = hf.Push(pushURL, pushReq)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			statusReq = hf.StatusRequest{
				Hash: pushRes.Hash,
			}

		LOOP2:
			for {
				select {
				case <-t.C:
					data, err = hf.Status(statusURL, statusReq)
					if err != nil {
						ch <- data
						break LOOP2
					}
					if gjson.ParseBytes(data).Get("status").String() == hf.CompleteStatus {
						ch <- data
						break LOOP2
					}
				case <-_ctx.Done():
					ch <- data
					break LOOP2
				}
			}

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
