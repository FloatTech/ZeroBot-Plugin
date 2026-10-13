package replystats

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type testDriver struct {
	ready chan func([]byte, zero.APICaller)
	stop  chan struct{}
	sent  chan zero.APIRequest
	done  chan struct{}
}

func (*testDriver) Connect()                                      {}
func (d *testDriver) Listen(receive func([]byte, zero.APICaller)) { d.ready <- receive; <-d.stop }
func (d *testDriver) CallAPI(_ context.Context, req zero.APIRequest) (zero.APIResponse, error) {
	time.Sleep(15 * time.Millisecond)
	d.sent <- req
	return receipt("123"), nil
}

func TestActualSDKIngressAndOriginalSendParameterEquivalence(t *testing.T) {
	c := New()
	driver := &testDriver{ready: make(chan func([]byte, zero.APICaller), 1), stop: make(chan struct{}), sent: make(chan zero.APIRequest, 16), done: make(chan struct{}, 16)}
	defer close(driver.stop)
	zero.OnMessage().SetBlock(true).Handle(func(ctx *zero.Ctx) {
		defer func() { driver.done <- struct{}{} }()
		var task *Task
		if ctx.Event.RawMessage == "observed" {
			task = c.Begin(ctx, GPT)
			defer task.End()
		}
		if ctx.Event.SubType == "forward" {
			msg := message.Message{{Type: "node", Data: map[string]string{"id": "42"}}}
			if task != nil {
				task.Send(ctx, msg)
			} else {
				ctx.Send(msg)
			}
		} else {
			msg := []message.Segment{message.At(111), message.Text("hello🙂\nworld")}
			if task != nil {
				task.SendChain(ctx, msg...)
			} else {
				ctx.SendChain(msg...)
			}
		}
	})
	zero.Run(&zero.Config{MaxProcessTime: time.Second, AddSpaceAfterAt: true, Driver: []zero.Driver{driver}})
	var receive func([]byte, zero.APICaller)
	select {
	case receive = <-driver.ready:
	case <-time.After(time.Second):
		t.Fatal("driver not ready")
	}
	for _, group := range []bool{false, true} {
		for _, forward := range []bool{false, true} {
			var requests []zero.APIRequest
			for _, mode := range []string{"original", "observed"} {
				event := map[string]any{"time": time.Now().Unix(), "self_id": 87654321, "post_type": "message", "message_type": "private", "sub_type": "friend", "message_id": 11, "user_id": 12345, "raw_message": mode, "message": message.Message{message.Text(mode)}, "sender": map[string]any{"user_id": 12345, "nickname": "Synthetic"}}
				if group {
					event["message_type"] = "group"
					event["group_id"] = 67890
				}
				if forward {
					event["sub_type"] = "forward"
				}
				raw, _ := json.Marshal(event)
				receive(raw, driver)
				select {
				case req := <-driver.sent:
					requests = append(requests, req)
				case <-time.After(2 * time.Second):
					t.Fatal("no send")
				}
				select {
				case <-driver.done:
				case <-time.After(time.Second):
					t.Fatal("handler not done")
				}
			}
			if requests[0].Action != requests[1].Action || !reflect.DeepEqual(requests[0].Params, requests[1].Params) {
				t.Fatalf("SDK parameters changed: %#v %#v", requests[0], requests[1])
			}
		}
	}
	got := c.Read()
	if got.Started != 4 || got.Completed != 4 || got.MissingIngress != 0 || got.ReplyAccepted != 4 || got.AverageFirstReplyMilliseconds < 15 {
		t.Fatalf("wrong SDK pipeline statistics: %+v", got)
	}
}
