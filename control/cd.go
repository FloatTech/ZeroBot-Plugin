package control

import (
	"encoding/binary"
	"strings"
	"time"

	b14 "github.com/fumiama/go-base16384"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

var startTime int64

func init() {
	// 插件冲突检测 会在本群发送一条消息并在约 1s 后撤回
	zero.OnFullMatch("插件冲突检测", zero.OnlyGroup, zero.AdminPermission, zero.OnlyToMe).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			tok, err := genToken()
			if err != nil {
				return
			}
			t := message.Text("●cd" + tok)
			startTime = time.Now().Unix()
			id := ctx.SendChain(t)
			process.SleepAbout1sTo2s()
			ctx.DeleteMessage(id)
		})

	zero.OnRegex("^●cd([\u4e00-\u8e00]{4})$", zero.OnlyGroup).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			if isValidToken(ctx.State["regex_matched"].([]string)[1]) {
				msg := ""
				gid := ctx.Event.GroupID
				ForEach(func(key string, manager *Control) bool {
					if manager.IsEnabledIn(gid) {
						msg += "\xfe\xff" + key
					}
					return true
				})
				if len(msg) > 2 {
					my, err := b14.UTF16be2utf8(b14.EncodeString(msg[2:]))
					mys := "●cd●" + helper.BytesToString(my)
					if err == nil {
						id := ctx.SendChain(message.Text(mys))
						process.SleepAbout1sTo2s()
						ctx.DeleteMessage(id)
					}
				}
			}
		})

	zero.OnRegex("^●cd●(([\u4e00-\u8e00]*[\u3d01-\u3d06]?))", zero.OnlyGroup).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			if time.Now().Unix()-startTime < 10 {
				msg, err := b14.UTF82utf16be(helper.StringToBytes(ctx.State["regex_matched"].([]string)[1]))
				if err == nil {
					gid := ctx.Event.GroupID
					for _, s := range strings.Split(b14.DecodeString(msg), "\xfe\xff") {
						mu.RLock()
						c, ok := managers[s]
						mu.RUnlock()
						if ok && c.IsEnabledIn(gid) {
							c.Disable(gid)
						}
					}
				}
			}
		})
}

func genToken() (tok string, err error) {
	timebytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timebytes, uint64(time.Now().Unix()))
	timebytes, err = b14.UTF16be2utf8(b14.Encode(timebytes[1:]))
	if err == nil {
		tok = helper.BytesToString(timebytes)
	}
	return
}

func isValidToken(tok string) (yes bool) {
	s, err := b14.UTF82utf16be(helper.StringToBytes(tok))
	if err == nil {
		timebytes := make([]byte, 1, 8)
		timebytes = append(timebytes, b14.Decode(s)...)
		yes = time.Now().Unix()-int64(binary.BigEndian.Uint64(timebytes)) < 10
	}
	return
}
