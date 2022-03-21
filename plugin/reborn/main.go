// Package reborn 投胎 来自 https://github.com/YukariChiba/tgbot/blob/main/modules/Reborn.py
package reborn

import (
	"fmt"
	"math/rand"

	control "github.com/FloatTech/zbputils/control"
	wr "github.com/mroth/weightedrand"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

func init() {
	en := control.Register("reborn", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "投胎\n- reborn",
		PublicDataFolder: "Reborn",
	})
	go func() {
		datapath := en.DataFolder()
		jsonfile := datapath + "rate.json"
		area := make(rate, 226)
		err := load(&area, jsonfile)
		if err != nil {
			panic(err)
		}
		choices := make([]wr.Choice, len(area))
		for i, a := range area {
			choices[i].Item = a.Name
			choices[i].Weight = uint(a.Weight * 1e9)
		}
		areac, err = wr.NewChooser(choices...)
		if err != nil {
			panic(err)
		}
		logrus.Printf("[Reborn]读取%d个国家/地区", len(area))
	}()
	en.OnFullMatch("reborn").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if rand.Int31() > 1<<27 {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("投胎成功！\n您出生在 %s, 是 %s。", randcoun(), randgen())))
			} else {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("投胎失败！\n您没能活到出生，祝您下次好运！"))
			}
		})
}
