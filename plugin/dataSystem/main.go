// Package datasystem 公用数据管理
package datasystem

import (
	"os"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	engine = control.Register("DataSystem", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "公用数据管理",
		Help: "----------昵 称 系 统---------\n" +
			"- @bot 叫我[xxx]\n- 注销昵称 [xxx/qq号/@QQ]\n" +
			"----------货 币 系 统---------\n" +
			"- 查看我的钱包\n" +
			"- 查看钱包排名\n" +
			"注:为本群排行，若群人数太多不建议使用该功能!!!\n" +
			"- /钱包 [QQ号|@群友]\n" +
			"- 支付 [QQ号|@群友] ATRI币值\n" +
			"- /记录 @群友 ATRI币值\n" +
			"- /记录 @加分群友 @减分群友 ATRI币值",
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
