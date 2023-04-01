// Package wenxin 百度文心AI
package wenxin

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 数据库
	sql "github.com/FloatTech/sqlite"
	// 百度文心
	// wenxin "github.com/FloatTech/AnimeAPI/wenxinAI/ernievilg"
)

type keydb struct {
	db *sql.Sqlite
	sync.RWMutex
}

// db内容
type apikey struct {
	ID         int64  // 群号
	APIKey     string // API Key
	SecretKey  string // Secret Key
	Token      string // AccessToken
	Updatetime int64  // token的有效时间
	MaxTimes   int    // 今日上限
}

var (
	wenxinvilg = &keydb{
		db: &sql.Sqlite{},
	}
	dtype = [...]string{
		"古风", "油画", "水彩画", "卡通画", "二次元", "浮世绘",
		"蒸汽波艺术", "low poly", "像素风格", "概念艺术", "未来主义",
		"赛博朋克", "写实风格", "洛丽塔风格", "巴洛克风格", "超现实主义",
	}
)

func init() { // 插件主体
	engine := control.Register("wenxinvilg", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "文心AI画图",
		Help: "基于百度文心的免费AI画图插件,\n因为是免费的,图片质量你懂的。\n" +
			"key申请链接:https://baidu.com/moduleApi/key\n" +
			"- 为[自己/本群/QQ号/群+群号]设置画图key [API Key] [Secret Key]\n" +
			"例:\n为自己设置画图key 123 456\n为10086设置画图key 123 456\n为群10010设置画图key 789 101\n" +
			"- 文心画图 ([图片类型] [图片尺寸] )[图片描述]\n" +
			"————————————————————\n" +
			"图片类型默认为二次元\n当前支持:\n" + strings.Join(dtype[:], "、") +
			"\n————————————————————\n" +
			"图片尺寸默认为方图\n当前支持:\n方图、长图、横图\n" +
			"————————————————————\n" +
			"图片描述指南:\n图片主体,细节词(请用逗号连接)\n官方prompt指南:https://baidu.com/wenxin/docs#Ol7ece95m\n" +
			"————————————————————\n" +
			"指令示例:\n" +
			"文心画图 金凤凰,背景绚烂,高饱和,古风,仙境,高清,4K,古风" +
			"文心画图 油画 方图 金凤凰,背景绚烂,高饱和,古风,仙境,高清,4K,古风",
		PrivateDataFolder: "wenxinAI",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Break()
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text(zero.BotConfig.NickName[0], "正在给别人画图,请不要打扰哦"),
				),
			)
		}),
	))
	cachePath := file.BOTPATH + "/" + engine.DataFolder() + "cache/"
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			return false
		}
		wenxinvilg.db.DBPath = engine.DataFolder() + "wenxin.db"
		err = wenxinvilg.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
			return false
		}
		return true
	})
	// 画图
	engine.OnRegex(`^文心画图\s+?(`+strings.Join(dtype[:], "|")+`)?\s?(方图|长图|横图)?\s?(.+)$`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := -ctx.Event.UserID
			gid := ctx.Event.GroupID
			// 获取个人和群的key
			userinfo, err1 := wenxinvilg.checkGroup(uid, "vilg")
			info, err2 := wenxinvilg.checkGroup(gid, "vilg")
			switch {
			// 如果是个人请求且报错
			case gid == 0 && err1 != nil:
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err1))
				return
			// 如果群报错而个人没有,就切换成个人的
			case err2 != nil && err1 == nil:
				info = userinfo
				if info.MaxTimes == 1 {
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("日访问量超限")))
					return
				}
			// 如果都报错就以群为优先级
			case err1 != nil && err2 != nil:
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err2))
				return
			default:
				if info.MaxTimes == 1 {
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("日访问量超限")))
					return
				}
			}
			// 创建任务
			keyword := ctx.State["regex_matched"].([]string)[3]
			if len([]rune(keyword)) >= 64 { // 描述不能超过64个字
				ctx.SendChain(message.Text("要求太多了啦！减少点！"))
				return
			}
			picType := ctx.State["regex_matched"].([]string)[1]
			if picType == "" {
				picType = "二次元"
			}
			chooseSize := ctx.State["regex_matched"].([]string)[2]
			wtime := 3
			picSize := "1024*1024"
			switch chooseSize {
			case "长图":
				wtime = 5
				picSize = "1024*1536"
			case "横图":
				wtime = 5
				picSize = "1536*1024"
			}
			taskID, err := BuildPicWork(info.Token, keyword, picType, picSize)
			if err != nil {
				if taskID == 17 {
					if err := wenxinvilg.setTimes(info.ID, "vilg", 1); err != nil {
						ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
					}
				}
				ctx.SendChain(message.Text("[wenxinvilg]ERROR", taskID, ":\n", err))
				return
			}
			if taskID < 1 {
				ctx.SendChain(message.Text("要求太复杂力！想不出来..."))
				return
			}
			// 开始画图
			ctx.SendChain(message.Text(zero.BotConfig.NickName[0], "知道了,我可能需要", time.Duration(wtime*10)*time.Second, "左右才能画好哦,请等待..."))
			i := 0
			for range time.NewTicker(10 * time.Second).C {
				// 等待 wtime * 10秒
				i++
				if i <= wtime {
					continue
				}
				/*
					if i > 60{// 十分钟还不出图就放弃
						ctx.SendChain(message.Text("呜呜呜,要求太复杂力！画不出来..."))
						return
					}
				// 获取结果*/
				picURL, status, err := GetPicResult(info.Token, taskID)
				if err != nil {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR", taskID, ":\n", err))
					return
				}
				if status == 1 {
					lastTime := time.Duration(i * 10 * int(time.Second))
					msg := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text("我画好了！\n本次绘画用了", lastTime))}
					for _, imginfo := range picURL {
						msg = append(msg,
							ctxext.FakeSenderForwardNode(ctx,
								message.Image(imginfo.Image)))
					}
					if id := ctx.Send(msg).ID(); id == 0 {
						ctx.SendChain(message.Text("ERROR: 可能被风控了"))
					}
					return
				}
			}
		})
	engine.OnRegex(`^文心仿图\s+?(`+strings.Join(dtype[:], "|")+`)?\s?(方图|长图|横图)?\s?(.(?:[^\]])*)`, getdb, zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := -ctx.Event.UserID
			gid := ctx.Event.GroupID
			picURL := ctx.State["image_url"].([]string)[0]
			picdata, err := web.GetData(picURL)
			if err != nil {
				return
			}
			cachePic := cachePath + strconv.FormatInt(gid, 10) + ".png"
			err = os.WriteFile(cachePic, picdata, 0644)
			if err != nil {
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
				return
			}
			// 获取个人和群的key
			userinfo, err1 := wenxinvilg.checkGroup(uid, "vilg")
			info, err2 := wenxinvilg.checkGroup(gid, "vilg")
			switch {
			// 如果是个人请求且报错
			case gid == 0 && err1 != nil:
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err1))
				return
			// 如果群报错而个人没有,就切换成个人的
			case err2 != nil && err1 == nil:
				info = userinfo
				if info.MaxTimes == 1 {
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("日访问量超限")))
					return
				}
			// 如果都报错就以群为优先级
			case err1 != nil && err2 != nil:
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err2))
				return
			default:
				if info.MaxTimes == 1 {
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("日访问量超限")))
					return
				}
			}
			// 创建任务
			keyword := ctx.State["regex_matched"].([]string)[3]
			if len([]rune(keyword)) >= 64 { // 描述不能超过64个字
				ctx.SendChain(message.Text("要求太多了啦！减少点！\n(请文字和图片分开发送)"))
				return
			}
			picType := ctx.State["regex_matched"].([]string)[1]
			if picType == "" {
				picType = "二次元"
			}
			chooseSize := ctx.State["regex_matched"].([]string)[2]
			wtime := 3
			picSize := "1024*1024"
			switch chooseSize {
			case "长图":
				wtime = 5
				picSize = "1024*1536"
			case "横图":
				wtime = 5
				picSize = "1536*1024"
			}
			taskID, err := BuildImgWork(info.Token, keyword, picType, picSize, cachePic)
			if err != nil {
				if taskID == 17 {
					if err := wenxinvilg.setTimes(info.ID, "vilg", 1); err != nil {
						ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
					}
				}
				ctx.SendChain(message.Text("[wenxinvilg]ERROR", taskID, ":\n", err))
				return
			}
			if taskID < 1 {
				ctx.SendChain(message.Text("要求太复杂力！想不出来..."))
				return
			}
			// 开始画图
			ctx.SendChain(message.Text(zero.BotConfig.NickName[0], "知道了,我可能需要", time.Duration(wtime*10)*time.Second, "左右才能画好哦,请等待..."))
			i := 0
			for range time.NewTicker(10 * time.Second).C {
				// 等待 wtime * 10秒
				i++
				if i <= wtime {
					continue
				}
				/*
					if i > 60{// 十分钟还不出图就放弃
						ctx.SendChain(message.Text("呜呜呜,要求太复杂力！画不出来..."))
						return
					}
				// 获取结果*/
				picURL, status, err := GetPicResult(info.Token, taskID)
				if err != nil {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR", status, ":\n", err))
					return
				}
				if status == 1 {
					lastTime := time.Duration(i * 10 * int(time.Second))
					msg := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text("我画好了！\n本次绘画用了", lastTime))}
					for _, imginfo := range picURL {
						msg = append(msg,
							ctxext.FakeSenderForwardNode(ctx,
								message.Image(imginfo.Image)))
					}
					if id := ctx.Send(msg).ID(); id == 0 {
						ctx.SendChain(message.Text("ERROR: 可能被风控了"))
					}
					return
				}
			}
		})
	engine.OnRegex(`^为(群)?(自己|本群|\d+)设置画图key\s(.*[^\s$])\s(.+)$`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			mode := ctx.State["regex_matched"].([]string)[1]
			user := ctx.State["regex_matched"].([]string)[2]
			aKey := ctx.State["regex_matched"].([]string)[3]
			sKey := ctx.State["regex_matched"].([]string)[4]
			dbID := -ctx.Event.UserID // 默认给自己
			switch {
			case mode != "": // 指定群的话
				gid, err := strconv.ParseInt(user, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
					return
				}
				dbID = gid
			case user == "本群": // 用于本群
				gid := ctx.Event.GroupID
				if gid == 0 {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", "请指定群聊,或者使用指令；\n为群xxx设置AI画图key xxx xxx"))
					return
				}
				dbID = gid
			case user != "自己": // 给别人开key
				uid, err := strconv.ParseInt(user, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
					return
				}
				dbID = -uid
			}
			err := wenxinvilg.insert(dbID, "vilg", aKey, sKey)
			if err != nil {
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
				return
			}
			ctx.SendChain(message.Text("成功!"))
		})
	/*********************************************************/
	en := control.Register("wenxinmodel", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "文心AI文本处理",
		Help: "基于百度文心AI的API文本处理\n" +
			"key申请链接:https://baidu.com/moduleApi/key\n" +
			"- 为[自己/本群/QQ号/群+群号]设置文心key [API Key] [Secret Key]\n" +
			"例:\n为自己设置文心key 123 456\n为10086设置文心key 123 456\n为群10010设置文心key 789 101\n" +
			"———————使用说明———————\n" +
			"---- {content} 表示给出此处文本----\n" +
			"---- [MASK] 表示给出此处问题----\n" +
			"---- {answer} 表示给出此处答案----\n" +
			"———————使用结构———————\n" +
			"结构1: [MASK]?\n" +
			"结构2: {content}。[MASK]?\n" +
			"结构3: {answer},[MASK]?\n" +
			"结构4: 要求:{content}\\n需求:[MASK]\n" +
			"结构5: 已知问题:{content}\\n求证:\n" +
			"结构6: 对对联:{content}" +
			"结构7: {content}\\n下一句:" +
			"———————使用示例———————\n" +
			"- 文心创作 今天天气为什么这么好?\n" +
			"- 文心创作 电车难题目前没人能解答。这是为什么?" +
			"- 文心创作 已知三边为3,4和5的三角形为直角三角形\\n求证:\n" +
			"- 文心创作 对对联:山清水秀地干净\n" +
			"- 文心创作 山清水秀地干净\\n下一句:\n" +
			"————————————————————\n" +
			"更多示例请阅读链接:\nhttps://baidu.com/wenxin/docs#xl75plkkg",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Break()
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text(zero.BotConfig.NickName[0], "正在给别人编辑,请不要打扰哦"),
				),
			)
		}),
	))
	en.OnRegex(`^为(群)?(自己|本群|\d+)设置文心key\s(.*[^\s$])\s(.+)$`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			mode := ctx.State["regex_matched"].([]string)[1]
			user := ctx.State["regex_matched"].([]string)[2]
			aKey := ctx.State["regex_matched"].([]string)[3]
			sKey := ctx.State["regex_matched"].([]string)[4]
			dbID := -ctx.Event.UserID // 默认给自己
			switch {
			case mode != "": // 指定群的话
				gid, err := strconv.ParseInt(user, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
					return
				}
				dbID = gid
			case user == "本群": // 用于本群
				gid := ctx.Event.GroupID
				if gid == 0 {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", "请指定群聊,或者使用指令；\n为群xxx设置AI画图key xxx xxx"))
					return
				}
				dbID = gid
			case user != "自己": // 给别人开key
				uid, err := strconv.ParseInt(user, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
					return
				}
				dbID = -uid
			}
			err := wenxinvilg.insert(dbID, "text", aKey, sKey)
			if err != nil {
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
				return
			}
			ctx.SendChain(message.Text("成功!"))
		})
	type style struct {
		world string
		ID    int
	}
	erniePrompt := map[string]style{
		"改写": {world: "SENT", ID: 20},
		"作文": {world: "zuowen", ID: 21},
		"文案": {world: "adtext", ID: 22},
		"摘要": {world: "Summarization", ID: 23},
		"对联": {world: "couplet", ID: 24},
		"问答": {world: "Dialogue", ID: 25},
		"小说": {world: "novel", ID: 26},
		"补全": {world: "cloze", ID: 27},
		"":   {world: "Misc", ID: 28},
		"抽取": {world: "Text2Annotation", ID: 30}}
	en.OnRegex(`^文心创作\s?(.*)$`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := -ctx.Event.UserID
			gid := ctx.Event.GroupID
			// 获取个人和群的key
			userinfo, err1 := wenxinvilg.checkGroup(uid, "vilg")
			info, err2 := wenxinvilg.checkGroup(gid, "vilg")
			switch {
			// 如果是个人请求且报错
			case gid == 0 && err1 != nil:
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err1))
				return
			// 如果群报错而个人没有,就切换成个人的
			case err2 != nil && err1 == nil:
				info = userinfo
				if info.MaxTimes == 1 {
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("日访问量超限")))
					return
				}
			// 如果都报错就以群为优先级
			case err1 != nil && err2 != nil:
				ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err2))
				return
			default:
				if info.MaxTimes == 1 {
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("日访问量超限")))
					return
				}
			}
			// 创建任务
			keyword := ctx.State["regex_matched"].([]string)[1]
			if len([]rune(keyword)) >= 1000 { // 描述不能超过1000
				ctx.SendChain(message.Text("你在写作文吗？减少点！"))
				return
			}
			prompt := erniePrompt[""]
			for promptType, promptWord := range erniePrompt {
				if strings.Contains(keyword, promptType) {
					prompt = promptWord
				}
			}
			taskID, err := BuildTextWork(info.Token, keyword, prompt.world, prompt.ID)
			if err != nil {
				if taskID == 17 {
					if err := wenxinvilg.setTimes(info.ID, "text", 1); err != nil {
						ctx.SendChain(message.Text("[wenxinvilg]ERROR:\n", err))
					}
				}
				ctx.SendChain(message.Text("[wenxinvilg]ERROR", taskID, ":\n", err))
				return
			}
			if taskID < 1 {
				ctx.SendChain(message.Text("要求太复杂力！想不出来..."))
				return
			}
			ctx.SendChain(message.Text(zero.BotConfig.NickName[0], "脑袋瓜不太行,让我想想..."))
			// 开始画图
			wtime := 3
			i := 0
			for range time.NewTicker(10 * time.Second).C {
				// 等待 wtime * 10秒
				i++
				if i <= wtime {
					continue
				}
				/*
					if i > 60{// 十分钟还不出图就放弃
						ctx.SendChain(message.Text("呜呜呜,要求太复杂力！画不出来..."))
						return
					}
				// 获取结果*/
				msgresult, status, err := GetTextResult(info.Token, taskID)
				if err != nil {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR", taskID, ":\n", err))
					return
				}
				if status == 1 {
					lastTime := time.Duration(i * 10 * int(time.Second))
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(msgresult, "\n(用时", lastTime, ")")))
					return
				} else {
					ctx.SendChain(message.Text("[wenxinvilg]ERROR: 请求超时!"))
					return
				}
			}
		})
}

