// Package easywife 简单本地老婆
package easywife

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
)

func init() {
	engine := control.Register("easywife", &control.Options{
		DisableOnDefault: false,
		Help: "本地老婆\n" +
			"抽老婆",
		PrivateDataFolder: "easywife",
	})
	cachePath := engine.DataFolder() + "wife/"
	os.MkdirAll(cachePath, 0755)
	engine.OnPrefix("抽老婆").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			wifes, _ := os.ReadDir(cachePath)
			name := ctx.NickName()
			now := time.Now()
			s := md5.Sum(helper.StringToBytes(fmt.Sprintf("%s%d%d%d", name, now.Year(), now.Month(), now.Day())))
			r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(s[:]))))
			n := r.Intn(len(wifes))
			wn := wifes[n].Name()
			reg := regexp.MustCompile(`[^\.]+`)
			list := reg.FindAllString(wn, -1)
			ctx.SendChain(
				message.Text(name, "さんが二次元で結婚するであろうヒロインは、", "\n"),
				message.Image("file:///"+file.BOTPATH+"/"+cachePath+wn),
				message.Text("\n【", list[0], "】です！"))
		})
	/*engine.OnPrefix("添加老婆", zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			name := ""
			for _, elem := range ctx.Event.Message {
				if elem.Type == "text" {
					name = strings.ReplaceAll(elem.Data["text"], " ", "")
					name = name[strings.LastIndex(name, "添加wife")+10:]
					name = strings.ReplaceAll(name, "/", "")
					name = strings.ReplaceAll(name, "\\", "")
					break
				}
			}
			url := ctx.State["image_url"].([]string)[0]
			err := file.DownloadTo(url, file.BOTPATH+"/"+cachePath+name, true)
			if err == nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功！"))
			} else {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("错误：", err.Error()))
			}
		})
	/*Todo.
	engine.OnPrefix("删除老婆", zero.OnlyGroup,zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
	*/
}
