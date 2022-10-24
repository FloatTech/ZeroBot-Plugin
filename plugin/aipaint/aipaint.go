// Package aipaint ai绘图
package aipaint

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	datapath  string
	predictRe = regexp.MustCompile(`{"steps".+?}`)
	// 参考host http://91.217.139.190:5010 http://91.216.169.75:5010
	aipaintTxt2ImgURL = "/got_image?token=%v&tags=%v"
	aipaintImg2ImgURL = "/got_image2image?token=%v&tags=%v"
	cfg               = newServerConfig("data/aipaint/config.json")
)

type result struct {
	Steps    int     `json:"steps"`
	Sampler  string  `json:"sampler"`
	Seed     int     `json:"seed"`
	Strength float64 `json:"strength"`
	Noise    float64 `json:"noise"`
	Scale    float64 `json:"scale"`
	Uc       string  `json:"uc"`
}

func (r *result) String() string {
	return fmt.Sprintf("steps: %v\nsampler: %v\nseed: %v\nstrength: %v\nnoise: %v\nscale: %v\nuc: %v\n", r.Steps, r.Sampler, r.Seed, r.Strength, r.Noise, r.Scale, r.Uc)
}

func init() { // 插件主体
	engine := control.Register("aipaint", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "ai绘图\n" +
			"- [ ai绘图 | 生成色图 | 生成涩图 | ai画图 ] xxx\n" +
			"- [ ai高级绘图 | 高级生成色图 | 高级生成涩图 | ai高级画图 ] [prompt]\n" +
			"- [ 以图绘图 | 以图生图 | 以图画图 ] xxx [图片]|@xxx|[qq号]\n" +
			"- 设置ai绘图配置 [server] [token]\n" +
			"例: 设置ai绘图配置 http://91.217.139.190:5010 abc\n" +
			"参考服务器 http://91.217.139.190:5010, http://91.216.169.75:5010, http://185.80.202.180:5010\n" +
			"通过 http://91.217.139.190:5010/token 获取token\n" +
			"[prompt]参数如下\n" +
			"tags:tag词条\nntags:ntag词条\nshape:[Portrait|Landscape|Square]\nscale:[6:20]\nseed:种子\n" +
			"参数与参数内容用:连接,每个参数之间用回车或者&分割",
		PrivateDataFolder: "aipaint",
	})
	datapath = file.BOTPATH + "/" + engine.DataFolder()
	engine.OnPrefixGroup([]string{`ai绘图`, `生成色图`, `生成涩图`, `ai画图`}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			server, token, err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("少女祈祷中..."))
			args := ctx.State["args"].(string)
			data, err := web.GetData(server + fmt.Sprintf(aipaintTxt2ImgURL, token, url.QueryEscape(strings.TrimSpace(strings.ReplaceAll(args, " ", "%20")))))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sendAiImg(ctx, data)
		})
	engine.OnRegex(`^(以图绘图|以图生图|以图画图)[\s\S]*?(\[CQ:(image\,file=([0-9a-zA-Z]{32}).*|at.+?(\d{5,11}))\].*|(\d+))$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			server, token, err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			c := newContext(ctx.Event.UserID)
			list := ctx.State["regex_matched"].([]string)
			err = c.prepareLogos(list[4]+list[5]+list[6], strconv.FormatInt(ctx.Event.UserID, 10))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			args := strings.TrimSuffix(strings.TrimPrefix(list[0], list[1]), list[2])
			if args == "" {
				ctx.SendChain(message.Text("ERROR: 以图绘图必须添加tag"))
				return
			}
			ctx.SendChain(message.Text("少女祈祷中..."))
			postURL := server + fmt.Sprintf(aipaintImg2ImgURL, token, url.QueryEscape(strings.TrimSpace(strings.ReplaceAll(args, " ", "%20"))))

			f, err := os.Open(c.headimgsdir[0])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			imageShape := ""
			switch {
			case img.Bounds().Dx() > img.Bounds().Dy():
				imageShape = "Landscape"
			case img.Bounds().Dx() == img.Bounds().Dy():
				imageShape = "Square"
			default:
				imageShape = "Portrait"
			}

			// 图片转base64
			base64Bytes, err := writer.ToBase64(img)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			data, err := web.PostData(postURL+"&shape="+imageShape, "text/plain", bytes.NewReader(base64Bytes))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sendAiImg(ctx, data)
		})
	engine.OnPrefixGroup([]string{`ai高级绘图`, `高级生成色图`, `高级生成涩图`, `ai高级画图`}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			server, token, err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			tags := make(map[string]string)
			args := strings.Split(ctx.State["args"].(string), "\n")
			if len(args) < 1 {
				ctx.SendChain(message.Text("ERROR: 请输入正确的参数"))
				return
			}
			for _, info := range args {
				value := strings.Split(info, ":")
				if len(value) > 1 {
					if value[0] == "R18" && value[1] == "1" {
						value[1] = "0"
						ctx.SendChain(message.Text("不准涩涩!已将R18设置为0。"))
					}
					tags[value[0]] = strings.Join(value[1:], ":")
				}
			}
			ctx.SendChain(message.Text("少女祈祷中..."))
			apiurl := "/got_image?token=" + token
			if _, ok := tags["tags"]; ok {
				apiurl += "&tags=" + url.QueryEscape(strings.ReplaceAll(strings.TrimSpace(tags["tags"]), " ", "%20"))
			}
			if _, ok := tags["ntags"]; ok {
				apiurl += "&ntags=" + url.QueryEscape(strings.ReplaceAll(strings.TrimSpace(tags["ntags"]), " ", "%20"))
			}
			if _, ok := tags["R18"]; ok {
				apiurl += "&R18=" + url.QueryEscape(strings.TrimSpace(tags["R18"]))
			}
			if _, ok := tags["shape"]; ok {
				apiurl += "&shape=" + url.QueryEscape(strings.TrimSpace(tags["shape"]))
			}
			if _, ok := tags["scale"]; ok {
				apiurl += "&scale=" + url.QueryEscape(strings.TrimSpace(tags["scale"]))
			}
			if _, ok := tags["seed"]; ok {
				apiurl += "&seed=" + url.QueryEscape(strings.TrimSpace(tags["seed"]))
			}
			data, err := web.GetData(server + apiurl)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sendAiImg(ctx, data)
		})
	engine.OnRegex(`^设置ai绘图配置\s(.*[^\s$])\s(.+)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			err := cfg.save(regexMatched[1], regexMatched[2])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置server为", regexMatched[1], ", token为", regexMatched[2]))
		})
}

func sendAiImg(ctx *zero.Ctx, data []byte) {
	var loadData string
	if predictRe.MatchString(binary.BytesToString(data)) {
		loadData = predictRe.FindStringSubmatch(binary.BytesToString(data))[0]
	}
	var r result
	if loadData != "" {
		err := json.Unmarshal(binary.StringToBytes(loadData), &r)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		r.Uc, err = url.QueryUnescape(r.Uc)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
	}
	encodeStr := base64.StdEncoding.EncodeToString(data)
	m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image("base64://"+encodeStr))}
	m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(r.String())))
	if mid := ctx.Send(m); mid.ID() == 0 {
		ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
	} else {
		go func(i message.MessageID) {
			time.Sleep(90 * time.Second)
			ctx.DeleteMessage(i)
		}(mid)
	}

}
