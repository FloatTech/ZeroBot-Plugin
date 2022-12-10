// Package kanban 打印版本信息
package kanban

import (
	"fmt"
	"sync"

	"github.com/FloatTech/zbputils/control"
	"github.com/fumiama/go-registry"
)

//go:generate go run github.com/FloatTech/ZeroBot-Plugin/kanban/gen

var once sync.Once

func init() {
	once.Do(PrintBanner)
}

var reg = registry.NewRegReader("reilia.fumiama.top:32664", control.Md5File, "fumiama")

// PrintBanner ...
func PrintBanner() {
	fmt.Print(
		"\n======================[ZeroBot-Plugin]======================",
		"\n", Banner, "\n",
		"----------------------[ZeroBot-公告栏]----------------------",
		"\n", Kanban(), "\n",
		"============================================================\n\n",
	)
}

// Kanban ...
func Kanban() string {
	err := reg.Connect()
	if err != nil {
		return err.Error()
	}
	defer reg.Close()
	text, err := reg.Get("ZeroBot-Plugin/kanban")
	if err != nil {
		return err.Error()
	}
	return text
}
