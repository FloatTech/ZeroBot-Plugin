// Package wenxin 百度文心AI
package wenxin

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 数据库
	sql "github.com/FloatTech/sqlite"
	// 百度文心大模型
	model "github.com/FloatTech/AnimeAPI/wenxinAI/erniemodle"
	// 百度文心AI画图API
	wenxin "github.com/FloatTech/AnimeAPI/wenxinAI/ernievilg"
)

const (
	serviceErr = "[wenxinvilg]ERROR:\n"
	modelErr   = "[wenxinmodel]ERROR:\n"
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
	MaxLimit   int    // 总使用次数
	DayLimit   int    // 当天的使用次数
	Lasttime   string // 记录使用的时间，用于刷新使用次数
}

var (
	name     = "椛椛"
	limit    int
	vilginfo = &keydb{
		db: &sql.Sqlite{},
	}
	modelinfo = &keydb{
		db: &sql.Sqlite{},
	}
	dtype = [...]string{
		"古风", "油画", "水彩画", "卡通画", "二次元", "浮世绘", "蒸汽波艺术", "low poly", "像素风格", "概念艺术", "未来主义", "赛博朋克", "写实风格", "洛丽塔风格", "巴洛克风格", "超现实主义",
	}
)

func init() { // 插件主体
	go func() {
		process.GlobalInitMutex.Lock()
		defer process.GlobalInitMutex.Unlock()
		name = zero.BotConfig.NickName[0]
	}()
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "文心AI画图",
		Help: "基于百度文心的免费AI画图插件,\n因为是免费的,图片质量你懂的。\n" +
			"key申请链接:https://wenxin.baidu.com/moduleApi/key\n" +
			"key和erniemodel插件的key相同。\n" +
			"注意:每个apikey每日上限50次,总上限500次请求。次数超过了请自行更新apikey\n" +
			"- 为[自己/本群/QQ号/群+群号]设置画图key [API Key] [Secret Key]\n" +
			"例：\n为自己设置画图key 123 456\n为10086设置画图key 123 456\n为群10010设置画图key 789 101\n" +
			"- [bot名称]画几张[图片描述]的[图片类型][图片尺寸]\n" +
			"————————————————————\n" +
			"图片描述指南:\n图片主体，细节词(请用逗号连接)\n官方prompt指南:https://wenxin.baidu.com/wenxin/docs#Ol7ece95m\n" +
			"————————————————————\n" +
			"图片类型当前支持：" + strings.Join(dtype[:], "、") +
			"\n————————————————————\n" +
			"图片尺寸当前只支持：方图/长图/横图\n" +
			"————————————————————\n" +
			"指令示例：\n" +
			name + "帮我画几张金凤凰，背景绚烂，高饱和，古风，仙境，高清，4K，古风的油画方图",
		PrivateDataFolder: "wenxinAI",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Break()
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text(zero.BotConfig.NickName[0], "正在给别人画图，请不要打扰哦"),
				),
			)
		}),
	))
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		vilginfo.db.DBPath = engine.DataFolder() + "ernieVilg.db"
		err := vilginfo.db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text(serviceErr, err))
			return false
		}
		return true
	})
	// 画图
	engine.OnRegex(`画几张(.*[^的$])的(.*[^\s$])(方图|长图|横图)$`, zero.OnlyToMe, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := -ctx.Event.UserID
			gid := ctx.Event.GroupID
			// 获取个人和群的key
			userinfo, err1 := vilginfo.checkGroup(uid, "vilg")
			info, err2 := vilginfo.checkGroup(gid, "vilg")
			switch {
			// 如果是个人请求且报错
			case gid == 0 && err1 != nil:
				ctx.SendChain(message.Text(serviceErr, err1))
				return
			// 如果群报错而个人没有,就切换成个人的
			case err2 != nil && err1 == nil:
				gid = uid
				info = userinfo
			// 如果都报错就以群为优先级
			case err1 != nil && err2 != nil:
				ctx.SendChain(message.Text(serviceErr, err2))
				return
			}
			// 判断使用次数
			check := false
			switch {
			// 群和个人都没有次数了
			case info.DayLimit == 0 && userinfo.DayLimit == 0:
				ctx.SendChain(message.Text("我已经画了", limit, "张了！我累了！不画不画，就不画！"))
				return
			// 个人还有次数的话
			case info.DayLimit == 0 && userinfo.DayLimit != 0:
				check = true
			}
			switch {
			// 群和个人都没有总次数了
			case info.MaxLimit == 0 && userinfo.MaxLimit == 0:
				ctx.SendChain(message.Text("设置的key使用次数超过了限额，请更换key。"))
				return
			// 个人还有总次数的话
			case info.MaxLimit == 0 && userinfo.MaxLimit != 0:
				check = true
			}
			if check { // 如果只有个人有次数就切换回个人key
				gid = uid
				info = userinfo
			}
			// 创建任务
			keyword := ctx.State["regex_matched"].([]string)[1]
			if len([]rune(keyword)) >= 64 { // 描述不能超过64个字
				ctx.SendChain(message.Text("要求太多了啦！减少点！"))
				return
			}
			picType := ctx.State["regex_matched"].([]string)[2]
			chooseSize := ctx.State["regex_matched"].([]string)[3]
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
			taskID, err := wenxin.BuildWork(info.Token, keyword, picType, picSize)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			if taskID < 1 {
				ctx.SendChain(message.Text("要求太复杂力！想不出来..."))
				return
			}
			// 开始画图
			ctx.SendChain(message.Text(zero.BotConfig.NickName[0], "知道了，我可能需要", time.Duration(wtime*10)*time.Second, "左右才能画好哦，请等待..."))
			i := 0
			for range time.NewTicker(10 * time.Second).C {
				// 等待 wtime * 10秒
				i++
				if i <= wtime {
					continue
				}
				/*
					if i > 60{// 十分钟还不出图就放弃
						ctx.SendChain(message.Text("呜呜呜，要求太复杂力！画不出来..."))
						return
					}
				// 获取结果*/
				picURL, status, err := wenxin.GetPic(info.Token, taskID)
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
				if status == "0" {
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
					break
				}
			}
			err = vilginfo.update(gid, 1)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
			}
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("累死了，今天我最多只能画", info.DayLimit-1, "张图哦"))
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
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
				dbID = gid
			case user == "本群": // 用于本群
				gid := ctx.Event.GroupID
				if gid == 0 {
					ctx.SendChain(message.Text(serviceErr, "请指定群聊，或者使用指令；\n为群xxx设置AI画图key xxx xxx"))
					return
				}
				dbID = gid
			case user != "自己": // 给别人开key
				uid, err := strconv.ParseInt(user, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
				dbID = -uid
			}
			err := vilginfo.insert(dbID, "vilg", aKey, sKey)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			ctx.SendChain(message.Text("成功!"))
		})
	/*********************************************************/
	en := control.Register("wenxinmodel", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "文心AI文本处理",
		Help: "基于百度文心AI的API文本处理\n" +
			"key申请链接:https://wenxin.baidu.com/moduleApi/key\n" +
			"key和ernievilg插件的key相同。\n" +
			"注意:每个apikey每日上限200条,总上限2000条。次数超过了请自行更新apikey\n" +
			"- 为[自己/本群/QQ号/群+群号]设置文心key [API Key] [Secret Key]\n" +
			"例：\n为自己设置文心key 123 456\n为10086设置文心key 123 456\n为群10010设置文心key 789 101\n" +
			"————————————————————\n" +
			"- 文心作文 (x字的)[作文题目]\n" +
			"————————————————————\n" +
			"- 文心提案 (x字的)[文案标题]\n" +
			"————————————————————\n" +
			"- 文心摘要 (x字的)[文章内容]\n" +
			"————————————————————\n" +
			"- 文心小说 (x字的)[小说上文]\n" +
			"————————————————————\n" +
			"- 文心对联 [上联]\n" +
			"————————————————————\n" +
			"- 文心问答 [问题]\n" +
			"————————————————————\n" +
			"- 文心补全 [带“_”的填空题]\n" +
			"————————————————————\n" +
			"- 文心自定义 [prompt]\n\n" +
			"prompt: [问题描述] [问题类型]:[题目] [解答类型]:[解题必带内容]\n" +
			"指令示例:\n" +
			"文心自定义 请写出下面这道题的解题过程。\\n题目:养殖场养鸭376只,养鸡的只数比鸭多258只,这个养殖场一共养鸭和鸡多少只?\\n解：\n\n" +
			"文心自定义 1+1=?\n" +
			"文心自定义 歌曲名：大风车转啊转\\n歌词：",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Break()
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text(zero.BotConfig.NickName[0], "正在给别人编辑，请不要打扰哦"),
				),
			)
		}),
	))
	getmodeldb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		modelinfo.db.DBPath = engine.DataFolder() + "ernieModel.db"
		err := modelinfo.db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text(modelErr, err))
			return false
		}
		return true
	})
	en.OnRegex(`^为(群)?(自己|本群|\d+)设置文心key\s(.*[^\s$])\s(.+)$`, getmodeldb).SetBlock(true).
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
					ctx.SendChain(message.Text(modelErr, err))
					return
				}
				dbID = gid
			case user == "本群": // 用于本群
				gid := ctx.Event.GroupID
				if gid == 0 {
					ctx.SendChain(message.Text(modelErr, "请指定群聊，或者使用指令；\n为群xxx设置AI画图key xxx xxx"))
					return
				}
				dbID = gid
			case user != "自己": // 给别人开key
				uid, err := strconv.ParseInt(user, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text(modelErr, err))
					return
				}
				dbID = -uid
			}
			err := modelinfo.insert(dbID, "model", aKey, sKey)
			if err != nil {
				ctx.SendChain(message.Text(modelErr, err))
				return
			}
			ctx.SendChain(message.Text("成功!"))
		})

	var erniemodel = map[string]int{
		"作文":  1,
		"提案":  2,
		"摘要":  3,
		"对联":  4,
		"问答":  5,
		"小说":  6,
		"补全":  7,
		"自定义": 8}
	var erniePrompt = map[string]string{
		"作文": "zuowen",
		"提案": "adtext",
		"摘要": "Summarization",
		"对联": "couplet",
		"问答": "Dialogue",
		"小说": "novel",
		"补全": "cloze"}
	en.OnRegex(`^文心(作文|提案|摘要|小说)\s?((\d+)字的)?(.*)$`, getmodeldb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := -ctx.Event.UserID
			gid := ctx.Event.GroupID
			// 获取个人和群的key
			userinfo, err1 := modelinfo.checkGroup(uid, "model")
			info, err2 := modelinfo.checkGroup(gid, "model")
			switch {
			// 如果是个人请求且报错
			case gid == 0 && err1 != nil:
				ctx.SendChain(message.Text(modelErr, err1))
				return
			// 如果群报错而个人没有,就切换成个人的
			case err2 != nil && err1 == nil:
				gid = uid
				info = userinfo
			// 如果都报错就以群为优先级
			case err1 != nil && err2 != nil:
				ctx.SendChain(message.Text(modelErr, err2))
				return
			}
			// 判断使用次数
			check := false
			switch {
			// 群和个人都没有次数了
			case info.DayLimit == 0 && userinfo.DayLimit == 0:
				ctx.SendChain(message.Text("今日请求次数已到200次了,明天在玩吧"))
				return
			// 个人还有次数的话
			case info.DayLimit == 0 && userinfo.DayLimit != 0:
				check = true
			}
			switch {
			// 群和个人都没有总次数了
			case info.MaxLimit == 0 && userinfo.MaxLimit == 0:
				ctx.SendChain(message.Text("设置的key使用次数超过了限额,请更换key。"))
				return
			// 个人还有总次数的话
			case info.MaxLimit == 0 && userinfo.MaxLimit != 0:
				check = true
			}
			if check { // 如果只有个人有次数就切换回个人key
				gid = uid
				info = userinfo
			}
			// 调用API
			modelStr := ctx.State["regex_matched"].([]string)[1]
			mun := ctx.State["regex_matched"].([]string)[3]
			minlen := 1
			maxlen := 128
			if mun != "" {
				max, err := strconv.Atoi(mun)
				if err != nil {
					ctx.SendChain(message.Text(modelErr, err))
					return
				}
				minlen = max
				if max > 128 {
					maxlen = max
				}
			}
			keyword := ctx.State["regex_matched"].([]string)[4]
			if len([]rune(keyword)) >= 1000 { // 描述不能超过1000
				ctx.SendChain(message.Text("是你写作文还是我写？减少点！"))
				return
			}
			result, err := model.GetResult(info.Token, erniemodel[modelStr], keyword, minlen, maxlen, erniePrompt[modelStr])
			if err != nil {
				ctx.SendChain(message.Text(modelErr, err))
				return
			}
			if id := ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(keyword, "，", result))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 请求超时!"))
			}
			err = modelinfo.update(gid, 1)
			if err != nil {
				ctx.SendChain(message.Text(modelErr, err))
			}
		})
	en.OnRegex(`^文心(对联|问答|补全|自定义)\s?(.*)$`, getmodeldb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := -ctx.Event.UserID
			gid := ctx.Event.GroupID
			// 获取个人和群的key
			userinfo, err1 := modelinfo.checkGroup(uid, "model")
			info, err2 := modelinfo.checkGroup(gid, "model")
			switch {
			// 如果是个人请求且报错
			case gid == 0 && err1 != nil:
				ctx.SendChain(message.Text(modelErr, err1))
				return
			// 如果群报错而个人没有,就切换成个人的
			case err2 != nil && err1 == nil:
				gid = uid
				info = userinfo
			// 如果都报错就以群为优先级
			case err1 != nil && err2 != nil:
				ctx.SendChain(message.Text(modelErr, err2))
				return
			}
			// 判断使用次数
			check := false
			switch {
			// 群和个人都没有次数了
			case info.DayLimit == 0 && userinfo.DayLimit == 0:
				ctx.SendChain(message.Text("今日请求次数已到200次了,明天在玩吧"))
				return
			// 个人还有次数的话
			case info.DayLimit == 0 && userinfo.DayLimit != 0:
				check = true
			}
			switch {
			// 群和个人都没有总次数了
			case info.MaxLimit == 0 && userinfo.MaxLimit == 0:
				ctx.SendChain(message.Text("设置的key使用次数超过了限额,请更换key。"))
				return
			// 个人还有总次数的话
			case info.MaxLimit == 0 && userinfo.MaxLimit != 0:
				check = true
			}
			if check { // 如果只有个人有次数就切换回个人key
				gid = uid
				info = userinfo
			}
			// 创建任务
			modelStr := ctx.State["regex_matched"].([]string)[1]
			keyword := ctx.State["regex_matched"].([]string)[2]
			if len([]rune(keyword)) >= 1000 { // 描述不能超过1000
				ctx.SendChain(message.Text("你在写作文吗？减少点！"))
				return
			}
			result, err := model.GetResult(info.Token, erniemodel[modelStr], keyword, 1, 128, erniePrompt[modelStr])
			if err != nil {
				ctx.SendChain(message.Text(modelErr, err))
				return
			}
			if id := ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(result))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 请求超时!"))
			}
			err = modelinfo.update(gid, 1)
			if err != nil {
				ctx.SendChain(message.Text(modelErr, err))
			}
		})
}

