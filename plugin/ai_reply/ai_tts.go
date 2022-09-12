package aireply

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/pkumza/numcn"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/aireply"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

const (
	ttsServiceName = "tts"
	cnapi          = "http://233366.proxy.nscc-gz.cn:8888?speaker=%s&text=%s"
	// testString     = "这是测试语言......"
)

// 每个角色的测试文案
var testRecord = map[string]string{
	"派蒙":    "哎，又是看不懂的东西。我完全不知道这些奇怪的问题和实验，能得到什么结果…",
	"凯亚":    "真是个急性子啊你。",
	"安柏":    "最初的鸟儿是不会飞翔的，飞翔是它们勇敢跃入峡谷的奖励。",
	"丽莎":    "嗨，小可爱，你是新来的助理吗？",
	"琴":     "蒲公英骑士，琴，申请入队。",
	"香菱":    "我是来自璃月的厨师香菱，最擅长的是做各种捞…捞，料理…哎呀，练了那么多次，还是会紧张，嘿。",
	"枫原万叶":  "飘摇风雨中，带刀归来赤脚行。",
	"迪卢克":   "在黎明来临之前，总要有人照亮黑暗。",
	"温迪":    "若你困于无风之地，我将为你奏响高天之歌。",
	"可莉":    "西风骑士团，火花骑士，可莉，前来报到！…呃—后面该说什么词来着？可莉背不下来啦...",
	"早柚":    "终末番,早柚，参上。 呼——",
	"托马":    "初次见面，异乡的旅人，你的名字我可是早就听说了。只要你不嫌弃，我托马，从今天起就是你的朋友了。",
	"芭芭拉":   "芭芭拉，闪耀登场~治疗就交给我吧，不会让你失望的！",
	"优菈":    "沉沦是很容易的一件事，但我仍想冻住这股潮流。",
	"云堇":    "曲高未必人不识，自有知音和清词。",
	"钟离":    "人间归离复归离，借一浮生逃浮生。",
	"魈":     "三眼五显仙人，魈，听召，前来守护",
	"凝光":    "就算古玩价值连城，给人的快乐，也只有刚拥有的一瞬",
	"雷电将军":  "浮世千百年来风景依旧，人之在世却如白露与泡影。",
	"北斗":    "不知道如何向前的话，总之先迈出第一步，后面的道路就会自然而然地展开了。",
	"甘雨":    "这项工作，该划掉了。",
	"七七":    "椰羊的奶，好喝!比一般的羊奶，好喝!",
	"刻晴":    "劳逸结合是不错，但也别放松过头。",
	"神里绫华":  "若知是梦何须醒，不比真如一相会。",
	"雷泽":    "你是朋友。我和你一起狩猎。",
	"神里绫人":  "此前听绫华屡次提起阁下，不料公务繁忙，直至今日才有机会相见。",
	"罗莎莉亚":  "哪怕如今你已经走上截然不同的道路，也不要否认从前的自己，从前的每一个你都是你脚下的基石，不要害怕过去，不要畏惧与它抗衡。",
	"阿贝多":   "用自己的双脚丈量土地，将未知变为知识。",
	"八重神子":  "我的神明，就托付给你了。",
	"宵宫":    "即使只是片刻的火花，也能在仰望黑夜的人心中留下久久不灭的美丽光芒。",
	"荒泷一斗":  "更好地活下去,绝不该靠牺牲同类换取，应该是,一起更好地活着,才对。",
	"九条裟罗":  "想要留住雪花。但在手心里，它只会融化的更快。",
	"夜兰":    "线人来信了，嗯，看来又出现了新的变数。",
	"珊瑚宫心海": "成为了现任人神巫女之后，我也慢慢习惯了这样的生活，更重要的是我也因此和你相遇了，不是吗？",
	"五郎":    "海祇岛反抗军大将，五郎，前来助阵！",
	"达达利亚":  "许下的诺言就好好遵守，做错了事情就承担责任，这才是家人应有的样子吧。",
	"莫娜":    "正是因为无法更改，无可违逆，只能接受，命运才会被称之为命运。",
	"班尼特":   "只要有大家在，伤口就不会痛!",
	"申鹤":    "不知道你是喜欢人间的灯火，还是山林的月光？",
	"行秋":    "有时明月无人夜，独向昭潭制恶龙。",
	"烟绯":    "律法即是约束，也是工具。",
	"久岐忍":   "有麻烦事要处理的话，直接告诉我就好，我来摆平。",
	"辛焱":    "马上就要演出了，你也一起来嗨吗？",
	"砂糖":    "我是砂糖，炼金术的…研究员。",
	"胡桃":    "阴阳有序，命运无常，死亡难以预测，却也有它的规矩。",
	"重云":    "我名重云，家族久居璃月，世代以驱邪除魔为业。",
	"菲谢尔":   "我即断罪之皇女，真名为菲谢尔。应命运的召唤降临在此间——哎？你也是，异世界的旅人吗…？",
	"诺艾尔":   "我是诺艾尔，西风骑士团的女仆，从今天起会陪你一起去冒险。",
	"迪奥娜":   "猫尾酒馆的招牌调酒师，迪奥娜，我的出场费可是很贵的。",
	"鹿野院平藏": "我叫鹿野院平藏，是天领奉行里破案最多最快的侦探……",
}

