package dice

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	rule  int
	win   string
	list  = [...]int{0, 1, 2, 3, 4, 5, 6}
	index = make(map[int]uint8)
)

func init() {
	engine := control.Register("dice", &control.Options{
		DisableOnDefault: true,
		Help:             "Dice! beta for zb ",
		PublicDataFolder: "Dice",
	})
	go func() {
		for i, s := range list {
			index[s] = uint8(i)
		}
	}()
	now := time.Now().Format("20060102")
	var signTF map[string](int)
	signTF = make(map[string](int))
	var result map[int64](int)
	result = make(map[int64](int))
	engine.OnFullMatch("今日人品").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			user := ctx.Event.UserID
			userS := strconv.FormatInt(user, 10)
			var si string = now + userS
			rand.Seed(time.Now().UnixNano())
			today := rand.Intn(100)
			if signTF[si] == 0 {
				signTF[si] = (1)
				result[user] = (today)
				ctx.SendChain(message.At(user), message.Text(" 阁下今日的人品值为", result[user], "呢~\n"), message.Image("https://img.qwq.nz/images/2022/04/04/aab2985d94e996558b303be42a954a4f.jpg"))
			} else {
				ctx.SendChain(message.At(user), message.Text(" 阁下今日的人品值为", result[user], "呢~\n"), message.Image("https://img.qwq.nz/images/2022/04/04/aab2985d94e996558b303be42a954a4f.jpg"))
			}
		})
	engine.OnRegex(`^.ra(\D+)(\d+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			c, ok := control.Lookup("dice")
			if ok {
				v := uint8(c.GetData(gid) & 0xff)
				if int(v) < len(list) {
					rule = list[v]
				}
			}
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			temp := ctx.State["regex_matched"].([]string)[1]
			math, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[2])
			r := rand.Intn(99) + 1
			switch rule {
			case 0:
				switch {
				case r == 1:
					win = "大成功"
				case math < 50 && r <= 100 && r >= 96 || math >= 50 && r == 100:
					win = "大失败"
				case r < math/5:
					win = "极难成功"
				case r < math/2:
					win = "困难成功"
				case r < math:
					win = "成功"
				default:
					win = "失败"
				}
			case 1:
				switch {
				case math < 50 && r == 1 || math >= 50 && r >= 1 && r <= 5:
					win = "大成功"
				case math < 50 && r < 100 && r > 96 || math >= 50 && r == 100:
					win = "大失败"
				case r < math/5:
					win = "极难成功"
				case r < math/2:
					win = "困难成功"
				case r < math:
					win = "成功"
				default:
					win = "失败"
				}
			case 2:
				switch {
				case r >= 1 && r <= 5 && r <= math:
					win = "大成功"
				case r >= 96 && r <= 100 && r > math:
					win = "大失败"
				case r < math/5:
					win = "极难成功"
				case r < math/2:
					win = "困难成功"
				case r < math:
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
				case r < math/5:
					win = "极难成功"
				case r < math/2:
					win = "困难成功"
				case r < math:
					win = "成功"
				default:
					win = "失败"
				}
			case 4:
				switch {
				case r >= 1 && r <= 5 && r <= math/10:
					win = "大成功"
				case math < 50 && r >= 96+math/10 || math >= 50 && r == 100:
					win = "大失败"
				case r < math/5:
					win = "极难成功"
				case r < math/2:
					win = "困难成功"
				case r < math:
					win = "成功"
				default:
					win = "失败"
				}
			case 5:
				switch {
				case r >= 1 && r <= 2 && r <= math/5:
					win = "大成功"
				case math < 50 && r >= 96 && r <= 100 || math >= 50 && r >= 99 && r <= 100:
					win = "大失败"
				case r < math/5:
					win = "极难成功"
				case r < math/2:
					win = "困难成功"
				case r < math:
					win = "成功"
				default:
					win = "失败"
				}
			case 6:
				switch {
				case r == 1 && r <= math || r%11 == 0 && r <= math:
					win = "大成功"
				case r == 100 && r > math || r%11 == 0 && r > math:
					win = "大失败"
				case r < math/5:
					win = "极难成功"
				case r < math/2:
					win = "困难成功"
				case r < math:
					win = "成功"
				default:
					win = "失败"
				}
			}
			msg := fmt.Sprintf("%s进行%s检定:\nD100=%d/%d %s", nickname, temp, r, math, win)
			ctx.Send(msg)
		})
	engine.OnRegex(`^.ra(\d+)(\D+)(\d+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			c, ok := control.Lookup("dice")
			if ok {
				v := uint8(c.GetData(gid) & 0xff)
				if int(v) < len(list) {
					rule = list[v]
				}
			}
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			i, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[1])
			temp := ctx.State["regex_matched"].([]string)[2]
			math, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[3])
			msg := fmt.Sprintf("%s进行%s检定:", nickname, temp)
			for i > 0 && i < 30 {
				i--
				r := rand.Intn(99) + 1
				switch rule {
				case 0:
					switch {
					case r == 1:
						win = "大成功"
					case math < 50 && r <= 100 && r >= 96 || math >= 50 && r == 100:
						win = "大失败"
					case r < math/5:
						win = "极难成功"
					case r < math/2:
						win = "困难成功"
					case r < math:
						win = "成功"
					default:
						win = "失败"
					}
				case 1:
					switch {
					case math < 50 && r == 1 || math >= 50 && r >= 1 && r <= 5:
						win = "大成功"
					case math < 50 && r < 100 && r > 96 || math >= 50 && r == 100:
						win = "大失败"
					case r < math/5:
						win = "极难成功"
					case r < math/2:
						win = "困难成功"
					case r < math:
						win = "成功"
					default:
						win = "失败"
					}
				case 2:
					switch {
					case r >= 1 && r <= 5 && r <= math:
						win = "大成功"
					case r >= 96 && r <= 100 && r > math:
						win = "大失败"
					case r < math/5:
						win = "极难成功"
					case r < math/2:
						win = "困难成功"
					case r < math:
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
					case r < math/5:
						win = "极难成功"
					case r < math/2:
						win = "困难成功"
					case r < math:
						win = "成功"
					default:
						win = "失败"
					}
				case 4:
					switch {
					case r >= 1 && r <= 5 && r <= math/10:
						win = "大成功"
					case math < 50 && r >= 96+math/10 || math >= 50 && r == 100:
						win = "大失败"
					case r < math/5:
						win = "极难成功"
					case r < math/2:
						win = "困难成功"
					case r < math:
						win = "成功"
					default:
						win = "失败"
					}
				case 5:
					switch {
					case r >= 1 && r <= 2 && r <= math/5:
						win = "大成功"
					case math < 50 && r >= 96 && r <= 100 || math >= 50 && r >= 99 && r <= 100:
						win = "大失败"
					case r < math/5:
						win = "极难成功"
					case r < math/2:
						win = "困难成功"
					case r < math:
						win = "成功"
					default:
						win = "失败"
					}
				case 6:
					switch {
					case r == 1 && r <= math || r%11 == 0 && r <= math:
						win = "大成功"
					case r == 100 && r > math || r%11 == 0 && r > math:
						win = "大失败"
					case r < math/5:
						win = "极难成功"
					case r < math/2:
						win = "困难成功"
					case r < math:
						win = "成功"
					default:
						win = "失败"
					}
				}
				msg += fmt.Sprintf("\nD100=%d/%d %s", r, math, win)
			}
			ctx.Send(msg)
		})
	engine.OnRegex(`^.setcoc(\d+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			atoi, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[1])
			gid := ctx.Event.GroupID
			rule, ok := index[atoi]
			if ok {
				c, ok := control.Lookup("dice")
				if ok {
					err := c.SetData(gid, int64(rule)&0xff)
					if err != nil {
						ctx.SendChain(message.Text("设置失败:", err))
						return
					}
					ctx.SendChain(message.Text("默认检定房规已设置:", rule))
					return
				}
				ctx.SendChain(message.Text("设置失败: 找不到插件"))
				return
			}
			ctx.SendChain(message.Text("没有这个规则哦～"))
		})
	engine.OnRegex("^.[rR](.*)[dD](.*)", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			r1 := ctx.State["regex_matched"].([]string)[1]
			d1 := ctx.State["regex_matched"].([]string)[2]
			if r1 == "" {
				r1 = "1"
			}
			if d1 == "" {
				d1 = "100"
			}
			r, _ := strconv.Atoi(r1)
			d, _ := strconv.Atoi(d1)
			if r < 1 || d <= 1 {
				ctx.SendChain(message.Text("阁下..你在让我骰什么啊？( ´_ゝ`)"))
				return
			}
			if r <= 100 && d <= 100 {
				sum := 0
				res := fmt.Sprintf("")
				for i := 0; i < r; i++ {
					rand := rand.Intn(d-1) + 1
					sum += rand
					if i == r-1 {
						res += fmt.Sprintf("%d", rand)
					} else {
						res += fmt.Sprintf("%d+", rand)
					}
				}
				msg := fmt.Sprintf("阁下掷出了R%dD%d=%d\n%s=%d", r, d, sum, res, sum)
				ctx.Send(msg)
			} else {
				ctx.SendChain(message.Text("骰子太多啦~~数不过来了！"))
			}
		})
}
