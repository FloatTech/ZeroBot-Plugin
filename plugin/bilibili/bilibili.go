// Package bilibili 查询b站用户信息
package bilibili

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"time"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	re             = regexp.MustCompile(`^\d+$`)
	danmakuTypeMap = map[int64]string{
		0: "普通消息",
		1: "礼物",
		2: "上舰",
		3: "Superchat",
		4: "进入直播间",
		5: "标题变动",
		6: "分区变动",
		7: "直播中止",
		8: "直播继续",
	}
	cfg = bz.NewCookieConfig("data/Bilibili/config.json")
)

// 查成分的
func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "b站查成分查弹幕",
		Help: "- >vup info [xxx]\n" +
			"- >user info [xxx]\n" +
			"- 查成分 [xxx]\n" +
			"- 查弹幕 [xxx]\n" +
			"- 设置b站cookie b_ut=7;buvid3=0;i-wanna-go-back=-1;innersign=0;\n" +
			"- 更新vup\n" +
			"Tips: (412就是拦截的意思,建议私聊把cookie设全)\n",
		PublicDataFolder: "Bilibili",
	})
	cachePath := engine.DataFolder() + "cache/"
	_ = os.RemoveAll(cachePath)
	_ = os.MkdirAll(cachePath, 0755)
	var getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		var err error
		_, _ = engine.GetLazyData("bilibili.db", false)
		vdb, err = initializeVup(engine.DataFolder() + "bilibili.db")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		return true
	})
	engine.OnRegex(`^>user info\s?(.{1,25})$`, getPara).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			id := ctx.State["uid"].(string)
			card, err := bz.GetMemberCard(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text(
				"uid: ", card.Mid, "\n",
				"name: ", card.Name, "\n",
				"sex: ", card.Sex, "\n",
				"sign: ", card.Sign, "\n",
				"level: ", card.LevelInfo.CurrentLevel, "\n",
				"birthday: ", card.Birthday,
			))
		})

	engine.OnRegex(`^>vup info\s?(.{1,25})$`, getPara).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			id := ctx.State["uid"].(string)
			// 获取详情
			fo, err := bz.GetVtbDetail(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text(
				"b站id: ", fo.Mid, "\n",
				"名字: ", fo.Uname, "\n",
				"当前粉丝数: ", fo.Follower, "\n",
				"24h涨粉数: ", fo.Rise, "\n",
				"视频投稿数: ", fo.Video, "\n",
				"直播间id: ", fo.Roomid, "\n",
				"舰队: ", fo.GuardNum, "\n",
				"直播总排名: ", fo.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", fo.Mid, "\n",
				"数据获取时间: ", time.Now().Format("2006-01-02 15:04:05"),
			))
		})

	engine.OnRegex(`^查成分\s?(.{1,25})$`, getPara, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			id := ctx.State["uid"].(string)
			today := time.Now().Format("20060102")
			drawedFile := cachePath + id + today + "vupLike.png"
			if file.IsExist(drawedFile) {
				ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
				return
			}
			u, err := bz.GetMemberCard(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			vups, err := vdb.filterVup(u.Attentions)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			vupLen := len(vups)
			medals, err := bz.GetMedalWall(cfg, id)
			sort.Sort(bz.MedalSorter(medals))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			frontVups := make([]vup, 0)
			medalMap := make(map[int64]bz.Medal)
			for _, v := range medals {
				up := vup{
					Mid:   v.Mid,
					Uname: v.Uname,
				}
				frontVups = append(frontVups, up)
				medalMap[v.Mid] = v
			}
			vups = append(vups, frontVups...)
			copy(vups[len(frontVups):], vups)
			copy(vups, frontVups)
			for i := len(frontVups); i < len(vups); i++ {
				if _, ok := medalMap[vups[i].Mid]; ok {
					vups = append(vups[:i], vups[i+1:]...)
					i--
				}
			}
			facePath := cachePath + id + "vupFace" + path.Ext(u.Face)
			backX := 500
			backY := 500
			var back image.Image
			if path.Ext(u.Face) != ".webp" {
				err = initFacePic(facePath, u.Face)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				back, err = gg.LoadImage(facePath)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				back = imgfactory.Size(back, backX, backY).Image()
			}
			if len(vups) > 50 {
				ctx.SendChain(message.Text(u.Name + "关注的up主太多了, 只展示前50个up"))
				vups = vups[:50]
			}
			canvas := gg.NewContext(1500, int(500*(1.1+float64(len(vups))/3)))
			fontSize := 50.0
			canvas.SetColor(color.White)
			canvas.Clear()
			if back != nil {
				canvas.DrawImage(back, 0, 0)
			}
			canvas.SetColor(color.Black)
			data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			if err = canvas.ParseFontFace(data, fontSize); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sl, _ := canvas.MeasureString("好")
			length, h := canvas.MeasureString(u.Mid)
			n, _ := canvas.MeasureString(u.Name)
			canvas.DrawString(u.Name, 550, 160-h)
			canvas.DrawRoundedRectangle(600+n-length*0.1, 160-h*2.5, length*1.2, h*2, fontSize*0.2)
			canvas.SetRGB255(221, 221, 221)
			canvas.Fill()
			canvas.SetColor(color.Black)
			canvas.DrawString(u.Mid, 600+n, 160-h)
			canvas.DrawString(fmt.Sprintf("粉丝：%d", u.Fans), 550, 240-h)
			canvas.DrawString(fmt.Sprintf("关注：%d", len(u.Attentions)), 1000, 240-h)
			canvas.DrawString(fmt.Sprintf("管人痴成分：%.2f%%（%d/%d）", float64(vupLen)/float64(len(u.Attentions))*100, vupLen, len(u.Attentions)), 550, 320-h)
			regtime := time.Unix(u.Regtime, 0).Format("2006-01-02 15:04:05")
			canvas.DrawString("注册日期："+regtime, 550, 400-h)
			canvas.DrawString("查询日期："+time.Now().Format("2006-01-02"), 550, 480-h)
			for i, v := range vups {
				if i%2 == 1 {
					canvas.SetRGB255(245, 245, 245)
					canvas.DrawRectangle(0, float64(backY)*1.1+float64(i)*float64(backY)/3, float64(backX*3), float64(backY)/3)
					canvas.Fill()
				}
				canvas.SetColor(color.Black)
				nl, _ := canvas.MeasureString(v.Uname)
				canvas.DrawString(v.Uname, float64(backX)*0.1, float64(backY)*1.1+float64(i+1)*float64(backY)/3-2*h)
				ml, _ := canvas.MeasureString(strconv.FormatInt(v.Mid, 10))
				canvas.DrawRoundedRectangle(nl-0.1*ml+float64(backX)*0.2, float64(backY)*1.1+float64(i+1)*float64(backY)/3-h*3.5, ml*1.2, h*2, fontSize*0.2)
				canvas.SetRGB255(221, 221, 221)
				canvas.Fill()
				canvas.SetColor(color.Black)
				canvas.DrawString(strconv.FormatInt(v.Mid, 10), nl+float64(backX)*0.2, float64(backY)*1.1+float64(i+1)*float64(backY)/3-2*h)
				if m, ok := medalMap[v.Mid]; ok {
					mnl, _ := canvas.MeasureString(m.MedalName)
					grad := gg.NewLinearGradient(nl+ml-sl/2+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-3.5*h, nl+ml+mnl+sl/2+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-1.5*h)
					r, g, b := int2rbg(m.MedalColorStart)
					grad.AddColorStop(0, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})
					r, g, b = int2rbg(m.MedalColorEnd)
					grad.AddColorStop(1, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})
					canvas.SetFillStyle(grad)
					canvas.SetLineWidth(4)
					canvas.MoveTo(nl+ml-sl/2+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-3.5*h)
					canvas.LineTo(nl+ml+mnl+sl/2+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-3.5*h)
					canvas.LineTo(nl+ml+mnl+sl/2+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-1.5*h)
					canvas.LineTo(nl+ml-sl/2+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-1.5*h)
					canvas.ClosePath()
					canvas.Fill()
					canvas.SetColor(color.White)
					canvas.DrawString(m.MedalName, nl+ml+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-2*h)
					r, g, b = int2rbg(m.MedalColorBorder)
					canvas.SetRGB255(int(r), int(g), int(b))
					canvas.DrawString(strconv.FormatInt(m.Level, 10), nl+ml+mnl+sl+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-2*h)
					mll, _ := canvas.MeasureString(strconv.FormatInt(m.Level, 10))
					canvas.SetLineWidth(4)
					canvas.MoveTo(nl+ml-sl/2+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-3.5*h)
					canvas.LineTo(nl+ml+mnl+mll+sl/2+float64(backX)*0.5, float64(backY)*1.1+float64(i+1)*float64(backY)/3-3.5*h)
					canvas.LineTo(nl+ml+mnl+mll+sl/2+float64(backX)*0.5, float64(backY)*1.1+float64(i+1)*float64(backY)/3-1.5*h)
					canvas.LineTo(nl+ml-sl/2+float64(backX)*0.4, float64(backY)*1.1+float64(i+1)*float64(backY)/3-1.5*h)
					canvas.ClosePath()
					canvas.Stroke()
				}
			}
			f, err := os.Create(drawedFile)
			if err != nil {
				log.Errorln("[bilibili]", err)
				data, err := imgfactory.ToBytes(canvas.Image())
				if err != nil {
					log.Errorln("[bilibili]", err)
					return
				}
				ctx.SendChain(message.ImageBytes(data))
				return
			}
			_, err = imgfactory.WriteTo(canvas.Image(), f)
			_ = f.Close()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
		})

	engine.OnRegex(`^查弹幕\s?(\S{1,25})\s?(\d*)$`, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		id := ctx.State["uid"].(string)
		pagenum := ctx.State["regex_matched"].([]string)[2]
		if pagenum == "" {
			pagenum = "0"
		}
		u, err := bz.GetMemberCard(id)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		var danmaku bz.Danmakusuki
		tr := &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{Transport: tr}
		data, err := web.RequestDataWith(client, fmt.Sprintf(bz.DanmakuAPI, id, pagenum), "GET", "", web.RandUA(), nil)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		err = json.Unmarshal(data, &danmaku)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		today := time.Now().Format("20060102150415")
		drawedFile := cachePath + id + today + "vupLike.png"
		facePath := cachePath + id + "vupFace" + path.Ext(u.Face)
		backX := 500
		backY := 500
		var back image.Image
		if path.Ext(u.Face) != ".webp" {
			err = initFacePic(facePath, u.Face)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			back, err = gg.LoadImage(facePath)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			back = imgfactory.Size(back, backX, backY).Image()
		}
		canvas := gg.NewContext(100, 100)
		fontSize := 50.0
		data, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
		}
		if err = canvas.ParseFontFace(data, fontSize); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		dz, h := canvas.MeasureString("好")
		danmuH := h * 2
		faceH := float64(510)

		totalDanmuku := 0
		for i := 0; i < len(danmaku.Data.Data.Records); i++ {
			totalDanmuku += len(danmaku.Data.Data.Records[i].Danmakus) + 1
		}
		cw := 3000
		mcw := float64(2000)
		ch := 550 + len(danmaku.Data.Data.Records)*int(faceH) + totalDanmuku*int(danmuH)
		canvas = gg.NewContext(cw, ch)
		canvas.SetColor(color.White)
		canvas.Clear()
		canvas.SetColor(color.Black)
		if err = canvas.ParseFontFace(data, fontSize); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		facestart := 100
		fontH := h * 1.6
		startWidth := float64(700)
		startWidth2 := float64(20)

		if back != nil {
			canvas.DrawImage(back, facestart, 0)
		}
		length, _ := canvas.MeasureString(u.Mid)
		n, _ := canvas.MeasureString(u.Name)
		canvas.DrawString(u.Name, startWidth, 122.5)
		canvas.DrawRoundedRectangle(900+n-length*0.1, 66, length*1.2, 75, fontSize*0.2)
		canvas.SetRGB255(221, 221, 221)
		canvas.Fill()
		canvas.SetColor(color.Black)
		canvas.DrawString(u.Mid, 900+n, 122.5)
		canvas.DrawString(fmt.Sprintf("粉丝：%d   关注：%d", u.Fans, u.Attention), startWidth, 222.5)
		canvas.DrawString(fmt.Sprintf("页码：[%d/%d]", danmaku.Data.PageNum, (danmaku.Data.Total-1)/5), startWidth, 322.5)
		canvas.DrawString("网页链接: "+fmt.Sprintf(bz.DanmakuURL, u.Mid), startWidth, 422.5)
		var channelStart float64
		channelStart = float64(550)
		for i := 0; i < len(danmaku.Data.Data.Records); i++ {
			item := danmaku.Data.Data.Records[i]
			facePath = cachePath + strconv.Itoa(item.Channel.UID) + "vupFace" + path.Ext(item.Channel.FaceURL)
			if path.Ext(item.Channel.FaceURL) != ".webp" {
				err = initFacePic(facePath, item.Channel.FaceURL)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				back, err = gg.LoadImage(facePath)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				back = imgfactory.Size(back, backX, backY).Image()
			}
			if back != nil {
				canvas.DrawImage(back, facestart, int(channelStart))
			}
			canvas.SetRGB255(24, 144, 255)
			canvas.DrawString("标题: "+item.Live.Title, startWidth, channelStart+fontH)
			canvas.DrawString("主播: "+item.Channel.UName, startWidth, channelStart+fontH*2)
			canvas.SetColor(color.Black)
			canvas.DrawString("开始时间: "+time.UnixMilli(item.Live.StartDate).Format("2006-01-02 15:04:05"), startWidth, channelStart+fontH*3)
			if item.Live.IsFinish {
				canvas.DrawString("结束时间: "+time.UnixMilli(item.Live.StopDate).Format("2006-01-02 15:04:05"), startWidth, channelStart+fontH*4)
				canvas.DrawString("直播时长: "+strconv.FormatFloat(float64(item.Live.StopDate-item.Live.StartDate)/3600000.0, 'f', 1, 64)+"小时", startWidth, channelStart+fontH*5)
			} else {
				t := "结束时间:"
				l, _ := canvas.MeasureString(t)
				canvas.DrawString(t, startWidth, channelStart+fontH*4)

				canvas.SetRGB255(0, 128, 0)
				t = "正在直播"
				canvas.DrawString(t, startWidth+l*1.1, channelStart+fontH*4)
				canvas.SetColor(color.Black)

				canvas.DrawString("直播时长: "+strconv.FormatFloat(float64(time.Now().UnixMilli()-item.Live.StartDate)/3600000.0, 'f', 1, 64)+"小时", startWidth, channelStart+fontH*5)
			}
			canvas.DrawString("弹幕数量: "+strconv.Itoa(item.Live.DanmakusCount), startWidth, channelStart+fontH*6)
			canvas.DrawString("观看次数: "+strconv.Itoa(item.Live.WatchCount), startWidth, channelStart+fontH*7)

			t := "收益:"
			l, _ := canvas.MeasureString(t)
			canvas.DrawString(t, startWidth, channelStart+fontH*8)

			t = "￥" + strconv.Itoa(int(item.Live.TotalIncome))
			canvas.SetRGB255(255, 0, 0)
			canvas.DrawString(t, startWidth+l*1.1, channelStart+fontH*8)
			canvas.SetColor(color.Black)

			DanmakuStart := channelStart + faceH
			for i := 0; i < len(item.Danmakus); i++ {
				moveW := startWidth2
				danmuNow := DanmakuStart + danmuH*float64(i+1)
				danItem := item.Danmakus[i]

				t := time.UnixMilli(danItem.SendDate).Format("15:04:05")
				l, _ := canvas.MeasureString(t)
				canvas.DrawString(t, moveW, danmuNow)
				moveW += l + dz

				t = danItem.UName
				l, _ = canvas.MeasureString(t)
				canvas.SetRGB255(24, 144, 255)
				canvas.DrawString(t, moveW, danmuNow)
				canvas.SetColor(color.Black)
				moveW += l + dz

				switch danItem.Type {
				case 0:
					t = danItem.Message
					l, _ = canvas.MeasureString(t)
					canvas.DrawString(t, moveW, danmuNow)
					moveW += l + dz
				case 1:
					t = danmakuTypeMap[danItem.Type]
					l, _ = canvas.MeasureString(t)
					canvas.SetRGB255(255, 0, 0)
					canvas.DrawString(t, moveW, danmuNow)
					moveW += l + dz

					t = danItem.Message
					l, _ = canvas.MeasureString(t)
					canvas.DrawString(t, moveW, danmuNow)
					canvas.SetColor(color.Black)
					moveW += l + dz
				case 2, 3:
					t = danmakuTypeMap[danItem.Type]
					l, _ = canvas.MeasureString(t)
					if danItem.Type == 3 {
						canvas.SetRGB255(0, 85, 255)
					} else {
						canvas.SetRGB255(128, 0, 128)
					}

					canvas.DrawString(t, moveW, danmuNow)
					moveW += l + dz

					t = danItem.Message
					l, _ = canvas.MeasureString(t)
					canvas.DrawString(t, moveW, danmuNow)
					moveW += l

					t = "["
					l, _ = canvas.MeasureString(t)
					canvas.DrawString(t, moveW, danmuNow)
					moveW += l

					t = "￥" + strconv.FormatFloat(danItem.Price, 'f', 1, 64)
					l, _ = canvas.MeasureString(t)
					canvas.SetRGB255(255, 0, 0)
					canvas.DrawString(t, moveW, danmuNow)
					if danItem.Type == 3 {
						canvas.SetRGB255(0, 85, 255)
					} else {
						canvas.SetRGB255(128, 0, 128)
					}
					moveW += l

					t = "]"
					l, _ = canvas.MeasureString(t)
					canvas.DrawString(t, moveW, danmuNow)
					canvas.SetColor(color.Black)
					moveW += l + dz
				case 4, 5, 6, 7, 8:
					t = danmakuTypeMap[danItem.Type]
					canvas.SetRGB255(0, 128, 0)
					l, _ = canvas.MeasureString(t)
					canvas.DrawString(t, moveW, danmuNow)
					canvas.SetColor(color.Black)
					moveW += l + dz
				default:
					canvas.SetRGB255(0, 128, 0)
					l, _ = canvas.MeasureString("未知类型" + strconv.Itoa(int(danItem.Type)))
					canvas.DrawString(t, moveW, danmuNow)
					canvas.SetColor(color.Black)
					moveW += l + dz
				}
				if moveW > mcw {
					mcw = moveW
				}
			}
			channelStart = DanmakuStart + float64(len(item.Danmakus)+1)*danmuH
		}
		im := canvas.Image().(*image.RGBA)
		nim := im.SubImage(image.Rect(0, 0, int(mcw), ch))
		f, err := os.Create(drawedFile)
		if err != nil {
			log.Errorln("[bilibili]", err)
			data, err := imgfactory.ToBytes(nim)
			if err != nil {
				log.Errorln("[bilibili]", err)
				return
			}
			ctx.SendChain(message.ImageBytes(data))
			return
		}
		_, err = imgfactory.WriteTo(nim, f)
		_ = f.Close()
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
	})

	engine.OnRegex(`^设置b站cookie?\s+(.*)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			cookie := ctx.State["regex_matched"].([]string)[1]
			err := cfg.Set(cookie)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置b站cookie为" + cookie))
		})

	engine.OnFullMatch("更新vup", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			err := updateVup()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("vup已更新"))
		})
}

func initFacePic(filename, faceURL string) error {
	if file.IsNotExist(filename) {
		data, err := web.GetData(faceURL)
		if err != nil {
			return err
		}
		err = os.WriteFile(filename, data, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

func int2rbg(t int64) (int64, int64, int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(t))
	b, g, r := int64(buf[0]), int64(buf[1]), int64(buf[2])
	return r, g, b
}

func getPara(ctx *zero.Ctx) bool {
	keyword := ctx.State["regex_matched"].([]string)[1]
	if !re.MatchString(keyword) {
		searchRes, err := bz.SearchUser(cfg, keyword)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		ctx.State["uid"] = strconv.FormatInt(searchRes[0].Mid, 10)
		return true
	}
	next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession())
	recv, cancel := next.Repeat()
	defer cancel()
	ctx.SendChain(message.Text("输入为纯数字, 请选择查询uid还是用户名, 输入对应序号：\n0. 查询uid\n1. 查询用户名"))
	for {
		select {
		case <-time.After(time.Second * 10):
			ctx.SendChain(message.Text("时间太久啦！", zero.BotConfig.NickName[0], "帮你选择查询uid"))
			ctx.State["uid"] = keyword
			return true
		case c := <-recv:
			msg := c.Event.Message.ExtractPlainText()
			num, err := strconv.Atoi(msg)
			if err != nil {
				ctx.SendChain(message.Text("请输入数字!"))
				continue
			}
			if num < 0 || num > 1 {
				ctx.SendChain(message.Text("序号非法!"))
				continue
			}
			if num == 0 {
				ctx.State["uid"] = keyword
				return true
			} else if num == 1 {
				searchRes, err := bz.SearchUser(cfg, keyword)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return false
				}
				ctx.State["uid"] = strconv.FormatInt(searchRes[0].Mid, 10)
				return true
			}
		}
	}
}
