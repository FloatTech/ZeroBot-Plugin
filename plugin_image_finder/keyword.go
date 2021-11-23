// Package imagefinder 关键字搜图
package imagefinder

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

type resultjson struct {
	Illusts []struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		Type      string `json:"type"`
		ImageUrls struct {
			SquareMedium string `json:"square_medium"`
			Medium       string `json:"medium"`
			Large        string `json:"large"`
		} `json:"image_urls"`
		Caption  string `json:"caption"`
		Restrict int    `json:"restrict"`
		User     struct {
			ID               int    `json:"id"`
			Name             string `json:"name"`
			Account          string `json:"account"`
			ProfileImageUrls struct {
				Medium string `json:"medium"`
			} `json:"profile_image_urls"`
			IsFollowed bool `json:"is_followed"`
		} `json:"user"`
		Tags []struct {
			Name           string      `json:"name"`
			TranslatedName interface{} `json:"translated_name"`
		} `json:"tags"`
		Tools          []interface{} `json:"tools"`
		PageCount      int           `json:"page_count"`
		Width          int           `json:"width"`
		Height         int           `json:"height"`
		SanityLevel    int           `json:"sanity_level"`
		XRestrict      int           `json:"x_restrict"`
		Series         interface{}   `json:"series"`
		MetaSinglePage struct {
			OriginalImageURL string `json:"original_image_url"`
		} `json:"meta_single_page,omitempty"`
		MetaPages      []interface{} `json:"meta_pages"`
		TotalView      int           `json:"total_view"`
		TotalBookmarks int           `json:"total_bookmarks"`
		IsBookmarked   bool          `json:"is_bookmarked"`
		Visible        bool          `json:"visible"`
		IsMuted        bool          `json:"is_muted"`
	} `json:"illusts"`
	NextURL         string `json:"next_url"`
	SearchSpanLimit int    `json:"search_span_limit"`
}

func init() {
	control.Register("imgfinder", &control.Options{
		DisableOnDefault: false,
		Help: "关键字搜图\n" +
			"- 来张 [xxx]",
	}).OnRegex(`^来张 (.*)$`, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			soutujson := soutuapi(keyword)
			pom1 := "https://i.pixiv.re"
			rannum := randintn(len(soutujson.Illusts))
			pom2 := soutujson.Illusts[rannum].ImageUrls.Medium[19:]
			ctx.SendChain(message.Image(pom1 + pom2))
		})
}

// soutuapi 请求api
func soutuapi(keyword string) *resultjson {
	url := "https://api.pixivel.moe/pixiv?type=search&page=0&mode=partial_match_for_tags&word=" + keyword
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	result := &resultjson{}
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}

// randintn 从json里的30条数据中随机获取一条返回
func randintn(len int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(len)
}
