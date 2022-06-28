// Package guessmusic 基于zbp的猜歌插件
package guessmusic

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	
	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/binary"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/web"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
)

const (
	ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66"
)

var (
	musicPath = file.BOTPATH + "/data/guessmusic/music/" // 绝对路径，歌库根目录,通过指令进行更改
	cuttime   = [...]string{"00:00:05", "00:00:30", "00:01:00"} // 音乐切割时间点，可自行调节时间（时：分：秒）
)

func init() { // 插件主体
	engine := control.Register("guessmusic", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "猜歌插件（该插件依赖ffmpeg）\n" +
			"- 个人猜歌\n" +
			"- 团队猜歌\n" +
			"- 设置缓存歌库路径 [绝对路径]\n" +
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
	cfgfile := engine.DataFolder() + "setpath.txt"
	if file.IsExist(cfgfile) {
		b, err := os.ReadFile(cfgfile)
		if err == nil {
			musicPath = binary.BytesToString(b)
			logrus.Infoln("[guessmusic] set dir to", musicPath)
		}
	}
	engine.OnRegex(`^设置缓存歌库路径(.*)$`, func(ctx *zero.Ctx) bool {
		if !zero.SuperUserPermission(ctx) {
			ctx.SendChain(message.Text("只有bot主人可以设置！"))
			return false
		}
		return true
	}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			musicPath = ctx.State["regex_matched"].([]string)[1]
			if musicPath == "" {
				ctx.SendChain(message.Text("请输入正确的路径!"))
			}
			musicPath = strings.ReplaceAll(musicPath, "\\", "/")
			if !strings.HasSuffix(musicPath, "/") {
				musicPath += "/"
			}
			err := os.WriteFile(cfgfile, binary.StringToBytes(musicPath), 0644)
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
			musicname, pathofmusic, err := musiclottery(mode, musicPath)
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			// 切割音频，生成3个10秒的音频
			outputPath := cachePath + gid + "/"
			err = musiccut(musicname, pathofmusic, outputPath)
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			// 进行猜歌环节
			ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + "0.wav"))
			answerstring := strings.Split(musicname, " - ")
			var next *zero.FutureEvent
			if ctx.State["regex_matched"].([]string)[1] == "个人" {
				next = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^-\S{1,}`), ctx.CheckSession())
			} else {
				next = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^-\S{1,}`), zero.CheckGroup(ctx.Event.GroupID))
			}
			var countofmusic = 0  // 音频数量
			var countofanswer = 0 // 问答次数
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
					msg = append(msg, message.Text("猜歌超时，游戏结束\n答案是:\n",
						"\n歌名:", answerstring[0],
						"\n歌手:", answerstring[1]))
					if mode == "-动漫2" {
						msg = append(msg, message.Text("\n歌曲出自:", answerstring[2]))
					}
					ctx.Send(msg)
					return
				case <-wait.C:
					wait.Reset(40 * time.Second)
					countofmusic++
					if countofmusic > 2 {
						wait.Stop()
						continue
					}
					ctx.SendChain(
						message.Text("好像有些难度呢，再听这段音频，要仔细听哦"),
					)
					ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + strconv.Itoa(countofmusic) + ".wav"))
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
								"\n歌名:", answerstring[0],
								"\n歌手:", answerstring[1]))
							if mode == "-动漫2" {
								msg = append(msg, message.Text("\n歌曲出自:", answerstring[2]))
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
						countofmusic++
						if countofmusic > 2 {
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
						ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + strconv.Itoa(countofmusic) + ".wav"))
					case strings.Contains(answerstring[0], answer) || strings.EqualFold(answerstring[0], answer):
						wait.Stop()
						tick.Stop()
						after.Stop()
						msg := make(message.Message, 0, 3)
						msg = append(msg, message.Reply(c.Event.MessageID))
						msg = append(msg, message.Text("太棒了，你猜对歌曲名了！答案是",
							"\n歌名:", answerstring[0],
							"\n歌手:", answerstring[1]))
						if mode == "-动漫2" {
							msg = append(msg, message.Text("\n歌曲出自:", answerstring[2]))
						}
						ctx.Send(msg)
						return
					case answerstring[1] == "未知" && answer == "未知":
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("该模式禁止回答“未知”"),
							),
						)
					case strings.Contains(answerstring[1], answer) || strings.EqualFold(answerstring[1], answer):
						wait.Stop()
						tick.Stop()
						after.Stop()
						msg := make(message.Message, 0, 3)
						msg = append(msg, message.Reply(c.Event.MessageID))
						msg = append(msg, message.Text("太棒了，你猜对歌手名了！答案是",
							"\n歌名:", answerstring[0],
							"\n歌手:", answerstring[1]))
						if mode == "-动漫2" {
							msg = append(msg, message.Text("\n歌曲出自:", answerstring[2]))
						}
						ctx.Send(msg)
						return
					default:
						if mode == "-动漫2" && (strings.Contains(answerstring[2], answer) || strings.EqualFold(answerstring[2], answer)) {
							wait.Stop()
							tick.Stop()
							after.Stop()
							ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
								message.Text("太棒了，你猜对番剧名了！答案是:",
									"\n歌名:", answerstring[0],
									"\n歌手:", answerstring[1],
									"\n歌曲出自:", answerstring[2]),
							))
							return
						}
						countofmusic++
						switch {
						case countofmusic > 2 && countofanswer < 6:
							wait.Stop()
							countofanswer++
							ctx.Send(
								message.ReplyWithMessage(c.Event.MessageID,
									message.Text("答案不对哦，加油啊~"),
								),
							)
						case countofmusic > 2:
							wait.Stop()
							tick.Stop()
							after.Stop()
							msg := make(message.Message, 0, 3)
							msg = append(msg, message.Reply(c.Event.MessageID))
							msg = append(msg, message.Text("次数到了，你没能猜出来。\n答案是:",
								"\n歌名:", answerstring[0],
								"\n歌手:", answerstring[1]))
							if mode == "-动漫2" {
								msg = append(msg, message.Text("\n歌曲出自:", answerstring[2]))
							}
							ctx.Send(msg)
							return
						default:
							wait.Reset(40 * time.Second)
							countofanswer++
							ctx.Send(
								message.ReplyWithMessage(c.Event.MessageID,
									message.Text("答案不对，再听这段音频，要仔细听哦"),
								),
							)
							ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + strconv.Itoa(countofmusic) + ".wav"))
						}
					}
				}
			}
		})
}

