// Package lolimi 来源于 https://api.lolimi.cn/ 的接口
package lolimi

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/FloatTech/AnimeAPI/tts/lolimi"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	lolimiURL = "https://api.lolimi.cn"
	raoURL    = lolimiURL + "/API/rao/api.php"
	yanURL    = lolimiURL + "/API/yan/?url=%v"
	xjjURL    = lolimiURL + "/API/tup/xjj.php"
	qingURL   = lolimiURL + "/API/qing/api.php"
	fabingURL = lolimiURL + "/API/fabing/fb.php?name=%v"
)

var (
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "桑帛云 API",
		Help: "- 让[嘉然|塔菲|东雪莲|懒羊羊|科比|孙笑川|陈泽|丁真|空|荧|派蒙|纳西妲|阿贝多|温迪|枫原万叶|钟离|荒泷一斗|八重神子|艾尔海森|提纳里|迪希雅|卡维|宵宫|莱依拉|赛诺|诺艾尔|托马|凝光|莫娜|北斗|神里绫华|雷电将军|芭芭拉|鹿野院平藏|五郎|迪奥娜|凯亚|安柏|班尼特|琴|柯莱|夜兰|妮露|辛焱|珐露珊|魈|香菱|达达利亚|砂糖|早柚|云堇|刻晴|丽莎|迪卢克|烟绯|重云|珊瑚宫心海|胡桃|可莉|流浪者|久岐忍|神里绫人|甘雨|戴因斯雷布|优菈|菲谢尔|行秋|白术|九条裟罗|雷泽|申鹤|迪娜泽黛|凯瑟琳|多莉|坎蒂丝|萍姥姥|罗莎莉亚|留云借风真君|绮良良|瑶瑶|七七|奥兹|米卡|夏洛蒂|埃洛伊|博士|女士|大慈树王|三月七|娜塔莎|希露瓦|虎克|克拉拉|丹恒|希儿|布洛妮娅|瓦尔特|杰帕德|佩拉|姬子|艾丝妲|白露|星|穹|桑博|伦纳德|停云|罗刹|卡芙卡|彦卿|史瓦罗|螺丝咕姆|阿兰|银狼|素裳|丹枢|黑塔|景元|帕姆|可可利亚|半夏|符玄|公输师傅|奥列格|青雀|大毫|青镞|费斯曼|绿芙蓉|镜流|信使|丽塔|失落迷迭|缭乱星棘|伊甸|伏特加女孩|狂热蓝调|莉莉娅|萝莎莉娅|八重樱|八重霞|卡莲|第六夜想曲|卡萝尔|姬子|极地战刃|布洛妮娅|次生银翼|理之律者|真理之律者|迷城骇兔|希儿|魇夜星渊|黑希儿|帕朵菲莉丝|天元骑英|幽兰黛尔|德丽莎|月下初拥|朔夜观星|暮光骑士|明日香|李素裳|格蕾修|梅比乌斯|渡鸦|人之律者|爱莉希雅|爱衣|天穹游侠|琪亚娜|空之律者|终焉之律者|薪炎之律者|云墨丹心|符华|识之律者|维尔薇|始源之律者|芽衣|雷之律者|苏莎娜|阿波尼亚|陆景和|莫弈|夏彦|左然]说我测尼玛\n- 随机绕口令\n- 颜值鉴定[图片]\n" +
			"- 随机妹子\n- 随机情话\n- 发病 嘉然\n\n说明: 颜值鉴定只能鉴定三次元图片",
	})
)

