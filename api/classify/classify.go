package classify

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/Yiwen-Chan/ZeroBot-Plugin/api/msgext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	BLOCK_REQUEST_CLASS = false
	CACHE_IMG_FILE      = "/tmp/setugt"
	CACHE_URI           = "file:///" + CACHE_IMG_FILE
	VOTE_API_URL        = "http://saki.fumiama.top/vote?uuid=零号&img=%s&class=%d"
	CLASSIFY_HEAD       = "http://saki.fumiama.top:62002/dice?url="
	msgofgrp            = make(map[int64]int64)
	dhashofmsg          = make(map[int64]string)
)

func Classify(ctx *zero.Ctx, targeturl string, noimg bool) {
	if BLOCK_REQUEST_CLASS {
		ctx.Send("请稍后再试哦")
	} else {
		BLOCK_REQUEST_CLASS = true
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
		BLOCK_REQUEST_CLASS = false
	}
}

func Vote(ctx *zero.Ctx, class int) {
	msg, ok := msgofgrp[ctx.Event.GroupID]
	if ok {
		ctx.DeleteMessage(msg)
		delete(msgofgrp, ctx.Event.GroupID)
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
		if dhash != "" {
			b14, err3 := url.QueryUnescape(dhash)
			if err3 == nil {
				ctx.Send("给你点提示哦：" + b14)
			}
		}
	} else {
		if !noimg {
			last_message_id := ctx.Send(msgext.ImageNoCache(CACHE_URI))
			last_group_id := ctx.Event.GroupID
			msgofgrp[last_group_id] = last_message_id
			dhashofmsg[last_message_id] = dhash
		}
		switch class {
		case 0:
			ctx.Send("[0]一堆像素点")
		case 1:
			ctx.Send("[1]普通")
		case 2:
			ctx.Send("[2]有点意思")
		case 3:
			ctx.Send("[3]不错")
		case 4:
			ctx.Send("[4]我好啦！")
		}
	}
}
