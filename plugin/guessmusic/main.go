// Package guessmusic 基于zbp的猜歌插件
package guessmusic

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/web"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
)

const (
	ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66"
)

var (
	cuttime = [...]string{"00:00:05", "00:00:30", "00:01:00"} // 音乐切割时间点，可自行调节时间（时：分：秒）
	cfg     = config{                                         // 默认 config
		MusicPath: file.BOTPATH + "/data/guessmusic/music/", // 绝对路径，歌库根目录,通过指令进行更改
		Local:     true,                                     // 是否使用本地音乐库
		API:       true,                                     // 是否使用 Api
	}
)

func init() { // 插件主体
	engine := control.Register("guessmusic", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "猜歌插件（该插件依赖ffmpeg）\n" +
			"- 个人猜歌\n" +
			"- 团队猜歌\n" +
			"- 设置猜歌缓存歌库路径 [绝对路径]\n" +
			"- 设置猜歌本地 [true/false]\n" +
			"- 设置猜歌Api [true/false]\n" +
			"注：默认歌库为网易云热歌榜\n" +
			"1.可在后面添加“-动漫”进行动漫歌猜歌\n-这个只能猜歌名和歌手\n" +
			"2.可在后面添加“-动漫2”进行动漫歌猜歌\n-这个可以猜番名，但歌手经常“未知”",
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
		err = saveConfig(cfgFile)
		if err != nil {
			panic(err)
		}
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
				cfg.MusicPath = musicPath
			case "本地":
				choice, err := strconv.ParseBool(value)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				cfg.Local = choice
			case "Api":
				choice, err := strconv.ParseBool(value)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				cfg.API = choice
			}
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text("ERROR:", err))
			}
		})
	engine.OnRegex(`^(个人|团队)猜歌(-动漫|-动漫2)?$`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			mode := ctx.State["regex_matched"].([]string)[2]
			gid := strconv.FormatInt(ctx.Event.GroupID, 10)
			if mode == "-动漫2" {
				ctx.SendChain(message.Text("正在准备歌曲,请稍等\n回答“-[歌曲名称|歌手|番剧|提示|取消]”\n一共3段语音，6次机会"))
			} else {
				ctx.SendChain(message.Text("正在准备歌曲,请稍等\n回答“-[歌曲名称|歌手|提示|取消]”\n一共3段语音，6次机会"))
			}
			// 随机抽歌
			musicName, pathOfMusic, err := musicLottery(mode, cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
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
			answerString := strings.Split(musicName, " - ")
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
					msg := make(message.Message, 0, 3)
					msg = append(msg, message.Reply(ctx.Event.MessageID))
					msg = append(msg, message.Text("猜歌超时，游戏结束\n答案是:",
						"\n歌名:", answerString[0],
						"\n歌手:", answerString[1]))
					if mode == "-动漫2" {
						msg = append(msg, message.Text("\n歌曲出自:", answerString[2]))
					}
					ctx.Send(msg)
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
							msg := make(message.Message, 0, 3)
							msg = append(msg, message.Reply(c.Event.MessageID))
							msg = append(msg, message.Text("游戏已取消，猜歌答案是",
								"\n歌名:", answerString[0],
								"\n歌手:", answerString[1]))
							if mode == "-动漫2" {
								msg = append(msg, message.Text("\n歌曲出自:", answerString[2]))
							}
							ctx.Send(msg)
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
					case strings.Contains(answerString[0], answer) || strings.EqualFold(answerString[0], answer):
						wait.Stop()
						tick.Stop()
						after.Stop()
						msg := make(message.Message, 0, 3)
						msg = append(msg, message.Reply(c.Event.MessageID))
						msg = append(msg, message.Text("太棒了，你猜对歌曲名了！答案是",
							"\n歌名:", answerString[0],
							"\n歌手:", answerString[1]))
						if mode == "-动漫2" {
							msg = append(msg, message.Text("\n歌曲出自:", answerString[2]))
						}
						ctx.Send(msg)
						return
					case answerString[1] == "未知" && answer == "未知":
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("该模式禁止回答“未知”"),
							),
						)
					case strings.Contains(answerString[1], answer) || strings.EqualFold(answerString[1], answer):
						wait.Stop()
						tick.Stop()
						after.Stop()
						msg := make(message.Message, 0, 3)
						msg = append(msg, message.Reply(c.Event.MessageID))
						msg = append(msg, message.Text("太棒了，你猜对歌手名了！答案是",
							"\n歌名:", answerString[0],
							"\n歌手:", answerString[1]))
						if mode == "-动漫2" {
							msg = append(msg, message.Text("\n歌曲出自:", answerString[2]))
						}
						ctx.Send(msg)
						return
					default:
						if mode == "-动漫2" && (strings.Contains(answerString[2], answer) || strings.EqualFold(answerString[2], answer)) {
							wait.Stop()
							tick.Stop()
							after.Stop()
							ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
								message.Text("太棒了，你猜对番剧名了！答案是:",
									"\n歌名:", answerString[0],
									"\n歌手:", answerString[1],
									"\n歌曲出自:", answerString[2]),
							))
							return
						}
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
							msg := make(message.Message, 0, 3)
							msg = append(msg, message.Reply(c.Event.MessageID))
							msg = append(msg, message.Text("次数到了，你没能猜出来。\n答案是:",
								"\n歌名:", answerString[0],
								"\n歌手:", answerString[1]))
							if mode == "-动漫2" {
								msg = append(msg, message.Text("\n歌曲出自:", answerString[2]))
							}
							ctx.Send(msg)
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

