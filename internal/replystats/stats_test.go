package replystats

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func testContext(group bool, stamped bool) *zero.Ctx {
	e := &zero.Event{PostType: "message", MessageType: "private", UserID: 12345}
	if group {
		e.MessageType = "group"
		e.GroupID = 67890
	}
	ctx := &zero.Ctx{Event: e, State: zero.State{}}
	if stamped {
		ctx.State[ingressKey] = time.Now().Add(-50 * time.Millisecond)
	}
	return ctx
}
func receipt(id string) zero.APIResponse {
	return zero.APIResponse{Status: "ok", Data: gjson.Parse(`{"message_id":` + id + `}`)}
}
func TestReplyTimingsIncludeIngressAndOnlyFinishedTasks(t *testing.T) {
	c := New()
	task := c.Begin(testContext(false, true), GPT)
	params := zero.Params{"user_id": int64(12345), "message": "SECRET-BODY"}
	response := task.Call("send_private_msg", params, func(action string, p zero.Params) zero.APIResponse {
		if action != "send_private_msg" || p["message"] != "SECRET-BODY" {
			t.Fatal("API parameters changed")
		}
		return receipt("-123")
	})
	if response.Data.Get("message_id").Int() != -123 {
		t.Fatal("response changed")
	}
	active := c.Read()
	if active.Active != 1 || active.WithReply != 0 || active.ReplyAccepted != 1 || active.AverageFirstReplyMilliseconds != 0 {
		t.Fatal("unfinished task mixed into average")
	}
	time.Sleep(10 * time.Millisecond)
	task.Call("send_private_msg", params, func(string, zero.Params) zero.APIResponse { return receipt("234") })
	task.End()
	task.End()
	got := c.Read()
	if got.Completed != 1 || got.WithReply != 1 || got.Active != 0 || got.MissingIngress != 0 || got.ReplyAccepted != 2 || got.AverageFirstReplyMilliseconds < 50 || got.AverageLastReplyMilliseconds < got.AverageFirstReplyMilliseconds+8 {
		t.Fatalf("wrong timings: %+v", got)
	}
	body, _ := json.Marshal(got)
	if strings.Contains(string(body), "SECRET") || strings.Contains(string(body), "12345") || strings.Contains(string(body), "67890") {
		t.Fatal("contents or targets retained")
	}
	task.Call("send_private_msg", params, func(string, zero.Params) zero.APIResponse { return receipt("345") })
	if c.Read().ReplyAccepted != 2 {
		t.Fatal("late call changed ended statistics")
	}
}
func TestReceiptBoundariesAndVoiceZeroID(t *testing.T) {
	for _, tc := range []struct {
		name, action string
		r            zero.APIResponse
		ok           bool
	}{
		{"positive", "send_group_msg", receipt("123"), true}, {"negative", "send_private_msg", receipt("-123"), true}, {"integer-string", "send_group_msg", receipt(`"123"`), true},
		{"zero", "send_group_msg", receipt("0"), false}, {"fraction", "send_group_msg", receipt("1.5"), false}, {"missing", "send_group_msg", zero.APIResponse{Status: "ok"}, false},
		{"status-failure", "send_group_msg", zero.APIResponse{Status: "failed", Data: gjson.Parse(`{"message_id":123}`)}, false}, {"code-failure", "send_group_msg", zero.APIResponse{Status: "ok", RetCode: 1, Data: gjson.Parse(`{"message_id":123}`)}, false},
		{"voice-zero", "send_group_ai_record", receipt("0"), true}, {"voice-failure", "send_group_ai_record", zero.APIResponse{Status: "failed"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if confirmed(tc.action, tc.r) != tc.ok {
				t.Fatal("receipt boundary differs")
			}
		})
	}
}
func TestFailedPartialReplyNoReplyAndMissingIngress(t *testing.T) {
	c := New()
	empty := c.Begin(testContext(false, false), Chat)
	empty.End()
	partial := c.Begin(testContext(false, true), Agent)
	partial.record(true)
	partial.record(false)
	partial.End()
	err := c.Begin(testContext(false, true), GPT)
	err.Fail()
	err.End()
	got := c.Read()
	if got.Started != 3 || got.NoReply != 1 || got.Failed != 2 || got.WithReply != 1 || got.MissingIngress != 1 || got.ReplyUnconfirmed != 1 || got.Active != 0 {
		t.Fatalf("misclassified tasks: %+v", got)
	}
	if c.Begin(nil, GPT) != nil || c.Begin(testContext(false, true), kindCount) != nil {
		t.Fatal("invalid task counted")
	}
	ctx := testContext(false, true)
	ctx.Event.MessageType = "guild"
	if c.Begin(ctx, GPT) != nil {
		t.Fatal("guild task counted")
	}
}
func TestCurrentConversationAndActionAllowlist(t *testing.T) {
	for _, group := range []bool{false, true} {
		c := New()
		task := c.Begin(testContext(group, true), Agent)
		for _, tc := range []struct {
			action string
			p      zero.Params
			want   bool
		}{
			{"send_group_msg", zero.Params{"group_id": int64(67890)}, group}, {"send_private_msg", zero.Params{"user_id": "12345"}, !group},
			{"send_group_forward_msg", zero.Params{"group_id": float64(67890)}, group}, {"send_private_forward_msg", zero.Params{"user_id": json.Number("12345")}, !group},
			{"send_group_ai_record", zero.Params{"group_id": 67890}, group}, {"send_msg", zero.Params{"message_type": "group", "group_id": 67890}, group},
			{"send_forward_msg", zero.Params{"user_id": 12345}, !group}, {"send_group_msg", zero.Params{"group_id": 67891}, false}, {"send_private_msg", zero.Params{"user_id": 12346}, false},
			{"get_group_info", zero.Params{"group_id": 67890}, false}, {"send_group_msg", zero.Params{"group_id": 67890.5}, false},
		} {
			if task.currentReply(tc.action, tc.p) != tc.want {
				t.Fatalf("wrong target selection: %s %v", tc.action, tc.p)
			}
		}
		task.Call("get_group_info", zero.Params{"group_id": 67890}, func(string, zero.Params) zero.APIResponse { return receipt("123") })
		task.End()
		if c.Read().ReplyAttempts != 0 || c.Read().NoReply != 1 {
			t.Fatal("non-message action counted as reply")
		}
	}
}
func TestPanicsArePreservedAndTasksEndOnce(t *testing.T) {
	c := New()
	func() {
		defer func() {
			if recover() != "original-panic" {
				t.Fatal("panic changed or swallowed")
			}
		}()
		task := c.Begin(testContext(false, true), GPT)
		defer task.End()
		task.Call("send_private_msg", zero.Params{"user_id": 12345}, func(string, zero.Params) zero.APIResponse { panic("original-panic") })
	}()
	got := c.Read()
	if got.Failed != 1 || got.ReplyUnconfirmed != 1 || got.Active != 0 {
		t.Fatal("panic leaked active task")
	}
}
func TestConcurrentTasksStayBoundedAndConserveCounts(t *testing.T) {
	c := New()
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			task := c.Begin(testContext(false, true), Kind(i%3))
			defer task.End()
			task.record(true)
		}(i)
	}
	wg.Wait()
	got := c.Read()
	if got.Started != 200 || got.Completed != 200 || got.WithReply != 200 || got.ReplyAccepted != 200 || got.Active != 0 {
		t.Fatalf("concurrent counts: %+v", got)
	}
	var started, accepted uint64
	for _, r := range got.Categories {
		started += r.Started
		accepted += r.ReplyAccepted
	}
	if started != got.Started || accepted != got.ReplyAccepted {
		t.Fatal("category totals do not conserve")
	}
}
