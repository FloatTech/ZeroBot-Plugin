// Package moehu 群友的API, 很好用
package moehu

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/process"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type storage int64

func (s *storage) setDownload(on bool) {
	if on {
		*s |= 0b001
	} else {
		*s &= 0b110
	}
}

func (s *storage) isDownload() bool {
	return *s&0b001 > 0
}

type result struct {
	Pic []string `json:"pic"`
}

const (
	// moehu api
	moehuAPI = "https://img.moehu.org/pic.php?return=json&id="
	ua       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36 Edg/102.0.1245.39"
)

var (
	groupSingle = single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 {
			return ctx.Event.GroupID
		}),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send("等一下，还有操作还未完成哦~")
		}),
	)
	allAPI = map[string]string{
		"兽耳":       moehuAPI + "kemonomimi",
		"白毛":       moehuAPI + "yin",
		"星空":       moehuAPI + "xing",
		"竖屏壁纸":     moehuAPI + "sjpic",
		"横屏壁纸":     moehuAPI + "pc",
		"樱岛麻衣":     moehuAPI + "ydmy",
		"猫羽雫":      moehuAPI + "myn",
		"椎名真白":     moehuAPI + "mashiro",
		"图来":       moehuAPI + "test",
		"狗子":       moehuAPI + "inugami-korone",
		"洛天依":      moehuAPI + "tianyi",
		"小狐狸":      moehuAPI + "fubuki",
		"三蹦子":      moehuAPI + "bh3",
		"mea":      moehuAPI + "kagura-mea",
		"黑猫":       moehuAPI + "gokou-ruri",
		"狗妈":       moehuAPI + "kagura-nana",
		"手机壁纸":     moehuAPI + "sjpic",
		"高清壁纸":     moehuAPI + "gqbz",
		"小鲨鱼":      moehuAPI + "gawr-gura",
		"雪妈":       moehuAPI + "yukihana",
		"马自立":      moehuAPI + "natsuiro",
		"粽子":       moehuAPI + "uruha-rushia",
		"花园猫猫":     moehuAPI + "hanazono-serena",
		"熊猫人":      moehuAPI + "sasaki-saku",
		"小绵羊":      moehuAPI + "tsunomaki-watame",
		"常暗永远":     moehuAPI + "tokoyami-towa",
		"天宫心":      moehuAPI + "amamiya-kokoro",
		"peko":     moehuAPI + "usada-pekora",
		"触手姬":      moehuAPI + "ninomae",
		"樱巫女":      moehuAPI + "sakura-miko",
		"miku":     moehuAPI + "miku",
		"鹿乃":       moehuAPI + "kano",
		"原神":       moehuAPI + "ys",
		"明日方舟":     moehuAPI + "mrfz",
		"碧蓝航线":     moehuAPI + "blhx",
		"车万":       moehuAPI + "dongf",
		"碧蓝档案":     moehuAPI + "blda",
		"赛马娘":      moehuAPI + "saima",
		"表情包":      moehuAPI + "bqb",
		"甘城猫猫":     moehuAPI + "gcmm",
		"mc":       moehuAPI + "mc",
		"kemomimi": moehuAPI + "kemomimi",
		"大神澪":      moehuAPI + "ookami-mio",
		"星川沙拉":     moehuAPI + "sara-hoshikawa",
		"木口EN":     moehuAPI + "holoen",
		"绊爱":       moehuAPI + "kizunaai",
		"少女前线":     moehuAPI + "snqx",
		"缘之空":      moehuAPI + "yzk",
		"鬼灭之刃":     moehuAPI + "gmzr",
		"妖尾":       moehuAPI + "yaowei",
		"re0":      moehuAPI + "re0",
		"sao":      moehuAPI + "sao",
		"萝莉":       moehuAPI + "loli",
		"白丝":       moehuAPI + "acgbs",
		"黑丝":       moehuAPI + "acghs",
		"saber":    moehuAPI + "saber",
		"四系乃":      moehuAPI + "yoshino",
		"见崎鸣":      moehuAPI + "misakimei",
		"阿卡林":      moehuAPI + "akari",
		"康娜":       moehuAPI + "kanna",
		"喵帕斯":      moehuAPI + "miaops",
		"妮姆芙":      moehuAPI + "nymph",
		"诺艾尔":      moehuAPI + "noel",
		"时崎狂三":     moehuAPI + "kurumi",
		"薇尔莉特":     moehuAPI + "violet",
		"忍野忍":      moehuAPI + "shinobu",
		"风见一姬":     moehuAPI + "kazuki",
		"伊莉雅":      moehuAPI + "iliya",
		"五等分的花嫁":   moehuAPI + "5huajia",
		"雷姆":       moehuAPI + "rem",
		"碧翠丝":      moehuAPI + "beatrice",
		"土间埋":      moehuAPI + "umr",
		"阿尼亚":      moehuAPI + "aniya",
		"02":       moehuAPI + "02",
		"约尔":       moehuAPI + "yor",
		"阿波连":      moehuAPI + "aharen",
		"御坂美琴":     moehuAPI + "misaka-mikoto",
		"高木":       moehuAPI + "takagi",
		"西片太太":     moehuAPI + "takagi",
		"唐可可":      moehuAPI + "tangkk",
		"水原千鹤":     moehuAPI + "mizuhara",
		"妮可":       moehuAPI + "nico",
		"智乃":       moehuAPI + "chiro",
		"亚丝娜":      moehuAPI + "asuna",
		"我很好奇":     moehuAPI + "eru",
		"臭鼬":       moehuAPI + "karyl",
		"灰原哀":      moehuAPI + "haibara",
		"龙王的牢饭":    moehuAPI + "hinatsuru",
		"志摩凛":      moehuAPI + "shimarin",
		"六花":       moehuAPI + "rikka",
		"冰菓":       moehuAPI + "bingg",
		"你的名字":     moehuAPI + "kiminame",
		"圣人惠":      moehuAPI + "katoumegumi",
		"谢丝塔":      moehuAPI + "siesta",
		"早坂爱":      moehuAPI + "hayasakaai",
		"凉宫春日":     moehuAPI + "haruhi",
		"藤原千花":     moehuAPI + "chika",
		"弥豆子":      moehuAPI + "nezuko",
		"小野寺":      moehuAPI + "onoderaoosaki",
		"伊蕾娜":      moehuAPI + "elaina",
		"三玖":       moehuAPI + "nakanomiku",
		"四宫辉夜":     moehuAPI + "kaguya",
		"佐天泪子":     moehuAPI + "ruiko",
		"白井黑子":     moehuAPI + "kuroko",
		"公主连接":     moehuAPI + "gongzhulj",
		"泉此方":      moehuAPI + "konata",
		"雪乃":       moehuAPI + "yukino",
		"弥子":       moehuAPI + "linomiko",
		"白银圭":      moehuAPI + "shiroganekei",
		"立华奏":      moehuAPI + "kanade",
		"摇曳露营":     moehuAPI + "camp",
		"摇曳百合":     moehuAPI + "yuruyuri",
		"喵内":       moehuAPI + "miyone",
		"学不来":      moehuAPI + "xuebulai",
		"悠哉日常大王":   moehuAPI + "nobiyori",
		"猫猫":       moehuAPI + "miao",
		"阿夸":       moehuAPI + "aqua",
		"猫宫日向":     moehuAPI + "nekomiya-hinata",
		"喜多川海梦":    moehuAPI + "kitagawa-marin",
		"熊污女":      moehuAPI + "amayadori-machi",
		"助手":       moehuAPI + "makise-kurisu",
		"艾拉":       moehuAPI + "lsla",
		"蝶祈":       moehuAPI + "yuzuriha-inori",
		"伊卡洛斯":     moehuAPI + "uranus-queen",
		"宁宁":       moehuAPI + "yashiro-nene",
		"菲洛":       moehuAPI + "filo",
		"食蜂操祈":     moehuAPI + "shokuho-isaki",
		"我妻由乃":     moehuAPI + "gasai-yuno",
		"长瀞同学":     moehuAPI + "nagatoro-hayase",
		"蜘蛛子":      moehuAPI + "noname-kumo",
		"flag大小姐":  moehuAPI + "flag-ojousama",
		"崛与宫村":     moehuAPI + "hori-to-miyamura",
		"路人女主":     moehuAPI + "saenai-heroine",
		"命运长椅":     moehuAPI + "mydcy",
		"高原魔女":     moehuAPI + "slime-300",
		"幼妻狐仙":     moehuAPI + "fox-senko",
	}
	en = control.Register("moehu", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  true,
		Brief:             "萌虎图片站",
		Help:              "- 随机<数量>张[类型]\n- 示例: 随机10张雷姆\n- 清空[所有|类型]缓存\n- [开启|关闭]使用缓存\n[兽耳|白毛|星空|竖屏壁纸|横屏壁纸|图来|小鲨鱼|雪妈|马自立|粽子|花园猫猫|熊猫人|小绵羊|常暗永远|天宫心|peko|触手姬|大神澪|星川莎拉|樱巫女|木口EN|绊爱|猫羽雫|樱岛麻衣|miku|鹿乃|原神|明日方舟|碧蓝航线|车万|在原七海|赛马娘|碧蓝档案|表情包|甘城猫猫|mc|kemomimi|手机壁纸|妖尾|少女前线|缘之空|鬼灭之刃|re0|sao|狗妈|黑猫|mea|三蹦子|洛天依|小狐狸|狗子|康娜|四系乃|时崎狂三|喵帕斯|见崎鸣|阿卡林|薇尔莉特|诺艾尔|saber|黑丝|白丝|伊莉雅|风见一姬|忍野忍|五等分的花嫁|妮姆芙|碧翠丝|雷姆|02|阿尼亚|高木|西片太太|约尔|御坂美琴|阿波连|土间埋|唐可可|水原千鹤|妮可|智乃|亚丝娜|我很好奇|臭鼬|灰原哀|龙王的牢饭|志摩凛|六花|冰菓|你的名字|圣人惠|谢丝塔|早坂爱|凉宫春日|藤原千花|弥子|祢豆子|小野寺|伊蕾娜|三玖|佐天泪子|白井黑子|公主连接|泉此方|雪乃|白银圭|间谍过家家|立华奏|摇曳露营|摇曳百合|学不来|悠哉日常大王|阿夸|喜多川海梦|flag大小姐|助手|崛与宫村|熊污女|命运长椅|路人女主|伊卡洛斯|蝶祈|艾拉|高原魔女|宁宁|猫宫日向|菲洛|食蜂操祈|我妻由乃|幼妻狐仙|长瀞同学|蜘蛛子] ",
		PrivateDataFolder: "moehu",
	}).ApplySingle(groupSingle)
	filepath = en.DataFolder()
)

