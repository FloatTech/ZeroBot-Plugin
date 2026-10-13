// Package qqvoice sends QQ's native AI voice directly into the current group.
package qqvoice

import (
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// Deliver never calls get_ai_record: on NapCat that API also posts into a group.
// A timeout/failure must not be retried because the first call may have sent.
func Deliver(text, modelID string, groupID int64, call func(string, zero.Params) zero.APIResponse) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "请在“发送AI语音”后填写要朗读的文字。"
	}
	if strings.TrimSpace(modelID) == "" {
		return "尚未选择AI语音声音，请主人在私聊发送“设置AI语音模型”。"
	}
	if groupID == 0 {
		return "QQ内置AI语音目前仅支持群聊，请在群内发送该命令；私聊内容不会发送到生成群。"
	}
	if groupID < 0 {
		return "AI语音群号无效。"
	}
	response := call("send_group_ai_record", zero.Params{"character": modelID, "group_id": groupID, "text": text})
	// NapCat v4.18.28 returns message_id=0 even after successful delivery.
	if response.Status == "ok" && response.RetCode == 0 {
		return ""
	}
	return "AI语音发送失败或回执未确认，未自动重试；请先检查群中是否已收到语音。"
}
