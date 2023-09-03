// Package gif 制图
package gif

import (
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	cmd      = make([]string, 0)
	datapath string
	cmdMap   = map[string]func(cc *context, args ...string) (string, error){
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
		"蒙蔽":     mengbi,
		"踩":      cai,
		"好玩":     haowan,
		"2转":     whirl,
		"2滚":     push,
		"踢球":     tiqiu,
		"2舔":     lick,
		"可莉吃":    klee,
		"胡桃啃":    hutaoken,
		"怀":      huai,
		"砰":      peng,
		"你犯法了":   fanfa,
		"炖":      dun,
		"2蹭":     ceng2,
		"诶嘿":     eihei,
		"膜拜":     worship,
		"吞":      ci,
		"揍":      zou,
		"给我变":    bian,
		"玩一下":    van,
		"不要看":    neko,
		"小天使":    xiaotianshi,
		"你的":     youer,
		"我老婆":    nowife,
		"远离":     yuanli,
		"抬棺":     taiguan,
		"一直":     alwaysDoGif,
	}
)

func init() { // 插件主体
	for k := range cmdMap {
		cmd = append(cmd, k)
	}
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "制图",
		Help: "下为制图命令:\n" +
			"- 搓|- 冲|- 摸|-拍|- 丢|- 吃|- 敲|- 啃|- 蹭|- 爬|- 撕\n" +
			"- 吸|- 嗦|- 扔|- 锤|- 紧贴|紧紧贴着|- 转|- 抬棺|- 远离\n" +
			"- 揍|- 吞|- 膜拜|- 诶嘿|- 2蹭|- 你犯法了|- 砰|- 注意力涣散\n" +
			"- 2敲|- 听音乐|- 永远爱你|- 2拍|- 顶|- 捣|- 打拳|- 滚\n" +
			"- 灰度|- 上翻|- 下翻|- 左翻|- 右翻|- 反色|- 浮雕|- 打码\n" +
			"- 负片|- 旋转|- 变形|- 亲|- 结婚申请|结婚登记|- 阿尼亚喜欢XXX\n" +
			"- 像只|- 我永远喜欢XXX|- 像样的亲亲|- 国旗|- 不要靠近\n" +
			"- 蒙蔽|- 踩|- 好玩|- 2转|- 踢球|- 2舔|- 可莉吃|- 胡桃啃|- 怀\n" +
			"- 小画家|- 完美|- 玩游戏|- 出警|- 警察|- 舔|舔屏|prpr\n" +
			"- 安全感|- 精神支柱|- 想什么|- 墙纸|- 为什么at我|- 交个朋友\n" +
			"- 打工人|- 继续干活|- 兑换券|- 炖|- 垃圾桶|- 垃圾|- 捶|- 啾啾\n" +
			"- 我老婆|- 小天使XXX|- 你的XXX|- 不要看|- 玩一下XXX|- 给我变\n" +
			"- 万能表情|- 空白表情|- 采访|- 需要|- 你可能需要|- 这像画吗\n" +
			"- 一直(支持动图)\n" +
			"例: 制图命令XXX[@用户|QQ号|图片]\n" +
			"Tips: XXX可以为限制长度的任何文字\n" +
			"对Bot使用为 @Bot制图命令[XXX]@Bot",
		PrivateDataFolder: "gif",
	}).ApplySingle(ctxext.DefaultSingle)
	datapath = file.BOTPATH + "/" + en.DataFolder()
	en.OnRegex(`^(` + strings.Join(cmd, "|") + `)[\s\S]*?(\[CQ:(image\,file=([0-9a-zA-Z]{32}).*|at.+?(\d{5,11}))\].*|(\d+))$`).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := newContext(ctx.Event.UserID)
		list := ctx.State["regex_matched"].([]string)
		err := c.prepareLogos(list[4]+list[5]+list[6], strconv.FormatInt(ctx.Event.UserID, 10))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		argslist := strings.Split(strings.TrimSuffix(strings.TrimPrefix(list[0], list[1]), list[2]), " ")
		picurl, err := cmdMap[list[1]](c, argslist...)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Image(picurl))
	})
}
