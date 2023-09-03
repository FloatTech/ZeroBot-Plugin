// Package qzone qq空间表白墙
package qzone

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/qzone"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/jinzhu/gorm"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	waitStatus = iota + 1
	agreeStatus
	disagreeStatus
	loveTag      = "表白"
	faceURL      = "http://q4.qlogo.cn/g?b=qq&nk=%v&s=640"
	anonymousURL = "https://gitcode.net/anto_july/avatar/-/raw/master/%v.png"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "QQ空间表白墙",
		Help: "- 登录QQ空间 (Cookie过期很快, 要经常登录)\n" +
			"- 发说说[xxx]\n" +
			"- (匿名)发表白墙[xxx]\n" +
			"- [ 同意 | 拒绝 ]表白墙 1,2,3 (最后一个参数是表白墙的序号数组, 用英文逗号连接)\n" +
			"- 查看[ 等待 | 同意 | 拒绝 | 所有 ]表白墙 0 (最后一个参数是页码, 建议私聊审稿)",
		PrivateDataFolder: "qzone",
	})
	go func() {
		qdb = initialize(engine.DataFolder() + "qzone.db")
	}()
	engine.OnFullMatch("登录QQ空间").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				qrsig           string
				ptqrtoken       string
				ptqrloginCookie string
				redirectCookie  string
				data            []byte
				err             error
			)
			data, qrsig, ptqrtoken, err = qzone.Ptqrshow()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("请扫描二维码, 登录QQ空间"))
			ctx.SendChain(message.ImageBytes(data))
			for {
				time.Sleep(2 * time.Second)
				data, ptqrloginCookie, err = qzone.Ptqrlogin(qrsig, ptqrtoken)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				text := binary.BytesToString(data)

				switch {
				case strings.Contains(text, "二维码已失效"):
					ctx.SendChain(message.Text("二维码已失效, 登录失败"))
					return
				case strings.Contains(text, "登录成功"):
					dealedCheckText := strings.ReplaceAll(text, "'", "")
					redirectURL := strings.Split(dealedCheckText, ",")[2]
					redirectCookie, err = qzone.LoginRedirect(redirectURL)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					m := qzone.NewManager(ptqrloginCookie + redirectCookie)
					qq, err := strconv.ParseInt(m.QQ, 10, 64)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					err = qdb.insertOrUpdate(qq, m.Cookie)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					ctx.SendChain(message.Text("登录成功"))
					return
				}
			}
		})
	engine.OnRegex(`^发说说.*?([\s\S]*)`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			text, base64imgs, err := parseTextAndImg(message.UnescapeCQCodeText(regexMatched[1]))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = publishEmotion(ctx.Event.SelfID, text, base64imgs)
			if err != nil {
				if gorm.IsRecordNotFoundError(err) {
					ctx.SendChain(message.Text(zero.BotConfig.NickName[0], "(", ctx.Event.SelfID, ")", "未登录QQ空间,请发送\"登录QQ空间\"初始化配置"))
					return
				}
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("发表成功"))
		})
	engine.OnRegex(`^(.{0,2})发表白墙.*?([\s\S]*)`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			if strings.TrimSpace(regexMatched[2]) == "" {
				ctx.SendChain(message.Text("请不要发送空内容"))
				return
			}
			qq := ctx.Event.UserID
			e := emotion{
				QQ:        qq,
				Msg:       message.UnescapeCQCodeText(regexMatched[2]),
				Status:    waitStatus,
				Tag:       loveTag,
				Anonymous: false,
			}
			if regexMatched[1] == "匿名" {
				e.Anonymous = true
			}
			_, err := qdb.saveEmotion(e)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("已收稿, 请耐心等待审核"))
		})
	engine.OnRegex(`^(同意|拒绝)表白墙\s?((?:\d+,){0,8}\d+)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var err error
			var ti int64
			regexMatched := ctx.State["regex_matched"].([]string)
			idStrList := strings.Split(regexMatched[2], ",")
			idList := make([]int64, 0, len(idStrList))
			for _, v := range idStrList {
				ti, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return
				}
				idList = append(idList, ti)
			}
			switch regexMatched[1] {
			case "同意":
				err = getAndPublishEmotion(ctx.Event.SelfID, idList)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				err = qdb.updateEmotionStatusByIDList(idList, agreeStatus)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text("同意表白墙", regexMatched[2], ", 发表成功"))
			case "拒绝":
				err = qdb.updateEmotionStatusByIDList(idList, disagreeStatus)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text("拒绝表白墙", regexMatched[2]))
			}
		})
	engine.OnRegex(`^查看(.{0,2})表白墙\s?(\d*)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				pageNum   int
				err       error
				base64Str []byte
			)
			regexMatched := ctx.State["regex_matched"].([]string)
			if regexMatched[2] == "" {
				pageNum = 0
			} else {
				pageNum, err = strconv.Atoi(regexMatched[2])
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
			}
			var status int
			switch regexMatched[1] {
			case "等待":
				status = 1
			case "同意":
				status = 2
			case "拒绝":
				status = 3
			case "所有":
				status = 0
			default:
				status = 1
			}
			el, err := qdb.getLoveEmotionByStatus(status, pageNum)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if len(el) == 0 {
				ctx.SendChain(message.Text("ERROR: 当前表白墙数量为0"))
				return
			}
			m := message.Message{}
			for _, v := range el {
				t := v.textBrief() + "\n呢称: " + ctx.CardOrNickName(v.QQ)
				base64Str, err = text.RenderToBase64(t, text.FontFile, 400, 20)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image("base64://"+binary.BytesToString(base64Str))))
				base64Str, err = renderForwardMsg(v.QQ, v.Msg)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image("base64://"+binary.BytesToString(base64Str))))
			}
			time.Sleep(time.Second)
			if id := ctx.Send(m).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
			}
		})
}

