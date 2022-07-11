// Package iw233 基于api制作的图插件
package iw233

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/math"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"
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

var (
	// GroupSingle 按群号反并发
	GroupSingle = single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 {
			return ctx.Event.GroupID
		}),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send("等一下，还有操作还未完成哦~")
		}),
	)
	en = control.Register("iw233", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  true,
		Help:              "iw233\n - 随机<数量>张[全部|兽耳|白毛|星空|竖屏壁纸|横屏壁纸]",
		PrivateDataFolder: "iw233",
	}).ApplySingle(GroupSingle)
	allAPI = map[string]string{
		"全部":   randomAPI,
		"兽耳":   animalAPI,
		"白毛":   whiteAPI,
		"星空":   starryAPI,
		"竖屏壁纸": verticalAPI,
		"横屏壁纸": horizontalAPI,
	}
)

func init() {
	en.OnRegex(`^随机(([0-9]+)[份|张])?(.*)`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.State["regex_matched"].([]string)[3]
			api, ok := allAPI[msg]
			if !ok {
				return
			}
			i := math.Str2Int64(ctx.State["regex_matched"].([]string)[2])
			m, err := getimage(ctx, api, msg, i)
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
			picURL := gjson.Get(binary.BytesToString(data), "pic").String()
			if id := ctx.SendGroupForwardMessage(ctx.Event.GroupID, message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image(picURL))}).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}

func getimage(ctx *zero.Ctx, api, rename string, i int64) (m message.Message, err error) {
	if i == 0 {
		i = 1
	}
	switch {
	case !zero.AdminPermission(ctx) && i > 15:
		i = 15
		ctx.SendChain(message.Text("普通成员最多只能随机15张图片哦~"))
	case !zero.SuperUserPermission(ctx) && i > 30:
		i = 30
		ctx.SendChain(message.Text("管理员最多只能随机30张图片哦~"))
	case zero.SuperUserPermission(ctx) && i > 100:
		i = 100
		ctx.SendChain(message.Text("太贪心啦！最多只能随机100张图片哦~"))
	}
	ctx.SendChain(message.Text("少女祈祷中..."))
	filepath := en.DataFolder()
	data, err := web.RequestDataWith(web.NewDefaultClient(), api+"&num="+strconv.FormatInt(i, 10), "GET", referer, ua)
	if err != nil {
		return
	}
	var r result
	err = json.Unmarshal(data, &r)
	if err != nil {
		return
	}
	err = os.Mkdir(file.BOTPATH+"/"+filepath+rename, 0664)
	if err != nil {
		return
	}
	md5 := md5.New()
	m = make(message.Message, 0, 100)
	for _, v := range r.Pic {
		md5.Write(binary.StringToBytes(v))
		name := hex.EncodeToString(md5.Sum(nil))[:8] + ".jpg"
		f := file.BOTPATH + "/" + filepath + rename + "/" + name
		if file.IsNotExist(f) {
			err = file.DownloadTo(v, f, false)
			if err != nil {
				return
			}
			m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+f)))
		} else {
			m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+f)))
		}
	}
	return
}