// 登记group的key
func (sql *keydb) insert(gid int64, model, akey, skey string) error {
	sql.Lock()
	defer sql.Unlock()
	// 给db文件创建表格(没有才创建)，表格名称groupinfo，表格结构apikey
	err := sql.db.Create("groupinfo", &apikey{})
	if err != nil {
		return err
	}
	// 获取group信息
	groupinfo := apikey{} // 用于暂存数据
	err = sql.db.Find("groupinfo", &groupinfo, "where ID is "+strconv.FormatInt(gid, 10))
	if err != nil {
		// 如果该group没有注册过
		err = sql.db.Find("groupinfo", &groupinfo, "where APIKey is '"+akey+"' and SecretKey is '"+skey+"'")
		if err == nil {
			// 如果key存在过将当前的数据迁移过去
			groupinfo.ID = gid
		} else {
			groupinfo = apikey{
				ID:        gid,
				APIKey:    akey,
				SecretKey: skey,
			}
			switch model {
			case "vilg":
				groupinfo.MaxLimit = 500
			case "model":
				groupinfo.MaxLimit = 2000
			}
		}
		return sql.db.Insert("groupinfo", &groupinfo)
	}
	// 进行更新
	groupinfo.APIKey = akey
	groupinfo.SecretKey = skey
	groupinfo.Token = ""
	groupinfo.Updatetime = 0
	switch model {
	case "vilg":
		groupinfo.MaxLimit = 500
	case "model":
		groupinfo.MaxLimit = 2000
	}
	return sql.db.Insert("groupinfo", &groupinfo)
}

