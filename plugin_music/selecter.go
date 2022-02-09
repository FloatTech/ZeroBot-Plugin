// Package music QQ音乐、网易云、酷狗、酷我 点歌
package music

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

func init() {
	control.Register("music", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "点歌\n" +
			"- 点歌[xxx]\n" +
			"- 网易点歌[xxx]\n" +
			"- 酷我点歌[xxx]\n" +
			"- 酷狗点歌[xxx]",
	}).OnRegex(`^(.{0,2})点歌\s?(.{1,25})$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			// switch 平台
			switch ctx.State["regex_matched"].([]string)[1] {
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
	music, _ := url.Parse("http://www.kuwo.cn/url")
	music.RawQuery = url.Values{
		"format":      []string{"mp3"},
		"rid":         []string{fmt.Sprintf("%d", info.Get("rid").Int())},
		"response":    []string{"url"},
		"type":        []string{"convert_url3"},
		"br":          []string{"128kmp3"},
		"from":        []string{"web"},
		"httpsStatus": []string{"1"},
	}.Encode()
	audio := gjson.ParseBytes(netGet(music.String(), headers))
	// 返回音乐卡片
	return message.CustomMusic(
		fmt.Sprintf("https://www.kuwo.cn/play_detail/%d", info.Get("rid").Int()),
		audio.Get("url").Str,
		info.Get("name").Str,
	).Add("content", info.Get("artist").Str).Add("image", info.Get("pic").Str)
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
	).Add("content", audio.Get("author_name").Str).Add("image", audio.Get("img").Str)
}

// cloud163 返回网易云音乐卡片
func cloud163(keyword string) message.MessageSegment {
	headers := http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded"},
		"User-Agent":   []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0"},
	}
	data := url.Values{
		"offset": []string{"0"},
		"total":  []string{"true"},
		"limit":  []string{"9"},
		"type":   []string{"1"},
		"s":      []string{keyword},
	}
	// 搜索音乐信息 第一首歌
	info := gjson.ParseBytes(netPost("http://music.163.com/api/search/pc", data, headers)).Get("result.songs.0")
	// 返回音乐卡片
	return message.CustomMusic(
		fmt.Sprintf("http://y.music.163.com/m/song?id=%d", info.Get("id").Int()),
		fmt.Sprintf("http://music.163.com/song/media/outer/url?id=%d.mp3", info.Get("id").Int()),
		info.Get("name").Str,
	).Add("content", info.Get("artists.0.name").Str).Add("image", info.Get("album.blurPicUrl").Str)
}

// qqmusic 返回QQ音乐卡片
func qqmusic(keyword string) message.MessageSegment {
	// 搜索音乐信息 第一首歌
	h1 := http.Header{
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0"},
	}
	search, _ := url.Parse("https://c.y.qq.com/soso/fcgi-bin/client_search_cp")
	search.RawQuery = url.Values{
		"w": []string{keyword},
	}.Encode()
	res := netGet(search.String(), h1)
	info := gjson.ParseBytes(res[9 : len(res)-1]).Get("data.song.list.0")
	// 获得音乐直链
	h2 := http.Header{
		"User-Agent": []string{"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"},
		"referer":    []string{"http://y.qq.com"},
	}
	music, _ := url.Parse("https://u.y.qq.com/cgi-bin/musicu.fcg")
	music.RawQuery = url.Values{
		"data": []string{`{"req": {"module": "CDN.SrfCdnDispatchServer", "method": "GetCdnDispatch", "param": {"guid": "3982823384", "calltype": 0, "userip": ""}}, "req_0": {"module": "vkey.GetVkeyServer", "method": "CgiGetVkey", "param": {"guid": "3982823384", "songmid": ["` + info.Get("songmid").Str + `"], "songtype": [0], "uin": "0", "loginflag": 1, "platform": "20"}}, "comm": {"uin": 0, "format": "json", "ct": 24, "cv": 0}}`},
	}.Encode()
	audio := gjson.ParseBytes(netGet(music.String(), h2))
	// 获得音乐封面
	image := "https://y.gtimg.cn/music/photo_new/" +
		find(
			`photo_new\u002F`,
			"?max_age",
			string(
				netGet("https://y.qq.com/n/yqq/song/"+info.Get("songmid").Str+".html", nil),
			),
		)
	// 返回音乐卡片
	return message.CustomMusic(
		"https://y.qq.com/n/yqq/song/"+info.Get("songmid").Str+".html",
		"https://isure.stream.qqmusic.qq.com/"+audio.Get("req_0.data.midurlinfo.0.purl").Str,
		info.Get("songname").Str,
	).Add("content", info.Get("singer.0.name").Str).Add("image", image)
}

// find 返回 pre 到 suf 之间的文本
func find(pre string, suf string, str string) string {
	n := strings.Index(str, pre)
	if n == -1 {
		n = 0
	} else {
		n += len(pre)
	}
	str = str[n:]
	m := strings.Index(str, suf)
	if m == -1 {
		m = len(str)
	}
	return str[:m]
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
	result, _ := ioutil.ReadAll(res.Body)
	return result
}

// netPost 返回请求数据
func netPost(url string, data url.Values, header http.Header) []byte {
	client := &http.Client{}
	request, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	request.Header = header
	res, err := client.Do(request)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	result, _ := ioutil.ReadAll(res.Body)
	return result
}
