// Package hyaku 百人一首
package hyaku

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/FloatTech/floatbox/binary"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const bed = "https://gitea.seku.su/fumiama/OguraHyakuninIsshu/raw/branch/master/"

type line struct {
	番号, 歌人, 上の句, 下の句, 上の句ひらがな, 下の句ひらがな string
}

func (l *line) String() string {
	b := binary.NewWriterF(func(w *binary.Writer) {
		r := reflect.ValueOf(l).Elem().Type()
		for i := 0; i < r.NumField(); i++ {
			switch i {
			case 0:
				w.WriteString("●")
			case 1:
				w.WriteString("◉")
			case 2, 3:
				w.WriteString("○")
			case 4, 5:
				w.WriteString("◎")
			}
			w.WriteString(r.Field(i).Name)
			w.WriteString("：")
			w.WriteString((*[6]string)(unsafe.Pointer(l))[i])
			w.WriteString("\n")
		}
	})
	return binary.BytesToString(b)
}

var lines [100]*line

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "百人一首",
		Help: "- 百人一首(随机发一首)\n" +
			"- 百人一首之n",
		PrivateDataFolder: "hyaku",
	})
	err := os.MkdirAll(engine.DataFolder()+"img", 0755)
	if err != nil {
		panic(err)
	}
	data, err := engine.GetCustomLazyData(bed, "小倉百人一首.csv")
	if err != nil {
		panic(err)
	}
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		panic(err)
	}
	records = records[1:] // skip title
	if len(records) != 100 {
		panic("invalid csvfile")
	}
	for j, r := range records {
		if len(r) != 6 {
			panic("invalid csvfile")
		}
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
	engine.OnFullMatch("百人一首").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		i := rand.Intn(100)
		img0, err := engine.GetCustomLazyData(bed, fmt.Sprintf("img/%03d.jpg", i+1))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		img1, err := engine.GetCustomLazyData(bed, fmt.Sprintf("img/%03d.png", i+1))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(
			message.ImageBytes(img0),
			message.Text("\n", lines[i]),
			message.ImageBytes(img1),
		)
	})
	engine.OnRegex(`^百人一首之\s?(\d+)$`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		i, err := strconv.Atoi(ctx.State["regex_matched"].([]string)[1])
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if i > 100 || i < 1 {
			ctx.SendChain(message.Text("ERROR: 超出范围"))
			return
		}
		img0, err := engine.GetCustomLazyData(bed, fmt.Sprintf("img/%03d.jpg", i))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		img1, err := engine.GetCustomLazyData(bed, fmt.Sprintf("img/%03d.png", i))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(
			message.ImageBytes(img0),
			message.Text("\n", lines[i-1]),
			message.ImageBytes(img1),
		)
	})
}