// 获取group信息
func (sql *keydb) checkGroup(gid int64, model string) (groupinfo apikey, err error) {
	sql.Lock()
	defer sql.Unlock()
	// 给db文件创建表格(没有才创建)，表格名称groupinfo，表格结构apikey
	err = sql.db.Create("groupinfo", &apikey{})
	if err != nil {
		return
	}
	switch model {
	case "vilg":
		limit = 50
		model = "画图"
	case "model":
		limit = 200
		model = "文心"
	}
	// 先判断该群是否已经设置过key了
	if ok := sql.db.CanFind("groupinfo", "where ID is "+strconv.FormatInt(gid, 10)); !ok {
		if gid > 0 {
			err = errors.New("该群没有设置过apikey，请前往https://wenxin.baidu.com/moduleApi/key获取key值后，发送指令:\n为本群设置" + model + "key [API Key] [Secret Key]\n或\n为自己设置" + model + "key [API Key] [Secret Key]")
		} else {
			err = errors.New("你没有设置过apikey，请前往https://wenxin.baidu.com/moduleApi/key获取key值后，发送指令:\n为自己设置" + model + "key [API Key] [Secret Key]")
		}
		return
	}
	// 获取group信息
	err = sql.db.Find("groupinfo", &groupinfo, "where ID is "+strconv.FormatInt(gid, 10))
	if err != nil {
		return
	}
	// 如果隔天使用刷新次数
	if time.Now().Format("2006/01/02") != groupinfo.Lasttime {
		groupinfo.DayLimit = limit
		groupinfo.Lasttime = time.Now().Format("2006/01/02")
	}
	if err = sql.db.Insert("groupinfo", &groupinfo); err != nil {
		return
	}
	// 如果token有效期过期
	if time.Since(time.Unix(groupinfo.Updatetime, 0)).Hours() > 24 || groupinfo.Token == "" {
		token, err1 := wenxin.GetToken(groupinfo.APIKey, groupinfo.SecretKey)
		if err1 != nil {
			err = err1
			return
		}
		groupinfo.Token = token
		groupinfo.Updatetime = time.Now().Unix()
		err = sql.db.Insert("groupinfo", &groupinfo)
		if err == nil {
			// 更新相同key的他人次数
			condition := "where not ID is " + strconv.FormatInt(gid, 10) +
				" and APIKey = '" + groupinfo.APIKey +
				"' and SecretKey = '" + groupinfo.SecretKey + "'"
			otherinfo := apikey{}
			var groups []int64 // 将相同的key的ID暂存
			// 无视没有找到相同的key的err
			_ = sql.db.FindFor("groupinfo", &otherinfo, condition, func() error {
				groups = append(groups, otherinfo.ID)
				return nil
			})
			if len(groups) != 0 { // 如果有相同的key就更新
				for _, group := range groups {
					err = sql.db.Find("groupinfo", &otherinfo, "where ID is "+strconv.FormatInt(group, 10))
					if err == nil {
						otherinfo.Token = groupinfo.Token
						otherinfo.Updatetime = groupinfo.Updatetime
						err = sql.db.Insert("groupinfo", &otherinfo)
					}
				}
			}
		}
	}
	return
}

