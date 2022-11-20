package baidufanyi

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

// config内容
type config struct {
	APPID     string `json:"APP ID"`
	Secretkey string `json:"密钥"`
	Maxchar   int    `json:"免费调用量(字符/月)"`
}

var (
	cfg  config
	lang = map[string]string{
		// A
		"阿拉伯语": "ara", "爱尔兰语": "gle", "奥克语": "oci", "阿尔巴尼亚语": "alb", "阿尔及利亚阿拉伯语": "arq",
		"阿肯语": "aka", "阿拉贡语": "arg", "阿姆哈拉语": "amh", "阿萨姆语": "asm", "艾马拉语": "aym",
		"阿塞拜疆语": "aze", "阿斯图里亚斯语": "ast", "奥塞梯语	": "oss", "爱沙尼亚语": "est", "奥杰布瓦语": "oji",
		"奥里亚语": "ori", "奥罗莫语": "orm",
		// B
		"波兰语": "pl", "波斯语": "per", "布列塔尼语": "bre", "巴什基尔语": "bak", "巴斯克语": "baq",
		"巴西葡萄牙语": "pot", "白俄罗斯语": "bel", "柏柏尔语": "ber", "邦板牙语": "pam", "保加利亚语": "bul",
		"北方萨米语": "sme", "北索托语": "ped", "本巴语": "bem", "比林语": "bli", "比斯拉马语": "bis",
		"俾路支语": "bal", "冰岛语": "ice", "波斯尼亚语": "bos", "博杰普尔语": "bho",
		// C
		"楚瓦什语": "chv", "聪加语": "tso",
		// D
		"丹麦语": "dan", "德语": "de", "鞑靼语": "tat", "掸语": "sha", "德顿语": "tet",
		"迪维希语": "div", "低地德语": "log",
		// E
		"俄语": "ru",
		// F
		"法语": "fra", "菲律宾语": "fil", "芬兰语": "fin", "梵语": "san", "弗留利语": "fri",
		"富拉尼语": "ful", "法罗语": "fao",
		// G
		"盖尔语": "gla", "刚果语": "kon", "高地索布语": "ups", "高棉语": "hkm", "格陵兰语": "kal",
		"格鲁吉亚语": "geo", "古吉拉特语": "guj", "古希腊语": "gra", "古英语": "eno", "瓜拉尼语": "grn",
		// H
		"韩语": "kor", "荷兰语": "nl", "胡帕语": "hup", "哈卡钦语": "hak", "海地语": "ht",
		"黑山语": "mot", "豪萨语": "hau",
		// J
		"吉尔吉斯语": "kir", "加利西亚语": "glg", "加拿大法语": "frn", "加泰罗尼亚语": "cat", "捷克语": "cs",
		// K
		"卡拜尔语": "kab", "卡纳达语": "kan", "卡努里语": "kau", "卡舒比语": "kah", "康瓦尔语": "cor",
		"科萨语": "xho", "科西嘉语": "cos", "克里克语": "cre", "克里米亚鞑靼语": "cri", "克林贡语": "kli",
		"克罗地亚语": "hrv", "克丘亚语": "que", "克什米尔语": "kas", "孔卡尼语": "kok", "库尔德语": "kur",
		// L
		"拉丁语": "lat", "老挝语": "lao", "罗马尼亚语": "rom", "拉特加莱语": "lag", "拉脱维亚语": "lav",
		"林堡语": "lim", "林加拉语": "lin", "卢干达语": "lug", "卢森堡语": "ltz", "卢森尼亚语": "ruy",
		"卢旺达语": "kin", "立陶宛语": "lit", "罗曼什语": "roh", "罗姆语": "ro", "逻辑语": "loj",
		// M
		"马来语": "may", "缅甸语": "bur", "马拉地语": "mar", "马拉加斯语": "mg", "马拉雅拉姆语": "mal",
		"马其顿语": "mac", "马绍尔语": "mah", "迈蒂利语": "mai", "曼克斯语": "glv", "毛里求斯克里奥尔语": "mau",
		"毛利语": "mao", "孟加拉语": "ben", "马耳他语": "mlt", "苗语": "hmn",
		// N
		"挪威语": "nor", "那不勒斯语": "nea", "南恩德贝莱语": "nbl", "南非荷兰语": "afr", "南索托语": "sot",
		"尼泊尔语": "nep",
		// P
		"葡萄牙语": "pt", "旁遮普语": "pan", "帕皮阿门托语": "pap", "普什图语": "pus",
		// Q
		"齐切瓦语": "nya", "契维语": "twi", "切罗基语": "chr",
		// R
		"日语": "jp", "瑞典语": "swe",
		// S
		"萨丁尼亚语": "srd", "萨摩亚语": "sm", "塞尔维亚-克罗地亚语": "sec", "塞尔维亚语": "srp", "桑海语": "sol",
		"僧伽罗语": "sin", "世界语": "epo", "书面挪威语": "nob", "斯洛伐克语": "sk", "斯洛文尼亚语": "slo",
		"斯瓦希里语": "swa", "塞尔维亚语（西里尔）": "src", "索马里语": "som",
		// T
		"泰语": "th", "土耳其语": "tr", "塔吉克语": "tgk", "泰米尔语": "tam", "他加禄语": "tgl",
		"提格利尼亚语": "tir", "泰卢固语": "tel", "突尼斯阿拉伯语": "tua", "土库曼语": "tuk",
		// W
		"乌克兰语": "ukr", "瓦隆语": "wln", "威尔士语": "wel", "文达语": "ven", "沃洛夫语": "wol", "乌尔都语": "urd",
		// X
		"西班牙语": "spa", "希伯来语": "heb", "希腊语": "el", "匈牙利语": "hu", "西弗里斯语": "fry",
		"西里西亚语": "sil", "希利盖农语": "hil", "下索布语": "los", "夏威夷语": "haw", "新挪威语": "nno",
		"西非书面语": "nqo", "信德语": "snd", "修纳语": "sna", "宿务语": "ceb", "叙利亚语": "syr",
		"巽他语": "sun",
		// Y
		"英语": "en", "印地语": "hi", "印尼语": "id", "意大利语": "it", "越南语": "vie",
		"意第绪语": "yid", "因特语": "ina", "亚齐语": "ach", "印古什语": "ing", "伊博语": "ibo",
		"伊多语": "ido", "约鲁巴语": "yor", "亚美尼亚语": "arm", "伊努克提图特语": "iku", "伊朗语": "ir",
		// Z
		"中文(简体)": "zh", "简中": "zh", "简体中文": "zh", "中文(文言文)": "wyw", "中文(粤语)": "yue",
		"中文(繁体)": "cht", "繁中": "cht", "繁体中文": "cht", "文言文": "wyw", "粤语": "yue",
		"祖鲁语": "zul", "爪哇语": "jav", "中古法语": "frm", "扎扎其语": "zaz",
	}
)

