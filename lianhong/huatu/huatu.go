// Package huatu 百度文心AI画图
package huatu

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/web"
)

type tokendata struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

// GetToken 获取当天的token
//
// 申请账号链接:https://wenxin.baidu.com/moduleApi/key
//
// clientID为API key,clientSecret为Secret key
//
// token有效时间为24小时
func GetToken(clientID, clientSecret string) (token string, err error) {
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/oauth/token?grant_type=client_credentials&client_id=" + url.QueryEscape(clientID) + "&client_secret=" + url.QueryEscape(clientSecret)
	postData := url.Values{}
	postData.Add("name", "雪儿")
	postData.Add("language", "golang")
	data, err := web.PostData(requestURL, "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
	if err != nil {
		return
	}
	var parsed tokendata
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	if parsed.Msg != "success" {
		err = errors.New(parsed.Msg + ",code:" + strconv.Itoa(parsed.Code))
		return
	}
	return parsed.Data, nil
}

type workstate struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		RequestID string `json:"requestId"`
		TaskID    int    `json:"taskId"`
	} `json:"data"`
}

// BuildWork 创建画图任务
//
// token:GetToken函数获取,
//
// keyword:图片描述,长度不超过64个字,prompt指南:https://wenxin.baidu.com/wenxin/docs#Ol7ece95m
//
// picType:图片风格，目前支持风格有：油画、水彩画、卡通、粉笔画、儿童画、蜡笔画
//
// picSize:图片尺寸，目前支持的有：1024*1024 方图、1024*1536 长图、1536*1024 横图。
// 传入的是尺寸数值，非文字。
//
// taskID:任务ID，用于查询结果
func BuildWork(toekn, keyword, picType, picSize string) (taskID int, err error) {
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernievilg/v1/txt2img?access_token=" + url.QueryEscape(toekn) +
		"&text=" + url.QueryEscape(keyword) + "&style=" + url.QueryEscape(picType) + "&resolution=" + url.QueryEscape(picSize)
	postData := url.Values{}
	postData.Add("name", "雪儿")
	postData.Add("language", "golang")
	data, err := web.PostData(requestURL, "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
	if err != nil {
		return
	}
	var parsed workstate
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	if parsed.Msg != "success" {
		err = errors.New(parsed.Msg + ",code:" + strconv.Itoa(parsed.Code))
		return
	}
	return parsed.Data.TaskID, nil
}

type picdata struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Img     string `json:"img"`
		Waiting string `json:"waiting"`
		ImgUrls []struct {
			Image string      `json:"image"`
			Score interface{} `json:"score"`
		} `json:"imgUrls"`
		CreateTime string `json:"createTime"`
		RequestID  string `json:"requestId"`
		Style      string `json:"style"`
		Text       string `json:"text"`
		Resolution string `json:"resolution"`
		TaskID     int    `json:"taskId"`
		Status     int    `json:"status"`
	} `json:"data"`
}

// GetPic 获取图片内容
//
// token由GetToken函数获取,taskID由BuildWork函数获取
//
// picurl:图片链接--
//
// stauts:结果状态,"30s"代表还在排队生成,"0"表示结果OK
func GetPic(toekn string, taskID int) (picurl string, status string, err error) {
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernievilg/v1/getImg?access_token=" + url.QueryEscape(toekn) +
		"&taskId=" + strconv.Itoa(taskID)
	postData := url.Values{}
	postData.Add("name", "雪儿")
	postData.Add("language", "golang")
	data, err := web.PostData(requestURL, "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
	if err != nil {
		return
	}
	var parsed picdata
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	if parsed.Msg != "success" {
		err = errors.New(parsed.Msg + ",code:" + strconv.Itoa(parsed.Code))
		return
	}
	status = parsed.Data.Waiting
	picurl = parsed.Data.Img
	return
}

var token = "24.6fc079766d2ad7c81895645aa75bd8dd.86400000.1665479417415.e30e7abdcca57e576e95eae27b9ba411-130258"

var clientID = "A6osGoF9f12UDbIl9UWMrS29XepSWtHE"
var clientSecret = "VbEXT5HRWRhPPxhIRXfpwNsjO698iFAX"

func init() {
	xx := control.Register("huatu", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "AI画图\n" +
			"画图XXX",
	})
	xx.OnRegex(`^画图\s*(.+)$`).SetBlock(true). // 正值输入
							Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			taskID, err := BuildWork(token, str, "卡通", "1024*1024")
			if err != nil {
				ctx.SendChain(message.Text("ERROR:画图错误"))
				return
			}
			ctx.SendChain(message.Text("请求画图成功,等待制作,倒计时120S"))
			time.Sleep(time.Duration(120) * time.Second)
			ctx.SendChain(message.Text("获得图片ID:", taskID))
			picurl, status, err := GetPic(token, taskID)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:输出错误"))
				return
			}
			ctx.SendChain(message.Text("获取状态:", status))
			data, err := web.GetData(picurl)
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	xx.OnFullMatch("获取token").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			token, err := GetToken(clientID, clientSecret)
			if err != nil {
				ctx.SendChain(message.Text("获取token失败惹", err))
				return
			}
			ctx.SendChain(message.Text(token))
		})
}
