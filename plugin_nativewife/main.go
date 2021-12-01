package nativewife

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/picture"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
)

const base = "data/nwife"

var baseuri = "file:///" + file.BOT_PATH + "/" + base

func init() {
	err := os.MkdirAll(base, 0755)
	if err != nil {
		panic(err)
	}
	engine := control.Register("nwife", &control.Options{
		DisableOnDefault: false,
		Help:             "nativewife\n- 抽wife[@xxx]\n- 添加wife[名字][图片]",
	})
	engine.OnPrefix("抽wife", zero.OnlyGroup).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			grpf := strconv.FormatInt(ctx.Event.GroupID, 36)
			wifes, err := os.ReadDir(base + "/" + grpf)
			if err != nil {
				ctx.SendChain(message.Text("一个wife也没有哦~"))
				return
			}
			switch len(wifes) {
			case 0:
				ctx.SendChain(message.Text("一个wife也没有哦~"))
			case 1:
				wn := wifes[0].Name()
				ctx.SendChain(message.Text("大家的wife都是", wn, "\n"), message.Image(baseuri+"/"+grpf+"/"+wn), message.Text("\n哦~"))
			default:
				// 获取名字
				name := ctx.State["args"].(string)
				if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
					qq, _ := strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
					name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, qq, false).Get("nickname").Str
				} else if name == "" {
					name = ctx.Event.Sender.NickName
				}
				now := time.Now()
				s := md5.Sum(helper.StringToBytes(fmt.Sprintf("%s%d%d%d", name, now.Year(), now.Month(), now.Day())))
				r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(s[:]))))
				n := r.Intn(len(wifes))
				wn := wifes[n].Name()
				ctx.SendChain(message.Text(name, "\n的wife是", wn, "\n"), message.Image(baseuri+"/"+grpf+"/"+wn), message.Text("\n哦~"))
			}
		})
	// 上传一张图
	engine.OnKeyword("添加wife", zero.OnlyGroup, zero.AdminPermission, picture.MustGiven).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			name := ""
			for _, elem := range ctx.Event.Message {
				if elem.Type == "text" {
					name = strings.ReplaceAll(elem.Data["text"], " ", "")
					name = name[strings.LastIndex(name, "添加wife")+10:]
				}
			}
			if name != "" {
				url := ctx.State["image_url"].([]string)[0]
				grpfolder := base + "/" + strconv.FormatInt(ctx.Event.GroupID, 36)
				if file.IsNotExist(grpfolder) {
					os.Mkdir(grpfolder, 0755)
				}
				err = file.DownloadTo(url, grpfolder+"/"+name, true)
				if err == nil {
					ctx.SendChain(message.Text("成功！"))
				} else {
					ctx.SendChain(message.Text("错误：", err.Error()))
				}
			} else {
				ctx.SendChain(message.Text("没有找到wife的名字！"))
			}
		})
}
