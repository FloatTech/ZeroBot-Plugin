package kanban

import (
	"fmt"

	"github.com/FloatTech/zbputils/control"
	"github.com/fumiama/go-registry"
)

var (
	// Banner ...
	Banner = "* OneBot + ZeroBot + Golang\n" +
		"* Version 1.6.0-beta1 - 2022-11-15 11:13:42 +0800 CST\n" +
		"* Copyright © 2020 - 2022 FloatTech. All Rights Reserved.\n" +
		"* Project: https://github.com/FloatTech/ZeroBot-Plugin"
	reg = registry.NewRegReader("reilia.fumiama.top:32664", control.Md5File, "fumiama")
)

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
