package setutime

import (
	"encoding/json"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"net/http"
)

type ServerResult struct {
	Error string `json:"error"`
	Data  []struct {
		Pid        int      `json:"pid"`
		P          int      `json:"p"`
		UID        int      `json:"uid"`
		Title      string   `json:"title"`
		Author     string   `json:"author"`
		R18        bool     `json:"r18"`
		Width      int      `json:"width"`
		Height     int      `json:"height"`
		Tags       []string `json:"tags"`
		Ext        string   `json:"ext"`
		UploadDate int64    `json:"uploadDate"`
		Urls       struct {
			Original string `json:"original"`
		} `json:"urls"`
	} `json:"data"`
}

func init() {
	zero.OnFullMatch("来张萝莉", zero.AdminPermission).
		Handle(func(ctx *zero.Ctx) {
			r18json := api()
			ctx.SendChain(message.Text(
				"pid:  ", r18json.Data[0].Pid, "\n",
				"title:  ", r18json.Data[0].Title, "\n",
				"author:  ", r18json.Data[0].Author, "\n",
				"r18:  ", r18json.Data[0].R18, "\n",
				"tags:  ", r18json.Data[0].Tags, "\n",
				"url:  ", r18json.Data[0].Urls.Original, "\n",
			),
			message.Image(r18json.Data[0].Urls.Original),
			)
		})
}

// !!!在群里慎用有封号风险：r18图太涩了号还想要的话最好别开放权限
func init() {
	zero.OnFullMatch("!来张萝莉r18", zero.SuperUserPermission).
		Handle(func(ctx *zero.Ctx) {
			r18json := r18api()
			ctx.SendChain(message.Text(
				"pid:  ", r18json.Data[0].Pid, "\n",
				"title:  ", r18json.Data[0].Title, "\n",
				"author:  ", r18json.Data[0].Author, "\n",
				"r18:  ", r18json.Data[0].R18, "\n",
				"tags:  ", r18json.Data[0].Tags, "\n",
				"url:  ", r18json.Data[0].Urls.Original, "\n",
			),
			message.Image(r18json.Data[0].Urls.Original),
			)
	})
}

// 发起api请求非r18
func api() *ServerResult {
	resp, err := http.Get("https://api.lolicon.app/setu/v2")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	result := &ServerResult{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}

// 发起api请求url带上了r18参数
func r18api() *ServerResult {
	resp, err := http.Get("https://api.lolicon.app/setu/v2?r18=1")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	result := &ServerResult{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}
