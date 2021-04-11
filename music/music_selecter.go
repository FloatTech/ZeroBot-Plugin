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

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	zero.OnRegex("^酷我点歌(.+?)$").SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(kuwo(ctx.State["regex_matched"].([]string)[1]))
			return
		})

	zero.OnRegex("^酷狗点歌(.+?)$").SetBlock(true).SetPriority(50).Handle(func(ctx *zero.Ctx) {
		ctx.Send(kugou(ctx.State["regex_matched"].([]string)[1]))
		return
	})

	zero.OnRegex("^网易点歌(.+?)$").SetBlock(true).SetPriority(50).Handle(func(ctx *zero.Ctx) {
		ctx.Send(WyCloud(ctx.State["regex_matched"].([]string)[1]))
		return
	})

	zero.OnRegex("^点歌(.+?)$").SetBlock(true).SetPriority(50).Handle(func(ctx *zero.Ctx) {
		ctx.Send(QQMusic(ctx.State["regex_matched"].([]string)[1]))
		return
	})
}

//-----------------------------------------------------------------------

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
	hash := GetMd5(
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
		strings.Replace(audio.Get("play_backup_url").Str, "\\/", "/", -1),
		audio.Get("audio_name").Str,
	).Add("content", audio.Get("author_name").Str).Add("image", audio.Get("img").Str)
}

func WyCloud(KeyWord string) string {
	res := NetPost(
		"http://music.163.com/api/search/pc",
		map[string]string{"offset": "0", "total": "true", "limit": "9", "type": "1", "s": KeyWord},
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36",
		},
	)
	info := gjson.ParseBytes(res).Get("result.songs.0")
	return fmt.Sprintf(
		"[CQ:music,type=custom,url=%s,audio=%s,title=%s,content=%s,image=%s]",
		fmt.Sprintf("http://y.music.163.com/m/song?id=%d", info.Get("id").Int()),
		fmt.Sprintf("http://music.163.com/song/media/outer/url?id=%d.mp3", info.Get("id").Int()),
		info.Get("name").Str,
		info.Get("artists.0.name").Str,
		info.Get("album.blurPicUrl").Str,
	)
}

func QQMusic(KeyWord string) string {
	params := `https://u.y.qq.com/cgi-bin/musicu.fcg?data=%7B%22req%22%3A+%7B%22module%22%3A+%22CDN.SrfCdnDispatchServer%22%2C+%22method%22%3A+%22GetCdnDispatch%22%2C+%22param%22%3A+%7B%22guid%22%3A+%223982823384%22%2C+%22calltype%22%3A+0%2C+%22userip%22%3A+%22%22%7D%7D%2C+%22req_0%22%3A+%7B%22module%22%3A+%22vkey.GetVkeyServer%22%2C+%22method%22%3A+%22CgiGetVkey%22%2C+%22param%22%3A+%7B%22guid%22%3A+%223982823384%22%2C+%22songmid%22%3A+%5B%22{}%22%5D%2C+%22songtype%22%3A+%5B0%5D%2C+%22uin%22%3A+%220%22%2C+%22loginflag%22%3A+1%2C+%22platform%22%3A+%2220%22%7D%7D%2C+%22comm%22%3A+%7B%22uin%22%3A+0%2C+%22format%22%3A+%22json%22%2C+%22ct%22%3A+24%2C+%22cv%22%3A+0%7D%7D`
	res := NetGet(
		"https://c.y.qq.com/soso/fcgi-bin/client_search_cp?w="+KeyWord,
		map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0"},
	)
	info := gjson.ParseBytes(res[9 : len(res)-1]).Get("data.song.list.0")
	return fmt.Sprintf(
		"[CQ:music,type=custom,url=%s,audio=%s,title=%s,content=%s,image=%s]",
		"https://y.qq.com/n/yqq/song/"+info.Get("songmid").Str+".html",
		"https://isure.stream.qqmusic.qq.com/"+gjson.ParseBytes(
			NetGet(
				strings.Replace(params, "{}", info.Get("songmid").Str, -1),
				map[string]string{
					"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
					"referer":    "http://y.qq.com",
				})).Get("req_0.data.midurlinfo.0.purl").Str,
		info.Get("songname").Str,
		info.Get("singer.0.name").Str,
		"https://y.gtimg.cn/music/photo_new"+StrMidGet("//y.gtimg.cn/music/photo_new", "?max_age", string(NetGet("https://y.qq.com/n/yqq/song/"+info.Get("songmid").Str+".html", map[string]string{}))),
	)
}

//-----------------------------------------------------------------------

func StrMidGet(pre string, suf string, str string) string {
	n := strings.Index(str, pre)
	if n == -1 {
		n = 0
	} else {
		n = n + len(pre)
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, suf)
	if m == -1 {
		m = len(str)
	}
	return string([]byte(str)[:m])
}

func GetMd5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	result := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	return result
}

func NetGet(get_url string, headers map[string]string) []byte {
	client := &http.Client{}
	request, _ := http.NewRequest("GET", get_url, nil)
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	res, _ := client.Do(request)
	defer res.Body.Close()
	result, _ := ioutil.ReadAll(res.Body)
	return result
}

// netGet 返回请求数据
func netGet(get_url string, header http.Header) []byte {
	client := &http.Client{}
	request, _ := http.NewRequest("GET", get_url, nil)
	request.Header = header
	res, _ := client.Do(request)
	defer res.Body.Close()
	result, _ := ioutil.ReadAll(res.Body)
	return result
}

func NetPost(post_url string, data map[string]string, headers map[string]string) []byte {
	client := &http.Client{}
	param := url.Values{}
	for key, value := range data {
		param.Set(key, value)
	}
	request, _ := http.NewRequest("POST", post_url, strings.NewReader(param.Encode()))
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	res, _ := client.Do(request)
	defer res.Body.Close()
	result, _ := ioutil.ReadAll(res.Body)
	return result
}
