package keywordban

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

var (
	matchers = map[int64]*zero.Matcher{}
	mu       = sync.RWMutex
	key      = control.Register("keywordban", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  true,
		Help:              "关键字检测",
		PrivateDataFolder: "keywordban",
	})
	filepath = key.DataFolder()
)

func init() {

}
