// Package dyloader 动态插件加载器
package dyloader

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/dyloader/plugin"
)

var typeIsSo bool
var visited bool
var pluginsmap = make(map[string]*plugin.Plugin)

func init() {
	zero.OnCommand("刷新插件", zero.SuperUserPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			err := scan()
			if err != nil {
				ctx.SendChain(message.Text("Error: " + err.Error()))
			} else {
				ctx.SendChain(message.Text("成功!"))
			}
		})
	zero.OnCommand("卸载插件", zero.SuperUserPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			model := extension.CommandModel{}
			_ = ctx.Parse(&model)
			_, ok := control.Lookup(model.Args)
			if ok {
				t := ".dll"
				if typeIsSo {
					t = ".so"
				}
				target := "plugin_" + model.Args + t
				logrus.Debugln("[dyloader] target:", target)
				p, ok := pluginsmap[target]
				if ok {
					err := plugin.Close(p)
					control.Delete(model.Args)
					delete(pluginsmap, target)
					if err != nil {
						ctx.SendChain(message.Text("Error: " + err.Error()))
					} else {
						ctx.SendChain(message.Text("成功!"))
					}
				} else {
					ctx.SendChain(message.Text("没有这个插件!"))
				}
			}
		})
	go func() {
		time.Sleep(time.Second * 2)
		_ = scan()
	}()
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
	if !visited {
		if strings.HasSuffix(n, ".so") {
			typeIsSo = true
			visited = true
		} else if strings.HasSuffix(n, ".dll") {
			visited = true
		}
	}
	if strings.HasSuffix(n, ".so") || strings.HasSuffix(n, ".dll") {
		target := path[strings.LastIndex(path, "/")+1:]
		logrus.Debugln("[dyloader] target:", target)
		_, ok := pluginsmap[target]
		if !ok {
			p, err := plugin.Open(path)
			if err == nil {
				logrus.Infoln("[dyloader]加载插件", path, "成功")
				pluginsmap[target] = p
			}
			if err != nil {
				logrus.Errorln("[dyloader]加载插件", path, "错误:", err)
			}
		}
	}
	return nil
}
