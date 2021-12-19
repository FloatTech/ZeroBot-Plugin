// Package rule zb 规则扩展
package rule

import zero "github.com/wdvxdr1123/ZeroBot"

// FirstValueInList 判断正则匹配的第一个参数是否在列表中
func FirstValueInList(list []string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		first := ctx.State["regex_matched"].([]string)[1]
		for _, v := range list {
			if first == v {
				return true
			}
		}
		return false
	}
}
