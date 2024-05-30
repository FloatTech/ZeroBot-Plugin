// Package ctxext zb context 扩展
package ctxext

import (
	"encoding/json"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/binary"
)

//nolint:revive
type (
	NoCtxGetMsg  func(int64) zero.Message
	NoCtxSendMsg func(any) int64
)

// GetMessage ...
func GetMessage(ctx *zero.Ctx) NoCtxGetMsg {
	return func(id int64) zero.Message {
		return ctx.GetMessage(message.NewMessageIDFromInteger(id))
	}
}

// GetFirstMessageInForward ...
func GetFirstMessageInForward(ctx *zero.Ctx) NoCtxGetMsg {
	return func(id int64) zero.Message {
		msg := GetMessage(ctx)(id)
		if len(msg.Elements) == 0 {
			return zero.Message{}
		}
		msgs := ctx.GetForwardMessage(msg.Elements[0].Data["id"]).Get("messages").Array()
		if len(msgs) == 0 {
			return zero.Message{}
		}
		m := zero.Message{
			Elements: message.ParseMessage(binary.StringToBytes(msgs[0].Get("content").Raw)),
			Sender:   &zero.User{},
		}
		err := json.Unmarshal(binary.StringToBytes(msgs[0].Get("sender").Raw), m.Sender)
		if err != nil {
			return zero.Message{}
		}
		return m
	}
}

// SendTo ...
func SendTo(ctx *zero.Ctx, user int64) NoCtxSendMsg {
	return func(msg any) int64 {
		return ctx.SendPrivateMessage(user, msg)
	}
}

// Send ...
func Send(ctx *zero.Ctx) NoCtxSendMsg {
	return func(msg any) int64 {
		return ctx.Send(msg).ID()
	}
}

// SendToSelf ...
func SendToSelf(ctx *zero.Ctx) NoCtxSendMsg {
	return func(msg any) int64 {
		return ctx.SendPrivateMessage(ctx.Event.SelfID, msg)
	}
}

// FakeSenderForwardNode ...
func FakeSenderForwardNode(ctx *zero.Ctx, msgs ...message.MessageSegment) message.MessageSegment {
	return message.CustomNode(
		ctx.CardOrNickName(ctx.Event.UserID),
		ctx.Event.UserID,
		msgs)
}

// SendFakeForwardToGroup ...
func SendFakeForwardToGroup(ctx *zero.Ctx, msgs ...message.MessageSegment) NoCtxSendMsg {
	return func(msg any) int64 {
		return ctx.SendGroupForwardMessage(ctx.Event.GroupID, message.Message{
			FakeSenderForwardNode(ctx, msg.(message.Message)...),
			FakeSenderForwardNode(ctx, msgs...),
		}).Get("message_id").Int()
	}
}
