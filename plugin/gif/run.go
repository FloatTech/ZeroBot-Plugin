// Package gif 制图
package gif

import (
	"reflect"
	"strconv"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	cmd      = make([]string, 0)
	datapath string
	cmdMap   = map[string]string{
		"搓":      "Cuo",
		"冲":      "Xqe",
		"摸":      "Mo",
		"拍":      "Pai",
		"丢":      "Diu",
		"吃":      "Chi",
		"敲":      "Qiao",
		"啃":      "Ken",
		"蹭":      "Ceng",
		"爬":      "Pa",
		"撕":      "Si",
		"灰度":     "Grayscale",
		"上翻":     "FlipV",
		"下翻":     "FlipV",
		"左翻":     "FlipH",
		"右翻":     "FlipH",
		"反色":     "Invert",
		"浮雕":     "Convolve3x3",
		"打码":     "Blur",
		"负片":     "InvertAndGrayscale",
		"旋转":     "Rotate",
		"变形":     "Deformation",
		"亲":      "Kiss",
		"结婚申请":   "Marriage",
		"结婚登记":   "Marriage",
		"阿尼亚喜欢":  "Anyasuki",
		"像只":     "Alike",
		"我永远喜欢":  "AlwaysLike",
		"永远喜欢":   "AlwaysLike",
		"像样的亲亲":  "DecentKiss",
		"国旗":     "ChinaFlag",
		"不要靠近":   "DontTouch",
		"万能表情":   "Universal",
		"空白表情":   "Universal",
		"采访":     "Interview",
		"需要":     "Need",
		"你可能需要":  "Need",
		"这像画吗":   "Paint",
		"小画家":    "Painter",
		"完美":     "Perfect",
		"玩游戏":    "PlayGame",
		"出警":     "Police",
		"警察":     "Police1",
		"舔":      "Prpr",
		"舔屏":     "Prpr",
		"prpr":   "Prpr",
		"安全感":    "SafeSense",
		"精神支柱":   "Support",
		"想什么":    "Thinkwhat",
		"墙纸":     "Wallpaper",
		"为什么at我": "Whyatme",
		"交个朋友":   "MakeFriend",
		"打工人":    "BackToWork",
		"继续干活":   "BackToWork",
		"兑换券":    "Coupon",
		"注意力涣散":  "Distracted",
		"垃圾桶":    "Garbage",
		"垃圾":     "Garbage",
		"捶":      "Thump",
		"啾啾":     "Jiujiu",
		"2敲":     "Knock",
		"听音乐":    "ListenMusic",
		"永远爱你":   "LoveYou",
		"2拍":     "Pat",
		"顶":      "JackUp",
		"捣":      "Pound",
		"打拳":     "Punch",
		"滚":      "Roll",
		"吸":      "Suck",
		"嗦":      "Suck",
		"扔":      "Throw",
		"锤":      "Hammer",
		"紧贴":     "Tightly",
		"紧紧贴着":   "Tightly",
		"转":      "Turn",
	}
)

func init() { // 插件主体
	for k := range cmdMap {
		cmd = append(cmd, k)
	}
	en := control.Register("gif", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "制图\n- 搓\n- 冲\n- 摸\n- 拍\n- 丢\n- 吃\n- 敲\n- 啃\n- 蹭\n- 爬\n- 撕\n- 灰度\n- 上翻|下翻\n" +
			"- 左翻|右翻\n- 反色\n- 浮雕\n- 打码\n- 负片\n- 旋转 45\n- 变形 100 100\n- 亲\n- 结婚申请|结婚登记\n- 阿尼亚喜欢\n- 像只\n" +
			"- 我永远喜欢|永远喜欢\n- 像样的亲亲\n- 国旗\n- 不要靠近\n- 万能表情|空白表情\n- 采访\n- 需要|你可能需要\n- 这像画吗\n- 小画家\n" +
			"- 完美\n- 玩游戏\n- 出警\n- 警察\n- 舔|舔屏|prpr\n- 安全感\n- 精神支柱\n- 想什么\n- 墙纸\n- 为什么at我\n- 交个朋友\n- 打工人|继续干活\n" +
			"- 兑换券\n- 注意力涣散\n- 垃圾桶|垃圾\n- 捶\n- 啾啾\n- 2敲\n- 听音乐\n- 永远爱你\n- 2拍\n- 顶\n- 捣\n- 打拳\n- 滚\n- 吸|嗦\n- 扔\n" +
			"- 锤\n- 紧贴|紧紧贴着\n- 转\n",
		PrivateDataFolder: "gif",
	}).ApplySingle(ctxext.DefaultSingle)
	datapath = file.BOTPATH + "/" + en.DataFolder()
	en.OnRegex(`^(` + strings.Join(cmd, "|") + `)[\s\S]*?(\[CQ:(image\,file=([0-9a-zA-Z]{32}).*|at.+?(\d{5,11}))\].*|(\d+))$`).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := newContext(ctx.Event.UserID)
		list := ctx.State["regex_matched"].([]string)
		err := c.prepareLogos(list[4]+list[5]+list[6], strconv.FormatInt(ctx.Event.UserID, 10))
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		argslist := strings.Split(strings.TrimSuffix(strings.TrimPrefix(list[0], list[1]), list[2]), " ")
		args := make([]reflect.Value, len(argslist))
		for i := 0; i < len(argslist); i++ {
			args[i] = reflect.ValueOf(argslist[i])
		}
		r := reflect.ValueOf(c).MethodByName(cmdMap[list[1]]).Call(args)
		picurl := r[0].String()
		if !r[1].IsNil() {
			err = r[1].Interface().(error)
		}
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Image(picurl))
	})
}
