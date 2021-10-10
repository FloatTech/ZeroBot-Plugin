// Package web 网页管理后端
package web

import "flag"

func init() {
	var en bool
	// 解析命令行参数，输入 `-g` 即可启用 gui
	flag.BoolVar(&en, "g", false, "Enable web gui.")
	if en {
		initGui()
	}
}
