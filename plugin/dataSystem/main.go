// Package datasystem 公用数据管理
package datasystem

import (
	"os"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	helpInfo []string
	engine   = control.Register("DataSystem", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "公用数据管理",
		Help:             strings.Join(helpInfo, "\n\n"),
	})
	cachePath = "data/wallet/cache/"
)

func init() {
	go func() {
		_ = os.RemoveAll(cachePath)
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
	}()
}
