package asoul

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func init() {
	engine.OnKeyword("来点然能量").
		Handle(func(ctx *zero.Ctx) {
			data := video(strconv.Itoa(diana))
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

func init() {
	engine.OnKeyword("来点晚能量").
		Handle(func(ctx *zero.Ctx) {
			data := video(strconv.Itoa(ava))
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

func init() {
	engine.OnKeyword("来点牛能量").
		Handle(func(ctx *zero.Ctx) {
			data := video(strconv.Itoa(kira))
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

func init() {
	engine.OnKeyword("来点乃能量").
		Handle(func(ctx *zero.Ctx) {
			data := video(strconv.Itoa(queen))
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

func init() {
	engine.OnKeyword("来点狼能量").
		Handle(func(ctx *zero.Ctx) {
			data := video(strconv.Itoa(carol))
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
		log.Errorln("[video]", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Errorln("[video]", err)
	}
	defer res.Body.Close()
	result := &vdInfo{}
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}