// 记录次数(-sub)
func (sql *keydb) update(gid int64, sub int) error {
	sql.Lock()
	defer sql.Unlock()
	// 给db文件创建表格(没有才创建)，表格名称groupinfo，表格结构apikey
	err := sql.db.Create("groupinfo", &apikey{})
	if err != nil {
		return err
	}
	groupinfo := apikey{} // 用于暂存数据
	// 获取group信息
	err = sql.db.Find("groupinfo", &groupinfo, "where ID is "+strconv.FormatInt(gid, 10))
	if err != nil {
		return err
	}
	groupinfo.MaxLimit -= sub
	groupinfo.DayLimit -= sub
	err = sql.db.Insert("groupinfo", &groupinfo)
	if err != nil {
		return err
	}
	// 更新相同key的他人次数
	condition := "where not ID is " + strconv.FormatInt(gid, 10) +
		" and APIKey = '" + groupinfo.APIKey +
		"' and SecretKey = '" + groupinfo.SecretKey + "'"
	otherinfo := apikey{}
	var groups []int64 // 将相同的key的ID暂存
	// 无视没有找到相同的key的err
	_ = sql.db.FindFor("groupinfo", &otherinfo, condition, func() error {
		groups = append(groups, otherinfo.ID)
		return nil
	})
	if len(groups) != 0 { // 如果有相同的key就更新
		for _, group := range groups {
			err = sql.db.Find("groupinfo", &otherinfo, "where ID is "+strconv.FormatInt(group, 10))
			if err == nil {
				otherinfo.MaxLimit = groupinfo.MaxLimit
				otherinfo.DayLimit = groupinfo.DayLimit
				otherinfo.Lasttime = groupinfo.Lasttime
				err = sql.db.Insert("groupinfo", &otherinfo)
			}
		}
	}
	return err
}
