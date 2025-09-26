package aichat

import (
	"errors"
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

var (
	apitypes = map[string]uint8{
		"OpenAI": 0,
		"OLLaMA": 1,
		"GenAI":  2,
	}
	apilist = [3]string{"OpenAI", "OLLaMA", "GenAI"}
)

// ModelType 支持打印 string 并生产 protocal
type ModelType int

func newModelType(typ string) (ModelType, error) {
	t, ok := apitypes[typ]
	if !ok {
		return 0, errors.New("未知类型 " + typ)
	}
	return ModelType(t), nil
}

func (mt ModelType) String() string {
	return apilist[mt]
}

func (mt ModelType) protocol(modn string, temp float32, topp float32, maxn uint) (mod model.Protocol, err error) {
	switch cfg.Type {
	case 0:
		mod = model.NewOpenAI(
			modn, cfg.Separator,
			temp, topp, maxn,
		)
	case 1:
		mod = model.NewOLLaMA(
			modn, cfg.Separator,
			temp, topp, maxn,
		)
	case 2:
		mod = model.NewGenAI(
			modn,
			temp, topp, maxn,
		)
	default:
		err = errors.New("unsupported model type " + strconv.Itoa(int(cfg.Type)))
	}
	return
}

// ModelBool 支持打印成 "是/否"
type ModelBool bool

func (mb ModelBool) String() string {
	if mb {
		return "是"
	}
	return "否"
}

// ModelKey 支持隐藏密钥
type ModelKey string

func (mk ModelKey) String() string {
	if len(mk) == 0 {
		return "未设置"
	}
	if len(mk) <= 4 {
		return "****"
	}
	key := string(mk)
	return key[:2] + strings.Repeat("*", len(key)-4) + key[len(key)-2:]
}

type config struct {
	ModelName      string
	ImageModelName string
	AgentModelName string
	Type           ModelType
	ImageType      ModelType
	AgentType      ModelType
	MaxN           uint
	TopP           float32
	SystemP        string
	API            string
	ImageAPI       string
	AgentAPI       string
	Key            ModelKey
	ImageKey       ModelKey
	AgentKey       ModelKey
	Separator      string
	NoSystemP      ModelBool
}

func newconfig() config {
	return config{
		ModelName: model.ModelDeepDeek,
		SystemP:   chat.SystemPrompt,
		API:       deepinfra.OpenAIDeepInfra,
	}
}

func (c *config) String() string {
	topp, maxn := c.mparams()
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("• 模型名：%s\n", c.ModelName))
	sb.WriteString(fmt.Sprintf("• 图像模型名：%s\n", c.ImageModelName))
	sb.WriteString(fmt.Sprintf("• Agent模型名：%s\n", c.AgentModelName))
	sb.WriteString(fmt.Sprintf("• 接口类型：%v\n", c.Type))
	sb.WriteString(fmt.Sprintf("• 图像接口类型：%v\n", c.ImageType))
	sb.WriteString(fmt.Sprintf("• Agent接口类型：%v\n", c.AgentType))
	sb.WriteString(fmt.Sprintf("• 最大长度：%d\n", maxn))
	sb.WriteString(fmt.Sprintf("• TopP：%.1f\n", topp))
	sb.WriteString(fmt.Sprintf("• 系统提示词：%s\n", c.SystemP))
	sb.WriteString(fmt.Sprintf("• 接口地址：%s\n", c.API))
	sb.WriteString(fmt.Sprintf("• 图像接口地址：%s\n", c.ImageAPI))
	sb.WriteString(fmt.Sprintf("• Agent接口地址：%s\n", c.AgentAPI))
	sb.WriteString(fmt.Sprintf("• 密钥：%v\n", c.Key))
	sb.WriteString(fmt.Sprintf("• 图像密钥：%v\n", c.ImageKey))
	sb.WriteString(fmt.Sprintf("• Agent密钥：%v\n", c.AgentKey))
	sb.WriteString(fmt.Sprintf("• 分隔符：%s\n", c.Separator))
	sb.WriteString(fmt.Sprintf("• 支持系统提示词：%v\n", !c.NoSystemP))
	return sb.String()
}

func (c *config) isvalid() bool {
	return c.ModelName != "" && c.API != "" && c.Key != ""
}

// 获取全局模型参数：TopP和最大长度
func (c *config) mparams() (topp float32, maxn uint) {
	// 处理TopP参数
	topp = c.TopP
	if topp == 0 {
		topp = 0.9
	}

	// 处理最大长度参数
	maxn = c.MaxN
	if maxn == 0 {
		maxn = 4096
	}

	return topp, maxn
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

func newextrasetstr[T ~string](ptr *T) func(ctx *zero.Ctx) {
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
		*ptr = T(args)
		err := c.SetExtra(&cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}

func newextrasetbool[T ~bool](ptr *T) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		args := ctx.State["regex_matched"].([]string)
		isno := args[1] == "不"
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: no such plugin"))
			return
		}
		*ptr = T(isno)
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

func newextrasetmodeltype(ptr *ModelType) func(ctx *zero.Ctx) {
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
		typ, err := newModelType(args)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		*ptr = typ
		err = c.SetExtra(&cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}
