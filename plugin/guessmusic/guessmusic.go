package guessmusic

import (
	"bytes"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/liuzl/gocc"
)

var cuttime = [...]string{"00:00:05", "00:00:30", "00:01:00"} // 音乐切割时间点,可自行调节时间（时：分：秒）
var t2s *gocc.OpenCC

func init() {
	// 初始化简繁体转换变量
	var err1 error
	t2s, err1 = gocc.New("t2s")
	if err1 != nil {
		log.Infof("[guessmusic]:%s", err1)
	}

	engine.OnRegex(`^(个人|团队)猜歌(-(.*))?$`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			mode := ctx.State["regex_matched"].([]string)[3]
			gid := ctx.Event.GroupID
			// 获取本地列表
			filelist, err := getlist(cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			// 加载默认歌单
			if mode == "" {
				index := -1
				for i, dlist := range cfg.Defaultlist {
					if dlist.GroupID == gid {
						index = i
						break
					}
				}
				if index == -1 {
					// 如果没有设置就默认第一个文件夹
					mode = filelist[0].Name
				} else {
					mode = cfg.Defaultlist[index].Name
					ok := true
					for _, listinfo := range filelist {
						if mode == listinfo.Name {
							ok = false
							break
						}
					}
					// 如果默认的歌单不存在了清空设置
					if ok {
						cfg.Defaultlist = append(cfg.Defaultlist[:index], cfg.Defaultlist[index+1:]...)
						_ = saveConfig(cfgFile)
						mode = filelist[0].Name
					}
				}
			}
			ctx.SendChain(message.Text("正在准备歌曲,请稍等\n回答“-[歌曲信息(歌名歌手等)|提示|取消]”\n一共3段语音,6次机会"))
			// 随机抽歌
			pathOfMusic, musicName, err := musicLottery(cfg.MusicPath, mode)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			// 解析歌曲信息
			music := strings.Split(musicName, ".")
			// 获取音乐后缀
			musictype := music[len(music)-1]
			if !strings.Contains(musictypelist, musictype) {
				ctx.SendChain(message.Text("抽取到了歌曲：\n",
					musicName, "\n该歌曲不是音乐后缀,请联系bot主人修改"))
				return
			}
			// 获取音乐信息
			musicInfo := strings.Split(strings.ReplaceAll(musicName, "."+musictype, ""), " - ")
			infoNum := len(musicInfo)
			if infoNum == 1 {
				ctx.SendChain(message.Text("抽取到了歌曲：\n",
					musicName, "\n该歌曲命名不符合命名规则,请联系bot主人修改"))
				return
			}
			answerString := "歌名:" + musicInfo[0] + "\n歌手:" + musicInfo[1]
			if infoNum > 2 {
				musicInfo[2] = strings.ReplaceAll(musicInfo[2], "&", "\n")
				answerString += "\n其他信息:\n" + musicInfo[2]
			}
			musicInfo = append(musicInfo, answerString)
			// 切割音频,生成3个10秒的音频
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
			recv, cancel := next.Repeat()
			defer cancel()
			wait := time.NewTimer(40 * time.Second)
			tick := time.NewTimer(105 * time.Second)
			after := time.NewTimer(120 * time.Second)
			wg := sync.WaitGroup{}
			var (
				messageStr  message.MessageSegment // 文本信息
				tickCount   = 0                    // 音频数量
				answerCount = 0                    // 问答次数
				win         bool                   // 是否赢得游戏
			)
			for {
				select {
				case <-tick.C:
					ctx.SendChain(message.Text("猜歌游戏,你还有15s作答时间"))
				case <-after.C:
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("时间超时,猜歌结束,公布答案：\n", answerString)))
					return
				case <-wait.C:
					wait.Reset(40 * time.Second)
					tickCount++
					if tickCount > 2 {
						wait.Stop()
						continue
					}
					ctx.SendChain(
						message.Text("好像有些难度呢,再听这段音频,要仔细听哦"),
					)
					ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + strconv.Itoa(tickCount) + ".wav"))
				case c := <-recv:
					wg.Add(1)
					go func() {
						messageStr, answerCount, tickCount, win = gameMatch(c, ctx.Event.UserID, musicInfo, answerCount, tickCount)
						if win { // 游戏结束的话
							wait.Stop()
							tick.Stop()
							after.Stop()
							ctx.SendChain(message.Reply(c.Event.MessageID), messageStr)
							ctx.SendChain(message.Record("file:///" + pathOfMusic + musicName))
						} else {
							wait.Reset(40 * time.Second)
							tick.Reset(105 * time.Second)
							after.Reset(120 * time.Second)
							if tickCount > 2 || messageStr.Data["text"] == "你无权限取消" {
								ctx.SendChain(message.Reply(c.Event.MessageID), messageStr)
							} else {
								ctx.SendChain(message.Reply(c.Event.MessageID), messageStr)
								ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + outputPath + strconv.Itoa(tickCount) + ".wav"))
							}
						}
						wg.Done()
					}()
					wg.Wait()
					if win {
						return
					}
				}
			}
		})
}

