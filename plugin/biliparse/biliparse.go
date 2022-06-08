package biliparse

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/web"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Bvid      string `json:"bvid"`
		Aid       int    `json:"aid"`
		Videos    int    `json:"videos"`
		Tid       int    `json:"tid"`
		Tname     string `json:"tname"`
		Copyright int    `json:"copyright"`
		Pic       string `json:"pic"`
		Title     string `json:"title"`
		Pubdate   int    `json:"pubdate"`
		Ctime     int    `json:"ctime"`
		Desc      string `json:"desc"`
		DescV2    []struct {
			RawText string `json:"raw_text"`
			Type    int    `json:"type"`
			BizID   int    `json:"biz_id"`
		} `json:"desc_v2"`
		State     int `json:"state"`
		Duration  int `json:"duration"`
		MissionID int `json:"mission_id"`
		Rights    struct {
			Bp            int `json:"bp"`
			Elec          int `json:"elec"`
			Download      int `json:"download"`
			Movie         int `json:"movie"`
			Pay           int `json:"pay"`
			Hd5           int `json:"hd5"`
			NoReprint     int `json:"no_reprint"`
			Autoplay      int `json:"autoplay"`
			UgcPay        int `json:"ugc_pay"`
			IsCooperation int `json:"is_cooperation"`
			UgcPayPreview int `json:"ugc_pay_preview"`
			NoBackground  int `json:"no_background"`
			CleanMode     int `json:"clean_mode"`
			IsSteinGate   int `json:"is_stein_gate"`
			Is360         int `json:"is_360"`
			NoShare       int `json:"no_share"`
			ArcPay        int `json:"arc_pay"`
			FreeWatch     int `json:"free_watch"`
		} `json:"rights"`
		Owner struct {
			Mid  int    `json:"mid"`
			Name string `json:"name"`
			Face string `json:"face"`
		} `json:"owner"`
		Stat struct {
			Aid        int    `json:"aid"`
			View       int    `json:"view"`
			Danmaku    int    `json:"danmaku"`
			Reply      int    `json:"reply"`
			Favorite   int    `json:"favorite"`
			Coin       int    `json:"coin"`
			Share      int    `json:"share"`
			NowRank    int    `json:"now_rank"`
			HisRank    int    `json:"his_rank"`
			Like       int    `json:"like"`
			Dislike    int    `json:"dislike"`
			Evaluation string `json:"evaluation"`
			ArgueMsg   string `json:"argue_msg"`
		} `json:"stat"`
		Dynamic   string `json:"dynamic"`
		Cid       int    `json:"cid"`
		Dimension struct {
			Width  int `json:"width"`
			Height int `json:"height"`
			Rotate int `json:"rotate"`
		} `json:"dimension"`
		Premiere           interface{} `json:"premiere"`
		TeenageMode        int         `json:"teenage_mode"`
		IsChargeableSeason bool        `json:"is_chargeable_season"`
		NoCache            bool        `json:"no_cache"`
		Pages              []struct {
			Cid       int    `json:"cid"`
			Page      int    `json:"page"`
			From      string `json:"from"`
			Part      string `json:"part"`
			Duration  int    `json:"duration"`
			Vid       string `json:"vid"`
			Weblink   string `json:"weblink"`
			Dimension struct {
				Width  int `json:"width"`
				Height int `json:"height"`
				Rotate int `json:"rotate"`
			} `json:"dimension"`
		} `json:"pages"`
		Subtitle struct {
			AllowSubmit bool          `json:"allow_submit"`
			List        []interface{} `json:"list"`
		} `json:"subtitle"`
		Staff []struct {
			Mid   int    `json:"mid"`
			Title string `json:"title"`
			Name  string `json:"name"`
			Face  string `json:"face"`
			Vip   struct {
				Type       int   `json:"type"`
				Status     int   `json:"status"`
				DueDate    int64 `json:"due_date"`
				VipPayType int   `json:"vip_pay_type"`
				ThemeType  int   `json:"theme_type"`
				Label      struct {
					Path                  string `json:"path"`
					Text                  string `json:"text"`
					LabelTheme            string `json:"label_theme"`
					TextColor             string `json:"text_color"`
					BgStyle               int    `json:"bg_style"`
					BgColor               string `json:"bg_color"`
					BorderColor           string `json:"border_color"`
					UseImgLabel           bool   `json:"use_img_label"`
					ImgLabelURIHans       string `json:"img_label_uri_hans"`
					ImgLabelURIHant       string `json:"img_label_uri_hant"`
					ImgLabelURIHansStatic string `json:"img_label_uri_hans_static"`
					ImgLabelURIHantStatic string `json:"img_label_uri_hant_static"`
				} `json:"label"`
				AvatarSubscript    int    `json:"avatar_subscript"`
				NicknameColor      string `json:"nickname_color"`
				Role               int    `json:"role"`
				AvatarSubscriptURL string `json:"avatar_subscript_url"`
				TvVipStatus        int    `json:"tv_vip_status"`
				TvVipPayType       int    `json:"tv_vip_pay_type"`
			} `json:"vip"`
			Official struct {
				Role  int    `json:"role"`
				Title string `json:"title"`
				Desc  string `json:"desc"`
				Type  int    `json:"type"`
			} `json:"official"`
			Follower   int `json:"follower"`
			LabelStyle int `json:"label_style"`
		} `json:"staff"`
		IsSeasonDisplay bool `json:"is_season_display"`
		UserGarb        struct {
			URLImageAniCut string `json:"url_image_ani_cut"`
		} `json:"user_garb"`
		HonorReply struct {
			Honor []struct {
				Aid                int    `json:"aid"`
				Type               int    `json:"type"`
				Desc               string `json:"desc"`
				WeeklyRecommendNum int    `json:"weekly_recommend_num"`
			} `json:"honor"`
		} `json:"honor_reply"`
	} `json:"data"`
}