var (
	re        = regexp.MustCompile(`(\-|\+)?\d+(\.\d+)?`)
	soundList = [...]string{
		"派蒙", "凯亚", "安柏", "丽莎", "琴",
		"香菱", "枫原万叶", "迪卢克", "温迪", "可莉",
		"早柚", "托马", "芭芭拉", "优菈", "云堇",
		"钟离", "魈", "凝光", "雷电将军", "北斗",
		"甘雨", "七七", "刻晴", "神里绫华", "雷泽",
		"神里绫人", "罗莎莉亚", "阿贝多", "八重神子", "宵宫",
		"荒泷一斗", "九条裟罗", "夜兰", "珊瑚宫心海", "五郎",
		"达达利亚", "莫娜", "班尼特", "申鹤", "行秋",
		"烟绯", "久岐忍", "辛焱", "砂糖", "胡桃",
		"重云", "菲谢尔", "诺艾尔", "迪奥娜", "鹿野院平藏",
	}
)

type ttsInstances struct {
	sync.RWMutex
	defaultSoundMode string
	soundMode        []string
}

func list(list []string, num int) string {
	s := ""
	for i, value := range list {
		s += value
		if (i+1)%num == 0 {
			s += "\n"
		} else {
			s += " | "
		}
	}
	return s
}

