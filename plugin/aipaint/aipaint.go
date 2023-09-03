// Package aipaint ai绘图
package aipaint

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
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
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "ai绘图",
		Help: "- [ ai绘图 | 生成色图 | 生成涩图 | ai画图 ] xxx\n" +
			"- [ ai高级绘图 | 高级生成色图 | 高级生成涩图 | ai高级画图 ] [prompt]\n" +
			"- 设置ai绘图配置 [server] [token]\n" +
			"- 设置ai绘图撤回时间90s\n" +
			"- 查看ai绘图配置\n" +
			"Tips: 使用前请先前往 http://91.217.139.190:5010/token 按提示获取token" +
			"设置token示例(请确保是主人并且响应): 设置ai绘图配置 http://91.217.139.190:5010 [token] (中括号无需输入)\n" +
			"参考服务器 http://91.217.139.190:5010, http://91.216.169.75:5010, http://185.80.202.180:5010\n" +
			"[prompt]参数如下\n" +
			"tags:tag词条\nntags:ntag词条\nshape:[Portrait|Landscape|Square]\nscale:[6:20]\nseed:种子\nstrength:[0-1] 建议0-0.7\nnoise:[0-1] 建议0-0.15" +
			"参数与参数内容用:连接,每个参数之间用回车分割",
		PrivateDataFolder: "aipaint",
	})
	datapath = file.BOTPATH + "/" + engine.DataFolder()
	if file.IsNotExist(cfg.file) {
		s := serverConfig{}
		data, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(cfg.file, data, 0666)
		if err != nil {
			panic(err)
		}
	}
	engine.OnPrefixGroup([]string{`ai绘图`, `生成色图`, `生成涩图`, `ai画图`}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("少女祈祷中..."))
			args := ctx.State["args"].(string)
			data, err := web.GetData(cfg.BaseURL + fmt.Sprintf(aipaintTxt2ImgURL, cfg.Token, url.QueryEscape(strings.TrimSpace(strings.ReplaceAll(args, " ", "%20")))))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sendAiImg(ctx, data, cfg.Interval)
		})
	engine.OnPrefixGroup([]string{`ai高级绘图`, `高级生成色图`, `高级生成涩图`, `ai高级画图`}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := cfg.load()
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
						ctx.SendChain(message.Text("不准涩涩! 已将R18设置为0. "))
					}
					tags[value[0]] = strings.Join(value[1:], ":")
				}
			}
			ctx.SendChain(message.Text("少女祈祷中..."))
			apiurl := "/got_image?token=" + cfg.Token
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
			if _, ok := tags["strength"]; ok {
				apiurl += "&strength=" + url.QueryEscape(strings.TrimSpace(tags["strength"]))
			}
			if _, ok := tags["noise"]; ok {
				apiurl += "&noise=" + url.QueryEscape(strings.TrimSpace(tags["noise"]))
			}
			data, err := web.GetData(cfg.BaseURL + apiurl)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sendAiImg(ctx, data, cfg.Interval)
		})
	engine.OnRegex(`^设置ai绘图配置\s(.*[^\s$])\s(.+)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = cfg.update(regexMatched[1], regexMatched[2], cfg.Interval)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置\nbase_url: ", cfg.BaseURL, "\ntoken: ", cfg.Token, "\ninterval: ", cfg.Interval))
		})
	engine.OnRegex(`^设置ai绘图撤回时间(\d{1,3})s$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			interval, err := strconv.Atoi(regexMatched[1])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = cfg.update(cfg.BaseURL, cfg.Token, interval)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置撤回时间为", cfg.Interval, "s"))
		})
	engine.OnFullMatch(`查看ai绘图配置`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("base_url: ", cfg.BaseURL, "\ntoken: ", cfg.Token, "\ninterval: ", cfg.Interval))
		})
}

func sendAiImg(ctx *zero.Ctx, data []byte, interval int) {
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
	} else if interval > 0 {
		go func(i message.MessageID) {
			time.Sleep(time.Duration(interval) * time.Second)
			ctx.DeleteMessage(i)
		}(mid)
	}
}