const (
	api    = "https://api.bilibili.com/x/web-interface/view?"
	origin = "https://www.bilibili.com/video/"
)

var reg = regexp.MustCompile(`https://www.bilibili.com/video/([0-9a-zA-Z]+)`)

func init() {
	en := control.Register("biliparse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "解析BILIBILI视频",
	})
	en.OnRegex(`(av[0-9]+|BV[0-9a-zA-Z]{10}){1}`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			m, err := parse(url)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(m)
		})
	en.OnRegex(`https://www.bilibili.com/video/([0-9a-zA-Z]+)`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			m, err := parse(url)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(m)
		})
	en.OnRegex(`(https://b23.tv/[0-9a-zA-Z]+)`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			realurl, err := getrealurl(url)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			url = realurl
			m, err := parse(url)
			ctx.Send(m)
		})
}

func parse(url string) (m message.Message, err error) {
	var id string
	switch {
	case url[:2] == "av" || url[:2] == "AV":
		id = "aid=" + url[2:]
	case url[:2] == "BV" || url[:2] == "bv":
		id = "bvid=BV" + url[2:]
	}
	data, err := web.GetData(api + id)
	if err != nil {
		return
	}
	var r result
	err = json.Unmarshal(data, &r)
	if err != nil {
		return
	}
	m = message.Message{message.Text("标题:", r.Data.Title),
		message.Text("\nUP主:", r.Data.Owner.Name),
		message.Text("\n播放量:", r.Data.Stat.View),
		message.Text(" · 弹幕数:", r.Data.Stat.Danmaku, "\n"),
		message.Image(r.Data.Pic),
		message.Text("\n点赞:", r.Data.Stat.Like),
		message.Text(",投币:", r.Data.Stat.Coin),
		message.Text("\n收藏:", r.Data.Stat.Favorite),
		message.Text(",分享:", r.Data.Stat.Share),
		message.Text("\n", origin, url)}
	return
}

func getrealurl(fakeurl string) (url string, err error) {
	data, err := http.Get(fakeurl)
	if err != nil {
		return
	}
	url = data.Request.URL.String()
	if !reg.MatchString(url) {
		err = errors.New("invalid video url")
		return
	}
	url = reg.FindStringSubmatch(url)[1]
	return
}
