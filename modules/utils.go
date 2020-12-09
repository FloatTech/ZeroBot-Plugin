package modules

import (
	"fmt"
	"gm/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func GetInt(state zero.State, index int64) int64 {
	fmt.Println(state["regex_matched"].([]string))
	return utils.Str2Int(state["regex_matched"].([]string)[index])
}

func GetStr(state zero.State, index int64) string {
	return state["regex_matched"].([]string)[index]
}
