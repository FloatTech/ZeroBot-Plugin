package qqvoice

import (
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func TestNativeVoiceSideEffectsAndAcknowledgement(t *testing.T) {
	for _, c := range []struct {
		name, text, model, status, data, want string
		group, retcode                        int64
		calls                                 int
	}{
		{"native zero ID success", "  中文🙂  ", "voice", "ok", `{"message_id":0}`, "", 627, 0, 1},
		{"positive ID success", "中文🙂", "voice", "ok", `{"message_id":123}`, "", 627, 0, 1},
		{"no ID success", "中文🙂", "voice", "ok", `{}`, "", 627, 0, 1},
		{"failed status despite ID", "中文🙂", "voice", "failed", `{"message_id":123}`, "未自动重试", 627, 0, 1},
		{"nonzero retcode", "中文🙂", "voice", "ok", `{}`, "未自动重试", 627, 1400, 1},
		{"timeout may have sent", "中文🙂", "voice", "", `{}`, "未自动重试", 627, 0, 1},
		{"async unknown receipt", "中文🙂", "voice", "async", `{}`, "未自动重试", 627, 1, 1},
		{"private never generates in group", "中文🙂", "voice", "ok", `{}`, "仅支持群聊", 0, 0, 0},
		{"negative group", "中文🙂", "voice", "ok", `{}`, "无效", -627, 0, 0},
		{"empty text", " \t ", "voice", "ok", `{}`, "填写", 627, 0, 0},
		{"missing voice", "中文🙂", "", "ok", `{}`, "未选择", 627, 0, 0},
		{"blank voice", "中文🙂", " \t", "ok", `{}`, "未选择", 627, 0, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			calls := 0
			feedback := Deliver(c.text, c.model, c.group, func(action string, params zero.Params) zero.APIResponse {
				calls++
				if action != "send_group_ai_record" || len(params) != 3 || params["group_id"] != c.group || params["character"] != c.model || params["text"] != "中文🙂" {
					t.Fatal("unexpected voice generation, forwarding, or destination")
				}
				return zero.APIResponse{Status: c.status, RetCode: c.retcode, Data: gjson.Parse(c.data)}
			})
			if calls != c.calls || (c.want == "" && feedback != "") || (c.want != "" && !strings.Contains(feedback, c.want)) {
				t.Fatalf("calls=%d feedback=%q", calls, feedback)
			}
		})
	}
}
