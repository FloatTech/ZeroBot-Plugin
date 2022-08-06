// Package hyaku 百人一首
package hyaku

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"unsafe"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const bed = "https://gitcode.net/u011570312/OguraHyakuninIsshu/-/raw/master/"

// nolint: asciicheck
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
	engine := control.Register("hyaku", &ctrl.Options[*zero.Ctx]{
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
		_ = f.Close()
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
	}()
	engine.OnFullMatch("百人一首").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		i := rand.Intn(100)
		ctx.SendChain(
			message.Image(fmt.Sprintf(bed+"img/%03d.jpg", i+1)),
			message.Text("\n", lines[i]),
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
		ctx.SendChain(
			message.Image(fmt.Sprintf(bed+"img/%03d.jpg", i)),
			message.Text("\n", lines[i-1]),
			message.Image(fmt.Sprintf(bed+"img/%03d.png", i)),
		)
	})
}
