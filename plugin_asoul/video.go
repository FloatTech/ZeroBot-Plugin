package asoul

import (
	"encoding/json"
	"fmt"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func init() {
	engine.OnMessage().
		Handle(func(ctx *zero.Ctx) {
			var data *vdInfo
			switch ctx.Event.Message.String() {
			case "来点然能量":
				data = video(strconv.Itoa(diana))
			case "来点晚能量":
				data = video(strconv.Itoa(ava))
			case "来点牛能量":
				data = video(strconv.Itoa(kira))
			case "来点乃能量":
				data = video(strconv.Itoa(queen))
			case "来点狼能量":
				data = video(strconv.Itoa(carol))
			}
			rand.Seed(time.Now().UnixNano())
			ranNub := rand.Intn(50)
			ctx.SendChain(message.Image(data.Data.List.Vlist[ranNub].Pic))
			ctx.SendChain(message.CustomMusic(
				"https://bilibili.com/video/"+data.Data.List.Vlist[ranNub].Bvid,
				"11111112355",
				data.Data.List.Vlist[ranNub].Title,
			))
		})
}

func video(uid string) *vdInfo {
	url := "https://api.bilibili.com/x/space/arc/search?&ps=50&pn=1&order=pubdate&mid=" + uid
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	result := &vdInfo{}
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}
