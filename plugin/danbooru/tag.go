package deepdanbooru

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"net/url"
	"sort"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text" // jpg png gif
	_ "golang.org/x/image/webp"              // webp
)

const api = "https://nsfwtag.azurewebsites.net/api/tag?limit=0.5&url="

type sorttags struct {
	tags map[string]float64
	tseq []string
}

func newsorttags(tags map[string]float64) (s *sorttags) {
	t := make([]string, 0, len(tags))
	for k := range tags {
		t = append(t, k)
	}
	return &sorttags{tags: tags, tseq: t}
}

func (s *sorttags) Len() int {
	return len(s.tags)
}

func (s *sorttags) Less(i, j int) bool {
	v1 := s.tseq[i]
	v2 := s.tseq[j]
	return s.tags[v1] >= s.tags[v2]
}

// Swap swaps the elements with indexes i and j.
func (s *sorttags) Swap(i, j int) {
	s.tseq[j], s.tseq[i] = s.tseq[i], s.tseq[j]
}

func tagurl(name, u string) (im image.Image, st *sorttags, err error) {
	ch := make(chan []byte, 1)
	go func() {
		var data []byte
		data, err = web.GetData(u)
		ch <- data
	}()

	data, err := web.GetData(api + url.QueryEscape(u))
	if err != nil {
		return
	}
	if len(data) < 4 {
		err = errors.New("data too short")
		return
	}
	tags := make(map[string]float64)
	err = json.Unmarshal(data[1:len(data)-1], &tags)
	if err != nil {
		return
	}

	longestlen := 0
	for k := range tags {
		if len(k) > longestlen {
			longestlen = len(k)
		}
	}
	longestlen++

	st = newsorttags(tags)
	sort.Sort(st)

	boldfd, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	consfd, err := file.GetLazyData(text.ConsolasFontFile, control.Md5File, true)
	if err != nil {
		return
	}

	data = <-ch
	if err != nil {
		return
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}

	img = imgfactory.Limit(img, 1280, 720)

	canvas := gg.NewContext(img.Bounds().Size().X, img.Bounds().Size().Y+int(float64(img.Bounds().Size().X)*0.2)+len(tags)*img.Bounds().Size().X/25)
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()
	canvas.DrawImage(img, 0, 0)
	if err = canvas.ParseFontFace(boldfd, float64(img.Bounds().Size().X)*0.1); err != nil {
		return
	}
	canvas.SetRGB(0, 0, 0)
	canvas.DrawString(name, float64(img.Bounds().Size().X)*0.02, float64(img.Bounds().Size().Y)+float64(img.Bounds().Size().X)*0.1)
	i := float64(img.Bounds().Size().Y) + float64(img.Bounds().Size().X)*0.2
	if err = canvas.ParseFontFace(consfd, float64(img.Bounds().Size().X)*0.04); err != nil {
		return
	}
	rate := float64(img.Bounds().Size().X) * 0.04
	for _, k := range st.tseq {
		canvas.DrawString(fmt.Sprintf("* %-*s -%.3f-", longestlen, k, tags[k]), float64(img.Bounds().Size().X)*0.04, i)
		i += rate
	}
	im = canvas.Image()
	return
}
