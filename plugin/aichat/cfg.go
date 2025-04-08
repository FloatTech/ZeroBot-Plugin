package aichat

import (
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var cfg = newconfig()

type config struct {
	ModelName string
	Type      int
	SystemP   string
	API       string
	Key       string
	Separator string
	NoReplyAT bool
	NoSystemP bool
}

func newconfig() config {
	return config{
		ModelName: model.ModelDeepDeek,
		SystemP:   chat.SystemPrompt,
		API:       deepinfra.OpenAIDeepInfra,
	}
}

func (c *config) isvalid() bool {
	return c.ModelName != "" && c.API != "" && c.Key != ""
}

func ensureconfig(ctx *zero.Ctx) bool {
	c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if !ok {
		return false
	}
	if !cfg.isvalid() {
		err := c.GetExtra(&cfg)
		if err != nil {
			logrus.Warnln("ERROR: get extra err:", err)
		}
		if !cfg.isvalid() {
			cfg = newconfig()
		}
	}
	return true
}

func newextrasetstr(ptr *string) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			ctx.SendChain(message.Text("ERROR: empty args"))
			return
		}
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: no such plugin"))
			return
		}
		*ptr = args
		err := c.SetExtra(&cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}

func newextrasetbool(ptr *bool) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		args := ctx.State["regex_matched"].([]string)
		isno := args[1] == "不"
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: no such plugin"))
			return
		}
		*ptr = isno
		err := c.SetExtra(&cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}
