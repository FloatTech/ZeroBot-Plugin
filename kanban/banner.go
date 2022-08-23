package kanban

import (
	"fmt"
	"strings"

	"github.com/fumiama/go-registry"
)

var (
	info = [...]string{
		"* OneBot + ZeroBot + Golang",
		"* Version 1.5.0 - 2022-08-16 11:16:21 +0800 CST",
		"* Copyright © 2020 - 2022 FloatTech. All Rights Reserved.",
		"* Project: https://github.com/FloatTech/ZeroBot-Plugin",
	}
	// Banner ...
	Banner = strings.Join(info[:], "\n")
	reg    = registry.NewRegReader("reilia.fumiama.top:32664", "fumiama")
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
