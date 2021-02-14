package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	xpath "github.com/antchfx/htmlquery"
)

// Ascii2dSearch Ascii2d 以图搜图
// 第一个参数 返回错误
// 第二个参数 返回的信息
func Ascii2dSearch(pic string) (text string, err error) {
	var (
		api = "https://ascii2d.net/search/uri"
	)
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: &transport,
	}

	// TODO 包装请求参数
	data := url.Values{}
	data.Set("uri", pic) // 图片链接
	fromData := strings.NewReader(data.Encode())

	// TODO 网络请求
	req, _ := http.NewRequest("POST", api, fromData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	// TODO 色合检索改变到特征检索
	var bovwUrl = strings.ReplaceAll(resp.Request.URL.String(), "color", "bovw")
	bovwReq, _ := http.NewRequest("POST", bovwUrl, nil)
	bovwReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	bovwReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36")
	bovwResp, err := client.Do(bovwReq)
	if err != nil {
		return "", err
	}
	defer bovwResp.Body.Close()
	// TODO 解析XPATH
	doc, err := xpath.Parse(resp.Body)
	if err != nil {
		return "", err
	}
	// TODO 取出每个返回的结果
	list := xpath.Find(doc, `//div[@class="row item-box"]`)
	var link string
	// TODO 遍历取出第一个返回的PIXIV结果
	for _, n := range list {
		linkPath := xpath.Find(n, `//div[2]/div[3]/h6/a[1]`)
		picPath := xpath.Find(n, `//div[1]/img`)
		if len(linkPath) != 0 && len(picPath) != 0 {
			link = xpath.SelectAttr(linkPath[0], "href")
			if strings.Contains(link, "www.pixiv.net") {
				break
			}
		}
	}
	// TODO 链接取出PIXIV id
	var index = strings.LastIndex(link, "/")
	if link == "" || index == -1 {
		return "", errors.New("Ascii2d not found")
	}
	var id = Str2Int(link[index+1:])
	if id == 0 {
		return "", errors.New("convert to pid error")
	}
	// TODO 根据PID查询插图信息
	var illust = &Illust{}
	if err := illust.IllustInfo(id); err != nil {
		return "", err
	}
	if illust.AgeLimit != "all-age" {
		return "", errors.New("Ascii2d not found")
	}
	// TODO 返回插图信息文本
	return fmt.Sprintf(
		`[SetuTime] emmm大概是这个？
标题：%s
插画ID：%d
画师：%s
画师ID：%d
直链：https://pixivel.moe/detail?id=%d`,
		illust.Title,
		illust.Pid,
		illust.UserName,
		illust.UserId,
		illust.Pid,
	), nil
}
