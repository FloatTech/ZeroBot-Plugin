package dice

import (
	"math/rand"
	"strconv"

	"github.com/FloatTech/floatbox/math"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type intn interface {
	int | int8 | int16 | int32 | int64
}

func init() {
	engine.OnRegex(`^[。.][Rr][AaCc]\s*([0-9]{1,2})?#?\s*([^[0-9]|.*])\s*([0-9]{1,2})$`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			i := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			var msg message.Message
			if i == 0 {
				i = 1
			} else if i > 10 {
				i = 10
				msg = append(msg, message.Text("最多检定10次哦~\n"))
			}
			word := ctx.State["regex_matched"].([]string)[2]
			num := math.Str2Int64(ctx.State["regex_matched"].([]string)[3])
			msg = append(msg, message.Text(nickname, "进行", word, "检定:"))
			var r rsl
			err := db.Find("rsl", &r, "where gid = "+strconv.FormatInt(ctx.Event.GroupID, 10))
			var rule int64
			if err == nil {
				rule = r.Rule
			}
			for ; i > 0; i-- {
				rs := rand.Int63n(100) + 1
				win := rules(rs, num, rule)
				msg = append(msg, message.Text("\nD100=", rs, "/", num, " ", win))
			}
			ctx.SendChain(msg...)
		})
	engine.OnRegex(`^[。.]setcoc\s*([0-6]{1})`, zero.OnlyGroup).SetBlock(true).
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
	engine.OnRegex(`^[。.]set\s*([0-9]+)`, zero.OnlyGroup).SetBlock(true).
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
	engine.OnRegex(`^[。.][Rr]\s*([0]*[1-9]+)?\s*[Dd]\s*([0]*[1-9]+)?`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			r1 := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			d1 := math.Str2Int64(ctx.State["regex_matched"].([]string)[2])
			if r1 == 0 {
				r1 = 1
			}
			if d1 == 0 {
				var d set
				err := db.Find("set", &d, "where uid = "+strconv.FormatInt(ctx.Event.UserID, 10))
				if err == nil {
					d1 = d.D
				}
			}
			if r1 <= 100 {
				var sum, i int64
				var res message.Message
				for ; i < r1-1; i++ {
					rand := rand.Int63n(d1) + 1
					sum += rand
					res = append(res, message.Text("+", rand))
				}
				rand := rand.Int63n(d1) + 1
				res = append(res, message.Text(rand))
				ctx.SendChain(message.Text("阁下掷出了R", r1, "D", d1, "=", sum, "\n", res.String(), "=", sum))
			} else {
				ctx.SendChain(message.Text("骰子太多啦~~数不过来了！"))
			}
		})
}

func rules[T intn](r, num, rule T) string {
	win := ""
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
