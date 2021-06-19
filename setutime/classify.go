package setutime

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/tidwall/gjson"
)

var (
	CACHE_IMG_FILE = "/tmp/setugt"
	CACHE_URI      = "file:///" + CACHE_IMG_FILE
	VOTE_API_URL   = "http://saki.fumiama.top/vote?uuid=零号&img=%s&class=%d"
	CLASSIFY_HEAD  = "http://saki.fumiama.top:62002/dice?url="
	MsgofGrp       = make(map[int64]int64)
	dhashofmsg     = make(map[int64]string)
)

func Classify(ctx *zero.Ctx, targeturl string, noimg bool) {
	get_url := CLASSIFY_HEAD + url.QueryEscape(targeturl)
	if noimg {
		get_url += "&noimg=true"
	}
	resp, err := http.Get(get_url)
	if err != nil {
		ctx.Send(fmt.Sprintf("ERROR: %v", err))
	} else {
		if noimg {
			data, err1 := ioutil.ReadAll(resp.Body)
			if err1 == nil {
				dhash := gjson.GetBytes(data, "img").String()
				class := int(gjson.GetBytes(data, "class").Int())
				replyClass(ctx, dhash, class, noimg)
			} else {
				ctx.Send(fmt.Sprintf("ERROR: %v", err1))
			}
		} else {
			class, err1 := strconv.Atoi(resp.Header.Get("Class"))
			dhash := resp.Header.Get("DHash")
			if err1 != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err1))
			}
			defer resp.Body.Close()
			// 写入文件
			data, _ := ioutil.ReadAll(resp.Body)
			f, _ := os.OpenFile(CACHE_IMG_FILE, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
			defer f.Close()
			f.Write(data)
			replyClass(ctx, dhash, class, noimg)
		}
	}
}

func Vote(ctx *zero.Ctx, class int) {
	msg, ok := MsgofGrp[ctx.Event.GroupID]
	if ok {
		ctx.DeleteMessage(msg)
		delete(MsgofGrp, ctx.Event.GroupID)
		dhash, ok2 := dhashofmsg[msg]
		if ok2 {
			http.Get(fmt.Sprintf(VOTE_API_URL, dhash, class))
			delete(dhashofmsg, msg)
		}
	}
}

func replyClass(ctx *zero.Ctx, dhash string, class int, noimg bool) {
	if class > 4 {
		switch class {
		case 5:
			ctx.Send("[5]影响不好啦！")
		case 6:
			ctx.Send("[6]太涩啦，🐛了！")
		}
		if dhash != "" && !noimg {
			b14, err3 := url.QueryUnescape(dhash)
			if err3 == nil {
				ctx.Send("给你点提示哦：" + b14)
			}
		}
	} else {
		var last_message_id int64
		if !noimg {
			last_message_id = ctx.SendChain(message.Image(CACHE_URI).Add("no_cache", "0"))
			last_group_id := ctx.Event.GroupID
			MsgofGrp[last_group_id] = last_message_id
			dhashofmsg[last_message_id] = dhash
		} else {
			last_message_id = ctx.Event.MessageID
		}
		switch class {
		case 0:
			ctx.SendChain(message.Reply(last_message_id), message.Text("[0]一堆像素点"))
		case 1:
			ctx.SendChain(message.Reply(last_message_id), message.Text("[1]普通"))
		case 2:
			ctx.SendChain(message.Reply(last_message_id), message.Text("[2]有点意思"))
		case 3:
			ctx.SendChain(message.Reply(last_message_id), message.Text("[3]不错"))
		case 4:
			ctx.SendChain(message.Reply(last_message_id), message.Text("[4]我好啦！"))
		}
	}
}
