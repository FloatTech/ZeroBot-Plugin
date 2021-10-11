// Package web 网页管理后端
package web

import "flag"

func init() {
	// 解析命令行参数，输入 `-g` 即可启用 gui
	if *flag.Bool("g", false, "Enable web gui.") {
		initGui()
	}
}
