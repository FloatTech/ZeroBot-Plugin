package antiabuse

import (
	"fmt"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func banRule(ctx *zero.Ctx) bool {
	if !ctx.Event.IsToMe {
		return false
	}
	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	uuid := fmt.Sprintf("%d-%d", gid, uid)
	if banSet.Include(uuid) {
		return false
	}
	gidPrefix := fmt.Sprintf("%d-", ctx.Event.GroupID)
	var words []string
	_ = wordSet.Iter(func(s string) error {
		trueWord := strings.SplitN(s, gidPrefix, 1)[1]
		words = append(words, trueWord)
		return nil
	})
	for _, word := range words {
		if strings.Contains(ctx.MessageString(), word) {
			if err := insertUser(gid, uid); err != nil {
				ctx.SendChain(message.Text("ban error: ", err))
			}
			banSet.Add(uuid)
			ctx.SetGroupBan(gid, uid, 4*3600)
			time.AfterFunc(4*time.Hour, func() {
				banSet.Remove(uuid)
				if err := deleteUser(gid, uid); err != nil {
					ctx.SendChain(message.Text("ban error: ", err))
				}
			})
			ctx.SendChain(message.Text("检测到违禁词，已封禁/屏蔽4小时"))
			return false
		}
	}
	return true
}