func init() {
	tts := &ttsInstances{
		defaultSoundMode: "派蒙",
		soundMode:        soundList[:],
	}
	engine := control.Register(ttsServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help: "语音回复(大家一起来炼丹)\n" +
			"- @Bot 任意文本(任意一句话回复)\n" +
			"- 设置语音模式[原神人物]\n" +
			"- 设置默认语音模式[原神人物]\n" +
			"- 恢复成默认语音模式\n" +
			"当前适用的原神人物含有以下：\n" + list(soundList[:], 5),
	})
	tts.RLock()
	defer tts.RUnlock()
	m, ok := control.Lookup(ttsServiceName)
	if ok {
		tts.defaultSoundMode = soundList[m.GetData(-2905)]
	}
	engine.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			// 获取回复模式
			r := aireply.NewAIReply(getReplyMode(ctx))
			// 获取回复的文本
			reply := r.TalkPlain(msg, zero.BotConfig.NickName[0])
			// 获取角色
			name := tts.getSoundMode(ctx)
			if _, ok := testRecord[name]; !ok {
				ctx.SendChain(message.Text("配置的语言人物数据丢失！请重新设置语言人物。"))
				return
			}
			// 获取语言
			record := message.Record(fmt.Sprintf(cnapi, url.QueryEscape(name), url.QueryEscape(
				// 将数字转文字
				re.ReplaceAllStringFunc(reply, func(s string) string {
					f, err := strconv.ParseFloat(s, 64)
					if err != nil {
						log.Errorln("[tts]:", err)
						return s
					}
					return numcn.EncodeFromFloat64(f)
				}),
			))).Add("cache", 0)
			if record.Data == nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
				return
			}
			// 发送语音
			if ID := ctx.SendChain(record); ID.ID() == 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
			}
		})
	engine.OnRegex(`^设置语音模式(.*)$`, zero.AdminPermission, func(ctx *zero.Ctx) bool {
		param := ctx.State["regex_matched"].([]string)[1]
		if _, ok := testRecord[param]; !ok {
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		param := ctx.State["regex_matched"].([]string)[1]
		// 保存设置
		err := tts.setSoundMode(ctx, param)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		// 设置验证
		name := tts.getSoundMode(ctx)
		if _, ok := testRecord[name]; !ok {
			ctx.SendChain(message.Text("配置的语言人物数据丢失！请重新设置语言人物。"))
			return
		}
		record := message.Record(fmt.Sprintf(cnapi, url.QueryEscape(name), url.QueryEscape(testRecord[name]))).Add("cache", 0)
		if ID := ctx.SendChain(record); ID.ID() == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置失败！无法发送测试语言，请重试。"))
			return
		}
		time.Sleep(time.Second * 2)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，默认模式为", name))
	})
	engine.OnRegex(`^设置默认语音模式(.*)$`, zero.SuperUserPermission, func(ctx *zero.Ctx) bool {
		param := ctx.State["regex_matched"].([]string)[1]
		if _, ok := testRecord[param]; !ok {
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		param := ctx.State["regex_matched"].([]string)[1]
		// 保存设置
		err := tts.setDefaultSoundMode(param)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		// 设置验证
		name := tts.defaultSoundMode
		record := message.Record(fmt.Sprintf(cnapi, url.QueryEscape(name), url.QueryEscape(testRecord[name]))).Add("cache", 0)
		if ID := ctx.SendChain(record); ID.ID() == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置失败！无法发送测试语言，请重试。"))
			return
		}
		time.Sleep(time.Second * 2)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，默认模式为", name))
	})
	engine.OnFullMatch("恢复成默认语音模式", zero.AdminPermission, func(ctx *zero.Ctx) bool {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		m, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if ok {
			tts.RLock()
			defer tts.RUnlock()
			index := m.GetData(gid)
			if int(index) < len(tts.soundMode) {
				return true
			}
		}
		return false
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		err := m.SetData(gid, 255)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		// 设置验证
		name := tts.getSoundMode(ctx)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，当前默认语音模式为", name))
	})
}

func (tts *ttsInstances) setSoundMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var index int64
	tts.RLock()
	for i, s := range tts.soundMode {
		if s == name {
			index = int64(i)
			break
		}
	}
	tts.RUnlock()
	m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	return m.SetData(gid, index)
}

func (tts *ttsInstances) getSoundMode(ctx *zero.Ctx) (name string) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if ok {
		tts.RLock()
		defer tts.RUnlock()
		index := m.GetData(gid)
		if int(index) < len(tts.soundMode) {
			return tts.soundMode[index]
		}
	}
	return tts.defaultSoundMode
}

func (tts *ttsInstances) setDefaultSoundMode(name string) error {
	var index int64
	tts.RLock()
	for i, s := range tts.soundMode {
		if s == name {
			index = int64(i)
			break
		}
	}
	tts.RUnlock()
	m, ok := control.Lookup(ttsServiceName)
	if !ok {
		return errors.New("[tts]service no fund ")
	}
	err := m.SetData(-2905, index)
	if err == nil {
		soundMode := tts.soundMode[m.GetData(-2905)]
		if soundMode != tts.soundMode[index] {
			return errors.New("[tts]检验数据失败,当前写入的数据为" + soundMode)
		}
		tts.defaultSoundMode = soundMode
	}
	return err
}
