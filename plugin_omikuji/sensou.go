// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_omikuji/model"
	"github.com/FloatTech/ZeroBot-Plugin/utils/web"

	log "github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	bed          = "https://codechina.csdn.net/u011570312/senso-ji-omikuji/-/raw/main/%d_%d.jpg"
	dbpath       = "data/omikuji/"
	dbfile       = dbpath + "signature.db"
	url          = "https://services.shen88.cn/chouqian/guanyinlingqian-${index}.html"
	ua           = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36"
	refer        = "https://services.shen88.cn/"
	signatureReg = `(<p><strong>.*?)<p class="center">`
)

var (
	engine = control.Register("omikuji", &control.Options{
		DisableOnDefault: false,
		Help: "浅草寺求签\n" +
			"- 求签|占卜\n- 解签",
	})
)

func init() { // 插件主体
	rand.Seed(time.Now().UnixNano())

	engine.OnFullMatchGroup([]string{"求签", "占卜"}).SetPriority(10).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			miku := int64(rand.Intn(100) + 1)
			userId := ctx.Event.UserID
			db, err := model.Open(dbfile)
			if err != nil {
				log.Errorln(err)
				return
			}
			miku, err = db.GetSignature(userId, miku)
			log.Println("miku为:", miku, "err为:", err)
			ctx.SendChain(
				message.At(userId),
				message.Image(fmt.Sprintf(bed, miku, 0)),
				message.Image(fmt.Sprintf(bed, miku, 1)),
			)
			db.Close()
		})
	engine.OnFullMatchGroup([]string{"解签"}).SetPriority(10).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			miku := int64(rand.Intn(100) + 1)
			userId := ctx.Event.UserID
			db, err := model.Open(dbfile)
			if err != nil {
				log.Errorln(err)
				return
			}
			miku, err = db.GetSignature(userId, miku)

			reqURL := strings.Replace(url, "${index}", strconv.FormatInt(miku, 10), -1)
			data, err := web.ReqWith(reqURL, "GET", refer, ua)
			if err != nil {
				log.Errorln("err为:", err)
				return
			}

			res := helper.BytesToString(data)
			reg := regexp.MustCompile(signatureReg)
			text := reg.FindStringSubmatch(res)[1]

			text = strings.Replace(text, "<p>", "", -1)
			text = strings.Replace(text, "</p>", "\n", -1)
			text = strings.Replace(text, "<strong>", "", -1)
			text = strings.Replace(text, "</strong>", "", -1)
			text = "\n" + text
			ctx.SendChain(
				message.At(userId),
				message.Text(text),
			)
			db.Close()

		})
}
