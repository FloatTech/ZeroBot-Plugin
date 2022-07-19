// Package iw233 基于api制作的图插件
package iw233

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"strings"

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

type config struct {
	IsDownload bool `json:"IsDownload"`
}

type result struct {
	Pic []string `json:"pic"`
}

const (
	// iw233 api
	iw233API = "https://mirlkoi.ifast3.vipnps.vip/api.php?type=json&"
	// moehu api
	moehuAPI = "https://img.moehu.org/pic.php?return=json&"
	// 色图api
	setuAPI = "http://iw233.fgimax2.fgnwctvip.com/API/Ghs.php?type=json"
	referer = "https://mirlkoi.ifast3.vipnps.vip"
	ua      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36 Edg/102.0.1245.39"
)

var (
	cfg = config{
		IsDownload: true,
	}
	groupSingle = single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 {
			return ctx.Event.GroupID
		}),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send("等一下，还有操作还未完成哦~")
		}),
	)
	allAPI = map[string]string{
		"全部":    iw233API + "sort=random",
		"兽耳":    iw233API + "sort=cat",
		"白毛":    iw233API + "sort=yin",
		"星空":    iw233API + "sort=xing",
		"竖屏壁纸":  iw233API + "sort=mp",
		"横屏壁纸":  iw233API + "sort=pc",
		"二次元图片": moehuAPI + "id=img",
		"萝莉":    moehuAPI + "id=loli",
		"古拉":    moehuAPI + "id=gawr-gura",
		"花园猫猫":  moehuAPI + "id=hanazono-serena",
		"天宫心":   moehuAPI + "id=amamiya-kokoro",
		"绊爱":    moehuAPI + "id=kizunaai",
		"神乐七奈":  moehuAPI + "id=kagura-nana",
		"白上吹雪":  moehuAPI + "id=fubuki",
		"猫羽雫":   moehuAPI + "id=myn",
		"樱岛麻衣":  moehuAPI + "id=ydmy",
		"初音未来":  moehuAPI + "id=miku",
		"洛天依":   moehuAPI + "id=tianyi",
		"五更琉璃":  moehuAPI + "id=gokou-ruri",
		"在原七海":  moehuAPI + "id=arihara-nanami",
		"鹿乃":    moehuAPI + "id=kona",
		"车万":    moehuAPI + "id=dongf",
		"赛马娘":   moehuAPI + "id=saima",
	}
	en = control.Register("iw233", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help: "iw233+moehu两个API\n" +
			" - 随机<数量>张[全部|兽耳|白毛|星空|竖屏壁纸|横屏壁纸]" +
			" - 随机<数量>张[二次元图片|萝莉|古拉|雪花菈米|花园猫猫|天宫心|绊爱|神乐七奈|白上吹雪|猫羽雫|樱岛麻衣|初音未来|洛天依|五更琉璃|在原七海|鹿乃|车万|赛马娘]\n" +
			" - 清空[与上面相同]缓存\n" +
			" - 清空所有缓存 " +
			" - [开启|关闭]使用缓存",
		PrivateDataFolder: "iw233",
	}).ApplySingle(groupSingle)
	filepath = en.DataFolder()
)

func init() {
	err := os.MkdirAll(filepath, 0755)
	if err != nil {
		panic(err)
	}
	cfgFile := en.DataFolder() + "config.json"
	if file.IsExist(cfgFile) {
		reader, err := os.Open(cfgFile)
		if err == nil {
			err = json.NewDecoder(reader).Decode(&cfg)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
		err = reader.Close()
		if err != nil {
			panic(err)
		}
	} else {
		err = saveConfig(cfgFile)
		if err != nil {
			panic(err)
		}
	}
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
			option := ctx.State["regex_matched"].([]string)[1]
			switch option {
			case "开启", "打开", "启用":
				cfg.IsDownload = true
			case "关闭", "关掉", "禁用":
				cfg.IsDownload = false
			default:
				return
			}
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("已设置图片模式为缓存" + option))
			} else {
				ctx.SendChain(message.Text("ERROR:", err))
			}
		})
}

func getimage(ctx *zero.Ctx, api, rename string, i int64) (m message.Message, err error) {
	if i == 0 {
		i = 1
	}
	switch {
	case strings.Contains(api, "https://img.moehu.org") && i > 10:
		i = 10
		ctx.SendChain(message.Text("这个api只能获取10张图片哦~"))
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
	data, err := web.RequestDataWith(web.NewDefaultClient(), api+"&num="+strconv.FormatInt(i, 10), "GET", referer, ua)
	if err != nil {
		return
	}
	var r result
	err = json.Unmarshal(data, &r)
	if err != nil {
		return
	}
	m = make(message.Message, 0, len(r.Pic))
	if cfg.IsDownload {
		_ = os.Mkdir(file.BOTPATH+"/"+filepath+rename, 0664)
		md5 := md5.New()
		for _, v := range r.Pic {
			_, err = md5.Write(binary.StringToBytes(v))
			if err != nil {
				return
			}
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
	for _, v := range r.Pic {
		m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image(v)))
	}
	return
}

func saveConfig(cfgFile string) (err error) {
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
