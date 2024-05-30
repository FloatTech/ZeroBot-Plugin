package pool

import (
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.OnMessage(zero.HasPicture).SetBlock(false).FirstPriority().Handle(func(ctx *zero.Ctx) {
		img, ok := ctx.State["image_url"].([]string)
		if !ok || len(img) == 0 {
			return
		}
		if !ntcachere.MatchString(img[0]) { // is not NTQQ
			return
		}
		rk, err := nturl(img[0]).rkey()
		if err != nil {
			logrus.Debugln("[imgpool] parse rkey error:", err, "image url:", img)
			return
		}
		err = rs.set(time.Minute, rk)
		if err != nil {
			logrus.Debugln("[imgpool] set rkey error:", err, "rkey:", rk)
			return
		}
		logrus.Debugln("[imgpool] set latest rkey:", rk)
	})
}
