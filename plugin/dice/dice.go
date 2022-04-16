package dice

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	rule int
	win  string
)

func init() {
	engine := control.Register("dice", &control.Options{
		DisableOnDefault:  true,
		Help:              "Dice! beta for zb ",
		PrivateDataFolder: "dice",
	})
	engine.OnRegex(`^.ra(\D+)(\d+)`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			list := []int{0, 1, 2, 3, 4, 5, 6}
			gid := ctx.Event.GroupID
			c, ok := control.Lookup("fortune")
			if ok {
				v := uint8(c.GetData(gid) & 0xff)
				if int(v) < 8 {
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
	engine.OnRegex(`^.ra(\d+)(\D+)(\d+)`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			list := []int{0, 1, 2, 3, 4, 5, 6}
			gid := ctx.Event.GroupID
			c, ok := control.Lookup("fortune")
			if ok {
				v := uint8(c.GetData(gid) & 0xff)
				if int(v) < 8 {
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
				r := rand.Intn(100) + 1
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
	engine.OnRegex(`^.set[0-6]`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			rule, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[1])
			gid := ctx.Event.GroupID
			c, ok := control.Lookup("fortune")
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
		})
}
