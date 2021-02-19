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

func CloudMusic(name string) (music CQMusic, err error) {
	var api = "http://music.163.com/api/search/pc"

	client := &http.Client{}

	// TODO 包装请求参数
	data := url.Values{}
	data.Set("offset", "0")
	data.Set("total", "true")
	data.Set("limit", "9")
	data.Set("type", "1")
	data.Set("s", name)
	fromData := strings.NewReader(data.Encode())

	// TODO 网络请求
	req, err := http.NewRequest("POST", api, fromData)
	if err != nil {
		return music, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return music, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return music, err
	}

	if code := resp.StatusCode; code != 200 {
		// 如果返回不是200则立刻抛出错误
		return music, errors.New(fmt.Sprintf("CloudMusic not found, code %d", code))
	}
	content := gjson.ParseBytes(body).Get("result.songs.0")
	music.Type = "custom"
	music.Url = fmt.Sprintf("http://y.music.163.com/m/song?id=%d", content.Get("id").Int())
	music.Audio = fmt.Sprintf("http://music.163.com/song/media/outer/url?id=%d.mp3", content.Get("id").Int())
	music.Title = content.Get("name").Str
	music.Content = content.Get("artists.0.name").Str
	music.Image = content.Get("album.blurPicUrl").Str
	return music, nil
}