// 随机抽取音乐
func musicLottery(mode, musicPath string) (musicName, pathOfMusic string, err error) {
	switch mode {
	case "-动漫":
		pathOfMusic = musicPath + "动漫/"
	case "-动漫2":
		pathOfMusic = musicPath + "动漫2/"
	default:
		pathOfMusic = musicPath + "歌榜/"
	}
	err = os.MkdirAll(pathOfMusic, 0755)
	if err != nil {
		err = errors.Errorf("[生成文件夹错误]ERROR:%s", err)
		return
	}
	files, err := ioutil.ReadDir(pathOfMusic)
	if err != nil {
		err = errors.Errorf("[读取本地列表错误]ERROR:%s", err)
		return
	}

	if cfg.Local && cfg.API {
		switch {
		case len(files) == 0:
			// 如果没有任何本地就下载歌曲
			musicName, err = getAPIMusic(mode, pathOfMusic)
			if err != nil {
				err = errors.Errorf("[本地数据为0，歌曲下载错误]ERROR:%s", err)
				return
			}
		case rand.Intn(2) == 0:
			// [0,1)只会取到0，rand不允许的
			musicName = getLocalMusic(files)
		default:
			musicName, err = getAPIMusic(mode, pathOfMusic)
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
	if cfg.API {
		musicName, err = getAPIMusic(mode, pathOfMusic)
		if err != nil {
			err = errors.Errorf("[获取API失败，未开启本地数据] ERROR:%s", err)
			return
		}
		return
	}
	err = errors.New("[未开启API以及本地数据]")
	return
}

func getAPIMusic(mode string, musicPath string) (musicName string, err error) {
	switch mode {
	case "-动漫":
		musicName, err = getPaugramData(musicPath)
	case "-动漫2":
		musicName, err = getAnimeData(musicPath)
	default:
		musicName, err = getNetEaseData(musicPath)
	}
	return
}

func getLocalMusic(files []fs.FileInfo) (musicName string) {
	if len(files) > 1 {
		musicName = strings.Replace(files[rand.Intn(len(files))].Name(), ".mp3", "", 1)
	} else {
		musicName = strings.Replace(files[0].Name(), ".mp3", "", 1)
	}
	return
}

// 下载保罗API的歌曲
func getPaugramData(musicPath string) (musicName string, err error) {
	api := "https://api.paugram.com/acgm/?list=1"
	referer := "https://api.paugram.com/"
	data, err := web.RequestDataWith(web.NewDefaultClient(), api, "GET", referer, ua)
	if err != nil {
		return
	}
	var parsed paugramData
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	name := parsed.Title
	artistsName := parsed.Artist
	musicURL := parsed.Link
	if name == "" || artistsName == "" {
		err = errors.New("无法获API取歌曲信息")
		return
	}
	musicName = name + " - " + artistsName
	downMusic := musicPath + "/" + musicName + ".mp3"
	response, err := http.Head(musicURL)
	if err != nil {
		err = errors.Errorf("下载音乐失败, ERROR: %s", err)
		return
	}
	if response.StatusCode != 200 {
		err = errors.Errorf("下载音乐失败, Status Code: %d", response.StatusCode)
		return
	}
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

// 下载animeMusic API的歌曲
func getAnimeData(musicPath string) (musicName string, err error) {
	api := "https://anime-music.jijidown.com/api/v2/music"
	referer := "https://anime-music.jijidown.com/"
	data, err := web.RequestDataWith(web.NewDefaultClient(), api, "GET", referer, ua)
	if err != nil {
		return
	}
	var parsed animeData
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	name := parsed.Res.Title
	artistName := parsed.Res.Author
	acgName := parsed.Res.AnimeInfo.Title
	// musicURL := parsed.Res.PlayURL
	if name == "" || artistName == "" {
		err = errors.New("无法获API取歌曲信息")
		return
	}
	requestURL := "https://autumnfish.cn/search?keywords=" + url.QueryEscape(name+" "+artistName) + "&limit=1"
	if artistName == "未知" {
		requestURL = "https://autumnfish.cn/search?keywords=" + url.QueryEscape(acgName+" "+name) + "&limit=1"
	}
	data, err = web.GetData(requestURL)
	if err != nil {
		err = errors.Errorf("API歌曲查询失败, ERROR: %s", err)
		return
	}
	var autumnfish autumnfishData
	err = json.Unmarshal(data, &autumnfish)
	if err != nil {
		return
	}
	if autumnfish.Code != 200 {
		err = errors.Errorf("下载音乐失败, Status Code: %d", autumnfish.Code)
		return
	}
	musicID := strconv.Itoa(autumnfish.Result.Songs[0].ID)
	if artistName == "未知" {
		artistName = strings.ReplaceAll(autumnfish.Result.Songs[0].Artists[0].Name, " - ", "-")
	}
	musicName = name + " - " + artistName + " - " + acgName
	downMusic := musicPath + "/" + musicName + ".mp3"
	musicURL := "http://music.163.com/song/media/outer/url?id=" + musicID
	response, err := http.Head(musicURL)
	if err != nil {
		err = errors.Errorf("下载音乐失败, ERROR: %s", err)
		return
	}
	if response.StatusCode != 200 {
		err = errors.Errorf("下载音乐失败, Status Code: %d", response.StatusCode)
		return
	}
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

// 下载网易云热歌榜音乐
func getNetEaseData(musicPath string) (musicName string, err error) {
	api := "https://api.uomg.com/api/rand.music?sort=%E7%83%AD%E6%AD%8C%E6%A6%9C&format=json"
	referer := "https://api.uomg.com/api/rand.music"
	data, err := web.RequestDataWith(web.NewDefaultClient(), api, "GET", referer, ua)
	if err != nil {
		return
	}
	var parsed netEaseData
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	name := parsed.Data.Name
	musicURL := parsed.Data.URL
	artistsName := parsed.Data.Artistsname
	if name == "" || artistsName == "" {
		err = errors.New("无法获API取歌曲信息")
		return
	}
	musicName = name + " - " + artistsName
	downMusic := musicPath + "/" + musicName + ".mp3"
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
		err = errors.Errorf("[生成歌曲目录错误]ERROR:%s", err)
		return
	}
	var stderr bytes.Buffer
	cmdArguments := []string{"-y", "-i", pathOfMusic + musicName + ".mp3",
		"-ss", cuttime[0], "-t", "10", file.BOTPATH + "/" + outputPath + "0.wav",
		"-ss", cuttime[1], "-t", "10", file.BOTPATH + "/" + outputPath + "1.wav",
		"-ss", cuttime[2], "-t", "10", file.BOTPATH + "/" + outputPath + "2.wav", "-hide_banner"}
	cmd := exec.Command("ffmpeg", cmdArguments...)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		err = errors.Errorf("[生成歌曲错误]ERROR:%s", stderr.String())
		return
	}
	return
}
