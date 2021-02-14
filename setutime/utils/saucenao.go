package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

// SauceNaoSearch SauceNao 以图搜图 需要链接 返回错误和信息
func SauceNaoSearch(pic string) (text string, err error) {
	var (
		api    = "https://saucenao.com/search.php"
		apiKey = "2cc2772ca550dbacb4c35731a79d341d1a143cb5"

		minSimilarity = 70.0 // 返回图片结果的最小相似度
	)

	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: &transport,
	}

	// TODO 包装请求参数
	data := url.Values{}
	data.Set("url", pic)         // 图片链接
	data.Set("api_key", apiKey)  // api_key
	data.Set("db", "5")          // 只搜索Pixiv
	data.Set("numres", "1")      // 返回一个结果
	data.Set("output_type", "2") // 返回JSON格式数据
	fromData := strings.NewReader(data.Encode())

	// TODO 网络请求
	req, err := http.NewRequest("POST", api, fromData)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if code := resp.StatusCode; code != 200 {
		// 如果返回不是200则立刻抛出错误
		return "", errors.New(fmt.Sprintf("SauceNAO not found, code %d", code))
	}
	content := gjson.ParseBytes(body)
	if status := content.Get("header.status").Int(); status != 0 {
		// 如果json信息返回status不为0则立刻抛出错误
		return "", errors.New(fmt.Sprintf("SauceNAO not found, status %d", status))
	}
	if content.Get("results.0.header.similarity").Float() < minSimilarity {
		return "", errors.New("SauceNAO not found")
	}
	// TODO 正常发送
	return fmt.Sprintf(
		`[SetuTime] 我有把握是这个！[CQ:image,file=%s]相似度：%s%%
标题：%s
插画ID：%d
画师：%s
画师ID：%d
直链：https://pixivel.moe/detail?id=%d`,
		content.Get("results.0.header.thumbnail").Str,
		content.Get("results.0.header.similarity").Str,
		content.Get("results.0.data.title").Str,
		content.Get("results.0.data.pixiv_id").Int(),
		content.Get("results.0.data.member_name").Str,
		content.Get("results.0.data.member_id").Int(),
		content.Get("results.0.data.pixiv_id").Int(),
	), nil
}
