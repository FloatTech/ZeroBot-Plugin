package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

// Illust 插画信息
type Illust struct {
	Pid         int64  `db:"pid"`
	Title       string `db:"title"`
	Caption     string `db:"caption"`
	Tags        string `db:"tags"`
	ImageUrls   string `db:"image_urls"`
	AgeLimit    string `db:"age_limit"`
	CreatedTime string `db:"created_time"`
	UserId      int64  `db:"user_id"`
	UserName    string `db:"user_name"`
}

// IllustInfo 根据p站插画id返回插画信息Illust
func (this *Illust) IllustInfo(id int64) (err error) {
	api := fmt.Sprintf("https://pixiv.net/ajax/illust/%d", id)
	transport := http.Transport{
		DisableKeepAlives: true,
		// 绕过sni审查
		TLSClientConfig: &tls.Config{
			ServerName:         "-",
			InsecureSkipVerify: true,
		},
		// 更改dns
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("tcp", "210.140.131.223:443")
		},
	}
	client := &http.Client{
		Transport: &transport,
	}

	// 网络请求
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Host", "pixiv.net")
	req.Header.Set("Referer", "pixiv.net")
	req.Header.Set("Accept", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if code := resp.StatusCode; code != 200 {
		return errors.New(fmt.Sprintf("Search illust's info failed, status %d", code))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	json := gjson.ParseBytes(body).Get("body")

	// 如果有"R-18"tag则判断为R-18（暂时）
	var ageLimit = "all-age"
	for _, tag := range json.Get("tags.tags.#.tag").Array() {
		if tag.Str == "R-18" {
			ageLimit = "r18"
			break
		}
	}
	// 解决json返回带html格式
	var caption = strings.ReplaceAll(json.Get("illustComment").Str, "<br />", "\n")
	if index := strings.Index(caption, "<"); index != -1 {
		caption = caption[:index]
	}
	// 解析返回插画信息
	this.Pid = json.Get("illustId").Int()
	this.Title = json.Get("illustTitle").Str
	this.Caption = caption
	this.Tags = fmt.Sprintln(json.Get("tags.tags.#.tag").Array())
	this.ImageUrls = json.Get("urls.original").Str
	this.AgeLimit = ageLimit
	this.CreatedTime = json.Get("createDate").Str
	this.UserId = json.Get("userId").Int()
	this.UserName = json.Get("userName").Str
	return nil
}
