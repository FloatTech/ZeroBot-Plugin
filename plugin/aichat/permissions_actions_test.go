package aichat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FloatTech/zbputils/chat"
	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	goba "github.com/fumiama/go-onebot-agent"
)

type permissionMemory struct{}

func (permissionMemory) Load(int64) []string      { return nil }
func (permissionMemory) Save(int64, string) error { return nil }

// Model output is deliberately hostile in rejected cases. The real Agent
// parser and action checks must reject it before any OneBot action can execute.
func TestAgentRoleActionBoundaries(t *testing.T) {
	for _, c := range []struct {
		name, action, params string
		role                 goba.PermRole
		group                int64
		sudo, allow          bool
	}{
		{"member sends here", "send_group_msg", `{"group_id":101,"message":"hello"}`, goba.PermRoleUser, 101, false, true},
		{"member cannot send elsewhere", "send_group_msg", `{"group_id":202,"message":"hello"}`, goba.PermRoleUser, 101, false, false},
		{"member cannot ban", "set_group_ban", `{"group_id":101,"user_id":123,"duration":1}`, goba.PermRoleUser, 101, false, false},
		{"admin can ban here", "set_group_ban", `{"group_id":101,"user_id":123,"duration":1}`, goba.PermRoleAdmin, 101, false, true},
		{"admin cannot ban elsewhere", "set_group_ban", `{"group_id":202,"user_id":123,"duration":1}`, goba.PermRoleAdmin, 101, false, false},
		{"admin cannot grant admin", "set_group_admin", `{"group_id":101,"user_id":123,"enable":true}`, goba.PermRoleAdmin, 101, false, false},
		{"admin cannot leave", "set_group_leave", `{"group_id":101}`, goba.PermRoleAdmin, 101, false, false},
		{"member reads here", "get_group_info", `{"group_id":101}`, goba.PermRoleUser, 101, false, true},
		{"member cannot read elsewhere", "get_group_info", `{"group_id":202}`, goba.PermRoleUser, 101, false, false},
		{"admin cannot read elsewhere", "get_group_info", `{"group_id":202}`, goba.PermRoleAdmin, 101, false, false},
		{"private member cannot query group", "get_group_info", `{"group_id":101}`, goba.PermRoleUser, -123, false, false},
		{"private member cannot send private action", "send_private_msg", `{"user_id":123,"message":"hello"}`, goba.PermRoleUser, -123, false, false},
		{"owner private action unchanged", "send_private_msg", `{"user_id":123,"message":"hello"}`, goba.PermRoleOwner, -123, true, true},
		{"owner other group unchanged", "get_group_info", `{"group_id":202}`, goba.PermRoleOwner, 101, true, true},
		{"zero group invalid", "get_group_info", `{"group_id":0}`, goba.PermRoleOwner, 101, true, false},
		{"negative group invalid", "get_group_info", `{"group_id":-123}`, goba.PermRoleOwner, 101, true, false},
		{"string group invalid", "get_group_info", `{"group_id":"101"}`, goba.PermRoleUser, 101, false, false},
		{"member array all denied", "send_group_msg", `{"group_id":101,"message":[{"type":"at","data":{"qq":"all"}}]}`, goba.PermRoleUser, 101, false, false},
		{"member CQ all denied", "send_group_msg", `{"group_id":101,"message":"[CQ:at,qq=all]hello"}`, goba.PermRoleUser, 101, false, false},
		{"member CQ person allowed", "send_group_msg", `{"group_id":101,"message":"[CQ:at,qq=123]hello"}`, goba.PermRoleUser, 101, false, true},
		{"member array person allowed", "send_group_msg", `{"group_id":101,"message":[{"type":"at","data":{"qq":"123"}}]}`, goba.PermRoleUser, 101, false, true},
		{"member invalid CQ at denied", "send_group_msg", `{"group_id":101,"message":"[CQ:at,qq=0]hello"}`, goba.PermRoleUser, 101, false, false},
		{"admin CQ all unchanged", "send_group_msg", `{"group_id":101,"message":"[CQ:at,qq=all]hello"}`, goba.PermRoleAdmin, 101, false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				content := `{"action":"` + c.action + `","params":` + c.params + `} ` + `{"action":"end_action","params":{}}`
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": content}}}})
			}))
			defer server.Close()
			ag := goba.NewAgent(&goba.Config{}, 16, 8, time.Hour, "", permissionMemory{}, true, true)
			reqs, _, err := chat.NextAgentActions(&ag, c.sudo, 1, deepinfra.NewAPI(server.URL, "synthetic"), model.NewOpenAI("test", "", .7, 1, 512, ""), c.group, c.role)
			if c.allow {
				if err != nil || len(reqs) != 1 || reqs[0].Action != c.action {
					t.Fatalf("allowed action lost: requests=%v error=%v", reqs, err)
				}
			} else if len(reqs) != 0 {
				t.Fatalf("unauthorized action reached execution boundary: %v", reqs)
			}
		})
	}
}