// 登记group的key
func (sql *keydb) insert(gid int64, genre, akey, skey string) error {
	sql.Lock()
	defer sql.Unlock()
	// 给db文件创建表格(没有才创建),表格名称groupinfo,表格结构apikey
	err := sql.db.Create(genre, &apikey{})
	if err != nil {
		return err
	}
	// 进行更新
	return sql.db.Insert(genre, &apikey{
		ID:        gid,
		APIKey:    akey,
		SecretKey: skey,
		MaxTimes:  0,
	})
}

// 登记group的key
func (sql *keydb) setTimes(gid int64, genre string, set int) error {
	sql.Lock()
	defer sql.Unlock()
	// 给db文件创建表格(没有才创建),表格名称groupinfo,表格结构apikey
	err := sql.db.Create(genre, &apikey{})
	if err != nil {
		return err
	}
	info := apikey{}
	_ = sql.db.Find(genre, &info, "where ID is "+strconv.FormatInt(gid, 10))
	info.MaxTimes = set
	// 进行更新
	return sql.db.Insert(genre, &info)
}

// 获取group信息
func (sql *keydb) checkGroup(gid int64, model string) (groupinfo apikey, err error) {
	sql.Lock()
	defer sql.Unlock()
	// 给db文件创建表格(没有才创建),表格名称groupinfo,表格结构apikey
	err = sql.db.Create(model, &apikey{})
	if err != nil {
		return
	}
	// 先判断该群是否已经设置过key了
	if ok := sql.db.CanFind(model, "where ID is "+strconv.FormatInt(gid, 10)); !ok {
		if gid > 0 {
			err = errors.New("该群没有设置过apikey,请前往https://baidu.com/moduleApi/key获取key值后,发送指令:\n为本群设置" + model + "key [API Key] [Secret Key]\n或\n为自己设置" + model + "key [API Key] [Secret Key]")
		} else {
			err = errors.New("你没有设置过apikey,请前往https://baidu.com/moduleApi/key获取key值后,发送指令:\n为自己设置" + model + "key [API Key] [Secret Key]")
		}
		return
	}
	// 获取group信息
	err = sql.db.Find(model, &groupinfo, "where ID is "+strconv.FormatInt(gid, 10))
	if err != nil {
		return
	}
	// 如果token有效期过期
	if time.Since(time.Unix(groupinfo.Updatetime, 0)).Hours() > 24 || groupinfo.Token == "" {
		token, code, err := GetToken(groupinfo.APIKey, groupinfo.SecretKey)
		if err != nil {
			if code == 17 {
				groupinfo.MaxTimes = 1
			}
			_ = sql.db.Insert(model, &groupinfo)
			return groupinfo, err
		}
		groupinfo.Token = token
		groupinfo.MaxTimes = 0
		groupinfo.Updatetime = time.Now().Unix()
		_ = sql.db.Insert(model, &groupinfo)
	}
	return
}
