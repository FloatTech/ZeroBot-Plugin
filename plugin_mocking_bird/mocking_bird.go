// Package mocking_bird 拟声鸟
package mocking_bird

import (
	"bytes"
	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_qingyunke"
	fileutil "github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/web"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	prio            = 250
	dbpath          = "data/MockingBird/"
	cachePath       = dbpath + "cache/"
	dbfile          = dbpath + "降噪3.wav"
	baseURL         = "http://aaquatri.com/sound/"
	synthesizersURL = baseURL + "api/synthesizers/"
	synthesizeURL   = baseURL + "api/synthesize"
)

var (
	engine = control.Register("mocking_bird", &control.Options{
		DisableOnDefault: false,
		Help:             "拟声鸟\n- @Bot 任意文本(任意一句话回复)",
	})
	vocoderList = []string{"WaveRNN", "HifiGAN"}
)

func init() {
	engine.OnMessage(zero.OnlyToMe).SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			// 调用青云客接口
			reply, err := qingyunke.GetMessage(msg)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 挑出 face 表情
			textReply, _ := qingyunke.DealReply(reply)
			// 拟声器生成音频
			syntPath := getSyntPath()
			fileName := getWav(textReply, syntPath, vocoderList[1], ctx.Event.UserID)
			// 回复
			ctx.SendChain(message.Record("file:///" + fileutil.BOTPATH + "/" + cachePath + fileName))
		})
}

func getSyntPath() (syntPath string) {
	data, err := web.ReqWith(synthesizersURL, "GET", "", "")
	if err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	syntPath = gjson.Get(helper.BytesToString(data), "0.path").String()
	return
}

func getWav(text, syntPath, vocoder string, uid int64) (fileName string) {
	fileName = strconv.FormatInt(uid, 10) + time.Now().Format("20060102150405") + ".wav"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your file
	f, err := os.Open(dbfile)
	if err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	defer f.Close()
	fw, err := w.CreateFormFile("file", dbfile)
	if err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	if _, err = io.Copy(fw, f); err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	if fw, err = w.CreateFormField("text"); err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	if _, err = fw.Write([]byte(text)); err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	if fw, err = w.CreateFormField("synt_path"); err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	if _, err = fw.Write([]byte(syntPath)); err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	if fw, err = w.CreateFormField("vocoder"); err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	if _, err = fw.Write([]byte(vocoder)); err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	w.Close()
	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", synthesizeURL, &b)
	if err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Errorln("[mocking_bird]:", err)
	}
	// Check the response
	if res.StatusCode != http.StatusOK {
		log.Errorf("bad status: %s", res.Status)
	}
	data, _ := ioutil.ReadAll(res.Body)
	ioutil.WriteFile(cachePath+fileName, data, 0666)
	return
}