func init() {
	engine.OnFullMatch("随机妹子").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Image(xjjURL))
	})
	engine.OnFullMatch("随机绕口令").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(raoURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(gjson.Get(binary.BytesToString(data), "data.Msg").String()))
	})
	engine.OnFullMatch("随机情话").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(qingURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(binary.BytesToString(data)))
	})
	engine.OnPrefix(`发病`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		name := ctx.NickName()
		data, err := web.GetData(fmt.Sprintf(fabingURL, url.QueryEscape(name)))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(gjson.Get(binary.BytesToString(data), "data").String()))
	})
	engine.OnKeywordGroup([]string{"颜值鉴定"}, zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			list := ctx.State["image_url"].([]string)
			if len(list) > 0 {
				ctx.SendChain(message.Text("少女祈祷中..."))
				data, err := web.GetData(fmt.Sprintf(yanURL, url.QueryEscape(list[0])))
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				assessment := gjson.Get(binary.BytesToString(data), "data.text").String()
				if assessment == "" {
					ctx.SendChain(message.Text("ERROR: 请输入正确的图片"))
					return
				}
				var text strings.Builder // 创建一个strings.Builder实例
				text.WriteString("评价: ")
				text.WriteString(assessment) // 添加评估信息

				for i := 0; i <= 2; i++ {
					key := gjson.Get(binary.BytesToString(data), "data.grade.key"+strconv.Itoa(i)).String()
					score := gjson.Get(binary.BytesToString(data), "data.grade.score"+strconv.Itoa(i)).String()
					if key != "" {
						text.WriteString("\n")
						text.WriteString(key)
						text.WriteString(": ")
						text.WriteString(score)
					}
				}

				ctx.SendChain(message.Text(text.String())) // 发送构建好的字符串
			}
		})
	engine.OnRegex("^让(嘉然|塔菲|东雪莲|懒羊羊|科比|孙笑川|陈泽|丁真|空|荧|派蒙|纳西妲|阿贝多|温迪|枫原万叶|钟离|荒泷一斗|八重神子|艾尔海森|提纳里|迪希雅|卡维|宵宫|莱依拉|赛诺|诺艾尔|托马|凝光|莫娜|北斗|神里绫华|雷电将军|芭芭拉|鹿野院平藏|五郎|迪奥娜|凯亚|安柏|班尼特|琴|柯莱|夜兰|妮露|辛焱|珐露珊|魈|香菱|达达利亚|砂糖|早柚|云堇|刻晴|丽莎|迪卢克|烟绯|重云|珊瑚宫心海|胡桃|可莉|流浪者|久岐忍|神里绫人|甘雨|戴因斯雷布|优菈|菲谢尔|行秋|白术|九条裟罗|雷泽|申鹤|迪娜泽黛|凯瑟琳|多莉|坎蒂丝|萍姥姥|罗莎莉亚|留云借风真君|绮良良|瑶瑶|七七|奥兹|米卡|夏洛蒂|埃洛伊|博士|女士|大慈树王|三月七|娜塔莎|希露瓦|虎克|克拉拉|丹恒|希儿|布洛妮娅|瓦尔特|杰帕德|佩拉|姬子|艾丝妲|白露|星|穹|桑博|伦纳德|停云|罗刹|卡芙卡|彦卿|史瓦罗|螺丝咕姆|阿兰|银狼|素裳|丹枢|黑塔|景元|帕姆|可可利亚|半夏|符玄|公输师傅|奥列格|青雀|大毫|青镞|费斯曼|绿芙蓉|镜流|信使|丽塔|失落迷迭|缭乱星棘|伊甸|伏特加女孩|狂热蓝调|莉莉娅|萝莎莉娅|八重樱|八重霞|卡莲|第六夜想曲|卡萝尔|姬子|极地战刃|布洛妮娅|次生银翼|理之律者|真理之律者|迷城骇兔|希儿|魇夜星渊|黑希儿|帕朵菲莉丝|天元骑英|幽兰黛尔|德丽莎|月下初拥|朔夜观星|暮光骑士|明日香|李素裳|格蕾修|梅比乌斯|渡鸦|人之律者|爱莉希雅|爱衣|天穹游侠|琪亚娜|空之律者|终焉之律者|薪炎之律者|云墨丹心|符华|识之律者|维尔薇|始源之律者|芽衣|雷之律者|苏莎娜|阿波尼亚|陆景和|莫弈|夏彦|左然)说([\\s\u4e00-\u9fa5\u3040-\u309F\u30A0-\u30FF\\w\\p{P}\u3000-\u303F\uFF00-\uFFEF]+)$").Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		name := ctx.State["regex_matched"].([]string)[1]
		msg := ctx.State["regex_matched"].([]string)[2]
		ctx.SendChain(message.Text("少女祈祷中......"))
		recordURL, err := lolimi.TTS(name, msg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Record(recordURL))
	})
}
