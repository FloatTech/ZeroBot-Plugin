package dice

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/FloatTech/zbputils/math"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	rule int64
	win  string
)

func init() {
	engine.OnRegex(`^[。.][Rr][Aa|Cc]\s*([0-9]+)[#]\s*(\S\D+)\s*([0-9]+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			i := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			word := ctx.State["regex_matched"].([]string)[2]
			num := math.Str2Int64(ctx.State["regex_matched"].([]string)[3])
			msg := fmt.Sprintf("%s进行%s检定:", nickname, word)
			var r rsl
			err := db.Find("rsl", &r, "where gid = "+strconv.FormatInt(ctx.Event.GroupID, 10))
			if err == nil {
				rule = r.Rule
			} else {
				rule = 0
			}
			if i <= 10 {
				for i > 0 {
					i--
					rs := rand.Intn(100) + 1
					win = rules(int64(rs), num, rule)
					msg += fmt.Sprintf("\nD100=%d/%d %s", rs, num, win)
				}
				ctx.SendChain(message.Text(msg))
			}
			ctx.SendChain(message.Text("最多检定10次哦~"))
		})
	engine.OnRegex(`^[。.][Rr][Aa|Cc]\s*(\S\D+)\s*([0-9]+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			word := ctx.State["regex_matched"].([]string)[1]
			num := math.Str2Int64(ctx.State["regex_matched"].([]string)[2])
			rs := rand.Intn(100) + 1
			var r rsl
			err := db.Find("rsl", &r, "where gid = "+strconv.FormatInt(ctx.Event.GroupID, 10))
			if err == nil {
				rule = r.Rule
			} else {
				rule = 0
			}
			win = rules(int64(rs), num, rule)
			ctx.SendChain(message.Text(fmt.Sprintf("%s进行%s检定:\nD100=%d/%d %s", nickname, word, rs, num, win)))
		})
	engine.OnRegex(`^[.。]setcoc\s*([0-6]{1})`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			r := &rsl{
				GrpID: ctx.Event.GroupID,
				Rule:  math.Str2Int64(ctx.State["regex_matched"].([]string)[1]),
			}
			err := db.Insert("rsl", r)
			if err == nil {
				ctx.SendChain(message.Text("当前群聊房规设置为了", r.Rule))
			} else {
				ctx.SendChain(message.Text("出错啦: ", err))
			}
		})
	engine.OnRegex(`^[.。]set\s*([0-9]+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			dint := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			if dint > 1000 {
				dint = 1000
				d := &set{
					UserID: ctx.Event.UserID,
					D:      dint,
				}
				err := db.Insert("set", d)
				if err == nil {
					ctx.SendChain(message.Text("最多1000哟~已自动设为1000"))
				} else {
					ctx.SendChain(message.Text("出错啦: ", err))
				}
				return
			}
			d := &set{
				UserID: ctx.Event.UserID,
				D:      dint,
			}
			err := db.Insert("set", d)
			if err == nil {
				ctx.SendChain(message.Text("阁下默认骰子被设定为了", d.D))
			} else {
				ctx.SendChain(message.Text("出错啦: ", err))
			}
		})
	engine.OnRegex(`^[。.][Rr][Dd]`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var r1, d1 int64
			r1 = 1
			var d set
			err := db.Find("set", &d, "where uid = "+strconv.FormatInt(ctx.Event.UserID, 10))
			if err == nil {
				d1 = d.D
			} else {
				d1 = 100
			}
			sum := rand.Intn(int(d1)) + 1
			ctx.SendChain(message.Text(fmt.Sprintf("阁下掷出了R%dD%d=%d", r1, d1, sum)))
		})
	engine.OnRegex(`^[。.][Rr]\s*([0-9]+).*?[Dd].*?([0-9]+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			r := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			d := math.Str2Int64(ctx.State["regex_matched"].([]string)[2])
			if r < 1 && d <= 1 {
				ctx.SendChain(message.Text("阁下..你在让我骰什么啊？( ´_ゝ`)"))
				return
			}
			if r <= 100 && d <= 100 {
				var sum, i int64
				res := ""
				for ; i < r; i++ {
					rand := rand.Intn(int(d)) + 1
					sum += int64(rand)
					if i == r-1 {
						res += fmt.Sprintf("%d", rand)
					} else {
						res += fmt.Sprintf("%d+", rand)
					}
				}
				ctx.SendChain(message.Text(fmt.Sprintf("阁下掷出了R%dD%d=%d\n%s=%d", r, d, sum, res, sum)))
			} else {
				ctx.SendChain(message.Text("骰子太多啦~~数不过来了！"))
			}
		})
}

func rules(r, num, rule int64) string {
	switch rule {
	case 0:
		switch {
		case r == 1:
			win = "大成功"
		case num < 50 && r <= 100 && r >= 96 || num >= 50 && r == 100:
			win = "大失败"
		case r < num/5:
			win = "极难成功"
		case r < num/2:
			win = "困难成功"
		case r < num:
			win = "成功"
		default:
			win = "失败"
		}
	case 1:
		switch {
		case num < 50 && r == 1 || num >= 50 && r >= 1 && r <= 5:
			win = "大成功"
		case num < 50 && r < 100 && r > 96 || num >= 50 && r == 100:
			win = "大失败"
		case r < num/5:
			win = "极难成功"
		case r < num/2:
			win = "困难成功"
		case r < num:
			win = "成功"
		default:
			win = "失败"
		}
	case 2:
		switch {
		case r >= 1 && r <= 5 && r <= num:
			win = "大成功"
		case r >= 96 && r <= 100 && r > num:
			win = "大失败"
		case r < num/5:
			win = "极难成功"
		case r < num/2:
			win = "困难成功"
		case r < num:
			win = "成功"
		default:
			win = "失败"
		}
	case 3:
		switch {
		case r >= 1 && r <= 5:
			win = "大成功"
		case r >= 96 && r <= 100:
			win = "大失败"
		case r < num/5:
			win = "极难成功"
		case r < num/2:
			win = "困难成功"
		case r < num:
			win = "成功"
		default:
			win = "失败"
		}
	case 4:
		switch {
		case r >= 1 && r <= 5 && r <= num/10:
			win = "大成功"
		case num < 50 && r >= 96+num/10 || num >= 50 && r == 100:
			win = "大失败"
		case r < num/5:
			win = "极难成功"
		case r < num/2:
			win = "困难成功"
		case r < num:
			win = "成功"
		default:
			win = "失败"
		}
	case 5:
		switch {
		case r >= 1 && r <= 2 && r <= num/5:
			win = "大成功"
		case num < 50 && r >= 96 && r <= 100 || num >= 50 && r >= 99 && r <= 100:
			win = "大失败"
		case r < num/5:
			win = "极难成功"
		case r < num/2:
			win = "困难成功"
		case r < num:
			win = "成功"
		default:
			win = "失败"
		}
	case 6:
		switch {
		case r == 1 && r <= num || r%11 == 0 && r <= num:
			win = "大成功"
		case r == 100 && r > num || r%11 == 0 && r > num:
			win = "大失败"
		case r < num/5:
			win = "极难成功"
		case r < num/2:
			win = "困难成功"
		case r < num:
			win = "成功"
		default:
			win = "失败"
		}
	}
	return win
}
