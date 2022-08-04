// Package gif 制图
package gif

import (
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
	cmdMap   = map[string]func(cc *context, args ...string) (string, error){
		"炖":      dun,
		"2蹭":     ceng2,
		"诶嘿":     eihei,
		"膜拜":     mobai,
		"吞":      tun,
		"揍":      zou,
		"给我变":    bian,
		"玩一下":    van,
		"不要看":    neko,
		"小天使":    xiaotianshi,
		"你的":     youer,
		"我老婆":    nowife,
		"远离":     yuanli,
		"抬棺":     taiguan,
		"搓":      cuo,
		"冲":      xqe,
		"摸":      mo,
		"拍":      pai,
		"丢":      diu,
		"吃":      chi,
		"敲":      qiao,
		"啃":      ken,
		"蹭":      ceng,
		"爬":      pa,
		"撕":      si,
		"灰度":     grayscale,
		"上翻":     flipV,
		"下翻":     flipV,
		"左翻":     flipH,
		"右翻":     flipH,
		"反色":     invert,
		"浮雕":     convolve3x3,
		"打码":     blur,
		"负片":     invertAndGrayscale,
		"旋转":     rotate,
		"变形":     deformation,
		"亲":      kiss,
		"结婚申请":   marriage,
		"结婚登记":   marriage,
		"阿尼亚喜欢":  anyasuki,
		"像只":     alike,
		"我永远喜欢":  alwaysLike,
		"永远喜欢":   alwaysLike,
		"像样的亲亲":  decentKiss,
		"国旗":     chinaFlag,
		"不要靠近":   dontTouch,
		"万能表情":   universal,
		"空白表情":   universal,
		"采访":     interview,
		"需要":     need,
		"你可能需要":  need,
		"这像画吗":   paint,
		"小画家":    painter,
		"完美":     perfect,
		"玩游戏":    playGame,
		"出警":     police,
		"警察":     police1,
		"舔":      prpr,
		"舔屏":     prpr,
		"prpr":   prpr,
		"安全感":    safeSense,
		"精神支柱":   support,
		"想什么":    thinkwhat,
		"墙纸":     wallpaper,
		"为什么at我": whyatme,
		"交个朋友":   makeFriend,
		"打工人":    backToWork,
		"继续干活":   backToWork,
		"兑换券":    coupon,
		"注意力涣散":  distracted,
		"垃圾桶":    garbage,
		"垃圾":     garbage,
		"捶":      thump,
		"啾啾":     jiujiu,
		"2敲":     knock,
		"听音乐":    listenMusic,
		"永远爱你":   loveYou,
		"2拍":     pat,
		"顶":      jackUp,
		"捣":      pound,
		"打拳":     punch,
		"滚":      roll,
		"吸":      suck,
		"嗦":      suck,
		"扔":      throw,
		"锤":      hammer,
		"紧贴":     tightly,
		"紧紧贴着":   tightly,
		"转":      turn,
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
		picurl, err := cmdMap[list[1]](c, argslist...)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Image(picurl))
	})
}