func getAndPublishEmotion(botqq int64, idList []int64) (err error) {
	var b []byte
	el, err := qdb.getEmotionByIDList(idList)
	if err != nil {
		return
	}
	base64imgs := make([]string, 0, 5)
	for _, v := range el {
		if v.Anonymous {
			v.QQ = 0
		}
		b, err = renderForwardMsg(v.QQ, v.Msg)
		if err != nil {
			return
		}
		base64imgs = append(base64imgs, binary.BytesToString(b))
	}
	return publishEmotion(botqq, "", base64imgs)
}

func publishEmotion(botqq int64, text string, base64imgs []string) (err error) {
	qc, err := qdb.getByUin(botqq)
	if err != nil {
		return
	}
	m := qzone.NewManager(qc.Cookie)
	_, err = m.EmotionPublish(text, base64imgs)
	return
}

func parseTextAndImg(raw string) (text string, base64imgs []string, err error) {
	base64imgs = make([]string, 0, 16)
	var imgdata []byte
	m := message.ParseMessageFromString(raw)
	for _, v := range m {
		if v.Type == "text" && v.Data["text"] != "" {
			text += v.Data["text"] + "\n"
		}
		if v.Type == "image" && v.Data["url"] != "" {
			imgdata, err = web.GetData(v.Data["url"])
			if err != nil {
				return
			}
			encodeStr := base64.StdEncoding.EncodeToString(imgdata)
			base64imgs = append(base64imgs, encodeStr)
		}
	}
	return
}

func renderForwardMsg(qq int64, raw string) (base64Bytes []byte, err error) {
	canvas := gg.NewContext(1000, 1000)
	canvas.SetRGB255(229, 229, 229)
	canvas.Clear()
	canvas.SetColor(color.Black)
	var (
		maxHeight = 0
		maxWidth  = 0
		backX     = 200
		backY     = 200
		margin    = 50
		face      []byte
		imgdata   []byte
		msgImg    image.Image
		faceImg   image.Image
	)
	if qq != 0 {
		face, err = web.GetData(fmt.Sprintf(faceURL, qq))
	} else {
		face, err = web.RequestDataWith(web.NewTLS12Client(), fmt.Sprintf(anonymousURL, rand.Intn(4)+1), "GET", "gitcode.net", web.RandUA(), nil)
	}
	if err != nil {
		return
	}
	faceImg, _, err = image.Decode(bytes.NewReader(face))
	if err != nil {
		return
	}
	back := imgfactory.Size(faceImg, backX, backY).Circle(0).Image()
	m := message.ParseMessageFromString(raw)
	maxHeight += margin

	for _, v := range m {
		switch {
		case v.Type == "text" && strings.TrimSpace(v.Data["text"]) != "":
			msgImg, err = text.Render(strings.TrimSuffix(v.Data["text"], "\r\n"), text.FontFile, 400, 40)
			if err != nil {
				return
			}
		case v.Type == "image" && v.Data["url"] != "":
			imgdata, err = web.GetData(v.Data["url"])
			if err != nil {
				return
			}
			msgImg, _, err = image.Decode(bytes.NewReader(imgdata))
			if err != nil {
				return
			}
		default:
			continue
		}
		canvas.DrawImage(back, margin, maxHeight)
		if msgImg.Bounds().Dx() > 500 {
			msgImg = imgfactory.Size(msgImg, 500, msgImg.Bounds().Dy()*500/msgImg.Bounds().Dx()).Image()
		}
		canvas.DrawImage(msgImg, 2*margin+backX, maxHeight)
		if 3*margin+backX+msgImg.Bounds().Dx() > maxWidth {
			maxWidth = 3*margin + backX + msgImg.Bounds().Dx()
		}
		if msgImg.Bounds().Dy() > backY {
			maxHeight += msgImg.Bounds().Dy() + margin
		} else {
			maxHeight += backY + margin
		}
	}
	im := canvas.Image().(*image.RGBA)
	nim := im.SubImage(image.Rect(0, 0, maxWidth, maxHeight))
	return imgfactory.ToBase64(nim)
}
