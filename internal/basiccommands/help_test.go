package basiccommands

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type localDriver struct {
	ready chan func([]byte, zero.APICaller)
	stop  chan struct{}
	sent  chan zero.APIRequest
}

func (*localDriver) Connect() {}
func (d *localDriver) Listen(receive func([]byte, zero.APICaller)) {
	d.ready <- receive
	<-d.stop
}
func (d *localDriver) CallAPI(_ context.Context, req zero.APIRequest) (zero.APIResponse, error) {
	d.sent <- req
	return zero.APIResponse{Status: "ok", Data: gjson.Parse(`{"message_id":99}`)}, nil
}

// A blocking chat fallback registered before help must not intercept help.
// Ordinary chat and group messages without @bot must still reach the fallback.
func TestHelpRouting(t *testing.T) {
	zero.OnMessage().SetBlock(true).ThirdPriority().Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("chat fallback"))
	})
	RegisterHelp("local help banner")
	driver := &localDriver{ready: make(chan func([]byte, zero.APICaller), 1), stop: make(chan struct{}), sent: make(chan zero.APIRequest, 16)}
	defer close(driver.stop)
	zero.Run(&zero.Config{CommandPrefix: "/", MaxProcessTime: time.Second, Driver: []zero.Driver{driver}})
	var receive func([]byte, zero.APICaller)
	select {
	case receive = <-driver.ready:
	case <-time.After(time.Second):
		t.Fatal("local driver did not start")
	}
	for i, tc := range []struct {
		name  string
		text  string
		group bool
		atBot bool
		want  string
	}{
		{"private slash", "/help", false, false, "local help banner"},
		{"trailing newline from QQ", "/help\n", false, false, "local help banner"},
		{"surrounding whitespace", "\r\n\t/help \t", false, false, "local help banner"},
		{"private plain", "help", false, false, "local help banner"},
		{"private dot", ".help", false, false, "local help banner"},
		{"private menu", "菜单", false, false, "local help banner"},
		{"ordinary chat", "hello", false, false, "chat fallback"},
		{"help with arguments", "/help extra", false, false, "chat fallback"},
		{"longer command", "/helper", false, false, "chat fallback"},
		{"group no at", "/help", true, false, "chat fallback"},
		{"group at bot", "/help", true, true, "local help banner"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			segments := message.Message{message.Text(tc.text)}
			if tc.atBot {
				segments = append(message.Message{message.At(87654321)}, segments...)
			}
			event := map[string]any{"time": time.Now().Unix(), "self_id": 87654321, "post_type": "message", "message_type": "private", "sub_type": "friend", "message_id": i + 1, "user_id": 12345678, "raw_message": tc.text, "message": segments, "sender": map[string]any{"user_id": 12345678, "nickname": "Local test"}}
			if tc.group {
				event["message_type"] = "group"
				event["group_id"] = 23456789
			}
			payload, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			receive(payload, driver)
			select {
			case request := <-driver.sent:
				encoded, err := json.Marshal(request.Params["message"])
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(encoded), tc.want) {
					t.Fatalf("reply = %s, want %q", encoded, tc.want)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("no reply to local message")
			}
		})
	}
}
