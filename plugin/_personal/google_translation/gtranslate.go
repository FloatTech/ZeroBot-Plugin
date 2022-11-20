package gtranslate

import (
	"net/url"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/PuerkitoBio/goquery"
)

var lang = map[string]string{
	"af":    "南非荷兰语",   // "afrikaans",
	"sq":    "阿尔巴尼亚语",  // "albanian",
	"am":    "阿姆哈拉语",   // "amharic",
	"ar":    "阿拉伯语",    // "arabic",
	"hy":    "亚美尼亚语",   // "armenian",
	"az":    "阿塞拜疆语",   // "azerbaijani",
	"eu":    "巴斯克语",    // "basque",
	"be":    "白俄罗斯语",   // "belarusian",
	"bn":    "孟加拉语",    // "bengali",
	"bs":    "波斯尼亚语",   // "bosnian",
	"bg":    "保加利亚语",   // "bulgarian",
	"ca":    "加泰罗尼亚语",  // "catalan",
	"ceb":   "宿务语",     // "cebuano",
	"ny":    "齐切瓦语",    // "chichewa",
	"中文":    "zh-CN",   // "chinese (simplified)",
	"汉语":    "zh-CN",   // "chinese (simplified)",
	"简中":    "zh-CN",   // "chinese (simplified)",
	"繁中":    "zh-TW",   // "chinese (traditional)",
	"zh-CN": "简体中文",    // "chinese (simplified)",
	"zh-TW": "繁体中文",    // "chinese (traditional)",
	"zh-cn": "简体中文",    // "chinese (simplified)",
	"zh-tw": "繁体中文",    // "chinese (traditional)",
	"co":    "科西嘉语",    // "corsican",
	"hr":    "克罗地亚语",   // "croatian",
	"cs":    "捷克语",     // "czech",
	"da":    "丹麦语",     // "danish",
	"nl":    "荷兰语",     // "dutch",
	"en":    "英语",      // "english",
	"eo":    "世界语",     // "esperanto",
	"et":    "爱沙尼亚语",   // "estonian",
	"tl":    "塔加路语",    // "filipino",
	"fi":    "芬兰语",     // "finnish",
	"fr":    "法语",      // "french",
	"fy":    "弗里斯兰语",   // "frisian",
	"gl":    "加利西亚语",   // "galician",
	"ka":    "格鲁吉亚语",   // "georgian",
	"de":    "德语",      // "german",
	"el":    "希腊语",     // "greek",
	"gu":    "古吉拉特语",   // "gujarati",
	"ht":    "海地克里奥尔语", // "haitian creole",
	"ha":    "豪萨语",     // "hausa",
	"haw":   "夏威夷语",    // "hawaiian",
	"iw":    "希伯来语",    // "hebrew",
	"hi":    "印地语",     // "hindi",
	"hmn":   "苗语",      // "hmong",
	"hu":    "匈牙利语",    // "hungarian",
	"is":    "冰岛语",     // "icelandic",
	"ig":    "伊博语",     // "igbo",
	"id":    "印度尼西亚语",  // "indonesian",
	"ga":    "爱尔兰语",    // "irish",
	"it":    "意大利语",    // "italian",
	"ja":    "日语",      // "japanese",
	"jw":    "爪哇语",     // "javanese",
	"kn":    "卡纳达语",    // "kannada",
	"kk":    "哈萨克语",    // "kazakh",
	"km":    "高棉语",     // "khmer",
	"ko":    "韩语",      // "korean",
	"ku":    "库尔德语",    // "kurdish (kurmanji)",
	"ky":    "吉尔吉斯语",   // "kyrgyz",
	"lo":    "老挝语",     // "lao",
	"la":    "拉丁语",     // "latin",
	"lv":    "拉脱维亚语",   // "latvian",
	"lt":    "立陶宛语",    // "lithuanian",
	"lb":    "卢森堡语",    // "luxembourgish",
	"mk":    "马其顿语",    // "macedonian",
	"mg":    "马尔加什语",   // "malagasy",
	"ms":    "马来语",     // "malay",
	"ml":    "马拉雅拉姆语",  // "malayalam",
	"mt":    "马耳他语",    // "maltese",
	"mi":    "毛利语",     // "maori",
	"mr":    "马拉语",     // "marathi",
	"mn":    "蒙语",      // "mongolian",
	"my":    "缅甸语",     // "myanmar (burmese)",
	"ne":    "尼泊尔语",    // "nepali",
	"no":    "挪威语",     // "norwegian",
	"ps":    "普什图语",    // "pashto",
	"fa":    "波斯语",     // "persian",
	"pl":    "波兰语",     // "polish",
	"pt":    "葡萄牙语",    // "portuguese",
	"pa":    "旁遮普语",    // "punjabi",
	"ro":    "罗马尼亚语",   // "romanian",
	"ru":    "俄语",      // "russian",
	"sm":    "萨摩亚语",    // "samoan",
	"gd":    "苏格兰盖尔语",  // "scots gaelic",
	"sr":    "塞尔维亚语",   // "serbian",
	"st":    "塞索托语",    // "sesotho",
	"sn":    "绍纳语",     // "shona",
	"sd":    "信德语",     // "sindhi",
	"si":    "僧伽罗语",    // "sinhala",
	"sk":    "斯洛伐克语",   // "slovak",
	"sl":    "斯洛文尼亚语",  // "slovenian",
	"so":    "索马里语",    // "somali",
	"es":    "西班牙语",    // "spanish",
	"su":    "巽他语",     // "sundanese",
	"sw":    "斯瓦希里语",   // "swahili",
	"sv":    "瑞典语",     // "swedish",
	"tg":    "塔吉克语",    // "tajik",
	"ta":    "泰米尔语",    // "tamil",
	"te":    "泰卢固语",    // "telugu",
	"th":    "泰语",      // "thai",
	"tr":    "土耳其语",    // "turkish",
	"uk":    "乌克兰语",    // "ukrainian",
	"ur":    "乌尔都语",    // "urdu",
	"uz":    "乌兹别克语",   // "uzbek",
	"vi":    "越南语",     // "vietnamese",
	"cy":    "威尔士语",    // "welsh",
	"xh":    "科萨语",     // "xhosa",
	"yi":    "意第绪语",    // "yiddish",
	"yo":    "约鲁巴语",    // "yoruba",
	"zu":    "祖鲁语",     // "zulu",
	"fil":   "菲律宾语",    // "Filipino",
	"he":    "希伯来语",    // "Hebrew"
}

