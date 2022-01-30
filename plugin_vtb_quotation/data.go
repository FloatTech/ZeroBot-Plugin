package vtbquotation

import (
	"os"

	"github.com/FloatTech/ZeroBot-Plugin/order"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
)

// 加载数据库
func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		_, _ = file.GetLazyData(dbfile, false, false)
	}()
}