func init() {
	engine := control.Register("baidufanyi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:           "百度翻译(支持回复翻译)",
		PrivateDataFolder: "baidufanyi",
		Help: 
			"-[/][从某语言]翻译[到某语言] [翻译的内容]\n" +
			"[]内容表示可选项\n但若不是回复翻译，翻译的内容不可少。\n" +
			"若是回复翻译，'/'符号不可少。\n例：\n" +
			"/翻译 hello the world\n" +
			"从英语翻译 hello the world\n" +
			"/翻译成中文 hello the world\n" +
			"从英语翻译成中文 hello the world",
	}).ApplySingle(ctxext.DefaultSingle)
	// 获取用户的配置
	cfgFile := engine.DataFolder() + "config.json"
	if file.IsExist(cfgFile) {
		reader, err := os.Open(cfgFile)
		if err == nil {
			err = json.NewDecoder(reader).Decode(&cfg)
		}
		if err != nil {
			panic("[baidufanyi]" + err.Error())
		}
		err = reader.Close()
		if err != nil {
			panic("[baidufanyi]" + err.Error())
		}
	} else {
		err := saveConfig(cfgFile)
		if err != nil {
			panic("[baidufanyi]" + err.Error())
		}
	}
	engine.OnRegex(`^\/?(从(\S+))?翻译(((到|成)(\S+))?\s)?(.+)`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		slang := ctx.State["regex_matched"].([]string)[2]
		tlang := ctx.State["regex_matched"].([]string)[6]
		txt := ctx.State["regex_matched"].([]string)[7]
		//txt := strings.ReplaceAll(ctx.State["regex_matched"].([]string)[7], "\n", "\"\n\"")
		switch {
		case txt == "":
			ctx.SendChain(message.Text("请输入翻译的内容"))
			return
		case strings.Contains(txt, "[CQ:"):
			ctx.SendChain(message.Text("仅支持文字翻译"))
			return
		}
		formlang := "auto"
		for key, value := range lang {
			switch slang {
			case key:
				formlang = value
			case value:
				formlang = key
			}
			if formlang != "auto" {
				break
			}
		}
		tolang := ""
		for key, value := range lang {
			switch tlang {
			case key:
				tolang = value
			case value:
				tolang = key
			}
			if tolang != "" {
				break
			}
		}
		if tolang == "" {
			var count int
			for _, v := range txt {
				if unicode.Is(unicode.Scripts["Han"], v) {
					count++
				}
			}
			if count > len([]rune(txt))/2 {
				tolang = "en"
			} else {
				tolang = "zh"
			}
		}
		slang, tlang, translated, err := translate(txt, formlang, tolang)
		if err != nil {
			ctx.SendChain(message.Text(err))
		} else {
			cfg.Maxchar -= len([]rune(txt))
			err := saveConfig(cfgFile)
			if err != nil {
				ctx.SendChain(message.Text("[baidufanyi]", err))
			}
			formlang := ""
			getlang := ""
			for key, value := range lang {
				if slang == value {
					formlang = key
				}
				if tlang == value {
					getlang = key
				}
				if formlang != "" && getlang != "" {
					break
				}
			}
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("尝试从", formlang, "翻译成", getlang, "：\n", translated)))
		}
	})
	engine.OnRegex(`^\[CQ:reply,id=.*](\s+)?\/(从(\S+))?翻译((到|成)(\S+))?`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		slang := ctx.State["regex_matched"].([]string)[3]
		tlang := ctx.State["regex_matched"].([]string)[6]
		txt := ctx.GetMessage(message.NewMessageIDFromString(ctx.Event.Message[0].Data["id"])).Elements[0].Data["text"]
		//txt = strings.ReplaceAll(txt, "\n", "\"\n\"")
		if txt == "" {
			ctx.SendChain(message.Text("仅支持文字翻译"))
			return
		}
		formlang := "auto"
		for key, value := range lang {
			switch slang {
			case key:
				formlang = value
			case value:
				formlang = key
			}
			if formlang != "auto" {
				break
			}
		}
		tolang := ""
		for key, value := range lang {
			switch tlang {
			case key:
				tolang = value
			case value:
				tolang = key
			}
			if tolang != "" {
				break
			}
		}
		if tolang == "" {
			var count int
			for _, v := range txt {
				if unicode.Is(unicode.Scripts["Han"], v) {
					count++
				}
			}
			if count > len([]rune(txt))/2 {
				tolang = "en"
			} else {
				tolang = "zh"
			}
		}
		slang, tlang, translated, err := translate(txt, formlang, tolang)
		if err != nil {
			ctx.SendChain(message.Text(err))
		} else {
			cfg.Maxchar -= len([]rune(txt))
			err := saveConfig(cfgFile)
			if err != nil {
				ctx.SendChain(message.Text("[baidufanyi]", err))
			}
			formlang := ""
			getlang := ""
			for key, value := range lang {
				if slang == value {
					formlang = key
				}
				if tlang == value {
					getlang = key
				}
				if formlang != "" && getlang != "" {
					break
				}
			}
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("尝试从", formlang, "翻译成", getlang, "：\n", translated)))
		}
	})
	engine.OnRegex(`^设置百度翻译key\s(.*[^\s$])\s(.+)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			cfg.APPID = ctx.State["regex_matched"].([]string)[1]
			cfg.Secretkey = ctx.State["regex_matched"].([]string)[2]
			cfg.Maxchar = 50000
			if err := saveConfig(cfgFile); err != nil {
				ctx.SendChain(message.Text(err))
			} else {
				ctx.SendChain(message.Text("成功！"))
			}
		})
}

// 保存用户配置
func saveConfig(cfgFile string) error {
	if reader, err := os.Create(cfgFile); err == nil {
		err = json.NewEncoder(reader).Encode(&cfg)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

type translation struct {
	From        string `json:"from"`
	To          string `json:"to"`
	TransResult []struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	} `json:"trans_result"`
}

func translate(txt, sl, tl string) (from, to, translated string, err error) {
	lastChar := cfg.Maxchar - len([]rune(txt))
	if lastChar <= 0 {
		err = errors.New("API翻译字符超过了本月免费调用字符数!你还可以翻译" + strconv.Itoa(cfg.Maxchar) + "个字符")
		return
	}
	now := strconv.FormatInt(time.Now().Unix(), 10)
	signinfo := cfg.APPID + txt + now + cfg.Secretkey
	m := fmt.Sprintf("%x", md5.Sum(helper.StringToBytes(signinfo)))
	requestURL := "http://api.fanyi.baidu.com/api/trans/vip/translate?q=" + url.QueryEscape(txt) +
		"&from=" + sl + "&to=" + tl + "&appid=" + cfg.APPID + "&salt=" + now + "&sign=" + m
	data, err := web.GetData(requestURL)
	if err != nil {
		return
	}
	var parsed translation
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	var result []string
	for _, transResult := range parsed.TransResult {
		result = append(result, transResult.Dst)
	}
	return parsed.From, parsed.To, strings.Join(result, ","), nil
}
