package modules

import (
	"gm/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func GetInt(state zero.State, index int64) int64 {
	return utils.Str2Int(state["regex_matched"].([]string)[index])
}

func GetStr(state zero.State, index int64) string {
	return state["regex_matched"].([]string)[index]
}

func GetNickname(groupID int64, userID int64) string {
	return zero.GetGroupMemberInfo(groupID, userID, false).Get("nickname").Str
}