func init() {
	engine := control.Register("gtranslate", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "谷歌翻译（支持回复翻译）\n" +
			"-[/][从某语言]翻译[到某语言] [翻译的内容]\n" +
			"[]内容表示可选项\n但若不是回复翻译，翻译的内容不可少。\n" +
			"若是回复翻译，'/'符号不可少。\n例：\n" +
			"/翻译 hello the world\n" +
			"从英语翻译 hello the world\n" +
			"/翻译成中文 hello the world\n" +
			"从英语翻译成中文 hello the world",
	})
	engine.OnRegex(`^\/?(从(\S+))?翻译(((到|成)(\S+))?\s)?(.+)`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		slang := ctx.State["regex_matched"].([]string)[2]
		tlang := ctx.State["regex_matched"].([]string)[6]
		txt := ctx.State["regex_matched"].([]string)[7]
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
				break
			case value:
				formlang = key
				break
			}
		}
		tolang := ""
		for key, value := range lang {
			switch tlang {
			case key:
				tolang = value
				break
			case value:
				tolang = key
				break
			}
		}
		if tolang == "" {
			tolang = "zh-CN"
			tlang = "中文"
		}
		translated, err := translate(txt, formlang, tolang)
		switch {
		case err != nil:
			ctx.SendChain(message.Text(err))
			return
		case translated == txt && tlang == "中文" && slang == "": // 如果本身是中文就换成英文
			slang = "中文"
			tlang = "英语"
			translated, err = translate(txt, "zh-CN", "en")
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			if translated == txt { // 如果还是本身
				ctx.SendChain(message.Text("我无法对其进行翻译，内容太高深力x"))
				return
			}
		}
		switch {
		case formlang == "auto" && tolang == "zh-CN":
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(translated)))
		case formlang == "auto" && tolang != "zh-CN":
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("尝试翻译成", tlang, "：\n", translated)))
		default:
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("尝试从", slang, "翻译成", tlang, "：\n", translated)))
		}
	})
	engine.OnRegex(`^\[CQ:reply,id=.*](\s+)?\/(从(\S+))?翻译((到|成)(\S+))?`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		slang := ctx.State["regex_matched"].([]string)[3]
		tlang := ctx.State["regex_matched"].([]string)[6]
		txt := ctx.GetMessage(message.NewMessageIDFromString(ctx.Event.Message[0].Data["id"])).Elements[0].Data["text"]
		if txt == "" {
			ctx.SendChain(message.Text("仅支持文字翻译"))
			return
		}
		formlang := "auto"
		for key, value := range lang {
			switch slang {
			case key:
				formlang = value
				break
			case value:
				formlang = key
				break
			}
		}
		tolang := ""
		for key, value := range lang {
			switch tlang {
			case key:
				tolang = value
				break
			case value:
				tolang = key
				break
			}
		}
		if tolang == "" {
			tolang = "zh-CN"
			tlang = "中文"
		}
		translated, err := translate(txt, formlang, tolang)
		switch {
		case err != nil:
			ctx.SendChain(message.Text(err))
			return
		case translated == txt && tlang == "中文" && (slang == ""): // 如果本身是中文就换成英文
			slang = "中文"
			tlang = "英语"
			translated, err = translate(txt, "zh-CN", "en")
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			if translated == txt { // 如果还是本身
				ctx.SendChain(message.Text("我无法对其进行翻译，内容太高深力x"))
				return
			}
		}
		switch {
		case formlang == "auto" && tolang == "zh-CN":
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(translated)))
		case formlang == "auto" && tolang != "zh-CN":
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("尝试翻译成", tlang, "：\n", translated)))
		default:
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("尝试从", slang, "翻译成", tlang, "：\n", translated)))
		}
	})
}
func translate(txt, sl, tl string) (translated string, err error) {
	url := "https://translate.google.cn/m?sl=" + sl + "&tl=" + tl + "&hl=zh-CN&q=" + url.QueryEscape(txt) // 空格以“+”连接
	// 请求html页面
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return
	}
	translated = doc.Find(".result-container").First().Text()
	return
}
