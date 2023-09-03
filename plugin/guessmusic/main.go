// Package guessmusic 基于zbp的猜歌插件
package guessmusic

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	wyy "github.com/FloatTech/AnimeAPI/neteasemusic"
	"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 图片输出
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/zbputils/img/text"
)

const serviceErr = "[guessmusic]error:"

var (
	// 用户数据
	cfg config
	// 插件主体
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "猜歌插件",
		Help: "------bot主人指令------\n" +
			"- 设置猜歌歌库路径 [绝对路径]\n" +
			"- [创建/删除]歌单 [歌单名称]\n" +
			"- 下载歌曲[歌曲名称/网易云歌曲ID]到[歌单名称]\n" +
			"------管 理 员 指 令------\n" +
			"- 设置猜歌默认歌单 [歌单名称]\n" +
			"- 上传歌曲[群文件的音乐名]到[歌单名称]\n" +
			"------公 用 指 令------\n" +
			"- 歌单列表\n" +
			"- [个人/团队]猜歌\n" +
			"\n------重 要 事 项------\n" +
			"1.本插件依赖ffmpeg\n" +
			"2.\"删除[歌单名称]\"是将本地歌单数据全部删除, 慎用\n" +
			"3.不支持下载VIP歌曲,如有需求请用群文件上传\n" +
			"4.未设置默认歌单的场合,猜歌歌单为歌单列表第一个。\n" +
			"此外可在\"[个人/团队]猜歌\"指令后面添加[-歌单名称]进行指定歌单猜歌\n" +
			"5.猜歌内容必须以[-]开头才会识别\n" +
			"6.歌曲命名规则为:\n歌名 - 歌手 - 其他(歌曲出处之类)" +
			"\n------插 件 扩 展------\n" +
			"内置了独角兽API,但API不保证可靠性。\n" +
			"可以自行搭建或寻找NeteaseCloudMusicApi框架的API,本插件支持该API以下指令\n" +
			"NeteaseCloudMusicApi项目地址:\nhttps://binaryify.github.io/NeteaseCloudMusicApi/#/\n" +
			"- 设置猜歌API帮助\n" +
			"- 设置猜歌API [API首页网址]\n" +
			"- 猜歌[开启/关闭][歌单/歌词]自动下载\n" +
			"- 登录网易云(这个指令目前不知道能干嘛,总之先保留了)\n" +
			"- 歌单信息 [网易云歌单链接/ID]\n" +
			"- [歌单名称]绑定网易云[网易云歌单链接/ID]\n" +
			"- 下载歌单[网易云歌单链接/ID]到[歌单名称]\n" +
			"- 解除绑定 [歌单名称]",
		PrivateDataFolder: "guessmusic",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Break()
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("已经有正在进行的游戏..."),
				),
			)
		}),
	))
	// 用于存放歌曲三个片段的缓存文件夹
	cachePath = engine.DataFolder() + "cache/"
	// 用于存放用户的配置
	cfgFile = engine.DataFolder() + "config.json"
	// ffmpeg支持的格式
	musictypelist = "mp3;MP3;wav;WAV;amr;AMR;3gp;3GP;3gpp;3GPP;acc;ACC"
)

