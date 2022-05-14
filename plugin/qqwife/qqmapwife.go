//Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品
package qqwife

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

var (
	qqwifegroup = make(map[int64]map[int64]*userinfo, 64)  // 64个群的预算大小
	singledogCD = rate.NewManager[string](time.Hour*24, 1) //技能CD时间
	lastdate    time.Time
	mu          sync.Mutex
	sendtext    = [...][]string{
		{ //表白成功
			"今天你向ta表白了，ta羞涩的点了点头同意了！\n",
			"你对ta说“以我之名，冠你指间，一天相伴，一天相随”.ta捂着嘴点了点头\n\n",
		},
		{ //表白失败
			"今天你向ta表白了，ta毫无感情的拒绝了你",
			"今天你向ta表白了，ta对你说“你是一个非常好的人”",
			"今天你向ta表白了，ta给了你一个拥抱后擦肩而过",
		},
		{ //ntr成功
			"你处心积虑的接近ta，ta最终选择跟随你\n",
		},
	}
)

type userinfo struct {
	target int64  //对象的QQ号
	uName  string //用户名称
	tName  string //对象名称
}

func init() {
	engine := control.Register("qqwife", &control.Options{
		DisableOnDefault: false,
		Help: "一群一天一夫一妻制群老婆\n（每天凌晨刷新CP）\n" +
			"- 娶群友\n- 群老婆列表\n" +
			"--------------------------------\n以下技能每人只能二选一\n   CD24H，不跨天刷新\n--------------------------------\n" +
			"- (娶|嫁)@对方QQ\n- 当[对方Q号|@对方QQ]的小三\n",
	})
	engine.OnFullMatch("娶群友", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			mu.Lock()
			defer mu.Unlock()
			if time.Now().Day() != lastdate.Day() {
				qqwifegroup = make(map[int64]map[int64]*userinfo, 64) // 跨天就重新初始化数据
			}
			//判断列表是否为空
			gid := ctx.Event.GroupID
			if qqwifegroup[gid] == nil {
				qqwifegroup[gid] = make(map[int64]*userinfo, 32)
			}
			// 先判断是否已经娶过或者被娶
			uid := ctx.Event.UserID
			// 如果娶过
			wife, ok := qqwifegroup[gid][uid]
			if ok {
				ctx.SendChain(
					message.At(uid),
					message.Text("今天你的群老婆是"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(wife.target, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", wife.tName, "]",
						"(", wife.target, ")哒",
					),
				)
				return
			}
			// 如果被娶过
			husband, ok := qqwifegroup[gid][-uid]
			if ok {
				ctx.SendChain(
					message.At(uid),
					message.Text("今天你被娶了，群老公是"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(husband.target, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", husband.tName, "]",
						"(", husband.target, ")哒",
					),
				)
				return
			}
			//  无缓存获取群员列表
			temp := ctx.GetThisGroupMemberListNoCache().Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-30):]
			// 将已经娶过的人剔除
			qqgrouplist := make([]int64, 0, len(temp))
			if len(qqwifegroup) == 0 {
				for k := 0; k < len(temp); k++ {
					qqgrouplist = append(qqgrouplist, temp[k].Get("user_id").Int())
				}
			} else {
				for k := 0; k < len(temp); k++ {
					_, ok := qqwifegroup[gid][temp[k].Get("user_id").Int()]
					if ok {
						continue
					}
					_, ok = qqwifegroup[gid][-temp[k].Get("user_id").Int()]
					if ok {
						continue
					}
					qqgrouplist = append(qqgrouplist, temp[k].Get("user_id").Int())
				}
			}
			// 没有人（只剩自己）的时候
			if len(qqgrouplist) == 0 {
				ctx.SendChain(message.Text("噢，此时此刻你还是一只单身狗，等待下一次情缘吧"))
				return
			}
			// 随机抽娶
			fiancee := qqgrouplist[rand.Intn(len(qqgrouplist))]
			if fiancee == uid { // 如果是自己
				ctx.SendChain(message.Text("噢，此时此刻你还是一只单身狗，等待下一次情缘吧"))
				return
			}
			//互相了解对方信息
			fianceename := ctx.CardOrNickName(fiancee)
			username := ctx.CardOrNickName(uid)
			husbandinfo := &userinfo{
				target: fiancee,
				tName:  fianceename,
				uName:  username,
			}
			wifeinfo := &userinfo{
				target: uid,
				tName:  username,
				uName:  fianceename,
			}
			//去民政局办证
			qqwifegroup[gid][uid] = husbandinfo
			qqwifegroup[gid][-fiancee] = wifeinfo
			// 让大家吃席
			ctx.SendChain(
				message.At(uid),
				message.Text("今天你的群老婆是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", fianceename, "]",
					"(", fiancee, ")哒",
				),
			)
			//记录结婚时间
			lastdate = time.Now()
		})
	//单生狗专属技能
	engine.OnRegex(`^(娶|嫁)\[CQ:at,qq=(\d+)\]`, zero.OnlyGroup, checkdog).SetBlock(true).Limit(cdcheck, iscding).
		Handle(func(ctx *zero.Ctx) {
			mu.Lock()
			defer mu.Unlock()
			choice := ctx.State["regex_matched"].([]string)[1]
			fiancee, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			uid := ctx.Event.UserID
			if uid == fiancee { //如果是自己
				ctx.SendChain(message.Text("今日获得成就：自恋狂"))
				return
			}
			if rand.Intn(2) == 1 {
				//二分之一的概率表白成功
				gid := ctx.Event.GroupID
				//判断列表是否为空
				if qqwifegroup[gid] == nil {
					qqwifegroup[gid] = make(map[int64]*userinfo, 32)
				}
				//记录数据
				fianceename := ctx.CardOrNickName(fiancee)
				username := ctx.CardOrNickName(uid)
				husband := &userinfo{
					target: fiancee,
					tName:  fianceename,
					uName:  username,
				}
				wife := &userinfo{
					target: uid,
					tName:  username,
					uName:  fianceename,
				}
				//根据技能分配0和1
				var choicetext string
				switch choice {
				case "娶":
					qqwifegroup[gid][uid] = husband
					qqwifegroup[gid][-fiancee] = wife
					choicetext = "今天你的群老婆是"
				default:
					qqwifegroup[gid][-uid] = husband
					qqwifegroup[gid][fiancee] = wife
					choicetext = "今天你的群老公是"
				}
				// 输出结果
				ctx.SendChain(
					message.Text(sendtext[0][rand.Intn(len(sendtext[0]))]),
					message.At(uid),
					message.Text(choicetext),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(fiancee), "]",
						"(", fiancee, ")哒",
					),
				)
			} else {
				ctx.SendChain(message.Text(sendtext[1][rand.Intn(len(sendtext[1]))]))
			}
		})
	//NTR技能
	engine.OnRegex(`^当(\[CQ:at,qq=(\d+)\] |(\d+))的小三`, zero.OnlyGroup, checkcp).SetBlock(true).Limit(cdcheck, iscding).
		Handle(func(ctx *zero.Ctx) {
			mu.Lock()
			defer mu.Unlock()
			fid := ctx.State["regex_matched"].([]string)
			fiancee, _ := strconv.ParseInt(fid[2]+fid[3], 10, 64)
			if rand.Intn(10)/4 != 0 { // 十分之三的概率NTR成功
				ctx.SendChain(message.Text("你的ntr计划失败了"))
				return
			}
			gid := ctx.Event.GroupID
			//判断对象是0还是1
			choicetext := "婆"
			target, ok := qqwifegroup[gid][fiancee]
			if !ok {
				target, _ = qqwifegroup[gid][-fiancee]
				choicetext = "公"
			}
			//重新绑定CP
			delete(qqwifegroup[gid], target.target)
			uid := ctx.Event.UserID
			fianceename := ctx.CardOrNickName(fiancee)
			username := ctx.CardOrNickName(uid)
			husband := &userinfo{
				target: fiancee,
				tName:  fianceename,
				uName:  username,
			}
			wife := &userinfo{
				target: uid,
				tName:  username,
				uName:  fianceename,
			}
			switch choicetext {
			case "婆":
				qqwifegroup[gid][uid] = husband
				qqwifegroup[gid][-fiancee] = wife
			default:
				qqwifegroup[gid][-uid] = husband
				qqwifegroup[gid][fiancee] = wife
			}
			// 输出结果
			ctx.SendChain(
				message.Text(sendtext[2][rand.Intn(len(sendtext[2]))]),
				message.At(uid),
				message.Text("今天你的群老"+choicetext+"是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(fiancee), "]",
					"(", fiancee, ")哒",
				),
			)
		})
	//显示群CP
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			mu.Lock()
			defer mu.Unlock()
			group, ok := qqwifegroup[ctx.Event.GroupID]
			if !ok {
				ctx.SendChain(message.Text("你群并没有任何的CP额"))
				return
			}
			if len(group) == 0 {
				ctx.SendChain(message.Text("你群并没有任何的CP额"))
				return
			}
			cplist := make([]string, 1, len(group)+1)
			cplist[0] = "群老公←———→群老婆\n--------------------------"
			for husband, target := range group {
				if husband > 0 {
					cplist = append(cplist, target.uName+" & "+target.tName)
				}
			}
			ctx.SendChain(message.Text(strings.Join(cplist, "\n")))
		})
}

