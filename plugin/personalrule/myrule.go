package personalrule

import (
	"os/exec"

	control "github.com/FloatTech/zbputils/control"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("mine", &control.Options{
		DisableOnDefault: false,
		Help: "Poweroff\n" +
			"- kill/[Maybe]reboot: [@bot#kill|@bot#关机|@bot#poweroff|@bot#killbot]",
	})
	engine.OnFullMatchGroup([]string{"#kill", "#关机", "#poweroff", "#killbot"}, zero.OwnerPermission, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(message.Poke(ctx.Event.UserID))
			err := exec.Command("rm", "-f", "/home/root/qqbot/session.token").Run()
			ctx.Send(message.Text(err))
			println(err)
			err = exec.Command("killall", "-9", "go-cqhttp").Run()
			ctx.Send(message.Text(err))
			println(err)
			err = exec.Command("killall", "-9", "zbp").Run()
			ctx.Send(message.Text(err))
			println(err)
			err = exec.Command("rm", "-rf", "/home/root/qqbot/data/{leveldb-v2,Reborn,SetuTime,videos,images,VtbQuotation,sleep,hs,voices,cache,acgimage,Funny}").Run()
			ctx.Send(message.Text(err))
			println(err)
			ctx.Send(message.Poke(ctx.Event.UserID))
		},
		)
	engine.OnFullMatchGroup([]string{"#update", "#upgrade", "#updata", "#更新"}, zero.OwnerPermission, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(message.Poke(ctx.Event.UserID))
			err := exec.Command("rm", "-rf", "/home/root/qqbot/data/{leveldb-v2,Reborn,SetuTime,videos,images,VtbQuotation,sleep,hs,voices,cache,acgimage,Funny}").Run()
			ctx.Send(message.Text(err))
			println(err)
			err = exec.Command("sh", "/home/root/qqbot/update.sh").Run()
			ctx.Send(message.Text(err))
			println(err)
			err = exec.Command("sh", "/home/root/qqbot/qqbot.sh").Run()
			ctx.Send(message.Text(err))
			println(err)
			ctx.Send(message.Poke(ctx.Event.UserID))
		},
		)
	engine.OnKeywordGroup([]string{"戳我", "揉", "揉揉", "戳戳", "~", "poke"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(message.Poke(ctx.Event.UserID))
		},
		)
}
