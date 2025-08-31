package aichat

import (
	"fmt"
	"strconv"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	cfg = newconfig()
)

type config struct {
	ModelName string
	Type      int
	MaxN      uint
	TopP      float32
	SystemP   string
	API       string
	Key       string
	Separator string
	NoReplyAT bool
	NoSystemP bool
	NoRecord  bool
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

func newextrasetuint(ptr *uint) func(ctx *zero.Ctx) {
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
		n, err := strconv.ParseUint(args, 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: parse args err: ", err))
			return
		}
		*ptr = uint(n)
		err = c.SetExtra(&cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}

func newextrasetfloat32(ptr *float32) func(ctx *zero.Ctx) {
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
		n, err := strconv.ParseFloat(args, 32)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: parse args err: ", err))
			return
		}
		*ptr = float32(n)
		err = c.SetExtra(&cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}

func printConfig(rate int64, temperature int64, cfg config) string {
	maxn := cfg.MaxN
	if maxn == 0 {
		maxn = 4096
	}
	topp := cfg.TopP
	if topp == 0 {
		topp = 0.9
	}
	var builder strings.Builder
	builder.WriteString("当前AI聊天配置：\n")
	builder.WriteString(fmt.Sprintf("• 模型名：%s\n", cfg.ModelName))
	builder.WriteString(fmt.Sprintf("• 接口类型：%d(%s)\n", cfg.Type, apilist[cfg.Type]))
	builder.WriteString(fmt.Sprintf("• 触发概率：%d%%\n", rate))
	builder.WriteString(fmt.Sprintf("• 温度：%.2f\n", float32(temperature)/100))
	builder.WriteString(fmt.Sprintf("• 最大长度：%d\n", maxn))
	builder.WriteString(fmt.Sprintf("• TopP：%.1f\n", topp))
	builder.WriteString(fmt.Sprintf("• 系统提示词：%s\n", cfg.SystemP))
	builder.WriteString(fmt.Sprintf("• 接口地址：%s\n", cfg.API))
	builder.WriteString(fmt.Sprintf("• 密钥：%s\n", maskKey(cfg.Key)))
	builder.WriteString(fmt.Sprintf("• 分隔符：%s\n", cfg.Separator))
	builder.WriteString(fmt.Sprintf("• 响应@：%s\n", yesNo(!cfg.NoReplyAT)))
	builder.WriteString(fmt.Sprintf("• 支持系统提示词：%s\n", yesNo(!cfg.NoSystemP)))
	builder.WriteString(fmt.Sprintf("• 以AI语音输出：%s\n", yesNo(!cfg.NoRecord)))
	return builder.String()
}

func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return key[:2] + strings.Repeat("*", len(key)-4) + key[len(key)-2:]
}

func yesNo(b bool) string {
	if b {
		return "是"
	}
	return "否"
}
