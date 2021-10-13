// Package dyloader 动态插件加载器
package dyloader

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/dyloader/plugin"
)

func init() {
	_ = scan()
	zero.OnCommand("刷新插件", zero.SuperUserPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			err := scan()
			if err != nil {
				ctx.SendChain(message.Text("Error: " + err.Error()))
			}
		})
}

func scan() error {
	return filepath.WalkDir("plugins/", load)
}

func load(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		return nil
	}
	n := d.Name()
	if strings.HasSuffix(n, ".so") || strings.HasSuffix(n, ".dll") {
		_, err = plugin.Open(path)
		if err == nil {
			logrus.Infoln("[dyloader]加载插件", path, "成功")
		}
		if err != nil {
			logrus.Errorln("[dyloader]加载插件", path, "错误:", err)
		}
	}
	return nil
}