//注入判断 是否为单身狗
func checkdog(ctx *zero.Ctx) bool {
	mu.Lock()
	defer mu.Unlock()
	gid := ctx.Event.GroupID
	if qqwifegroup[gid] == nil {
		return true
	}
	fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，你的对象好像不存在？"))
		return false
	}
	uid := ctx.Event.UserID
	// 如果用户娶过
	husband, ok := qqwifegroup[gid][uid]
	if ok {
		switch husband.target {
		case fiancee:
			ctx.SendChain(message.Text("笨蛋~你明明已经娶了啊w"))
		default:
			ctx.SendChain(message.Text("笨蛋~你家里还有个吃白饭的w"))
		}
		return false
	}
	// 如果用户被娶过
	wife, ok := qqwifegroup[gid][-uid]
	if ok {
		switch wife.target {
		case fiancee:
			ctx.SendChain(message.Text("笨蛋~你明明已经嫁给他了啊w"))
		default:
			ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
		}
		return false
	}
	// 如果未婚妻娶过
	wife, ok = qqwifegroup[gid][fiancee]
	if ok {
		switch wife.target {
		case uid:
			ctx.SendChain(message.Text("笨蛋~你明明已经嫁给他了啊w"))
		default:
			ctx.SendChain(message.Text("他有别的女人了，你该放下了"))
		}
		return false
	}
	// 如果未婚妻被娶过
	husband, ok = qqwifegroup[gid][-fiancee]
	if ok {
		switch husband.target {
		case uid:
			ctx.SendChain(message.Text("笨蛋~你明明已经娶了啊w"))
		default:
			ctx.SendChain(message.Text("这是一个纯爱的世界，拒绝NTR"))
		}
		return false
	}
	return true
}