// 随机抽取音乐
func musiclottery(mode, musicPath string) (musicname, pathofmusic string, err error) {
	switch mode {
	case "-动漫":
		pathofmusic = musicPath + "动漫/"
	case "-动漫2":
		pathofmusic = musicPath + "动漫2/"
	default:
		pathofmusic = musicPath + "歌榜/"
	}
	err = os.MkdirAll(pathofmusic, 0755)
	if err != nil {
		err = errors.Errorf("[生成文件夹错误]ERROR:%s", err)
		return
	}
	files, err := ioutil.ReadDir(pathofmusic)
	if err != nil {
		err = errors.Errorf("[读取本地列表错误]ERROR:%s", err)
		return
	}
	// 随机抽取音乐从本地或者线上
	switch {
	case len(files) == 0:
		// 如果没有任何本地就下载歌曲
		switch mode {
		case "-动漫":
			musicname, err = getpaugramdata(pathofmusic)
		case "-动漫2":
			musicname, err = getanimedata(pathofmusic)
		default:
			musicname, err = getuomgdata(pathofmusic)
		}
		if err != nil {
			err = errors.Errorf("[本地数据为0，歌曲下载错误]ERROR:%s", err)
			return
		}
	case rand.Intn(2) == 0:
		// [0,1)只会取到0，rand不允许的
		if len(files) > 1 {
			musicname = strings.Replace(files[rand.Intn(len(files))].Name(), ".mp3", "", 1)
		} else {
			musicname = strings.Replace(files[0].Name(), ".mp3", "", 1)
		}
	default:
		switch mode {
		case "-动漫":
			musicname, err = getpaugramdata(pathofmusic)
		case "-动漫2":
			musicname, err = getanimedata(pathofmusic)
		default:
			musicname, err = getuomgdata(pathofmusic)
		}
		if err != nil {
			// 如果下载失败就从本地抽一个歌曲
			if len(files) > 1 {
				musicname = strings.Replace(files[rand.Intn(len(files))].Name(), ".mp3", "", 1)
			} else {
				musicname = strings.Replace(files[0].Name(), ".mp3", "", 1)
			}
			err = nil
		}
	}
	return
}

