package dice

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

var (
	rule  int
	win   string
)

func init() {
	engine := control.Register("dice", &control.Options{
		DisableOnDefault: true,
		Help:             "试图移植的dice\n-.jrrp\n-.ra\n-.rd",
		PublicDataFolder: "Dice",
	})
	engine.OnFullMatchGroup([]string{".jrrp", "。jrrp"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now()
			uid := ctx.Event.UserID
			seed := md5.Sum(helper.StringToBytes(fmt.Sprintf("%d%d%d%d", uid, now.Year(), now.Month(), now.Day())))
			r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed[:]))))
			jrrp := r.Intn(100)+1
			ctx.SendChain(message.At(uid), message.Text(" 阁下今日的人品值为", jrrp, "呢~"))
		})
	engine.OnRegex(`^[。.][Rr][Aa|Cc].*?(\D+).*?([0-9]+).*?`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			temp := ctx.State["regex_matched"].([]string)[1]
			math, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[2])
			r := rand.Intn(100) + 1
			win = rules(r, math)
			msg := fmt.Sprintf("%s进行%s检定:\nD100=%d/%d %s", nickname, temp, r, math, win)
			ctx.Send(msg)
		})
	engine.OnRegex(`^[。.][Rr][Aa|Cc].*?([0-9]+)#.*?(\D+).*?([0-9]+).*?`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			i, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[1])
			temp := ctx.State["regex_matched"].([]string)[2]
			math, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[3])
			msg := fmt.Sprintf("%s进行%s检定:", nickname, temp)
			if i <= 10 {
				for i > 0 {
					i--
					r := rand.Intn(100) + 1
					win = rules(r, math)
					msg += fmt.Sprintf("\nD100=%d/%d %s", r, math, win)
				}
				ctx.Send(msg)
			} else {
				ctx.SendChain(message.Text("最多检定10次哟~"))
			}
		})
	engine.OnRegex(`[.。]setcoc(\d+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 这里应该是设置规则的，但是咕咕咕
		})
	engine.OnRegex(`^[。.][Rr].*?([0-9]+).*?[Dd].*?([0-9]+).*?`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var r1,d1 string
			if r1 = ctx.State["regex_matched"].([]string)[1] ; r1 == "" {
				r1 = "1"
			}
			if d1 = ctx.State["regex_matched"].([]string)[2] ; d1 == "" {
				d1 = "100"
			}
			r, _ := strconv.Atoi(r1)
			d, _ := strconv.Atoi(d1)
			if r < 1 && d <= 1 {
				ctx.SendChain(message.Text("阁下..你在让我骰什么啊？( ´_ゝ`)"))
				return
			}
			if r <= 100 && d <= 100 {
				sum := 0
				res := fmt.Sprintf("")
				for i := 0; i < r; i++ {
					rand := rand.Intn(d) + 1
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

func rules(r, math int) (win string) {
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
	return win
}
