// Package robbery 打劫群友  基于“qqwife”插件魔改
package robbery

import (
	"math/rand"
	"strconv"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/wdvxdr1123/ZeroBot/extension/single"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type robberyRepo struct {
	sync.RWMutex
	db sql.Sqlite
}

type robberyRecord struct {
	UserID   int64  `db:"user_id"`   // 劫匪
	VictimID int64  `db:"victim_id"` // 受害者
	Time     string `db:"time"`      // 时间
}

func init() {
	var police robberyRepo
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "打劫别人的钱包",
		Help: "- 打劫[对方Q号|@对方QQ]\n" +
			"1. 受害者钱包少于1000不能被打劫\n" +
			"2. 打劫成功率 40%\n" +
			"4. 打劫失败罚款1000（钱不够，钱包归零）\n" +
			"5. 保险赔付0-80%\n" +
			"6. 打劫成功获得对方0-5%+500的财产（最高1W）\n" +
			"7. 每日可打劫或被打劫一次\n" +
			"8. 打劫失败不计入次数\n",
		PrivateDataFolder: "robbery",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("别着急，警察局门口排长队了！"),
				),
			)
		}),
	))
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		police.db = sql.New(engine.DataFolder() + "robbery.db")
		err := police.db.Open(time.Hour)
		if err == nil {
			// 创建CD表
			err = police.db.Create("criminal_record", &robberyRecord{})
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return false
			}
			return true
		}
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	})

	// 打劫功能
	engine.OnRegex(`^打劫\s?(\[CQ:at,(?:\S*,)?qq=(\d+)(?:,\S*)?\]|(\d+))`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			fiancee := ctx.State["regex_matched"].([]string)
			victimID, _ := strconv.ParseInt(fiancee[2]+fiancee[3], 10, 64)
			if victimID == uid {
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.At(uid), message.Text("不能打劫自己")))
				return
			}

			// 查询记录
			ok, err := police.getRecord(victimID, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}

			if ok == 1 {
				ctx.SendChain(message.Text("对方今天已经被打劫了，给人家留点后路吧"))
				return
			}
			if ok >= 2 {
				ctx.SendChain(message.Text("你今天已经成功打劫过了，贪心没有好果汁吃！"))
				return
			}

			// 穷人保护
			victimWallet := wallet.GetWalletOf(victimID)
			if victimWallet < 1000 {
				ctx.SendChain(message.Text("对方太穷了！打劫失败"))
				return
			}

			// 判断打劫是否成功
			if rand.Intn(100) > 60 {
				updateMoney := wallet.GetWalletOf(uid)
				if updateMoney >= 1000 {
					updateMoney = 1000
				}
				ctx.SendChain(message.Text("打劫失败,罚款1000"))
				err := wallet.InsertWalletOf(uid, -updateMoney)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:罚款失败，钱包坏掉力:\n", err))
					return
				}
				return
			}
			userIncrMoney := math.Min(rand.Intn(victimWallet/20)+500, 10000)
			victimDecrMoney := userIncrMoney / (rand.Intn(4) + 1)

			// 记录结果
			err = wallet.InsertWalletOf(victimID, -victimDecrMoney)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:钱包坏掉力:\n", err))
				return
			}
			err = wallet.InsertWalletOf(uid, +userIncrMoney)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:打劫失败，脏款掉入虚无\n", err))
				return
			}

			// 写入记录
			err = police.insertRecord(victimID, uid)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[ERROR]:犯罪记录写入失败\n", err))
			}

			ctx.SendChain(message.At(uid), message.Text("打劫成功，钱包增加：", userIncrMoney, wallet.GetWalletName()))
			ctx.SendChain(message.At(victimID), message.Text("保险公司对您进行了赔付，您实际损失：", victimDecrMoney, wallet.GetWalletName()))
		})
}

// ok==0 可以打劫；ok==1 程序错误 or 受害者进入CD；ok==2 用户进入CD; ok==3 用户和受害者都进入CD；
func (sql *robberyRepo) getRecord(victimID, uid int64) (ok int, err error) {
	sql.Lock()
	defer sql.Unlock()
	// 创建群表格
	err = sql.db.Create("criminal_record", &robberyRecord{})
	if err != nil {
		return 1, err
	}
	// 拼接查询SQL
	limitID := "WHERE victim_id = ? OR user_id = ?"
	if !sql.db.CanFind("criminal_record", limitID, victimID, uid) {
		// 没有记录即不用比较
		return 0, nil
	}
	cdInfo := robberyRecord{}

	err = sql.db.FindFor("criminal_record", &cdInfo, limitID, func() error {
		if time.Now().Format("2006/01/02") != cdInfo.Time {
			// // 如果跨天了就删除
			err = sql.db.Del("criminal_record", limitID, victimID, uid)
			return nil
		}
		// 俩个if是为了保证，重复打劫同一个人，ok == 3
		if cdInfo.UserID == uid {
			ok += 2
		}
		if cdInfo.VictimID == victimID {
			// lint 不允许使用 ok += 1
			ok++
		}
		return nil
	}, victimID, uid)
	return ok, err
}

func (sql *robberyRepo) insertRecord(vid int64, uid int64) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Insert("criminal_record", &robberyRecord{
		UserID:   uid,
		VictimID: vid,
		Time:     time.Now().Format("2006/01/02"),
	})
}
