package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// SauceNaoSearch SauceNao 以图搜图 需要链接 返回错误和信息
func SauceNaoSearch(pic string) (message.Message, error) {
	var (
		api    = "https://saucenao.com/search.php"
		apiKey = "2cc2772ca550dbacb4c35731a79d341d1a143cb5"

		minSimilarity = 70.0 // 返回图片结果的最小相似度
	)

	// 包装请求参数
	link, _ := url.Parse(api)
	link.RawQuery = url.Values{
		"url":         []string{pic},
		"api_key":     []string{apiKey},
		"db":          []string{"5"},
		"numres":      []string{"1"},
		"output_type": []string{"2"},
	}.Encode()

	// 网络请求
	client := &http.Client{}
	req, err := http.NewRequest("GET", link.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// 如果返回不是200则立刻抛出错误
		return nil, fmt.Errorf("SauceNAO not found, code %d", resp.StatusCode)
	}
	content := gjson.ParseBytes(body)
	if status := content.Get("header.status").Int(); status != 0 {
		// 如果json信息返回status不为0则立刻抛出错误
		return nil, fmt.Errorf("SauceNAO not found, status %d", status)
	}
	if content.Get("results.0.header.similarity").Float() < minSimilarity {
		return nil, fmt.Errorf("SauceNAO not found")
	}
	result := content.Get("results.0")
	// 正常发送
	return message.Message{
		message.Text("[SetuTime] 我有把握是这个！"),
		message.Image(result.Get("header.thumbnail").Str),
		message.Text(
			"\n",
			"相似度：", result.Get("header.similarity").Str, "\n",
			"标题：", result.Get("data.title").Str, "\n",
			"插画ID：", result.Get("data.pixiv_id").Int(), "\n",
			"画师：", result.Get("data.member_name").Str, "\n",
			"画师ID：", result.Get("data.member_id").Int(), "\n",
			"直链：", "https://pixivel.moe/detail?id=", result.Get("data.pixiv_id").Int(),
		),
	}, nil
}
