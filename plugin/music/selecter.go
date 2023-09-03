// Package music QQ音乐、网易云、酷狗、酷我、咪咕 点歌
package music

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/web"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "点歌",
		Help: "- 点歌[xxx]\n" +
			"- 网易点歌[xxx]\n" +
			"- 酷我点歌[xxx]\n" +
			"- 酷狗点歌[xxx]\n" +
			"- 咪咕点歌[xxx]",
	}).OnRegex(`^(.{0,2})点歌\s?(.{1,25})$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			// switch 平台
			switch ctx.State["regex_matched"].([]string)[1] {
			case "咪咕":
				ctx.SendChain(migu(ctx.State["regex_matched"].([]string)[2]))
			case "酷我":
				ctx.SendChain(kuwo(ctx.State["regex_matched"].([]string)[2]))
			case "酷狗":
				ctx.SendChain(kugou(ctx.State["regex_matched"].([]string)[2]))
			case "网易":
				ctx.SendChain(cloud163(ctx.State["regex_matched"].([]string)[2]))
			default: // 默认 QQ音乐
				ctx.SendChain(qqmusic(ctx.State["regex_matched"].([]string)[2]))
			}
		})
}

// migu 返回咪咕音乐卡片
func migu(keyword string) message.MessageSegment {
	headers := http.Header{
		"Cookie":     []string{"audioplayer_exist=1; audioplayer_open=0; migu_cn_cookie_id=3ad476db-f021-4bda-ab91-c485ac3d56a0; Hm_lvt_ec5a5474d9d871cb3d82b846d861979d=1671119573; Hm_lpvt_ec5a5474d9d871cb3d82b846d861979d=1671119573; WT_FPC=id=279ef92eaf314cbb8d01671116477485:lv=1671119583092:ss=1671116477485"},
		"csrf":       []string{"LWKACV45JSQ"},
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"},
		"Referer":    []string{"http://m.music.migu.cn"},
		"proxy":      []string{"false"},
	}
	// 搜索音乐信息 第一首歌
	search, _ := url.Parse("http://m.music.migu.cn/migu/remoting/scr_search_tag")
	search.RawQuery = url.Values{
		"keyword": []string{keyword},
		"type":    []string{"2"},
		"pgc":     []string{"1"},
		"rows":    []string{"10"},
	}.Encode()
	info := gjson.ParseBytes(netGet(search.String(), headers)).Get("musics.0")
	// 返回音乐卡片
	return message.CustomMusic(
		fmt.Sprintf("https://music.migu.cn/v3/music/song/%s", info.Get("copyrightId").String()),
		info.Get("mp3").String(),
		info.Get("songName").String(),
	).Add("content", info.Get("artist").Str).Add("image", info.Get("cover").Str).Add("subtype", "migu")
}

// kuwo 返回酷我音乐卡片
func kuwo(keyword string) message.MessageSegment {
	headers := http.Header{
		"Cookie":     []string{"Hm_lvt_cdb524f42f0ce19b169a8071123a4797=1610284708,1610699237; _ga=GA1.2.1289529848.1591618534; kw_token=LWKACV45JSQ; Hm_lpvt_cdb524f42f0ce19b169a8071123a4797=1610699468; _gid=GA1.2.1868980507.1610699238; _gat=1"},
		"csrf":       []string{"LWKACV45JSQ"},
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0"},
		"Referer":    []string{"https://www.kuwo.cn/search/list?key="},
	}
	// 搜索音乐信息 第一首歌
	search, _ := url.Parse("https://www.kuwo.cn/api/www/search/searchMusicBykeyWord")
	search.RawQuery = url.Values{
		"key":         []string{keyword},
		"pn":          []string{"1"},
		"rn":          []string{"1"},
		"httpsStatus": []string{"1"},
	}.Encode()
	info := gjson.ParseBytes(netGet(search.String(), headers)).Get("data.list.0")
	// 获得音乐直链
	music, _ := url.Parse("http://www.kuwo.cn/api/v1/www/music/playUrl")
	music.RawQuery = url.Values{
		"mid":         []string{fmt.Sprintf("%d", info.Get("rid").Int())},
		"type":        []string{"convert_url3"},
		"br":          []string{"320kmp3"},
		"httpsStatus": []string{"1"},
	}.Encode()
	audio := gjson.ParseBytes(netGet(music.String(), headers))
	// 返回音乐卡片
	return message.CustomMusic(
		fmt.Sprintf("https://www.kuwo.cn/play_detail/%d", info.Get("rid").Int()),
		audio.Get("data.url").Str,
		info.Get("name").Str,
	).Add("content", info.Get("artist").Str).Add("image", info.Get("pic").Str).Add("subtype", "kuwo")
}

// kugou 返回酷狗音乐卡片
func kugou(keyword string) message.MessageSegment {
	stamp := time.Now().UnixNano() / 1e6
	hash := md5str(
		fmt.Sprintf(
			"NVPh5oo715z5DIWAeQlhMDsWXXQV4hwtbitrate=0callback=callback123clienttime=%dclientver=2000dfid=-inputtype=0iscorrection=1isfuzzy=0keyword=%smid=%dpage=1pagesize=30platform=WebFilterprivilege_filter=0srcappid=2919tag=emuserid=-1uuid=%dNVPh5oo715z5DIWAeQlhMDsWXXQV4hwt",
			stamp, keyword, stamp, stamp,
		),
	)
	// 搜索音乐信息 第一首歌
	h1 := http.Header{
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0"},
	}
	search, _ := url.Parse("https://complexsearch.kugou.com/v2/search/song")
	search.RawQuery = url.Values{
		"callback":         []string{"callback123"},
		"keyword":          []string{keyword},
		"page":             []string{"1"},
		"pagesize":         []string{"30"},
		"bitrate":          []string{"0"},
		"isfuzzy":          []string{"0"},
		"tag":              []string{"em"},
		"inputtype":        []string{"0"},
		"platform":         []string{"WebFilter"},
		"userid":           []string{"-1"},
		"clientver":        []string{"2000"},
		"iscorrection":     []string{"1"},
		"privilege_filter": []string{"0"},
		"srcappid":         []string{"2919"},
		"clienttime":       []string{fmt.Sprintf("%d", stamp)},
		"mid":              []string{fmt.Sprintf("%d", stamp)},
		"uuid":             []string{fmt.Sprintf("%d", stamp)},
		"dfid":             []string{"-"},
		"signature":        []string{hash},
	}.Encode()
	res := netGet(search.String(), h1)
	info := gjson.ParseBytes(res[12 : len(res)-2]).Get("data.lists.0")
	// 获得音乐直链
	h2 := http.Header{
		"Cookie":     []string{"kg_mid=d8e70a262c93d47599c6196c612d6f4f; Hm_lvt_aedee6983d4cfc62f509129360d6bb3d=1610278505,1611631363,1611722252; kg_dfid=33ZWee1kircl0jcJ1h0WF1fX; Hm_lpvt_aedee6983d4cfc62f509129360d6bb3d=1611727348; kg_dfid_collect=d41d8cd98f00b204e9800998ecf8427e"},
		"Host":       []string{"wwwapi.kugou.com"},
		"TE":         []string{"Trailers"},
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0"},
	}
	music := "https://wwwapi.kugou.com/yy/index.php?r=play%2Fgetdata&hash=" + info.Get("FileHash").Str + "&album_id=" + info.Get("AlbumID").Str
	audio := gjson.ParseBytes(netGet(music, h2)).Get("data")
	// 返回音乐卡片
	return message.CustomMusic(
		"https://www.kugou.com/song/#hash="+audio.Get("hash").Str+"&album_id="+audio.Get("album_id").Str,
		strings.ReplaceAll(audio.Get("play_backup_url").Str, "\\/", "/"),
		audio.Get("audio_name").Str,
	).Add("content", audio.Get("author_name").Str).Add("image", audio.Get("img").Str).Add("subtype", "kugou")
}

// cloud163 返回网易云音乐卡片
func cloud163(keyword string) (msg message.MessageSegment) {
	requestURL := "http://music.163.com/api/search/get/web?type=1&limit=1&s=" + url.QueryEscape(keyword)
	data, err := web.GetData(requestURL)
	if err != nil {
		msg = message.Text("ERROR: ", err)
		return
	}
	msg = message.Music("163", gjson.ParseBytes(data).Get("result.songs.0.id").Int())
	return
}

// qqmusic 返回QQ音乐卡片
func qqmusic(keyword string) (msg message.MessageSegment) {
	requestURL := "https://c.y.qq.com/splcloud/fcgi-bin/smartbox_new.fcg?platform=yqq.json&key=" + url.QueryEscape(keyword)
	data, err := web.RequestDataWith(web.NewDefaultClient(), requestURL, "GET", "", web.RandUA(), nil)
	if err != nil {
		msg = message.Text("ERROR: ", err)
		return
	}
	msg = message.Music("qq", gjson.ParseBytes(data).Get("data.song.itemlist.0.id").Int())
	return
}

// md5str 返回字符串 MD5
func md5str(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	result := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	return result
}

// netGet 返回请求数据
func netGet(url string, header http.Header) []byte {
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	request.Header = header
	res, err := client.Do(request)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	result, _ := io.ReadAll(res.Body)
	return result
}
