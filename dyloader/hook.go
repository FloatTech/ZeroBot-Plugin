package dyloader

import zero "github.com/wdvxdr1123/ZeroBot"

func sendGroupMessage(ctx *zero.Ctx, groupID int64, message interface{}) int64 {
	return ctx.SendGroupMessage(groupID, message)
}
func sendPrivateMessage(ctx *zero.Ctx, userID int64, message interface{}) int64 {
	return ctx.SendPrivateMessage(userID, message)
}
func getMessage(ctx *zero.Ctx, messageID int64) zero.Message {
	return ctx.GetMessage(messageID)
}
func parse(ctx *zero.Ctx, model interface{}) (err error) {
	return ctx.Parse(model)
}
