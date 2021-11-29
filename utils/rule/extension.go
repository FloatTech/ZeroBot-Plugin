package rule

import zero "github.com/wdvxdr1123/ZeroBot"

// FirstValueInList 判断正则匹配的第一个参数是否在列表中
func FirstValueInList(list []string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		first := ctx.State["regex_matched"].([]string)[1]
		for i := range list {
			if first == list[i] {
				return true
			}
		}
		return false
	}
}