// 下载保罗API的歌曲
func getpaugramdata(musicPath string) (musicname string, err error) {
	api := "https://api.paugram.com/acgm/?list=1"
	referer := "https://api.paugram.com/"
	data, err := web.RequestDataWith(web.NewDefaultClient(), api, "GET", referer, ua)
	if err != nil {
		return
	}
	name := gjson.Get(binary.BytesToString(data), "title").String()
	artistsname := gjson.Get(binary.BytesToString(data), "artist").String()
	musicurl := gjson.Get(binary.BytesToString(data), "link").String()
	if name == "" || artistsname == "" {
		err = errors.Errorf("the music is missed")
		return
	}
	musicname = name + " - " + artistsname
	downmusic := musicPath + "/" + musicname + ".mp3"
	response, err := http.Head(musicurl)
	if err != nil || response.StatusCode != 200 {
		err = errors.Errorf("the music is missed")
		return
	}
	if file.IsNotExist(downmusic) {
		data, err = web.GetData(musicurl + ".mp3")
		if err != nil {
			return
		}
		err = os.WriteFile(downmusic, data, 0666)
		if err != nil {
			return
		}
	}
	return
}

// 下载animeMusic API的歌曲
func getanimedata(musicPath string) (musicname string, err error) {
	api := "https://anime-music.jijidown.com/api/v2/music"
	referer := "https://anime-music.jijidown.com/"
	data, err := web.RequestDataWith(web.NewDefaultClient(), api, "GET", referer, ua)
	if err != nil {
		return
	}
	name := gjson.Get(binary.BytesToString(data), "res").Get("title").String()
	artistsname := gjson.Get(binary.BytesToString(data), "res").Get("author").String()
	acgname := gjson.Get(binary.BytesToString(data), "res").Get("anime_info").Get("title").String()
	musicurl := gjson.Get(binary.BytesToString(data), "res").Get("play_url").String()
	if name == "" || artistsname == "" {
		err = errors.Errorf("the music is missed")
		return
	}
	musicname = name + " - " + artistsname + " - " + acgname
	downmusic := musicPath + "/" + musicname + ".mp3"
	response, err := http.Head(musicurl)
	if err != nil || response.StatusCode != 200 {
		err = errors.Errorf("the music is missed")
		return
	}
	if file.IsNotExist(downmusic) {
		data, err = web.GetData(musicurl + ".mp3")
		if err != nil {
			return
		}
		err = os.WriteFile(downmusic, data, 0666)
		if err != nil {
			return
		}
	}
	return
}

// 下载网易云热歌榜音乐
func getuomgdata(musicPath string) (musicname string, err error) {
	api := "https://api.uomg.com/api/rand.music?sort=%E7%83%AD%E6%AD%8C%E6%A6%9C&format=json"
	referer := "https://api.uomg.com/api/rand.music"
	data, err := web.RequestDataWith(web.NewDefaultClient(), api, "GET", referer, ua)
	if err != nil {
		return
	}
	musicdata := gjson.Get(binary.BytesToString(data), "data")
	name := musicdata.Get("name").String()
	musicurl := musicdata.Get("url").String()
	artistsname := musicdata.Get("artistsname").String()
	musicname = name + " - " + artistsname
	downmusic := musicPath + "/" + musicname + ".mp3"
	if file.IsNotExist(downmusic) {
		data, err = web.GetData(musicurl + ".mp3")
		if err != nil {
			return
		}
		err = os.WriteFile(downmusic, data, 0666)
		if err != nil {
			return
		}
	}
	return
}

// 切割音乐成三个10s音频
func musiccut(musicname, pathofmusic, outputPath string) (err error) {
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		err = errors.Errorf("[生成歌曲目录错误]ERROR:%s", err)
		return
	}
	var stderr bytes.Buffer
	cmdArguments := []string{"-y", "-i", pathofmusic + musicname + ".mp3",
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
