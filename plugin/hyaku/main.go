package hyaku

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"unsafe"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const bed = "https://gitcode.net/u011570312/OguraHyakuninIsshu/-/raw/master/"

type line struct {
	no, 歌人, 上の句, 下の句, 上の句ひらがな, 下の句ひらがな string
}

var lines [100]*line

func init() {
	engine := control.Register("hyaku", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "百人一首\n" +
			"- 百人一首(随机发一首)\n" +
			"- 百人一首之n",
		PrivateDataFolder: "hyaku",
	})
	csvfile := engine.DataFolder() + "hyaku.csv"
	go func() {
		if file.IsNotExist(csvfile) {
			err := file.DownloadTo(bed+"小倉百人一首.csv", csvfile, true)
			if err != nil {
				_ = os.Remove(csvfile)
				panic(err)
			}
		}
		f, err := os.Open(csvfile)
		if err != nil {
			panic(err)
		}
		records, err := csv.NewReader(f).ReadAll()
		if err != nil {
			panic(err)
		}
		records = records[1:] // skip title
		if len(records) != 100 {
			panic("invalid csvfile")
		}
		for j, r := range records {
			i, err := strconv.Atoi(r[0])
			if err != nil {
				panic(err)
			}
			i--
			if j != i {
				panic("invalid csvfile")
			}
			lines[i] = (*line)(*(*unsafe.Pointer)(unsafe.Pointer(&r)))
		}
	}()
	engine.OnFullMatch("百人一首").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		i := rand.Intn(100)
		ctx.SendChain(
			message.Image(fmt.Sprintf(bed+"img/%03d.jpg", i+1)),
			message.Text("\n",
				"●番　号: ", lines[i].no, "\n",
				"●歌　人: ", lines[i].歌人, "\n",
				"●上の句: ", lines[i].上の句, "\n",
				"●下の句: ", lines[i].下の句, "\n",
				"●上の句ひらがな: ", lines[i].上の句ひらがな, "\n",
				"●下の句ひらがな: ", lines[i].下の句ひらがな, "\n",
			),
			message.Image(fmt.Sprintf(bed+"img/%03d.png", i+1)),
		)
	})
	engine.OnRegex(`^百人一首之\s?(\d+)$`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		i, err := strconv.Atoi(ctx.State["regex_matched"].([]string)[1])
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		if i > 100 || i < 1 {
			ctx.SendChain(message.Text("ERROR:超出范围"))
			return
		}
		i--
		ctx.SendChain(
			message.Image(fmt.Sprintf(bed+"img/%03d.jpg", i+1)),
			message.Text("\n",
				"●番　号: ", lines[i].no, "\n",
				"●歌　人: ", lines[i].歌人, "\n",
				"●上の句: ", lines[i].上の句, "\n",
				"●下の句: ", lines[i].下の句, "\n",
				"●上の句ひらがな: ", lines[i].上の句ひらがな, "\n",
				"●下の句ひらがな: ", lines[i].下の句ひらがな, "\n",
			),
			message.Image(fmt.Sprintf(bed+"img/%03d.png", i+1)),
		)
	})
}
