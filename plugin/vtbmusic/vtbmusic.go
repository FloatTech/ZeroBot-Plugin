// Package vtbmusic vtb点歌
package vtbmusic

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	getGroupListURL = "https://aqua.chat/v1/GetGroupsList"
	getMusicListURL = "https://aqua.chat/v1/GetMusicList"
	fileURL         = "https://cdn.aqua.chat/"
	musicListBody   = `{"search":{"condition":"VocalId","keyword":"%v"},"sortField":"CreateTime","sortType":"desc","pageIndex":1,"pageRows":10000}`
)

type groupsList struct {
	Total int `json:"Total"`
	Data  []struct {
		ID         string `json:"Id"`
		CreateTime string `json:"CreateTime"`
		Name       string `json:"Name"`
		GroupImg   string `json:"GroupImg"`
		VocalList  []struct {
			ID          string `json:"Id"`
			CreateTime  string `json:"CreateTime"`
			ChineseName string `json:"ChineseName"`
			OriginName  string `json:"OriginName"`
			AvatarImg   string `json:"AvatarImg"`
		} `json:"VocalList"`
	} `json:"Data"`
	Success   bool        `json:"Success"`
	ErrorCode int         `json:"ErrorCode"`
	Msg       interface{} `json:"Msg"`
}

type musicList struct {
	Total int `json:"Total"`
	Data  []struct {
		ID              string      `json:"Id"`
		CreateTime      string      `json:"CreateTime"`
		PublishTime     interface{} `json:"PublishTime"`
		CreatorID       interface{} `json:"CreatorId"`
		CreatorRealName interface{} `json:"CreatorRealName"`
		Deleted         bool        `json:"Deleted"`
		OriginName      string      `json:"OriginName"`
		VocalID         string      `json:"VocalId"`
		VocalName       string      `json:"VocalName"`
		CoverImg        string      `json:"CoverImg"`
		Music           string      `json:"Music"`
		Lyric           interface{} `json:"Lyric"`
		CDN             string      `json:"CDN"`
		BiliBili        interface{} `json:"BiliBili"`
		YouTube         interface{} `json:"YouTube"`
		Twitter         interface{} `json:"Twitter"`
		Likes           interface{} `json:"Likes"`
		Length          float64     `json:"Length"`
		Label           interface{} `json:"Label"`
		IsLike          bool        `json:"isLike"`
		Duration        float64     `json:"Duration"`
		Source          interface{} `json:"Source"`
		SourceName      interface{} `json:"SourceName"`
		Statis          struct {
			PlayCount    int `json:"PlayCount"`
			CommentCount int `json:"CommentCount"`
			LikeCount    int `json:"LikeCount"`
			ShareCount   int `json:"ShareCount"`
		} `json:"Statis"`
		VocalList []struct {
			ID         string `json:"Id"`
			Cn         string `json:"cn"`
			Jp         string `json:"jp"`
			En         string `json:"en"`
			Originlang string `json:"originlang"`
		} `json:"VocalList"`
	} `json:"Data"`
	Success   bool        `json:"Success"`
	ErrorCode int         `json:"ErrorCode"`
	Msg       interface{} `json:"Msg"`
}

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "vtbmusic.com点歌",
		Help: "- vtb点歌\n" +
			"- vtb随机点歌",
		PrivateDataFolder: "vtbmusic",
	})
	storePath := engine.DataFolder()
	// 开启
	engine.OnFullMatch(`vtb点歌`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession(), zero.RegexRule(`^\d+$`))
			recv, cancel := next.Repeat()
			defer cancel()
			i := 0
			paras := [3]int{}
			data, err := web.PostData(getGroupListURL, "application/json", strings.NewReader(`{"PageIndex":1,"PageRows":9999}`))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			var (
				gl         groupsList
				ml         musicList
				num        int
				imageBytes []byte
			)
			err = json.Unmarshal(data, &gl)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			tex := "请输入群组序号\n"
			for i, v := range gl.Data {
				tex += fmt.Sprintf("%d. %s\n", i, v.Name)
			}
			imageBytes, err = text.RenderToBase64(tex, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+binary.BytesToString(imageBytes))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("vtb点歌超时"))
					return
				case c := <-recv:
					msg := c.Event.Message.ExtractPlainText()
					num, err = strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请输入数字!"))
						continue
					}
					switch i {
					case 0:
						if num < 0 || num >= len(gl.Data) {
							ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("序号非法!"))
							continue
						}
						if len(gl.Data[num].VocalList) == 0 {
							ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("无内容, 点歌失败"))
							return
						}
						paras[0] = num
						tex = "请输入vtb序号\n"
						for i, v := range gl.Data[paras[0]].VocalList {
							tex += fmt.Sprintf("%d. %s\n", i, v.OriginName)
						}
						imageBytes, err = text.RenderToBase64(tex, text.FontFile, 400, 20)
						if err != nil {
							ctx.SendChain(message.Text("ERROR: ", err))
							return
						}
						if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+binary.BytesToString(imageBytes))); id.ID() == 0 {
							ctx.SendChain(message.Text("ERROR: 可能被风控了"))
						}
					case 1:
						if num < 0 || num >= len(gl.Data[paras[0]].VocalList) {
							ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("序号非法!"))
							continue
						}
						paras[1] = num
						data, err := web.PostData(getMusicListURL, "application/json", strings.NewReader(fmt.Sprintf(musicListBody, gl.Data[paras[0]].VocalList[paras[1]].ID)))
						if err != nil {
							ctx.SendChain(message.Text("ERROR: ", err))
							return
						}
						err = json.Unmarshal(data, &ml)
						if err != nil {
							ctx.SendChain(message.Text("ERROR: ", err))
							return
						}
						if len(ml.Data) == 0 {
							ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("无内容, 点歌失败"))
							return
						}
						tex = "请输入歌曲序号\n"
						for i, v := range ml.Data {
							tex += fmt.Sprintf("%d. %s\n", i, v.OriginName)
						}
						imageBytes, err = text.RenderToBase64(tex, text.FontFile, 400, 20)
						if err != nil {
							ctx.SendChain(message.Text("ERROR: ", err))
							return
						}
						if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+binary.BytesToString(imageBytes))); id.ID() == 0 {
							ctx.SendChain(message.Text("ERROR: 可能被风控了"))
						}
					case 2:
						if num < 0 || num >= len(ml.Data) {
							ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("序号非法!"))
							continue
						}
						paras[2] = num
						// 最后播放歌曲
						groupName := gl.Data[paras[0]].Name
						vtbName := gl.Data[paras[0]].VocalList[paras[1]].OriginName
						musicName := ml.Data[paras[2]].OriginName
						recURL := fileURL + ml.Data[paras[2]].Music
						ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请欣赏", groupName, "-", vtbName, "的《", musicName, "》"))
						recordFile := storePath + fmt.Sprintf("%d-%d-%d", paras[0], paras[1], paras[2]) + path.Ext(recURL)
						if file.IsExist(recordFile) {
							ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
							return
						}
						err = dlrec(recordFile, recURL)
						if err != nil {
							ctx.SendChain(message.Text("ERROR: ", err))
							return
						}
						ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
						return
					}
					i++
				}
			}
		})
	engine.OnFullMatch(`vtb随机点歌`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				paras = [3]int{}
				gl    groupsList
				ml    musicList
			)
			data, err := web.PostData(getGroupListURL, "application/json", strings.NewReader(`{"PageIndex":1,"PageRows":9999}`))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = json.Unmarshal(data, &gl)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if len(gl.Data) == 0 {
				ctx.SendChain(message.Text("ERROR: 数组为空"))
				return
			}
			paras[0] = rand.Intn(len(gl.Data))
			for len(gl.Data[paras[0]].VocalList) == 0 {
				paras[0] = rand.Intn(len(gl.Data))
			}
			paras[1] = rand.Intn(len(gl.Data[paras[0]].VocalList))
			data, err = web.PostData(getMusicListURL, "application/json", strings.NewReader(fmt.Sprintf(musicListBody, gl.Data[paras[0]].VocalList[paras[1]].ID)))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = json.Unmarshal(data, &ml)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			for len(ml.Data) == 0 {
				paras[1] = rand.Intn(len(gl.Data[paras[0]].VocalList))
				data, err = web.PostData(getMusicListURL, "application/json", strings.NewReader(fmt.Sprintf(musicListBody, gl.Data[paras[0]].VocalList[paras[1]].ID)))
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				err = json.Unmarshal(data, &ml)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
			}
			paras[2] = rand.Intn(len(ml.Data))
			// 最后播放歌曲
			groupName := gl.Data[paras[0]].Name
			vtbName := gl.Data[paras[0]].VocalList[paras[1]].OriginName
			musicName := ml.Data[paras[2]].OriginName
			recURL := fileURL + ml.Data[paras[2]].Music
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请欣赏", groupName, "-", vtbName, "的《", musicName, "》"))
			recordFile := storePath + fmt.Sprintf("%d-%d-%d", paras[0], paras[1], paras[2]) + path.Ext(recURL)
			if file.IsExist(recordFile) {
				ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
				return
			}
			err = dlrec(recordFile, recURL)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
		})
}

func dlrec(recordFile, recordURL string) error {
	if file.IsNotExist(recordFile) {
		data, err := web.RequestDataWithHeaders(web.NewTLS12Client(), recordURL, "GET", func(r *http.Request) error {
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
			return nil
		}, nil)
		if err != nil {
			return err
		}
		return os.WriteFile(recordFile, data, 0666)
	}
	return nil
}
