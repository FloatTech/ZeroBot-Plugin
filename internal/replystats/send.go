package replystats

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func targetID(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		if !math.IsNaN(v) && !math.IsInf(v, 0) && math.Trunc(v) == v && math.Abs(v) <= 9007199254740991 {
			return int64(v), true
		}
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		return n, err == nil
	}
	return 0, false
}
func (t *Task) currentReply(action string, params zero.Params) bool {
	if t == nil {
		return false
	}
	groupAction := false
	privateAction := false
	switch action {
	case "send_group_msg", "send_group_forward_msg", "send_group_ai_record":
		groupAction = true
	case "send_private_msg", "send_private_forward_msg":
		privateAction = true
	case "send_msg", "send_forward_msg":
		if params["message_type"] == "group" {
			groupAction = true
		} else if params["message_type"] == "private" {
			privateAction = true
		} else if gid, ok := targetID(params["group_id"]); ok && gid != 0 {
			groupAction = true
		} else {
			privateAction = true
		}
	default:
		return false
	}
	if t.groupID != 0 {
		gid, ok := targetID(params["group_id"])
		return groupAction && ok && gid == t.groupID
	}
	uid, ok := targetID(params["user_id"])
	return privateAction && ok && uid == t.userID
}
func confirmed(action string, response zero.APIResponse) bool {
	if response.Status != "ok" || response.RetCode != 0 {
		return false
	}
	if action == "send_group_ai_record" {
		return true
	} // Existing NapCat voice receipt may contain zero ID.
	id := response.Data.Get("message_id")
	if !id.Exists() {
		return false
	}
	n, err := strconv.ParseInt(id.String(), 10, 64)
	return err == nil && n != 0
}

// Call observes the original API response without modifying params, results or retry policy.
func (t *Task) Call(action string, params zero.Params, call func(string, zero.Params) zero.APIResponse) (response zero.APIResponse) {
	observe := t.currentReply(action, params)
	returned := false
	defer func() {
		if observe && !returned {
			t.record(false)
		}
	}()
	response = call(action, params)
	returned = true
	if observe {
		t.record(confirmed(action, response))
	}
	return response
}

// Send follows ZeroBot's existing private/group and merged-forward routing.
// Unsupported event types retain the SDK path and are outside these statistics.
func (t *Task) Send(ctx *zero.Ctx, msg message.Message) message.ID {
	if t == nil {
		return ctx.Send(msg)
	}
	group := ctx.Event.MessageType == "group"
	action := "send_private_msg"
	key := "message"
	params := zero.Params{"user_id": ctx.Event.UserID}
	if group {
		action = "send_group_msg"
		params = zero.Params{"group_id": ctx.Event.GroupID}
	}
	if len(msg) > 0 && msg[0].Type == "node" {
		key = "messages"
		action = strings.Replace(action, "_msg", "_forward_msg", 1)
	}
	params[key] = msg
	response := t.Call(action, params, ctx.CallAction)
	return message.NewMessageIDFromInteger(response.Data.Get("message_id").Int())
}
func (t *Task) SendChain(ctx *zero.Ctx, msg ...message.Segment) message.ID {
	if t == nil {
		return ctx.SendChain(msg...)
	}
	if zero.BotConfig.AddSpaceAfterAt && len(msg) > 0 {
		spaced := make(message.Message, 0, len(msg)*2)
		for i := 0; i < len(msg)-1; i++ {
			spaced = append(spaced, msg[i])
			if msg[i].Type == "at" && (msg[i+1].Type != "text" || (len(msg[i+1].Data["text"]) > 0 && msg[i+1].Data["text"][0] != ' ')) {
				spaced = append(spaced, message.Text(" "))
			}
		}
		msg = append(spaced, msg[len(msg)-1])
	}
	return t.Send(ctx, msg)
}
