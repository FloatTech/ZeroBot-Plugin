// Package emozi 颜文字抽象转写
package emozi

import (
	"os"
	"strconv"
	"strings"

	"github.com/FloatTech/AnimeAPI/emozi"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "颜文字抽象转写",
		Help:              "- 抽象转写[文段]\n- 抽象还原[文段]\n- 抽象登录[用户名]",
		PrivateDataFolder: "emozi",
	})
	usr := emozi.Anonymous()
	data, err := os.ReadFile(en.DataFolder() + "user.txt")
	hasaccount := false
	if err == nil {
		arr := strings.Split(string(data), "\n")
		if len(arr) >= 2 {
			usr = emozi.NewUser(arr[0], arr[1])
			err = usr.Login()
			if err != nil {
				logrus.Infoln("[emozi]", "以", usr, "身份登录失败:", err)
				usr = emozi.Anonymous()
			} else {
				logrus.Infoln("[emozi]", "以", usr, "身份登录成功")
				hasaccount = true
			}
		}
	}

	en.OnPrefix("抽象转写").Limit(ctxext.LimitByUser).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		txt := strings.TrimSpace(ctx.State["args"].(string))
		out, chs, err := usr.Marshal(false, txt)
		if err != nil {
			if hasaccount {
				err = usr.Login()
				if err == nil {
					out, chs, err = usr.Marshal(false, txt)
				}
			}
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		}
		if len(chs) == 0 {
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(out)))
			return
		}
		for i, c := range chs {
			ch := ctx.Get("请选择第" + strconv.Itoa(i) + "个多音字(1~" + strconv.Itoa(c) + ")")
			n, err := strconv.Atoi(ch)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if n < 1 || n > c {
				ctx.SendChain(message.Text("ERROR: 输入越界"))
				return
			}
			chs[i] = n - 1
		}
		out, _, err = usr.Marshal(false, txt, chs...)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(out)))
	})
	en.OnPrefix("抽象还原").Limit(ctxext.LimitByUser).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		txt := strings.TrimSpace(ctx.State["args"].(string))
		out, err := usr.Unmarshal(false, txt)
		if err != nil {
			if hasaccount {
				err = usr.Login()
				if err == nil {
					out, err = usr.Unmarshal(false, txt)
				}
			}
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		}
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(out)))
	})
	en.OnPrefix("抽象登录", zero.OnlyPrivate).Limit(ctxext.LimitByUser).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		name := strings.TrimSpace(ctx.State["args"].(string))
		pswd := strings.TrimSpace(ctx.Get("请输入密码"))
		newusr := emozi.NewUser(name, pswd)
		err := newusr.Login()
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		err = os.WriteFile(en.DataFolder()+"user.txt", []byte(name+"\n"+pswd), 0644)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		usr = newusr
		hasaccount = true
		ctx.SendChain(message.Text("成功"))
	})
}
