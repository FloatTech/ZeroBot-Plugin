// Package iw233 基于api制作的图插件
package iw233

import (
	"encoding/json"
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/math"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

type result struct {
	Pic []string `json:"pic"`
}

const (
	// 随机壁纸api
	randomAPI = "https://mirlkoi.ifast3.vipnps.vip/api.php?sort=random&type=json"
	// 兽耳api
	animalAPI = "https://mirlkoi.ifast3.vipnps.vip/api.php?sort=cat&type=json"
	// 白毛api
	whiteAPI = "https://mirlkoi.ifast3.vipnps.vip/api.php?sort=yin&type=json"
	// 星空api
	starryAPI = "https://mirlkoi.ifast3.vipnps.vip/api.php?sort=xing&type=json"
	// 竖屏壁纸api
	verticalAPI = "https://mirlkoi.ifast3.vipnps.vip/api.php?sort=mp&type=json"
	// 横屏壁纸api
	horizontalAPI = "https://mirlkoi.ifast3.vipnps.vip/api.php?sort=pc&type=json"
	// 色图api
	setuAPI = "http://iw233.fgimax2.fgnwctvip.com/API/Ghs.php?type=json"
	referer = "https://mirlkoi.ifast3.vipnps.vip"
	ua      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36 Edg/102.0.1245.39"
)

// GroupSingle 按群号反并发
var GroupSingle = single.New(
	single.WithKeyFn(func(ctx *zero.Ctx) int64 {
		return ctx.Event.GroupID
	}),
	single.WithPostFn[int64](func(ctx *zero.Ctx) {
		ctx.Send("等一下，还有操作还未完成哦~")
	}),
)

var (
	en = control.Register("iw233", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help:             "iw233\n - 随机<数量>张[全部|兽耳|白毛|星空|竖屏壁纸|横屏壁纸]",
		PublicDataFolder: "Iw233",
	}).ApplySingle(GroupSingle)
)

func init() {
	en.OnRegex(`^随机([0-9]+)?[份|张]全部`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			i := int(math.Str2Int64(ctx.State["regex_matched"].([]string)[1]))
			m, err := getimage(ctx, randomAPI, i)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, m).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	en.OnRegex(`^随机([0-9]+)?[份|张]兽耳`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			i := int(math.Str2Int64(ctx.State["regex_matched"].([]string)[1]))
			m, err := getimage(ctx, animalAPI, i)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, m).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	en.OnRegex(`^随机([0-9]+)?[份|张]白毛`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			i := int(math.Str2Int64(ctx.State["regex_matched"].([]string)[1]))
			m, err := getimage(ctx, whiteAPI, i)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, m).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	en.OnRegex(`^随机([0-9]+)?[份|张]星空`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			i := int(math.Str2Int64(ctx.State["regex_matched"].([]string)[1]))
			m, err := getimage(ctx, starryAPI, i)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, m).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	en.OnRegex(`^随机([0-9]+)?[份|张]竖屏壁纸`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			i := int(math.Str2Int64(ctx.State["regex_matched"].([]string)[1]))
			m, err := getimage(ctx, verticalAPI, i)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, m).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	en.OnRegex(`^随机([0-9]+)?[份|张]横屏壁纸`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			i := int(math.Str2Int64(ctx.State["regex_matched"].([]string)[1]))
			m, err := getimage(ctx, horizontalAPI, i)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, m).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	en.OnRegex(`^随机色图`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.RequestDataWith(web.NewDefaultClient(), setuAPI, "GET", referer, ua)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			picURL := gjson.Get(helper.BytesToString(data), "pic").String()
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image(picURL))}).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}

func getimage(ctx *zero.Ctx, api string, i int) (m message.Message, err error) {
	switch {
	case !zero.AdminPermission(ctx) && i > 15:
		i = 15
		ctx.SendChain(message.Text("普通成员最多只能随机15张图片哦~"))
	case !zero.SuperUserPermission(ctx) && i > 30:
		i = 30
		ctx.SendChain(message.Text("管理员最多只能随机30张图片哦~"))
	case i > 50:
		ctx.SendChain(message.Text("那么多真的好吗（￣▽￣）"))
	}
	filepath := en.DataFolder()
	data, err := web.RequestDataWith(web.NewDefaultClient(), api+"&num="+strconv.Itoa(i), "GET", referer, ua)
	if err != nil {
		return
	}
	var r result
	err = json.Unmarshal(data, &r)
	if err != nil {
		return
	}
	m = make(message.Message, 0, 100)
	for _, v := range r.Pic {
		name := filepath + v[40:]
		f := "file:///" + file.BOTPATH + "/" + name
		if file.IsExist(file.BOTPATH + "/" + name) {
			m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image(f)))
		} else {
			err = file.DownloadTo(v, name, false)
			if err != nil {
				return
			}
			m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image(f)))
		}
	}
	return
}
