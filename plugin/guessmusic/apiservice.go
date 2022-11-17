package guessmusic

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	wyy "github.com/FloatTech/AnimeAPI/neteasemusic"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	// API配置
	engine.OnPrefix("设置猜歌API", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["args"].(string)
			if option == "帮助" {
				ctx.SendChain(message.Text(
					"项目地址:binaryify.github.io/NeteaseCloudMusicApi" +
						"\n网上有基于该框架的API,可以自行搜索白嫖。\n" +
						"添加API指令:\n设置猜歌API [API首页网址]"))
				return
			}
			if !strings.HasSuffix(option, "/") {
				option += "/"
			}
			cfg.APIURL = option
			err := saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
	// API配置
	engine.OnRegex(`^猜歌(开启|关闭)(歌单|歌词)自动下载`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			swtich := ctx.State["regex_matched"].([]string)[1]
			option := ctx.State["regex_matched"].([]string)[2]
			chose := true
			if swtich == "关闭" {
				chose = false
			}
			if option == "歌单" {
				cfg.API = chose
			} else {
				cfg.Local = chose
			}
			err := saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
	engine.OnFullMatch("登录网易云", zero.SuperUserPermission, func(ctx *zero.Ctx) bool {
		if !zero.OnlyPrivate(ctx) {
			ctx.SendChain(message.Text("为了保护登录过程,请bot主人私聊。"))
			return false
		}
		return true
	}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyURL := cfg.APIURL + "login/qr/key"
			data, err := web.GetData(keyURL)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "获取网易云key失败,", err))
				return
			}
			var keyInfo keyInfo
			err = json.Unmarshal(data, &keyInfo)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "解析网易云key失败,", err))
				return
			}
			qrURL := cfg.APIURL + "login/qr/create?key=" + keyInfo.Data.Unikey + "&qrimg=1"
			data, err = web.GetData(qrURL)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "获取网易云二维码失败,", err))
				return
			}
			var qrInfo qrInfo
			err = json.Unmarshal(data, &qrInfo)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "解析网易云二维码失败,", err))
				return
			}
			ctx.SendChain(message.Text("[请使用手机APP扫描二维码或者进入网页扫码登录]\n", qrInfo.Data.Qrurl),
				message.Image("base64://"+strings.ReplaceAll(qrInfo.Data.Qrimg, "data:image/png;base64,", "")),
				message.Text("二维码有效时间为6分钟,登陆后请耐心等待结果,获取cookie过程有些漫长。"))
			i := 0
			for range time.NewTicker(10 * time.Second).C {
				APIURL := cfg.APIURL + "login/qr/check?key=" + url.QueryEscape(keyInfo.Data.Unikey)
				data, err := web.GetData(APIURL)
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, "无法获取登录状态,", err))
					return
				}
				var cookiesInfo cookyInfo
				err = json.Unmarshal(data, &cookiesInfo)
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, "解析登录状态失败,", err))
					return
				}
				switch cookiesInfo.Code {
				case 803:
					cfg.Cookie = cookiesInfo.Cookie
					err = saveConfig(cfgFile)
					if err == nil {
						ctx.SendChain(message.Text("成功！"))
					} else {
						ctx.SendChain(message.Text(serviceErr, err))
					}
					return
				case 801:
					i++
					if i%6 == 0 { // 每1分钟才提醒一次,减少提示(380/60=6次)
						ctx.SendChain(message.Text("状态:", cookiesInfo.Message))
					}
					continue
				case 800:
					ctx.SendChain(message.Text("状态:", cookiesInfo.Message))
					return
				default:
					ctx.SendChain(message.Text("状态:", cookiesInfo.Message))
					continue
				}
			}
		})
	engine.OnRegex(`^歌单信息\s*((https:\/\/music\.163\.com\/#\/playlist\?id=)?(\d+)|http:\/\/music\.163\.com\/playlist\/(\d+).*)$`).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			listID := ctx.State["regex_matched"].([]string)[3] + ctx.State["regex_matched"].([]string)[4]
			_, err := strconv.ParseInt(listID, 10, 64)
			if err != nil {
				ctx.SendChain(message.Text("请输入正确的歌单ID或者歌单连接"))
				return
			}
			APIURL := cfg.APIURL + "playlist/detail?id=" + listID
			data, err := web.GetData(APIURL)
			if err != nil {
				ctx.SendChain(message.Text("无法连接歌单,", err))
				return
			}
			var parsed listInfoOfAPI
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				ctx.SendChain(message.Text("无法解析歌单ID内容,", err))
				return
			}
			ctx.SendChain(
				message.Image(parsed.Playlist.CoverImgURL),
				message.Text(
					"歌单名称:", parsed.Playlist.Name,
					"\n歌单ID:", parsed.Playlist.ID,
					"\n创建人:", parsed.Playlist.Creator.Nickname,
					"\n创建时间:", time.Unix(parsed.Playlist.CreateTime/1000, 0).Format("2006-01-02"),
					"\n标签:", strings.Join(parsed.Playlist.Tags, ";"),
					"\n歌曲数量:", parsed.Playlist.TrackCount,
					"\n歌单简介:\n", parsed.Playlist.Description,
					"\n更新时间:", time.Unix(parsed.Playlist.UpdateTime/1000, 0).Format("2006-01-02"),
				))
		})
	// 本地绑定网易云歌单ID
	engine.OnRegex(`^(.*)绑定网易云\s*((https:\/\/music\.163\.com\/#\/playlist\?id=)?(\d+)|http:\/\/music\.163\.com\/playlist\/(\d+).*)$`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			listName := ctx.State["regex_matched"].([]string)[1]
			listID := ctx.State["regex_matched"].([]string)[4] + ctx.State["regex_matched"].([]string)[5]
			ctx.SendChain(message.Text("正在校验歌单信息,请稍等"))
			pathOfMusic := cfg.MusicPath + listName + "/"
			if file.IsNotExist(pathOfMusic) {
				ctx.SendChain(message.Text(serviceErr, "歌单不存在于本地"))
				return
			}
			// 是否存在该歌单
			APIURL := cfg.APIURL + "playlist/track/all?id=" + listID
			data, err := web.GetData(APIURL)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			var parsed musicListOfApI
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, "无法解析歌单ID内容,", err))
				return
			}
			if parsed.Code != 200 {
				ctx.SendChain(message.Text(serviceErr, parsed.Code))
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
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
	engine.OnPrefix("解除绑定", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			delList := ctx.State["args"].(string)
			filelist, err := getlist(cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			var playID int64
			for _, listinfo := range filelist {
				if delList == listinfo.Name {
					playID = listinfo.ID
					break
				}
			}
			// 删除ID
			if playID == 0 { // 如果ID没有且没删除文件
				ctx.SendChain(message.Text("歌单名称错误或者该歌单并没有绑定网易云,可以发送“歌单列表”获取歌单名称"))
				return
			}
			index := -1
			for i, list := range cfg.Playlist {
				if playID == list.ID {
					index = i
					break
				}
			}
			if index == -1 {
				ctx.SendChain(message.Text("歌单名称错误或者该歌单并没有绑定网易云,可以发送“歌单列表”获取歌单名称"))
				return
			}
			cfg.Playlist = append(cfg.Playlist[:index], cfg.Playlist[index+1:]...)
			err = saveConfig(cfgFile)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
	// 下载歌曲到对应的歌单里面
	engine.OnRegex(`^下载歌单\s*((https:\/\/music\.163\.com\/#\/playlist\?id=)?(\d+)|http:\/\/music\.163\.com\/playlist\/(\d+).*[^\s$])\s*到\s*(.*)$`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[3] + ctx.State["regex_matched"].([]string)[4]
			listName := ctx.State["regex_matched"].([]string)[5]
			ctx.SendChain(message.Text("正在校验歌单信息,请稍等"))
			// 是否存在该歌单
			if file.IsNotExist(cfg.MusicPath + listName) {
				ctx.SendChain(message.Text("歌单不存在,是否创建？(是/否)"))
				next := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`(是|否)`), ctx.CheckSession())
				recv, cancel := next.Repeat()
				defer cancel()
				wait := time.NewTimer(120 * time.Second)
				answer := ""
				for {
					select {
					case <-wait.C:
						wait.Stop()
						ctx.SendChain(message.Text("等待超时,取消下载"))
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
				err := os.MkdirAll(cfg.MusicPath+listName, 0755)
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
			}
			ctx.SendChain(message.Text("开始下载歌曲,需要一定时间下载,请稍等"))
			listID, err := strconv.ParseInt(keyword, 10, 64)
			if err == nil {
				err = downloadlist(listID, cfg.MusicPath+listName+"/")
			}
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
}

// 随机从歌单下载歌曲(歌单ID, 音乐保存路径)
func drawByAPI(playlistID int64, musicPath string) (musicName string, err error) {
	APIURL := cfg.APIURL + "playlist/track/all?id=" + strconv.FormatInt(playlistID, 10)
	data, err := web.GetData(APIURL)
	if err != nil {
		err = errors.Errorf("无法获取歌单列表\n%s", err)
		return
	}
	var parsed musicListOfApI
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		err = errors.Errorf("无法读取歌单列表\n%s", err)
		return
	}
	listlen := len(parsed.Songs)
	randidx := rand.Intn(listlen)
	// 将"/"符号去除,不然无法生成文件
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
		// 将"/"符号去除,不然无法下载
		cource = strings.ReplaceAll(cource, "/", "&")
	}
	if name == "" || musicID == 0 {
		err = errors.New("无法获API取歌曲信息")
		return
	}
	if cource != "" {
		name += " - " + artistName + " - " + cource
	} else {
		name += " - " + artistName
	}
	// 下载歌曲
	err = wyy.DownloadMusic(musicID, name, musicPath)
	if err == nil {
		musicName = name + ".mp3"
		if cfg.Local {
			// 下载歌词
			_ = wyy.DownloadLrc(musicID, name, musicPath+"歌词/")
		}
	}
	return
}

// 下载歌单歌曲(歌单ID, 音乐保存路径)
func downloadlist(playlistID int64, musicPath string) error {
	APIURL := cfg.APIURL + "playlist/track/all?id=" + strconv.FormatInt(playlistID, 10)
	data, err := web.GetData(APIURL)
	if err != nil {
		return err
	}
	var parsed musicListOfApI
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return err
	}
	if parsed.Code != 200 {
		err = errors.Errorf("requset code : %d", parsed.Code)
		return err
	}
	for _, info := range parsed.Songs {
		// 将"/"符号去除,不然无法生成文件
		musicName := strings.ReplaceAll(info.Name, "/", "·")
		musicID := info.ID
		artistName := ""
		for i, ARInfo := range info.Ar {
			if i != 0 {
				artistName += "&" + ARInfo.Name
			} else {
				artistName += ARInfo.Name
			}
		}
		cource := ""
		if info.Alia != nil {
			cource = strings.Join(info.Alia, "&")
			// 将"/"符号去除,不然无法下载
			cource = strings.ReplaceAll(cource, "/", "&")
		}
		if musicName == "" || musicID == 0 {
			err = errors.New("无法获API取歌曲信息")
			return err
		}
		if cource != "" {
			musicName += " - " + artistName + " - " + cource
		} else {
			musicName += " - " + artistName
		}
		// 下载歌曲
		err = wyy.DownloadMusic(musicID, musicName, musicPath)
		if err == nil {
			if cfg.Local {
				// 下载歌词
				_ = wyy.DownloadLrc(musicID, musicName, musicPath+"歌词/")
			}
		}
	}
	return nil
}

/*****************************************************************/
/**************************独角兽API*******************************/
/*****************************************************************/
// 下载从独角兽抽到的歌曲ID(歌单ID, 音乐保存路径, 歌词保存路径)
func downloadByOvooa(playlistID int64, musicPath string) (musicName string, err error) {
	// 抽取歌曲
	mid, err := drawByOvooa(playlistID)
	if err != nil {
		err = errors.Errorf("API%s", err)
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
		err = errors.Errorf("music IDThis music ID sreached munber is %d", mun)
	}
	return
}

// 通过独角兽API随机抽取歌单歌曲ID(参数：歌单ID)
func drawByOvooa(playlistID int64) (musicID int, err error) {
	APIURL := "https://ovooa.com/API/163_Music_Rand/api.php?id=" + strconv.FormatInt(playlistID, 10)
	data, err := web.GetData(APIURL)
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
