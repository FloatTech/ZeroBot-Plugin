package mocking_bird

import (
	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	"github.com/sirupsen/logrus"
	"os"
)

// 加载数据库
func init() {
	go func() {
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		os.RemoveAll(cachePath)
		_ = os.MkdirAll(cachePath, 0755)
		_, err := file.GetLazyData(dbfile, false, true)
		if err != nil {
			panic(err)
		}
		logrus.Infoln("[mocking_bird]加载实例音频")
	}()
}