func init() {
	// 新建缓存文件夹
	err := os.MkdirAll(cachePath, 0755)
	if err != nil {
		panic(serviceErr + err.Error())
	}
	// 载入用户配置
	if file.IsExist(cfgFile) {
		reader, err := os.Open(cfgFile)
		if err == nil {
			err = json.NewDecoder(reader).Decode(&cfg)
		}
		if err != nil {
			panic(serviceErr + err.Error())
		}
		err = reader.Close()
		if err != nil {
			panic(serviceErr + err.Error())
		}
	} else {
		// 配置默认 config
		cfg = config{
			MusicPath: file.BOTPATH + "/data/guessmusic/music/", // 绝对路径，歌库根目录,通过指令进行更改
			Playlist: []listRaw{
				{
					Name: "这里是歌单名称,id为网易云歌单ID",
					ID:   123456,
				},
			},
			Defaultlist: []dlist{
				{
					GroupID: 123456,
					Name:    "这里是歌单名称,gid是群号",
				},
			},
			API:   true,
			Local: true,
		}
		err = saveConfig(cfgFile)
		if err != nil {
			panic(serviceErr + err.Error())
		}
	}
	// 用户配置
	engine.OnPrefix("设置猜歌歌库路径", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["args"].(string)
			musicPath := strings.ReplaceAll(option, "\\", "/")
			if !strings.HasSuffix(musicPath, "/") {
				musicPath += "/"
			}
			err := os.MkdirAll(musicPath, 0755)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			cfg.MusicPath = musicPath
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
	engine.OnPrefix("创建歌单", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			newList := cfg.MusicPath + ctx.State["args"].(string)
			if file.IsNotExist(newList) {
				err := os.MkdirAll(newList, 0755)
				if err == nil {
					ctx.SendChain(message.Text("成功！"))
				} else {
					ctx.SendChain(message.Text(serviceErr, err))
				}
			} else {
				ctx.SendChain(message.Text("歌单已存在！"))
			}
		})
	engine.OnPrefix("删除歌单", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			delList := ctx.State["args"].(string)
			err = os.RemoveAll(cfg.MusicPath + delList)
			if err != nil {
				ctx.SendChain(message.Text("删除失败，可能是歌单名称错误。\n可以发送“歌单列表”获取歌单名称"))
				return
			}
			// 删除绑定的网易云ID
			index := -1
			for i, list := range cfg.Playlist {
				if delList == list.Name {
					index = i
					break
				}
			}
			if index == -1 {
				ctx.SendChain(message.Text("成功！"))
				return
			}
			cfg.Playlist = append(cfg.Playlist[:index], cfg.Playlist[index+1:]...)
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
	// 下载歌曲到对应的歌单里面
	engine.OnRegex(`^下载歌曲\s*(.*)\s*到\s*(.*)$`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			listName := ctx.State["regex_matched"].([]string)[2]
			ctx.SendChain(message.Text("正在校验歌单信息，请稍等"))
			// 是否存在该歌单
			if file.IsNotExist(cfg.MusicPath + listName) {
				ctx.SendChain(message.Text("歌单不存在，是否创建？(是/否)"))
				next := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(是|否)`), ctx.CheckSession())
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
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
			}
			searchlist, err := wyy.SearchMusic(keyword, 5)
			if err != nil {
				ctx.SendChain(message.Text("查询歌曲失败！\nerr:", err))
				return
			}
			if len(searchlist) == 0 {
				ctx.SendChain(message.Text("歌曲没有查询到，请确认信息正确"))
				return
			}
			var musicchoose []string
			for musicName := range searchlist {
				musicchoose = append(musicchoose, musicName)
			}
			savePath := cfg.MusicPath + listName + "/"
			index := 0
			if len(musicchoose) > 1 {
				var msg []string
				msg = append(msg, "搜索到相近的歌曲，请回复对应序号进行下载或回复取消")
				for i, musicName := range musicchoose {
					msg = append(msg, strconv.Itoa(i)+"."+musicName)
				}
				ctx.SendChain(message.Text(strings.Join(msg, "\n")))
				next := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`[0-4]|取消`), ctx.CheckSession())
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
						if answer == "取消" {
							ctx.SendChain(message.Text("已取消下载"))
							return
						}
						index, _ = strconv.Atoi(answer)
					}
					if answer != "" {
						break
					}
				}
			}
			musicName := musicchoose[index]
			// 下载歌曲
			err = wyy.DownloadMusic(searchlist[musicName], musicName, savePath)
			if err == nil {
				if cfg.Local {
					// 下载歌词
					_ = wyy.DownloadLrc(searchlist[musicName], musicName, savePath+"歌词/")
				}
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
	// 从群文件下载歌曲
	engine.OnRegex(`^上传歌曲\s*(.*)\s*到\s*(.*)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			fileName := ctx.State["regex_matched"].([]string)[1]
			listName := ctx.State["regex_matched"].([]string)[2]
			// 判断群文件是否存在
			fileSearchName, fileURL := getFileURLbyFileName(ctx, fileName)
			if fileSearchName == "" {
				ctx.SendChain(message.Text(serviceErr, "请确认群文件文件名称是否正确或存在"))
				return
			}
			// 解析歌曲信息
			music := strings.Split(fileSearchName, ".")
			// 获取音乐后缀
			musictype := music[len(music)-1]
			if !strings.Contains(musictypelist, musictype) {
				ctx.SendChain(message.Text(fileSearchName, "不是插件支持的后缀,请更改后缀"))
				return
			}
			// 获取音乐信息
			musicInfo := strings.Split(strings.ReplaceAll(fileSearchName, "."+musictype, ""), " - ")
			infoNum := len(musicInfo)
			if infoNum == 1 {
				ctx.SendChain(message.Text(fileSearchName, "不符合命名规则,请更改名称"))
				return
			}
			fileName = "歌名:" + musicInfo[0] + "\n歌手:" + musicInfo[1]
			musicAlia := ""
			if infoNum > 2 {
				musicAlia = musicInfo[2]
				fileName += "\n其他信息:\n" + strings.ReplaceAll(musicAlia, "&", "\n")
			}
			// 是否存在该歌单
			if file.IsNotExist(cfg.MusicPath + listName) {
				if !zero.SuperUserPermission(ctx) {
					ctx.SendChain(message.Text("歌单名称错误。\n可以发送“歌单列表”获取歌单名称"))
					return
				}
				ctx.SendChain(message.Text("歌单不存在，是否创建？(是/否)"))
				next := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(是|否)`), ctx.CheckSession())
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
				err = os.MkdirAll(cfg.MusicPath+listName, 0755)
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
			}
			// 下载歌曲
			ctx.SendChain(message.Text("在群文件中找到了歌曲,信息如下:\n", fileName, "\n确认正确后回复“是/否”进行上传"))
			next := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(是|否)`), ctx.CheckSession())
			recv, cancel := next.Repeat()
			defer cancel()
			wait := time.NewTimer(120 * time.Second)
			answer := ""
			for {
				select {
				case <-wait.C:
					wait.Stop()
					ctx.SendChain(message.Text("等待超时，取消上传"))
					return
				case c := <-recv:
					wait.Stop()
					answer = c.Event.Message.String()
				}
				if answer == "否" {
					ctx.SendChain(message.Text("上传已经取消"))
					return
				}
				if answer != "" {
					break
				}
			}
			err = file.DownloadTo(fileURL, cfg.MusicPath+listName+"/"+fileSearchName)
			if err == nil {
				process.SleepAbout1sTo2s()
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
			}
		})
	engine.OnFullMatch("歌单列表").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			filelist, err := getlist(cfg.MusicPath)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
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
			boldfd, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
			}
			fd, err := file.GetLazyData(text.FontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.ParseFontFace(boldfd, fontSize); err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			_, h := canvas.MeasureString("序号\t\t歌单名\t\t\t歌曲数量\t\t网易云歌单ID")
			/***********绘制标题***********/
			canvas.DrawString("序号\t\t歌单名\t\t歌曲数量\t\t网易云歌单ID", 20, 50-h) // 放置在中间位置
			canvas.DrawString("——————————————————————", 20, 70-h)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.ParseFontFace(fd, fontSize); err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
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
			data, err := imgfactory.ToBytes(canvas.Image())
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			if id := ctx.SendChain(message.ImageBytes(data)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	engine.OnPrefix("设置猜歌默认歌单", zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["args"].(string)
			gid := ctx.Event.GroupID
			if file.IsNotExist(cfg.MusicPath + option) {
				ctx.SendChain(message.Text("歌单名称错误，可以发送“歌单列表”获取歌单名称"))
				return
			}
			cfg.Defaultlist = append(cfg.Defaultlist, dlist{
				GroupID: gid,
				Name:    option,
			})
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text(serviceErr, err))
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
		wyyID[wyyinfo.Name] = wyyinfo.ID
	}
	err = os.MkdirAll(pathOfMusic, 0755)
	if err != nil {
		return
	}
	files, err := os.ReadDir(pathOfMusic)
	if err != nil {
		return
	}
	if len(files) == 0 {
		err = errors.New("所设置的歌库不存在任何歌单！")
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

// 遍历群文件
func getFileURLbyFileName(ctx *zero.Ctx, fileName string) (fileSearchName, fileURL string) {
	filesOfGroup := ctx.GetThisGroupRootFiles()
	files := filesOfGroup.Get("files").Array()
	folders := filesOfGroup.Get("folders").Array()
	// 遍历当前目录的文件名
	if len(files) != 0 {
		for _, fileNameOflist := range files {
			if strings.Contains(fileNameOflist.Get("file_name").String(), fileName) {
				fileSearchName = fileNameOflist.Get("file_name").String()
				fileURL = ctx.GetThisGroupFileUrl(fileNameOflist.Get("busid").Int(), fileNameOflist.Get("file_id").String())
				return
			}
		}
	}
	// 遍历子文件夹
	if len(folders) != 0 {
		for _, folderNameOflist := range folders {
			folderID := folderNameOflist.Get("folder_id").String()
			fileSearchName, fileURL = getFileURLbyfolderID(ctx, fileName, folderID)
			if fileSearchName != "" {
				return
			}
		}
	}
	return
}
func getFileURLbyfolderID(ctx *zero.Ctx, fileName, folderid string) (fileSearchName, fileURL string) {
	filesOfGroup := ctx.GetThisGroupFilesByFolder(folderid)
	files := filesOfGroup.Get("files").Array()
	folders := filesOfGroup.Get("folders").Array()
	// 遍历当前目录的文件名
	if len(files) != 0 {
		for _, fileNameOflist := range files {
			if strings.Contains(fileNameOflist.Get("file_name").String(), fileName) {
				fileSearchName = fileNameOflist.Get("file_name").String()
				fileURL = ctx.GetThisGroupFileUrl(fileNameOflist.Get("busid").Int(), fileNameOflist.Get("file_id").String())
				return
			}
		}
	}
	// 遍历子文件夹
	if len(folders) != 0 {
		for _, folderNameOflist := range folders {
			folderID := folderNameOflist.Get("folder_id").String()
			fileSearchName, fileURL = getFileURLbyfolderID(ctx, fileName, folderID)
			if fileSearchName != "" {
				return
			}
		}
	}
	return
}