//注入判断 是否满足小三要求
func checkcp(ctx *zero.Ctx) bool {
	mu.Lock()
	defer mu.Unlock()
	gid := ctx.Event.GroupID
	if qqwifegroup[gid] == nil {
		ctx.SendChain(message.Text("ta无法达成你当小三的条件"))
		return false
	}
	fid := ctx.State["regex_matched"].([]string)
	fiancee, err := strconv.ParseInt(fid[2]+fid[3], 10, 64)
	for i, v := range ctx.State["regex_matched"].([]string) {
		fmt.Println(i, " : ", v)
	}
	if err != nil {
		ctx.SendChain(message.Text("额，你的对象好像不存在?"))
		return false
	}
	uid := ctx.Event.UserID
	// 如果用户娶过
	_, ok := qqwifegroup[gid][uid]
	if ok {
		ctx.SendChain(message.Text("抱歉，建国之后不支持后宫"))
		return false
	}
	// 如果用户被娶过
	_, ok = qqwifegroup[gid][-uid]
	if ok {
		ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
		return false
	}
	// 如果未婚妻没使用过插件
	_, ok1 := qqwifegroup[gid][fiancee]
	_, ok2 := qqwifegroup[gid][fiancee]
	if !ok1 || !ok2 {
		ctx.SendChain(message.Text("ta无法达成你当小三的条件"))
		return false
	}
	return true
}

//以群号和昵称为限制
func cdcheck(ctx *zero.Ctx) *rate.Limiter {
	limitID := strconv.FormatInt(ctx.Event.GroupID, 10) + strconv.FormatInt(ctx.Event.UserID, 10)
	return singledogCD.Load(limitID)
}
func iscding(ctx *zero.Ctx) {
	ctx.SendChain(message.Text("今日你的技能正在CD中"))
}
