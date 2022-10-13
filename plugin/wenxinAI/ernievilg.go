// Package ernie AI画图
package ernie

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
	// 百度文心AI画图API
	wenxin "github.com/FloatTech/AnimeAPI/wenxinAI/ernievilg"
)

const (
	serviceName = "AIdraw"
	serviceErr  = "[" + serviceName + "]ERROR:\n"
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
	groupinfo = &keydb{
		db: &sql.Sqlite{},
	}
	limit = 50
	dtype = [...]string{
		"古风", "油画", "水彩画", "卡通画", "二次元", "浮世绘", "蒸汽波艺术", "low poly", "像素风格", "概念艺术", "未来主义", "赛博朋克", "写实风格", "洛丽塔风格", "巴洛克风格", "超现实主义",
	}
)

func init() { // 插件主体
	engine := control.Register(serviceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "AI画图\n" +
			"基于百度文心的免费AI画图插件,\n因为是免费的,图片质量你懂的。\n" +
			"key申请链接：https://wenxin.baidu.com/moduleApi/key\n" +
			"注意：每个apikey每日上限50次,总上限500次请求。次数超过了请自行更新apikey\n" +
			"- 为[自己/本群/QQ号/群+群号]设置AI画图key [API Key] [Secret Key]\n" +
			"例：\n[为10086设置AI画图key 123 456]\n[为群10010设置AI画图key 789 101]\n" +
			"- [bot名称]画几张[图片描述]的[图片类型][图片尺寸]\n" +
			"————————————————————\n" +
			"图片描述指南:\n图片主体，细节词(请用逗号连接)\n官方prompt指南:https://wenxin.baidu.com/wenxin/docs#Ol7ece95m\n" +
			"————————————————————\n" +
			"图片类型当前支持：" + strings.Join(dtype[:], "、") +
			"\n————————————————————\n" +
			"图片尺寸当前只支持：方图/长图/横图\n" +
			"————————————————————\n" +
			"指令示例：\n" +
			"椛椛帮我画几张金凤凰，背景绚烂，高饱和，古风，仙境，高清，4K，古风的油画方图",
		PrivateDataFolder: "ernievilg",
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
		groupinfo.db.DBPath = engine.DataFolder() + "keydb.db"
		err := groupinfo.db.Open(time.Hour * 24)
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
			userinfo, err1 := groupinfo.checkGroup(uid)
			info, err2 := groupinfo.checkGroup(gid)
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
					msg := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text("我画好了！"))}
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
			err = groupinfo.update(gid)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
			}
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("累死了，今天我最多只能画", info.DayLimit-1, "张图哦"))
		})
	engine.OnRegex(`^为(群)?(自己|本群|\d+)设置AI画图key\s(.*[^\s$])\s(.+)$`, getdb).SetBlock(true).
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
			err := groupinfo.insert(dbID, aKey, sKey)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			ctx.SendChain(message.Text("成功!"))
		})
}

// 登记group的key
func (sql *keydb) insert(gid int64, akey, skey string) error {
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
				MaxLimit:  500,
			}
		}
		return sql.db.Insert("groupinfo", &groupinfo)
	}
	// 进行更新
	groupinfo.APIKey = akey
	groupinfo.SecretKey = skey
	groupinfo.MaxLimit = 500
	return sql.db.Insert("groupinfo", &groupinfo)
}

// 获取group信息
func (sql *keydb) checkGroup(gid int64) (groupinfo apikey, err error) {
	sql.Lock()
	defer sql.Unlock()
	// 给db文件创建表格(没有才创建)，表格名称groupinfo，表格结构apikey
	err = sql.db.Create("groupinfo", &apikey{})
	if err != nil {
		return
	}
	// 先判断该群是否已经设置过key了
	if ok := sql.db.CanFind("groupinfo", "where ID is "+strconv.FormatInt(gid, 10)); !ok {
		if gid > 0 {
			err = errors.New("该群没有设置过apikey，请前往https://wenxin.baidu.com/moduleApi/key获取key值后，发送指令:\n为本群设置AI画图key [API Key] [Secret Key]")
		} else {
			err = errors.New("你没有设置过apikey，请前往https://wenxin.baidu.com/moduleApi/key获取key值后，发送指令:\n为自己设置AI画图key [API Key] [Secret Key]")
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
	if time.Since(time.Unix(groupinfo.Updatetime, 0)).Hours() > 24 {
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

// 记录次数(-1)
func (sql *keydb) update(gid int64) error {
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
	groupinfo.MaxLimit--
	groupinfo.DayLimit--
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
