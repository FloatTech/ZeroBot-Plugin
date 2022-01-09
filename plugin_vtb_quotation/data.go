package vtbquotation

import (
	"os"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

// 加载数据库
func init() {
	go func() {
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		_, _ = file.GetLazyData(dbfile, false, false)
	}()
}
