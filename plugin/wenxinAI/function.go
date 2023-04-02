// Package wenxin 百度文心AI
package wenxin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/web"
)

type tokendata struct {
	Code int64  `json:"code"`
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
func GetToken(clientID, clientSecret string) (token string, code int64, err error) {
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/oauth/token?grant_type=client_credentials&client_id=" + url.QueryEscape(clientID) + "&client_secret=" + url.QueryEscape(clientSecret)
	data, err := web.PostData(requestURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return
	}
	var parsed tokendata
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	if parsed.Msg != "success" {
		return "", parsed.Code, errors.New(parsed.Msg)
	}
	return parsed.Data, parsed.Code, nil
}

type workstate struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		RequestID string `json:"requestId"`
		TaskID    int64  `json:"taskId"`
	} `json:"data"`
}

// BuildPicWork 创建画图任务
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
// taskID:任务ID，用于查询结果,如果报错为错误代码
func BuildPicWork(token, keyword, picType, picSize string) (taskID int64, err error) {
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernievilg/v1/txt2img?access_token=" + url.QueryEscape(token)
	postData := url.Values{}
	postData.Add("text", keyword)
	postData.Add("style", picType)
	postData.Add("resolution", picSize)
	// postData.Add("num", "6")
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
		return parsed.Code, errors.New(parsed.Msg)
	}
	return parsed.Data.TaskID, nil
}

// BuildImgWork 创建以图绘图任务
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
// image:参考图片,本地文件
//
// taskID:任务ID，用于查询结果,如果报错为错误代码
func BuildImgWork(token, keyword, picType, picSize, image string) (taskID int64, err error) {
	picfile, err := os.Open(image)
	if err != nil {
		return
	}
	//创建一个multipart类型的写文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	//使用给出的属性名paramName和文件名filePath创建一个新的form-data头
	part, err := writer.CreateFormFile("image/png", filepath.Base(picfile.Name()))
	if err != nil {
		return
	}
	//将源复制到目标，将file写入到part   是按默认的缓冲区32k循环操作的，不会将内容一次性全写入内存中,这样就能解决大文件的问题
	_, err = io.Copy(part, picfile)
	_ = writer.Close()
	_ = picfile.Close()
	if err != nil {
		return
	}
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernievilg/v1/txt2img?access_token=" + url.QueryEscape(token)
	postData := url.Values{}
	postData.Add("text", keyword)
	postData.Add("style", picType)
	postData.Add("resolution", picSize)
	postData.Add("image", body.String())
	// postData.Add("num", "6")
	data, err := web.PostData(requestURL, "application/form-data", strings.NewReader(postData.Encode()))
	if err != nil {
		return
	}
	var parsed workstate
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	if parsed.Msg != "success" {
		return parsed.Code, errors.New(parsed.Msg)
	}
	return parsed.Data.TaskID, nil
}

// BuildTextWork 创建文字异步任务
//
// token:GetToken函数获取,
//
// keyword:问题描述
//
// style:版本
//
// taskID:任务ID，用于查询结果,如果报错为错误代码
func BuildTextWork(token, keyword, prompt string, style int) (taskID int64, err error) {
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernie/3.0." + strconv.Itoa(style) + "/zeus?access_token=" + url.QueryEscape(token)
	postData := url.Values{}
	postData.Add("text", keyword)        // 模型的输入文本，为prompt形式的输入。
	postData.Add("async", "1")           // 异步标识，现阶段必传且传1
	postData.Add("typeId", "1")          // 模型类型
	postData.Add("seq_len", "512")       // 最大生成长度[1, 1000]
	postData.Add("min_dec_len", "2")     // 最小生成长度[1, seq_len]
	postData.Add("topp", "0.8")          // 影响输出文本的多样性，取值越大，生成文本的多样性越强。
	postData.Add("task_prompt", prompt)  // 任务类型
	postData.Add("style", "1")           // 任务类型
	postData.Add("penalty_score", "1.2") // 减少重复生成的现象。值越大表示惩罚越大。设置过大会导致长文本生成效果变差。
	postData.Add("is_unidirectional", "0")
	postData.Add("min_dec_penalty_text", "。?：！[<S>]")
	postData.Add("mask_type", "sentence") // 设置该值可以控制模型生成粒度。可选参数为word, sentence, paragraph
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
		return parsed.Code, errors.New(parsed.Msg)
	}
	return parsed.Data.TaskID, nil
}

// GetPicResult 获取图片内容
//
// token由GetToken函数获取,taskID由BuildWork函数获取
//
// PicURL:[x]struct{Image:图片链接,Score:评分}
//
// API会返回x张图片,数量不确定的,随机的。
//
// 评分目前都是null,我不知道有什么用，既然API预留了，我也预留吧
//
// stauts:结果状态,如果报错为错误代码
func GetPicResult(token string, taskID int64) (picurls map[string]any, status int64, err error) {
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernievilg/v1/getImg?access_token=" + url.QueryEscape(token)
	postData := url.Values{}
	postData.Add("taskId", strconv.FormatInt(taskID, 10))
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
		return nil, parsed.Code, errors.New(parsed.Msg)
	}
	picurls = make(map[string]any, 2*len(parsed.Data.ImgUrls))
	for _, picurl := range parsed.Data.ImgUrls {
		picurls[picurl.Image] = picurl.Score
	}
	return picurls, parsed.Data.Status, nil
}

type picdata struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Img        string   `json:"img"`
		Waiting    string   `json:"waiting"`
		ImgUrls    []picURL `json:"imgUrls"`
		CreateTime string   `json:"createTime"`
		RequestID  string   `json:"requestId"`
		Style      string   `json:"style"`
		Text       string   `json:"text"`
		Resolution string   `json:"resolution"`
		TaskID     int64    `json:"taskId"`
		Status     int64    `json:"status"`
	} `json:"data"`
}

// picURL ...
type picURL struct {
	Image string      `json:"image"`
	Score interface{} `json:"score"`
}

// GetTextResult 获取结果
//
// token由GetToken函数获取,taskID由BuildWork函数获取
//
// stauts:结果状态,如果报错为错误代码
func GetTextResult(token string, taskID int64) (result string, status int64, err error) {
	requestURL := "https://wenxin.baidu.com/moduleApi/portal/api/rest/1.0/ernie/v1/getResult?access_token=" + url.QueryEscape(token)
	postData := url.Values{}
	postData.Add("taskId", strconv.FormatInt(taskID, 10))
	data, err := web.PostData(requestURL, "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
	if err != nil {
		return
	}
	var parsed msgresult
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return
	}
	if parsed.Msg != "success" {
		return "", parsed.Code, errors.New(parsed.Msg)
	}
	return parsed.Data.Result, parsed.Data.Status, nil
}

type msgresult struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Result     string `json:"result"`
		CreateTime string `json:"createTime"`
		RequestID  string `json:"requestId"`
		Text       string `json:"text"`
		TaskID     int64  `json:"taskId"`
		Status     int64  `json:"status"`
	} `json:"data"`
}