func init() {
	en.OnRegex(`^随机(([0-9]+)[份|张])?(.*)`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				return
			}
			msg := ctx.State["regex_matched"].([]string)[3]
			api, ok := allAPI[msg]
			if !ok {
				return
			}
			i := math.Str2Int64(ctx.State["regex_matched"].([]string)[2])
			switch {
			case i == 0:
				i = 1
			case !zero.AdminPermission(ctx) && i > 15:
				i = 15
				ctx.SendChain(message.Text("普通成员最多只能随机15张图片哦~"))
			case !zero.SuperUserPermission(ctx) && i > 30:
				i = 30
				ctx.SendChain(message.Text("管理员最多只能随机30张图片哦~"))
			case zero.SuperUserPermission(ctx) && i > 100:
				i = 100
				ctx.SendChain(message.Text("太贪心啦！最多只能随机100张图片哦~"))
			default:
				ctx.SendChain(message.Text("少女祈祷中..."))
			}
			data, err := web.RequestDataWith(web.NewDefaultClient(), api+"&num="+strconv.FormatInt(i, 10), "GET", "", ua, nil)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			var r result
			err = json.Unmarshal(data, &r)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			m := make(message.Message, len(r.Pic))
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			gdata := (storage)(c.GetData(gid))
			var wg sync.WaitGroup
			if gdata.isDownload() {
				_ = os.Mkdir(file.BOTPATH+"/"+filepath+msg, 0664)
				md5 := md5.New()
				wg.Add(len(r.Pic))
				for i, v := range r.Pic {
					go func(i int, v string) {
						defer wg.Done()
						_, err = md5.Write(binary.StringToBytes(v))
						if err != nil {
							return
						}
						name := hex.EncodeToString(md5.Sum(nil))[:8] + ".jpg"
						filename := file.BOTPATH + "/" + filepath + msg + "/" + name
						if file.IsNotExist(filename) {
							err = file.DownloadTo(v, filename)
							if err != nil {
								return
							}
							m[i] = ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+filename))
							process.SleepAbout1sTo2s()
							return
						}
						m[i] = ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+filename))
					}(i, v)
				}
			} else {
				for i, v := range r.Pic {
					m[i] = ctxext.FakeSenderForwardNode(ctx, message.Image(v))
				}
			}
			wg.Wait()
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, m).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	en.OnRegex(`^清空(.*)缓存`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			rm := ctx.State["regex_matched"].([]string)[1]
			if rm == "所有" {
				if err := os.RemoveAll(file.BOTPATH + "/" + filepath); err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				_ = os.Mkdir(file.BOTPATH+"/"+filepath, 0664)
				ctx.SendChain(message.Text("清空所有缓存成功~"))
				return
			}
			_, ok := allAPI[rm]
			if !ok {
				return
			}
			if err := os.RemoveAll(file.BOTPATH + "/" + filepath + rm); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("清空", rm, "缓存成功~"))
		})
	en.OnRegex(`^(.*)使用缓存$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				return
			}
			option := ctx.State["regex_matched"].([]string)[1]
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			gdata := (storage)(c.GetData(gid))
			switch option {
			case "开启", "打开", "启用":
				gdata.setDownload(true)
			case "关闭", "关掉", "禁用":
				gdata.setDownload(false)
			default:
				return
			}
			ctx.SendChain(message.Text("已设置图片模式为缓存" + option))
		})
}
