package deepdanbooru

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"sort"
	"strings"
	"time"

	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	imgutils "github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/text" // jpg png gif
	"github.com/tidwall/gjson"
	_ "golang.org/x/image/webp" // webp
)

const (
	envURL               = "https://hf.space/embed/hysts/DeepDanbooru/api/queue"
	pushURL              = envURL + "/push/"
	statusURL            = envURL + "/status/"
	normalScoreThreshold = 0.5
	sessionHash          = "zerobot"
	predictAction        = "predict"
)

type hfRequest struct {
	Action      string        `json:"action"`
	FnIndex     int           `json:"fn_index"`
	Data        []interface{} `json:"data"`
	SessionHash string        `json:"session_hash"`
}

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
	data, err := web.GetData(u)
	if err != nil {
		return
	}
	hs, err := pushData(data)
	if err != nil {
		return
	}
	if hs == "" {
		err = errors.New("ERROR: 图片上传失败,返回的hash为空")
		return
	}
	tags, err := statusData(hs)
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

	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return
	}
	_, err = file.GetLazyData(text.ConsolasFontFile, true)
	if err != nil {
		return
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}

	img = imgutils.Limit(img, 1280, 720)

	canvas := gg.NewContext(img.Bounds().Size().X, img.Bounds().Size().Y+int(float64(img.Bounds().Size().X)*0.2)+len(tags)*img.Bounds().Size().X/25)
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()
	canvas.DrawImage(img, 0, 0)
	if err = canvas.LoadFontFace(text.BoldFontFile, float64(img.Bounds().Size().X)*0.1); err != nil {
		return
	}
	canvas.SetRGB(0, 0, 0)
	canvas.DrawString(name, float64(img.Bounds().Size().X)*0.02, float64(img.Bounds().Size().Y)+float64(img.Bounds().Size().X)*0.1)
	i := float64(img.Bounds().Size().Y) + float64(img.Bounds().Size().X)*0.2
	if err = canvas.LoadFontFace(text.ConsolasFontFile, float64(img.Bounds().Size().X)*0.04); err != nil {
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

func pushData(data []byte) (hash string, err error) {
	encodeStr := base64.StdEncoding.EncodeToString(data)
	encodeStr = "data:image/jpeg;base64," + encodeStr
	r := hfRequest{
		Action:      predictAction,
		FnIndex:     0,
		Data:        []interface{}{encodeStr, normalScoreThreshold},
		SessionHash: sessionHash,
	}
	b, err := json.Marshal(r)
	if err != nil {
		return
	}
	data, err = web.PostData(pushURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return
	}
	time.Sleep(1 * time.Second)
	hash = gjson.ParseBytes(data).Get("hash").String()
	return
}

func statusData(hash string) (tags map[string]float64, err error) {
	tags = make(map[string]float64)
	data, err := web.PostData(statusURL, "application/json", strings.NewReader(fmt.Sprintf(`{"hash": "%v"}`, hash)))
	if err != nil {
		return
	}
	gjson.ParseBytes(data).Get("data.data.0.confidences").ForEach(func(_, v gjson.Result) bool {
		tags[v.Get("label").String()] = v.Get("confidence").Float()
		return true
	})
	return
}
