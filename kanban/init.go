// Package kanban 打印版本信息
package kanban

import (
	"fmt"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/kanban/banner"
)

//go:generate go run github.com/FloatTech/ZeroBot-Plugin/kanban/gen

func init() {
	PrintBanner()
}

// PrintBanner ...
func PrintBanner() {
	fmt.Print(
		"\n======================[ZeroBot-Plugin]======================",
		"\n", banner.Banner, "\n",
		"----------------------[ZeroBot-公告栏]----------------------",
		"\n", Kanban(), "\n",
		"============================================================\n\n",
	)
}

// Kanban ...
func Kanban() string {
	text, err := fetchKanban("reilia.fumiama.top:32664", 5*time.Second)
	if err != nil {
		return "公告暂不可用: " + err.Error()
	}
	return text
}
