package music

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init()  {
	zero.OnCommand("点歌").SetBlock(true).SetPriority(50).Handle(func(ctx *zero.Ctx) {
		ctx.Send(QQMusic(ctx.State["args"].(string)))
		return
	})

	zero.OnCommand("酷我点歌").SetBlock(true).SetPriority(50).Handle(func(ctx *zero.Ctx) {
		ctx.Send(KuWo(ctx.State["args"].(string)))
		return
	})

	zero.OnCommand("酷狗点歌").SetBlock(true).SetPriority(50).Handle(func(ctx *zero.Ctx) {
		ctx.Send(KuGou(ctx.State["args"].(string)))
		return
	})

	zero.OnCommand("网易点歌").SetBlock(true).SetPriority(50).Handle(func(ctx *zero.Ctx) {
		ctx.Send(WyCloud(ctx.State["args"].(string)))
		return
	})
}

//-----------------------------------------------------------------------

func KuWo(KeyWord string) string {
	headers := map[string]string{
		"Cookie": "Hm_lvt_cdb524f42f0ce19b169a8071123a4797=1610284708,1610699237; _ga=GA1.2.1289529848.1591618534; kw_token=LWKACV45JSQ; Hm_lpvt_cdb524f42f0ce19b169a8071123a4797=1610699468; _gid=GA1.2.1868980507.1610699238; _gat=1",
		"csrf": "LWKACV45JSQ",
		"User-Agent": "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0",
		"Referer": "https://www.kuwo.cn/search/list?key=",
	}
	api := "https://www.kuwo.cn/api/www/search/searchMusicBykeyWord?key="+url.QueryEscape(KeyWord)+"&pn=1&rn=1&httpsStatus=1"
	info := gjson.ParseBytes(NetGet(api,headers)).Get("data.list.0")
	return fmt.Sprintf("[CQ:music,type=custom,url=%s,audio=%s,title=%s,content=%s,image=%s]",
		fmt.Sprintf("https://www.kuwo.cn/play_detail/%d",info.Get("rid").Int()),
		gjson.ParseBytes(
			NetGet(fmt.Sprintf(
				"http://www.kuwo.cn/url?format=mp3&rid=%d&response=url&type=convert_url3&br=128kmp3&from=web&httpsStatus=1", info.Get("rid").Int()),headers)).Get("url").Str,
		info.Get("name").Str,
		info.Get("artist").Str,
		info.Get("pic").Str,
		)
}

func KuGou(KeyWord string) string {
	stamp := time.Now().UnixNano()/1e6
	api := fmt.Sprintf(
		"https://complexsearch.kugou.com/v2/search/song?callback=callback123&keyword=%s&page=1&pagesize=30&bitrate=0&isfuzzy=0&tag=em&inputtype=0&platform=WebFilter&userid=-1&clientver=2000&iscorrection=1&privilege_filter=0&srcappid=2919&clienttime=%d&mid=%d&uuid=%d&dfid=-&signature=%s",
		KeyWord,stamp,stamp,stamp,GetMd5(fmt.Sprintf(
			"NVPh5oo715z5DIWAeQlhMDsWXXQV4hwtbitrate=0callback=callback123clienttime=%dclientver=2000dfid=-inputtype=0iscorrection=1isfuzzy=0keyword=%smid=%dpage=1pagesize=30platform=WebFilterprivilege_filter=0srcappid=2919tag=emuserid=-1uuid=%dNVPh5oo715z5DIWAeQlhMDsWXXQV4hwt",
			stamp,KeyWord,stamp,stamp)),
		)
	res := NetGet(api, map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0"})
	info := gjson.ParseBytes(res[12:len(res)-2]).Get("data.lists.0")
	res = NetGet(
		fmt.Sprintf("https://wwwapi.kugou.com/yy/index.php?r=play/getdata&hash=%s&album_id=%s",info.Get("FileHash").Str,info.Get("AlbumID").Str),
		map[string]string{
			"Cookie": "kg_mid=d8e70a262c93d47599c6196c612d6f4f; Hm_lvt_aedee6983d4cfc62f509129360d6bb3d=1610278505,1611631363,1611722252; kg_dfid=33ZWee1kircl0jcJ1h0WF1fX; Hm_lpvt_aedee6983d4cfc62f509129360d6bb3d=1611727348; kg_dfid_collect=d41d8cd98f00b204e9800998ecf8427e",
			"Host": "wwwapi.kugou.com",
			"TE": "Trailers",
			"User-Agent": "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0",
		})
	jump_url := fmt.Sprintf("https://www.kugou.com/song/#hash=%s&album_id=%s",info.Get("FileHash").Str,info.Get("AlbumID").Str)
	info = gjson.ParseBytes(res).Get("data")
	return fmt.Sprintf("[CQ:music,type=custom,url=%s,audio=%s,title=%s,content=%s,image=%s]",
		jump_url,
		strings.Replace(info.Get("play_backup_url").Str,"\\/","/",-1),
		info.Get("song_name").Str,
		info.Get("author_name").Str,
		info.Get("img").Str,
		)
}

func WyCloud(KeyWord string) string {
	res := NetPost(
		"http://music.163.com/api/search/pc",
		map[string]string{"offset": "0", "total": "true", "limit": "9", "type": "1", "s": KeyWord},
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36",
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
	info := gjson.ParseBytes(res[9:len(res)-1]).Get("data.song.list.0")
	return fmt.Sprintf(
		"[CQ:music,type=custom,url=%s,audio=%s,title=%s,content=%s,image=%s]",
		"https://y.qq.com/n/yqq/song/"+info.Get("songmid").Str+".html",
		"https://isure.stream.qqmusic.qq.com/" + gjson.ParseBytes(
			NetGet(
				strings.Replace(params,"{}",info.Get("songmid").Str,-1),
				map[string]string{
					"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
					"referer": "http://y.qq.com",
				})).Get("req_0.data.midurlinfo.0.purl").Str,
		info.Get("songname").Str,
		info.Get("singer.0.name").Str,
		"https://y.gtimg.cn/music/photo_new"+StrMidGet("//y.gtimg.cn/music/photo_new","?max_age",string(NetGet("https://y.qq.com/n/yqq/song/"+info.Get("songmid").Str+".html", map[string]string{}))),
		)
}

//-----------------------------------------------------------------------

func StrMidGet(pre string,suf string,str string) string {
	n := strings.Index(str, pre)
	if n == -1 {n = 0} else {n = n + len(pre)}
	str = string([]byte(str)[n:])
	m := strings.Index(str, suf)
	if m == -1 {m = len(str)}
	return string([]byte(str)[:m])
}

func GetMd5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	result := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	return result
}

func NetGet(get_url string,headers map[string]string) []byte {
	client := &http.Client{}
	request,_ := http.NewRequest("GET",get_url,nil)
	for key,value := range headers{request.Header.Add(key,value)}
	res,_ := client.Do(request)
	defer res.Body.Close()
	result,_ := ioutil.ReadAll(res.Body)
	return result
}

func NetPost(post_url string,data map[string]string,headers map[string]string) []byte {
	client := &http.Client{}
	param := url.Values{}
	for key,value := range data{param.Set(key,value)}
	request,_ := http.NewRequest("POST",post_url,strings.NewReader(param.Encode()))
	for key,value := range headers{request.Header.Add(key,value)}
	res,_ := client.Do(request)
	defer res.Body.Close()
	result,_ := ioutil.ReadAll(res.Body)
	return result
}