// 随机抽取音乐
func musicLottery(musicPath, listName string) (pathOfMusic, musicName string, err error) {
	// 读取歌单文件
	pathOfMusic = musicPath + listName + "/"
	if file.IsNotExist(pathOfMusic) {
		err = errors.New("指定的歌单不存在,可发送“歌单列表”查看歌单列表")
		return
	}
	files, err := os.ReadDir(pathOfMusic)
	if err != nil {
		return
	}
	// 获取绑定的网易云
	var playlistID int64
	for _, listinfo := range cfg.Playlist {
		if listinfo.Name == listName {
			playlistID = listinfo.ID
		}
	}
	// 如果本地列表为空
	if len(files) == 0 {
		if playlistID == 0 || !cfg.API {
			err = errors.New("本地歌单数据为0")
			return
		}
		// 如果绑定了歌单ID
		if cfg.APIURL == "" {
			// 如果没有配置过API地址,尝试连接独角兽
			musicName, err = downloadByOvooa(playlistID, pathOfMusic)
			if err != nil {
				err = errors.Errorf("本地歌单数据为0,API下载歌曲失败\n%s", err)
			}
		} else {
			// 从API中抽取歌曲
			musicName, err = drawByAPI(playlistID, pathOfMusic)
			if err != nil {
				err = errors.Errorf("本地歌单数据为0,API下载歌曲失败\n%s", err)
			}
		}
		return
	}
	// 进行随机抽取
	if playlistID == 0 || !cfg.API {
		musicName = getLocalMusic(files, 10)
	} else {
		switch rand.Intn(3) { // 三分二概率抽取API的
		case 1:
			musicName = getLocalMusic(files, 10)
		default:
			if cfg.APIURL == "" {
				// 如果没有配置过API地址,尝试连接独角兽
				musicName, err = downloadByOvooa(playlistID, pathOfMusic)
			} else {
				// 从API中抽取歌曲
				musicName, err = drawByAPI(playlistID, pathOfMusic)
			}
			if err != nil {
				musicName = getLocalMusic(files, 10)
			}
		}
	}
	if musicName == "" {
		err = errors.New("抽取歌曲轮空了,请重试")
	}
	return
}

// 从本地列表中随机抽取一首( indexMax : 最大递归次数 )
func getLocalMusic(files []fs.DirEntry, indexMax int) (musicName string) {
	if len(files) > 1 {
		music := files[rand.Intn(len(files))]
		// 如果是文件夹就递归
		if music.IsDir() {
			indexMax--
			if indexMax <= 0 {
				return
			}
			musicName = getLocalMusic(files, indexMax)
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

// 数据匹配（结果信息，答题次数，提示次数，是否结束游戏）
func gameMatch(c *zero.Ctx, beginner int64, musicInfo []string, answerTimes, tickTimes int) (message.MessageSegment, int, int, bool) {
	answer := strings.Replace(c.Event.Message.String(), "-", "", 1)
	// 大小写，简繁体转换
	answer = ConvertText(answer)
	for i, element := range musicInfo {
		musicInfo[i] = ConvertText(element)
	}
	switch {
	case answer == "取消":
		if c.Event.UserID == beginner {
			return message.Text("游戏已取消,猜歌答案是\n", musicInfo[len(musicInfo)-1], "\n\n\n下面欣赏猜歌的歌曲"), answerTimes, tickTimes, true
		}
		return message.Text("你无权限取消"), answerTimes, tickTimes, false
	case answer == "提示":
		tickTimes++
		if tickTimes > 2 {
			return message.Text("已经没有提示了哦"), answerTimes, tickTimes, false
		}
		return message.Text("再听这段音频,要仔细听哦"), answerTimes, tickTimes, false
	case strings.Contains(musicInfo[0], answer) || strings.EqualFold(musicInfo[0], answer):
		return message.Text("太棒了,你猜对歌曲名了！答案是\n", musicInfo[len(musicInfo)-1], "\n\n下面欣赏猜歌的歌曲"), answerTimes, tickTimes, true
	case strings.Contains(musicInfo[1], answer) || strings.EqualFold(musicInfo[1], answer):
		return message.Text("太棒了,你猜对歌手名了！答案是\n", musicInfo[len(musicInfo)-1], "\n\n下面欣赏猜歌的歌曲"), answerTimes, tickTimes, true
	case len(musicInfo) == 4 && (strings.Contains(musicInfo[2], answer) || strings.EqualFold(musicInfo[2], answer)):
		return message.Text("太棒了,你猜对相关信息了！答案是\n", musicInfo[len(musicInfo)-1], "\n\n下面欣赏猜歌的歌曲"), answerTimes, tickTimes, true
	default:
		answerTimes++
		tickTimes++
		switch {
		case tickTimes > 2 && answerTimes < 6:
			return message.Text("答案不对哦,还有", 6-answerTimes, "次答题,加油啊~"), answerTimes, tickTimes, false
		case tickTimes > 2:
			return message.Text("次数到了,没能猜出来。答案是\n", musicInfo[len(musicInfo)-1], "\n\n下面欣赏猜歌的歌曲"), answerTimes, tickTimes, true
		default:
			return message.Text("答案不对,再听这段音频,要仔细听哦"), answerTimes, tickTimes, false
		}
	}
}

// ConvertText 将传入字符串中的英文转为小写，繁体中文转为简体中文
func ConvertText(input string) string {
	// 将字符串中的英文转为小写
	toLower := strings.ToLower(input)
	toLower, err := t2s.Convert(toLower)
	if err != nil {
		message.Text("简繁转换失败")
		return toLower
	}
	return toLower
}
