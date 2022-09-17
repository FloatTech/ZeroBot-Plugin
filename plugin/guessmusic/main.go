// Package guessmusic 基于zbp的猜歌插件
package guessmusic

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/wdvxdr1123/ZeroBot/extension/single"

	// 图片输出
	"github.com/FloatTech/zbputils/img/text"
)

const (
	ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66"
)

var (
	catlist       = make(map[string]int64, 100)
	filelist      []string
	musictypelist = "mp3;MP3;wav;WAV;amr;AMR;3gp;3GP;3gpp;3GPP;acc;ACC"
	cuttime       = [...]string{"00:00:05", "00:00:30", "00:01:00"} // 音乐切割时间点，可自行调节时间（时：分：秒）
	cfg           config
)

func init() { // 插件主体
	engine := control.Register("guessmusic", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "猜歌插件（该插件依赖ffmpeg）\n" +
			"------bot主人指令------\n" +
			"- 设置猜歌缓存歌库路径 [绝对路径]\n" +
			"- 设置猜歌[本地/Api] [true/false]\n" +
			"- 登录网易云\n" +
			"- 添加歌单 [网易云歌单ID] [歌单名称]\n" +
			"- 删除歌单 [网易云歌单ID/API歌单名称]\n" +
			"注：\n1.不登陆也能用，API有几率返回400\n" +
			"2.[歌单名称]可为空，默认原标题\n" +
			"------公 用 指 令------\n" +
			"- 获取歌单列表\n" +
			"- [网易云歌单ID/API歌单名称]歌单信息\n" +
			"- [个人/团队]猜歌\n" +
			"注：默认歌库为网易云ACG动画榜\n" +
			"可在后面添加[-歌单名称]进行指定歌单猜歌\n" +
			"歌单的歌曲命名规则为:\n歌名 - 歌手 - 其他(歌曲出处之类)",
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
	cachePath := engine.DataFolder() + "cache/"
	err := os.MkdirAll(cachePath, 0755)
	if err != nil {
		panic(err)
	}
	cfgFile := engine.DataFolder() + "config.json"
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
		var plist = []listRaw{
			{
				Name: "动画榜",
				ID:   3001835560,
			},
		}
		cfg = config{ // 默认 config
			MusicPath: file.BOTPATH + "/data/guessmusic/music/", // 绝对路径，歌库根目录,通过指令进行更改
			Local:     true,                                     // 是否使用本地音乐库
			API:       true,                                     // 是否使用 Api
			Cookie:    "",
			Playlist:  plist,
		}
		err = saveConfig(cfgFile)
		if err != nil {
			panic(err)
		}
	}
	err = getcatlist(cfg.MusicPath)
	if err != nil {
		logrus.Infof("[guessmusic2]无法获取歌单列表,[error]：%s", err)
	}
	engine.OnRegex(`^设置猜歌(缓存歌库路径|本地|Api)\s*(.*)$`, func(ctx *zero.Ctx) bool {
		if !zero.SuperUserPermission(ctx) {
			ctx.SendChain(message.Text("只有bot主人可以设置！"))
			return false
		}
		return true
	}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["regex_matched"].([]string)[1]
			value := ctx.State["regex_matched"].([]string)[2]
			switch option {
			case "缓存歌库路径":
				if value == "" {
					ctx.SendChain(message.Text("请输入正确的路径!"))
					return
				}
				musicPath := strings.ReplaceAll(value, "\\", "/")
				if !strings.HasSuffix(musicPath, "/") {
					musicPath += "/"
				}
				err = os.MkdirAll(cfg.MusicPath, 0755)
				if err != nil {
					ctx.SendChain(message.Text("[生成文件夹错误]ERROR: ", err))
					return
				}
				cfg.MusicPath = musicPath
			case "本地":
				choice, err := strconv.ParseBool(value)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				cfg.Local = choice
			case "Api":
				choice, err := strconv.ParseBool(value)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				cfg.API = choice
			}
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
		})
	engine.OnFullMatch("登录网易云", zero.SuperUserPermission, func(ctx *zero.Ctx) bool {
		if !zero.OnlyPrivate(ctx) {
			ctx.SendChain(message.Text("为了保护登录过程，请bot主人私聊。"))
			return false
		}
		return true
	}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyURL := "https://music.cyrilstudio.top/login/qr/key"
			data, err := web.GetData(keyURL)
			if err != nil {
				ctx.SendChain(message.Text("获取网易云key失败, ERROR: ", err))
				return
			}
			var keyInfo keyInfo
			err = json.Unmarshal(data, &keyInfo)
			if err != nil {
				ctx.SendChain(message.Text("解析网易云key失败, ERROR: ", err))
				return
			}
			qrURL := "https://music.cyrilstudio.top/login/qr/create?key=" + keyInfo.Data.Unikey + "&qrimg=1"
			data, err = web.GetData(qrURL)
			if err != nil {
				ctx.SendChain(message.Text("获取网易云二维码失败, ERROR: ", err))
				return
			}
			var qrInfo qrInfo
			err = json.Unmarshal(data, &qrInfo)
			if err != nil {
				ctx.SendChain(message.Text("解析网易云二维码失败, ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("[请使用手机APP扫描二维码或者进入网页扫码登录]\n", qrInfo.Data.Qrurl),
				message.Image("base64://"+strings.ReplaceAll(qrInfo.Data.Qrimg, "data:image/png;base64,", "")),
				message.Text("二维码有效时间为6分钟,登陆后请耐心等待结果，获取cookie过程有些漫长。"))
			i := 0
			for range time.NewTicker(10 * time.Second).C {
				apiURL := "https://music.cyrilstudio.top/login/qr/check?key=" + url.QueryEscape(keyInfo.Data.Unikey)
				referer := "https://music.cyrilstudio.top"
				data, err := web.RequestDataWith(web.NewDefaultClient(), apiURL, "GET", referer, ua)
				if err != nil {
					ctx.SendChain(message.Text("无法获取登录状态, ERROR: ", err))
					return
				}
				var cookiesInfo cookyInfo
				err = json.Unmarshal(data, &cookiesInfo)
				if err != nil {
					ctx.SendChain(message.Text("解析登录状态失败, ERROR: ", err))
					return
				}
				switch cookiesInfo.Code {
				case 803:
					cfg.Cookie = cookiesInfo.Cookie
					err = saveConfig(cfgFile)
					if err == nil {
						ctx.SendChain(message.Text("成功！"))
					} else {
						ctx.SendChain(message.Text("ERROR: ", err))
					}
					return
				case 801:
					i++
					if i%6 == 0 { // 每1分钟才提醒一次,减少提示(380/60=6次)
						ctx.SendChain(message.Text("状态：", cookiesInfo.Message))
					}
					continue
				case 800:
					ctx.SendChain(message.Text("状态：", cookiesInfo.Message))
					return
				default:
					ctx.SendChain(message.Text("状态：", cookiesInfo.Message))
					continue
				}
			}
		})
	engine.OnRegex(`^添加歌单\s?(\d+)(\s(.*))?$`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			listID := ctx.State["regex_matched"].([]string)[1]
			listName := ctx.State["regex_matched"].([]string)[3]
			ctx.SendChain(message.Text("正在校验歌单信息，请稍等"))
			// 是否存在该歌单
			apiURL := "https://music.cyrilstudio.top/playlist/detail?id=" + listID + "&cookie=" + cfg.Cookie
			referer := "https://music.cyrilstudio.top"
			data, err := web.RequestDataWith(web.NewDefaultClient(), apiURL, "GET", referer, ua)
			if err != nil {
				ctx.SendChain(message.Text("无法连接歌单,[error]", err))
				return
			}
			var parsed topList
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				ctx.SendChain(message.Text("无法解析歌单ID内容,[error]", err))
				return
			}
			// 是否有权限访问歌单列表内容
			apiURL = "https://music.cyrilstudio.top/playlist/track/all?id=" + listID + "&cookie=" + cfg.Cookie
			referer = "https://music.163.com/"
			data, err = web.RequestDataWith(web.NewDefaultClient(), apiURL, "GET", referer, ua)
			if err != nil {
				ctx.SendChain(message.Text("无法获取歌单列表\n ERROR: ", err))
				return
			}
			var musiclist topMusicInfo
			err = json.Unmarshal(data, &musiclist)
			if err != nil {
				ctx.SendChain(message.Text("你的cookie在API中无权访问该歌单\n该歌单有可能是用户私人歌单"))
				return
			}
			// 获取列表名字
			if listName == "" {
				listName = parsed.Playlist.Name
			}
			playID, _ := strconv.ParseInt(listID, 10, 64)
			catlist[listName] = playID
			cfg.Playlist = append(cfg.Playlist, listRaw{
				Name: listName,
				ID:   playID,
			})
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
		})
	engine.OnRegex(`^删除歌单\s?(.*)$`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			delList := ctx.State["regex_matched"].([]string)[1]
			var playlist []listRaw
			var newCatList = make(map[string]int64)
			var ok = false
			for name, musicID := range catlist {
				if delList == name || delList == strconv.FormatInt(musicID, 10) {
					ok = true
					continue
				}
				newCatList[name] = musicID
				playlist = append(playlist, listRaw{
					Name: name,
					ID:   musicID,
				})
			}
			if !ok {
				ctx.SendChain(message.Text("目标歌单未找到，请确认是否正确"))
				return
			}
			catlist = newCatList
			cfg.Playlist = playlist
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
		})
	engine.OnFullMatch("获取歌单列表").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			var msg []string
			// 获取网易云歌单列表
			if cfg.API {
				catlist = make(map[string]int64, 100)
				msg = append(msg, "当前添加的API歌单含有以下：\n")
				for i, listInfo := range cfg.Playlist {
					catlist[listInfo.Name] = listInfo.ID
					msg = append(msg, strconv.Itoa(i)+":"+listInfo.Name)
					if i%3 == 2 {
						msg = append(msg, "\n")
					}
				}
			}
			// 获取本地歌单列表*/
			if cfg.Local {
				err = os.MkdirAll(cfg.MusicPath, 0755)
				if err == nil {
					files, err := os.ReadDir(cfg.MusicPath)
					if err == nil {
						if len(files) == 0 {
							ctx.SendChain(message.Text("缓存目录没有读取到任何歌单"))
							filelist = nil
						} else {
							msg = append(msg, "\n当前本地歌单含有以下：\n")
							i := 0
							for _, name := range files {
								if !name.IsDir() {
									continue
								}
								filelist[i] = strconv.Itoa(i) + ":" + name.Name()
								msg = append(msg, filelist[i])
								if i%3 == 2 {
									msg = append(msg, "\n")
								}
								i++
							}
						}
					} else {
						ctx.SendChain(message.Text("[读取本地列表错误]ERROR: ", err))
					}
				} else {
					ctx.SendChain(message.Text("[生成文件夹错误]ERROR: ", err))
				}
			}
			if msg == nil {
				ctx.SendChain(message.Text("本地和API均未开启！"))
				return
			}
			msgs, err := text.RenderToBase64(strings.Join(msg, "    "), text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("生成列表图片失败，请重试"))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + helper.BytesToString(msgs))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	engine.OnSuffix("歌单信息").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			list := ctx.State["args"].(string)
			if list == "" {
				ctx.SendChain(message.Text("请输入歌单ID或者API歌单名称\n歌单ID为(网页/分享)链接的“playlist”后面的第一串数字"))
				return
			}
			var listIDStr string
			for listName, listID := range catlist {
				if list == listName || list == strconv.FormatInt(listID, 10) {
					listIDStr = strconv.FormatInt(listID, 10)
					break
				}
			}
			if listIDStr == "" {
				_, err := strconv.ParseInt(list, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("仅支持歌单ID查询"))
					return
				}
				listIDStr = list
			}
			apiURL := "https://music.cyrilstudio.top/playlist/detail?id=" + listIDStr + "&cookie=" + cfg.Cookie
			referer := "https://music.cyrilstudio.top"
			data, err := web.RequestDataWith(web.NewDefaultClient(), apiURL, "GET", referer, ua)
			if err != nil {
				ctx.SendChain(message.Text("无法连接歌单,[error]", err))
				return
			}
			var parsed topList
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				ctx.SendChain(message.Text("无法解析歌单ID内容,[error]", err))
				return
			}
			ctx.SendChain(
				message.Image(parsed.Playlist.CoverImgURL),
				message.Text(
					"歌单名称：", parsed.Playlist.Name,
					"\n歌单ID：", parsed.Playlist.ID,
					"\n创建人：", parsed.Playlist.Creator.Nickname,
					"\n创建时间：", time.Unix(parsed.Playlist.CreateTime/1000, 0).Format("2006-01-02"),
					"\n标签：", strings.Join(parsed.Playlist.Tags, ";"),
					"\n歌曲数量：", parsed.Playlist.TrackCount,
					"\n歌单简介:\n", parsed.Playlist.Description,
					"\n更新时间：", time.Unix(parsed.Playlist.UpdateTime/1000, 0).Format("2006-01-02"),
				))
		})
	engine.OnRegex(`^(个人|团队)猜歌(-(.*))?$`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			mode := ctx.State["regex_matched"].([]string)[3]
			if mode == "" {
				mode = "动画榜"
				catlist[mode] = 3001835560
			}
			_, ok := catlist[mode]
			// 如果本地和API不存在该歌单
			if !strings.Contains(strings.Join(filelist, " "), mode) && !ok {
				ctx.SendChain(message.Text("歌单名称错误，可以发送“获取歌单列表”获取歌单名称"))
				return
			}
			gid := strconv.FormatInt(ctx.Event.GroupID, 10)
			ctx.SendChain(message.Text("正在准备歌曲,请稍等\n回答“-[歌曲信息(歌名歌手等)|提示|取消]”\n一共3段语音，6次机会"))
			// 随机抽歌
			musicName, pathOfMusic, err := musicLottery(mode, cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			// 解析歌曲信息
			music := strings.Split(musicName, ".")
			// 获取音乐后缀
			musictype := music[len(music)-1]
			if !strings.Contains(musictypelist, musictype) {
				ctx.SendChain(message.Text("抽取到了本地歌曲：\n",
					musicName, "\n该歌曲不是音乐后缀，请联系bot主人修改"))
				return
			}
			// 获取音乐信息
			musicInfo := strings.Split(strings.ReplaceAll(musicName, "."+musictype, ""), " - ")
			infoNum := len(musicInfo)
			if infoNum == 1 {
				ctx.SendChain(message.Text("抽取到了本地歌曲：\n",
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
			outputPath := cachePath + gid + "/"
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

func getcatlist(pathOfMusic string) error {
	catlist = make(map[string]int64, 100)
	for _, listInfo := range cfg.Playlist {
		catlist[listInfo.Name] = listInfo.ID
	}
	err := os.MkdirAll(pathOfMusic, 0755)
	if err != nil {
		err = errors.Errorf("[生成文件夹错误]ERROR: %s", err)
		return err
	}
	files, err := os.ReadDir(pathOfMusic)
	if err != nil {
		err = errors.Errorf("[读取本地列表错误]ERROR: %s", err)
		return err
	}
	for i, name := range files {
		filelist = append(filelist, strconv.Itoa(i)+":"+name.Name())
	}
	return nil
}

// 随机抽取音乐
func musicLottery(mode, musicPath string) (musicName, pathOfMusic string, err error) {
	pathOfMusic = musicPath + mode + "/"
	err = os.MkdirAll(pathOfMusic, 0755)
	if err != nil {
		err = errors.Errorf("[生成文件夹错误]ERROR: %s", err)
		return
	}
	files, err := os.ReadDir(pathOfMusic)
	if err != nil {
		err = errors.Errorf("[读取本地列表错误]ERROR: %s", err)
		return
	}
	listID, ok := catlist[mode]
	listIDstr := strconv.FormatInt(listID, 10)
	if cfg.Local && cfg.API {
		switch {
		case len(files) == 0:
			if !ok {
				// 如果歌单是本地歌单
				err = errors.New("本地歌单数据为0")
				return
			}
			// 如果没有任何本地就下载歌曲
			musicName, err = getListMusic(listIDstr, pathOfMusic)
			if err != nil {
				err = errors.Errorf("[本地数据为0，歌曲下载错误]ERROR: %s", err)
				return
			}
		case rand.Intn(2) == 0 || !ok:
			// 1/2概率抽本地或者歌单只有本地有时
			musicName = getLocalMusic(files)
		default:
			musicName, err = getListMusic(listIDstr, pathOfMusic)
			if err != nil {
				// 如果下载失败就从本地抽一个歌曲
				musicName = getLocalMusic(files)
				err = nil
			}
		}
		return
	}
	if cfg.Local {
		if len(files) == 0 {
			err = errors.New("[本地数据为0，未开启API数据]")
			return
		}
		musicName = getLocalMusic(files)
		return
	}
	if cfg.API && ok {
		musicName, err = getListMusic(listIDstr, pathOfMusic)
		if err != nil {
			err = errors.Errorf("[获取API失败，未开启本地数据] ERROR: %s", err)
			return
		}
		return
	}
	err = errors.New("[请确认本地和API设置已开启或歌单存在]")
	return
}

func getLocalMusic(files []fs.DirEntry) (musicName string) {
	if len(files) > 1 {
		musicName = files[rand.Intn(len(files))].Name()
	} else {
		musicName = files[0].Name()
	}
	return
}

// 下载网易云歌单音乐
func getListMusic(listID, pathOfMusic string) (musicName string, err error) {
	apiURL := "https://music.cyrilstudio.top/playlist/track/all?id=" + listID + "&cookie=" + cfg.Cookie
	referer := "https://music.163.com/"
	data, err := web.RequestDataWith(web.NewDefaultClient(), apiURL, "GET", referer, ua)
	if err != nil {
		err = errors.Errorf("无法获取歌单列表\n ERROR: %s", err)
		return
	}
	var parsed topMusicInfo
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		err = errors.Errorf("无法读取歌单列表\n ERROR: %s", err)
		return
	}
	listlen := len(parsed.Songs)
	randidx := rand.Intn(listlen)
	// 将"/"符号去除，不然无法生成文件
	name := strings.ReplaceAll(parsed.Songs[randidx].Name, "/", "·")
	musicID := parsed.Songs[randidx].ID
	artistName := ""
	for i, ARInfo := range parsed.Songs[randidx].Ar {
		if i != 0 {
			artistName += "&" + ARInfo.Name
		} else {
			artistName += ARInfo.Name
		}
	}
	cource := ""
	if parsed.Songs[randidx].Alia != nil {
		cource = strings.Join(parsed.Songs[randidx].Alia, "&")
		// 将"/"符号去除，不然无法下载
		cource = strings.ReplaceAll(cource, "/", "&")
	}
	if name == "" || musicID == 0 {
		err = errors.New("无法获API取歌曲信息")
		return
	}
	musicURL := "http://music.163.com/song/media/outer/url?id=" + strconv.Itoa(musicID)
	response, err := http.Head(musicURL)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			err = errors.Errorf("歌曲丢失, 可能歌曲已下架或者登录状态已过期。\n可尝试重新登录排除后者问题。")
		} else {
			err = errors.Errorf("下载音乐失败, ERROR: %s", err)
		}
		return
	}
	_ = response.Body.Close()
	if response.StatusCode != 200 {
		err = errors.Errorf("下载音乐失败, Status Code: %d", response.StatusCode)
		return
	}
	if cource != "" {
		musicName = name + " - " + artistName + " - " + cource + ".mp3"
	} else {
		musicName = name + " - " + artistName + ".mp3"
	}
	downMusic := pathOfMusic + musicName
	if file.IsNotExist(downMusic) {
		data, err = web.GetData(musicURL)
		if err != nil {
			return
		}
		err = os.WriteFile(downMusic, data, 0666)
		if err != nil {
			return
		}
	}
	return
}

// 切割音乐成三个10s音频
func cutMusic(musicName, pathOfMusic, outputPath string) (err error) {
	err = os.MkdirAll(outputPath, 0755)
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
