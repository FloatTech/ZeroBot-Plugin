// Package guessmusic 基于zbp的猜歌插件
package guessmusic

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 网易云插件
	wyy "github.com/FloatTech/AnimeAPI/neteasemusic"

	// 图片输出
	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/zbputils/img/text"
)

const servicename = "guessmusic"

var (
	filelist      []listinfo
	musictypelist = "mp3;MP3;wav;WAV;amr;AMR;3gp;3GP;3gpp;3GPP;acc;ACC"
	cuttime       = [...]string{"00:00:05", "00:00:30", "00:01:00"} // 音乐切割时间点，可自行调节时间（时：分：秒）
	cfg           config
)

func init() { // 插件主体
	engine := control.Register(servicename, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "猜歌插件（该插件依赖ffmpeg）\n" +
			"由于不可抗因素无法获取网易云歌单内容,\n插件改为本地猜歌了，但保留了下歌功能\n" +
			"------bot主人指令------\n" +
			"- 设置猜歌歌库路径 [绝对路径]\n" +
			"-(指令仅歌词有效) 猜歌[开启/关闭][歌单/歌词]自动下载\n" +
			"-(指令已失效) 添加歌单 [网易云歌单链接/ID] [歌单名称]\n" +
			"- 下载歌曲 [歌曲名称/网易云歌曲ID] [歌单名称]\n" +
			"- 删除歌单 [网易云歌单ID/歌单名称]\n" +
			"注：\n删除网易云歌单ID仅只是解除绑定\n删除歌单名称是将本地数据全部删除，慎用\n" +
			"------管 理 员 指 令------\n" +
			"- 设置猜歌默认歌单 [歌单名称]\n" +
			"------公 用 指 令------\n" +
			"- 歌单列表\n" +
			"- [个人/团队]猜歌\n" +
			"注：默认歌库为歌单列表第一个\n如果设置了默认歌单变为指定的歌单\n" +
			"可在“[个人/团队]猜歌指令”后面添加[-歌单名称]进行指定歌单猜歌\n" +
			"猜歌内容必须以[-]开头才会识别\n" +
			"本地歌曲命名规则为:\n歌名 - 歌手 - 其他(歌曲出处之类)",
		PrivateDataFolder: "guessmusic",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("已经有正在进行的游戏..."),
				),
			)
		}),
	))
	serviceErr := "[" + servicename + "]"
	// 用于存放歌曲三个片段的文件夹
	cachePath := engine.DataFolder() + "cache/"
	err := os.MkdirAll(cachePath, 0777)
	if err != nil {
		panic(serviceErr + "ERROR:" + err.Error())
	}
	// 获取用户的配置
	cfgFile := engine.DataFolder() + "config.json"
	if file.IsExist(cfgFile) {
		reader, err := os.Open(cfgFile)
		if err == nil {
			err = json.NewDecoder(reader).Decode(&cfg)
		}
		if err != nil {
			panic(serviceErr + "ERROR:" + err.Error())
		}
		err = reader.Close()
		if err != nil {
			panic(serviceErr + "ERROR:" + err.Error())
		}
	} else {
		cfg = config{ // 配置默认 config
			MusicPath: file.BOTPATH + "/data/guessmusic/music/", // 绝对路径，歌库根目录,通过指令进行更改
			API:       true,
			Local:     true,
			Playlist: []listRaw{
				{
					Name: "FM",
					ID:   3136952023,
				}},
		}
		err = saveConfig(cfgFile)
		if err != nil {
			panic(serviceErr + "ERROR:" + err.Error())
		}
	}
	filelist, err = getlist(cfg.MusicPath)
	if err != nil {
		logrus.Errorln(serviceErr + "ERROR:" + err.Error())
	}
	// 用户配置
	engine.OnRegex(`^设置猜歌(歌库路径|默认歌单)\s*(.*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["regex_matched"].([]string)[1]
			value := ctx.State["regex_matched"].([]string)[2]
			var err error
			switch option {
			case "歌库路径":
				if !zero.SuperUserPermission(ctx) {
					ctx.SendChain(message.Text("只有bot主人可以设置！"))
					return
				}
				musicPath := strings.ReplaceAll(value, "\\", "/")
				if !strings.HasSuffix(musicPath, "/") {
					musicPath += "/"
				}
				err = os.MkdirAll(musicPath, 0777)
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, "生成文件夹ERROR:\n", err))
					return
				}
				cfg.MusicPath = musicPath
			case "默认歌单":
				gid := ctx.Event.GroupID
				if gid == 0 || !zero.AdminPermission(ctx) {
					ctx.SendChain(message.Text("无权设置！"))
					return
				}
				index := ""
				for _, listinfo := range filelist {
					if listinfo.Name == value {
						index = value
						break
					}
				}
				if index == "" {
					ctx.SendChain(message.Text("歌单名称错误，可以发送“歌单列表”获取歌单名称"))
					return
				}
				cfg.Defaultlist = append(cfg.Defaultlist, dlist{
					GroupID: gid,
					Name:    value,
				})
			}
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, "ERROR:\n", err))
			}
		})
	engine.OnRegex(`^猜歌(开启|关闭)(歌单|歌词)自动下载`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			swtich := ctx.State["regex_matched"].([]string)[1]
			option := ctx.State["regex_matched"].([]string)[1]
			chose := true
			if swtich == "关闭" {
				chose = false
			}
			if option == "歌单" {
				cfg.API = chose
			} else {
				cfg.Local = chose
			}
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, "ERROR:\n", err))
			}
		})
	// 本地绑定网易云歌单ID
	engine.OnRegex(`^添加歌单\s?(https:.*id=)?(\d+)\s?(.*)$`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			listID := ctx.State["regex_matched"].([]string)[2]
			listName := ctx.State["regex_matched"].([]string)[3]
			ctx.SendChain(message.Text("正在校验歌单信息，请稍等"))
			// 是否存在该歌单
			apiURL := "https://ovooa.com/API/163_Music_Rand/api.php?id=" + listID
			data, err := web.GetData(apiURL)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "error:", err))
				return
			}
			var parsed ovooaData
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "无法解析歌单ID内容:", err))
				return
			}
			if parsed.Code != 1 {
				ctx.SendChain(message.Text(serviceErr, "error:", parsed.Text))
				return
			}
			pathOfMusic := cfg.MusicPath + listName + "/"
			err = os.MkdirAll(pathOfMusic, 0777)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "歌单不存在于本地，尝试创建该歌单失败:\n", err))
				return
			}
			mid, _ := strconv.ParseInt(listID, 10, 64)
			cfg.Playlist = append(cfg.Playlist, listRaw{
				Name: listName,
				ID:   mid,
			})
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, "error:", err))
			}
		})
	engine.OnRegex(`^删除歌单\s?(.*)$`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			delList := ctx.State["regex_matched"].([]string)[1]
			filelist, err = getlist(cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "歌单列表获取error:", err))
				return
			}
			index := 1024
			for i, listinfo := range filelist {
				if delList == listinfo.Name || delList == strconv.FormatInt(listinfo.ID, 10) {
					if delList == listinfo.Name {
						err = os.RemoveAll(cfg.MusicPath + delList)
						if err != nil {
							ctx.SendChain(message.Text("歌单文件删除失败：\n", err))
							return
						}
					}
					index = i
					break
				}
			}
			if index == 1024 {
				ctx.SendChain(message.Text("歌单名称错误，可以发送“歌单列表”获取歌单名称"))
				return
			}
			var newCatList []listRaw
			for _, list := range cfg.Playlist {
				if list.Name == filelist[index].Name {
					continue
				}
				newCatList = append(newCatList, list)
			}
			cfg.Playlist = newCatList
			err = saveConfig(cfgFile)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "ERROR:", err))
			}
			filelist, err = getlist(cfg.MusicPath)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, "ERROR:", err))
			}
		})
	// 下载歌曲到对应的歌单里面
	engine.OnRegex(`^下载歌曲\s?(\d+|.*[^\s$])\s(.*[^\s$])$`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			listName := ctx.State["regex_matched"].([]string)[2]
			ctx.SendChain(message.Text("正在校验歌单信息，请稍等"))
			// 是否存在该歌单
			filelist, err := getlist(cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "获取歌单列表ERROR:", err))
				return
			}
			ok := true
			for _, listinfo := range filelist {
				if listName == listinfo.Name {
					ok = false
					break
				}
			}
			if ok {
				ctx.SendChain(message.Text("歌单不存在，是否创建？(是/否)"))
				next := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`(是|否)`), ctx.CheckSession())
				recv, cancel := next.Repeat()
				defer cancel()
				wait := time.NewTimer(120 * time.Second)
				answer := ""
				for {
					select {
					case <-wait.C:
						wait.Stop()
						ctx.SendChain(message.Text("等待超时，取消下载"))
						return
					case c := <-recv:
						wait.Stop()
						answer = c.Event.Message.String()
					}
					if answer == "否" {
						ctx.SendChain(message.Text("下载已经取消"))
						return
					}
					if answer != "" {
						break
					}
				}
				err = os.MkdirAll(cfg.MusicPath+listName, 0777)
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, "生成文件夹ERROR:\n", err))
					return
				}
			}
			searchlist, err := wyy.SearchMusic(keyword, 5)
			if err != nil {
				ctx.SendChain(message.Text("查询歌曲失败！\nerr:", err))
				return
			}
			listmun := len(searchlist)
			if listmun == 0 {
				ctx.SendChain(message.Text("歌曲没有查询到，请确认信息正确"))
				return
			}
			musicList := make([]string, listmun)
			i := 0
			for musicName := range searchlist {
				musicList[i] = musicName
				i++
			}
			savePath := cfg.MusicPath + listName + "/"
			if listmun == 1 {
				musicName := musicList[0]
				musicID := searchlist[musicName]
				// 下载歌曲
				err = wyy.DownloadMusic(musicID, musicName, savePath)
				if err == nil {
					if cfg.Local {
						// 下载歌词
						_ = wyy.DownloadLrc(musicID, musicName, savePath+"歌词/")
					}
					ctx.SendChain(message.Text("成功！"))
				} else {
					ctx.SendChain(message.Text(serviceErr, "error:", err))
				}
				return
			}
			var msg []string
			msg = append(msg, "搜索到相近的歌曲，请回复对应序号进行下载或回复取消")
			for j, musicName := range musicList {
				msg = append(msg, strconv.Itoa(j)+"."+musicName)
			}
			ctx.SendChain(message.Text(strings.Join(msg, "\n")))
			next := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`[0-4]|取消`), ctx.CheckSession())
			recv, cancel := next.Repeat()
			defer cancel()
			wait := time.NewTimer(120 * time.Second)
			for {
				select {
				case <-wait.C:
					wait.Stop()
					ctx.SendChain(message.Text("等待超时，取消下载"))
					return
				case c := <-recv:
					wait.Stop()
					answer := c.Event.Message.String()
					if answer == "取消" {
						ctx.SendChain(message.Text("已取消下载"))
						return
					}
					index, _ := strconv.Atoi(answer)
					// 下载歌曲
					musicName := musicList[index]
					err = wyy.DownloadMusic(searchlist[musicName], musicName, savePath)
					if err == nil {
						if cfg.Local {
							// 下载歌词
							_ = wyy.DownloadLrc(searchlist[musicName], musicName, savePath+"歌词/")
						}
						ctx.SendChain(message.Text("成功！"))
					} else {
						ctx.SendChain(message.Text(serviceErr, "error:", err))
					}
					return
				}
			}
		})
	engine.OnFullMatch("歌单列表").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			filelist, err := getlist(cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "获取歌单列表ERROR:", err))
				return
			}
			/***********设置图片的大小和底色***********/
			number := len(filelist)
			fontSize := 20.0
			if number < 10 {
				number = 10
			}
			canvas := gg.NewContext(480, int(80+fontSize*float64(number)))
			canvas.SetRGB(1, 1, 1) // 白色
			canvas.Clear()
			/***********下载字体，可以注销掉***********/
			_, err = file.GetLazyData(text.BoldFontFile, true)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "ERROR:", err))
			}
			_, err = file.GetLazyData(text.FontFile, true)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "ERROR:", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
				ctx.SendChain(message.Text(serviceErr, "ERROR:", err))
				return
			}
			_, h := canvas.MeasureString("序号\t\t歌单名\t\t\t歌曲数量\t\t网易云歌单ID")
			/***********绘制标题***********/
			canvas.DrawString("序号\t\t歌单名\t\t歌曲数量\t\t网易云歌单ID", 20, 50-h) // 放置在中间位置
			canvas.DrawString("——————————————————————", 20, 70-h)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.FontFile, fontSize); err != nil {
				ctx.SendChain(message.Text(serviceErr, "ERROR:", err))
				return
			}
			_, h = canvas.MeasureString("焯")
			j := 0
			for i, listinfo := range filelist {
				canvas.DrawString(strconv.Itoa(i), 15, float64(85+20*i)-h)
				canvas.DrawString(listinfo.Name, 85, float64(85+20*i)-h)
				canvas.DrawString(strconv.Itoa(listinfo.Number), 220, float64(85+20*i)-h)
				if listinfo.ID != 0 {
					canvas.DrawString(strconv.FormatInt(listinfo.ID, 10), 320, float64(85+20*i)-h)
				}
				j = i + 2
			}
			for _, dlist := range cfg.Defaultlist {
				if dlist.GroupID == ctx.Event.GroupID {
					canvas.DrawString("当前设置的默认歌单为: "+dlist.Name, 80, float64(85+20*j)-h)
				}
			}
			data, cl := writer.ToBytes(canvas.Image())
			if id := ctx.SendChain(message.ImageBytes(data)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
			cl()
		})
	engine.OnRegex(`^(个人|团队)猜歌(-(.*))?$`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			mode := ctx.State["regex_matched"].([]string)[3]
			gid := ctx.Event.GroupID
			filelist, err := getlist(cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "获取歌单列表ERROR:", err))
				return
			}
			if mode == "" {
				for _, dlist := range cfg.Defaultlist {
					if dlist.GroupID == gid {
						mode = dlist.Name
						break
					}
				}
			}
			if mode == "" {
				mode = filelist[0].Name
			} else {
				ok := true
				for _, listinfo := range filelist {
					if mode == listinfo.Name {
						ok = false
						break
					}
				}
				if ok {
					ctx.SendChain(message.Text("歌单名称错误，可以发送“歌单列表”获取歌单名称"))
					return
				}
			}
			ctx.SendChain(message.Text("正在准备歌曲,请稍等\n回答“-[歌曲信息(歌名歌手等)|提示|取消]”\n一共3段语音，6次机会"))
			// 随机抽歌
			pathOfMusic, musicName, err := musicLottery(cfg.MusicPath, mode)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "ERROR:", err))
				return
			}
			// 解析歌曲信息
			music := strings.Split(musicName, ".")
			// 获取音乐后缀
			musictype := music[len(music)-1]
			if !strings.Contains(musictypelist, musictype) {
				ctx.SendChain(message.Text("抽取到了歌曲：\n",
					musicName, "\n该歌曲不是音乐后缀，请联系bot主人修改"))
				return
			}
			// 获取音乐信息
			musicInfo := strings.Split(strings.ReplaceAll(musicName, "."+musictype, ""), " - ")
			infoNum := len(musicInfo)
			if infoNum == 1 {
				ctx.SendChain(message.Text("抽取到了歌曲：\n",
					musicName, "\n该歌曲命名不符合命名规则，请联系bot主人修改"))
				return
			}
			answerString := "歌名:" + musicInfo[0] + "\n歌手:" + musicInfo[1]
			musicAlia := ""
			if infoNum > 2 {
				musicAlia = musicInfo[2]
				answerString += "\n其他信息:\n" + strings.ReplaceAll(musicAlia, "&", "\n")
			}
			// 切割音频，生成3个10秒的音频
			outputPath := cachePath + strconv.FormatInt(gid, 10) + "/"
			err = cutMusic(musicName, pathOfMusic, outputPath)
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			// 进行猜歌环节
			ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + "0.wav"))
			var next *zero.FutureEvent
			if ctx.State["regex_matched"].([]string)[1] == "个人" {
				next = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^-\S{1,}`), ctx.CheckSession())
			} else {
				next = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^-\S{1,}`), zero.CheckGroup(ctx.Event.GroupID))
			}
			var musicCount = 0  // 音频数量
			var answerCount = 0 // 问答次数
			recv, cancel := next.Repeat()
			defer cancel()
			wait := time.NewTimer(40 * time.Second)
			tick := time.NewTimer(105 * time.Second)
			after := time.NewTimer(120 * time.Second)
			for {
				select {
				case <-tick.C:
					ctx.SendChain(message.Text("猜歌游戏，你还有15s作答时间"))
				case <-after.C:
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("时间超时，猜歌结束，公布答案：\n", answerString)))
					return
				case <-wait.C:
					wait.Reset(40 * time.Second)
					musicCount++
					if musicCount > 2 {
						wait.Stop()
						continue
					}
					ctx.SendChain(
						message.Text("好像有些难度呢，再听这段音频，要仔细听哦"),
					)
					ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + strconv.Itoa(musicCount) + ".wav"))
				case c := <-recv:
					wait.Reset(40 * time.Second)
					tick.Reset(105 * time.Second)
					after.Reset(120 * time.Second)
					answer := strings.Replace(c.Event.Message.String(), "-", "", 1)
					switch {
					case answer == "取消":
						if c.Event.UserID == ctx.Event.UserID {
							wait.Stop()
							tick.Stop()
							after.Stop()
							ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
								message.Text("游戏已取消，猜歌答案是\n", answerString, "\n\n\n下面欣赏猜歌的歌曲")))
							ctx.SendChain(message.Record("file:///" + pathOfMusic + musicName))
							return
						}
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("你无权限取消"),
							),
						)
					case answer == "提示":
						musicCount++
						if musicCount > 2 {
							wait.Stop()
							ctx.Send(
								message.ReplyWithMessage(c.Event.MessageID,
									message.Text("已经没有提示了哦"),
								),
							)
							continue
						}
						wait.Reset(40 * time.Second)
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("再听这段音频，要仔细听哦"),
							),
						)
						ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + strconv.Itoa(musicCount) + ".wav"))
					case strings.Contains(musicInfo[0], answer) || strings.EqualFold(musicInfo[0], answer):
						wait.Stop()
						tick.Stop()
						after.Stop()
						ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
							message.Text("太棒了，你猜对歌曲名了！答案是\n", answerString, "\n\n下面欣赏猜歌的歌曲")))
						ctx.SendChain(message.Record("file:///" + pathOfMusic + musicName))
						return
					case strings.Contains(musicInfo[1], answer) || strings.EqualFold(musicInfo[1], answer):
						wait.Stop()
						tick.Stop()
						after.Stop()
						ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
							message.Text("太棒了，你猜对歌手名了！答案是\n", answerString, "\n\n下面欣赏猜歌的歌曲")))
						ctx.SendChain(message.Record("file:///" + pathOfMusic + musicName))
						return
					case strings.Contains(musicAlia, answer) || strings.EqualFold(musicAlia, answer):
						wait.Stop()
						tick.Stop()
						after.Stop()
						ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
							message.Text("太棒了，你猜对出处了！答案是\n", answerString, "\n\n下面欣赏猜歌的歌曲")))
						ctx.SendChain(message.Record("file:///" + pathOfMusic + musicName))
						return
					default:
						musicCount++
						switch {
						case musicCount > 2 && answerCount < 6:
							wait.Stop()
							answerCount++
							ctx.Send(
								message.ReplyWithMessage(c.Event.MessageID,
									message.Text("答案不对哦，加油啊~"),
								),
							)
						case musicCount > 2:
							wait.Stop()
							tick.Stop()
							after.Stop()
							ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
								message.Text("次数到了，没能猜出来。答案是\n", answerString, "\n\n下面欣赏猜歌的歌曲")))
							ctx.SendChain(message.Record("file:///" + pathOfMusic + musicName))
							return
						default:
							wait.Reset(40 * time.Second)
							answerCount++
							ctx.Send(
								message.ReplyWithMessage(c.Event.MessageID,
									message.Text("答案不对，再听这段音频，要仔细听哦"),
								),
							)
							ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + strconv.Itoa(musicCount) + ".wav"))
						}
					}
				}
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

// 获取本地歌单列表
func getlist(pathOfMusic string) (list []listinfo, err error) {
	wyyID := make(map[string]int64, 100)
	for _, wyyinfo := range cfg.Playlist {
		if wyyinfo.ID != 0 {
			wyyID[wyyinfo.Name] = wyyinfo.ID
		}
	}
	err = os.MkdirAll(pathOfMusic, 0777)
	if err != nil {
		return
	}
	files, err := os.ReadDir(pathOfMusic)
	if err != nil {
		return
	}
	if len(files) == 0 {
		err = errors.Errorf("所设置的歌库不存在任何歌单！")
		return
	}
	for _, name := range files {
		if !name.IsDir() {
			continue
		}
		listName := name.Name()
		listfiles, err := os.ReadDir(pathOfMusic + listName)
		if err != nil {
			continue
		}
		list = append(list, listinfo{
			Name:   listName,
			Number: len(listfiles),
			ID:     wyyID[listName],
		})
	}
	return
}

// 随机抽取音乐
func musicLottery(musicPath, listName string) (pathOfMusic, musicName string, err error) {
	filelist, err := getlist(musicPath)
	if err != nil {
		err = errors.Errorf("获取列表错误,%s", err)
		return
	}
	var fileList = make(map[string]int64, 100)
	for _, listinfo := range filelist {
		fileList[listinfo.Name] = listinfo.ID
	}
	playlistID, ok := fileList[listName]
	if !ok {
		err = errors.Errorf("指定的歌单不存在与列表当中")
		return
	}
	pathOfMusic = musicPath + listName + "/"
	err = os.MkdirAll(pathOfMusic, 0777)
	if err != nil {
		return
	}
	files, err := os.ReadDir(pathOfMusic)
	if err != nil {
		return
	}
	//如果本地列表为空
	if len(files) == 0 {
		if playlistID == 0 || !cfg.API {
			err = errors.New("本地歌单数据为0")
			return
		}
		// 如果绑定了歌单ID
		musicName, err = downloadByOvooa(playlistID, pathOfMusic)
		err = errors.Errorf("本地歌单数据为0,API下载歌曲失败\n%s", err)
		return
	}
	// 进行随机抽取
	if playlistID == 0 || !cfg.API {
		musicName = getLocalMusic(files)
	} else {
		switch rand.Intn(3) { //三分二概率抽取API的
		case 1:
			musicName = getLocalMusic(files)
		default:
			musicName, err = downloadByOvooa(playlistID, pathOfMusic)
			if err != nil {
				musicName = getLocalMusic(files)
				err = nil
				return
			}
		}
	}
	return
}

// 从本地列表中随机抽取一首
func getLocalMusic(files []fs.DirEntry) (musicName string) {
	if len(files) > 1 {
		music := files[rand.Intn(len(files))]
		// 如果是文件夹就递归
		if music.IsDir() {
			musicName = getLocalMusic(files)
		} else {
			musicName = music.Name()
		}
	} else {
		music := files[0]
		if !music.IsDir() {
			musicName = files[0].Name()
		}
	}
	return
}

// 下载从独角兽抽到的歌曲ID(歌单ID, 音乐保存路径, 歌词保存路径)
func downloadByOvooa(playlistID int64, musicPath string) (musicName string, err error) {
	// 抽取歌曲
	mid, err := drawByOvooa(playlistID)
	if err != nil {
		err = errors.Errorf("API ERROR: %s", err)
		return
	}
	// 获取完成的歌名
	musiclist, err := wyy.SearchMusic(strconv.Itoa(mid), 1)
	if err != nil {
		err = errors.Errorf("API歌曲下载ERROR: %s", err)
		return
	}
	// 歌曲ID理论是唯一的
	mun := len(musiclist)
	if mun == 1 {
		// 拉取歌名
		musicList := make([]string, mun)
		i := 0
		for musicName := range musiclist {
			musicList[i] = musicName
		}
		name := musicList[0]
		// 下载歌曲
		err = wyy.DownloadMusic(mid, name, musicPath)
		if err == nil {
			musicName = name + ".mp3"
			if cfg.Local {
				// 下载歌词
				_ = wyy.DownloadLrc(mid, name, musicPath+"歌词/")
			}
		}
	} else {
		err = errors.Errorf("music ID ERROR: This music ID sreached munber is %d", mun)
	}
	return
}

// 通过独角兽API随机抽取歌单歌曲ID(参数：歌单ID)
func drawByOvooa(playlistID int64) (musicID int, err error) {
	apiURL := "https://ovooa.com/API/163_Music_Rand/api.php?id=" + strconv.FormatInt(playlistID, 10)
	data, err := web.GetData(apiURL)
	if err != nil {
		return
	}
	var parsed ovooaData
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	if parsed.Code != 1 {
		return
	}
	return parsed.Data.ID, nil
}

// 切割音乐成三个10s音频
func cutMusic(musicName, pathOfMusic, outputPath string) (err error) {
	err = os.MkdirAll(outputPath, 0777)
	if err != nil {
		err = errors.Errorf("[生成歌曲目录错误]ERROR: %s", err)
		return
	}
	var stderr bytes.Buffer
	cmdArguments := []string{"-y", "-i", pathOfMusic + musicName,
		"-ss", cuttime[0], "-t", "10", file.BOTPATH + "/" + outputPath + "0.wav",
		"-ss", cuttime[1], "-t", "10", file.BOTPATH + "/" + outputPath + "1.wav",
		"-ss", cuttime[2], "-t", "10", file.BOTPATH + "/" + outputPath + "2.wav", "-hide_banner"}
	cmd := exec.Command("ffmpeg", cmdArguments...)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		err = errors.Errorf("[生成歌曲错误]ERROR: %s", stderr.String())
		return
	}
	return
}
