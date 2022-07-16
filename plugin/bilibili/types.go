package bilibili

const (
	// tURL bilibili动态前缀
	tURL = "https://t.bilibili.com/"
	// dynamicDetailURL 当前动态信息,一个card
	dynamicDetailURL = "https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/get_dynamic_detail?dynamic_id=%v"
	// memberCardURL 个人信息
	memberCardURL = "https://account.bilibili.com/api/member/getCardByMid?mid=%v"
	// articleInfoURL 查看专栏信息
	articleInfoURL = "https://api.bilibili.com/x/article/viewinfo?id=%v"
	// cvURL b站专栏前缀
	cvURL = "https://www.bilibili.com/read/cv"
	// liveRoomInfoURL 查看直播间信息
	liveRoomInfoURL = "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%v"
	// lURL b站直播间前缀
	lURL = "https://live.bilibili.com/"
	// videoInfoURL 查看视频信息
	videoInfoURL = "https://api.bilibili.com/x/web-interface/view?aid=%v&bvid=%v"
	// vURL 视频网址前缀
	vURL = "https://www.bilibili.com/video/"
	// searchUserURL 查找b站用户
	searchUserURL = "http://api.bilibili.com/x/web-interface/search/type?search_type=bili_user&keyword=%v"
	// vtbDetailURL 查找vtb信息
	vtbDetailURL = "https://api.vtbs.moe/v1/detail/%v"
	// medalwallURL 查找牌子
	medalwallURL = "https://api.live.bilibili.com/xlive/web-ucenter/user/MedalWall?target_id=%v"
)

// dynamicCard 总动态结构体,包括desc,card
type dynamicCard struct {
	Desc      Desc   `json:"desc"`
	Card      string `json:"card"`
	Extension struct {
		VoteCfg struct {
			VoteID  int    `json:"vote_id"`
			Desc    string `json:"desc"`
			JoinNum int    `json:"join_num"`
		} `json:"vote_cfg"`
		Vote string `json:"vote"`
	} `json:"extension"`
}

// Card 卡片结构体
type Card struct {
	Item struct {
		Content     string `json:"content"`
		UploadTime  int    `json:"upload_time"`
		Description string `json:"description"`
		Pictures    []struct {
			ImgSrc string `json:"img_src"`
		} `json:"pictures"`
		Timestamp int `json:"timestamp"`
		Cover     struct {
			Default string `json:"default"`
		} `json:"cover"`
		OrigType int `json:"orig_type"`
	} `json:"item"`
	AID             interface{} `json:"aid"`
	BvID            interface{} `json:"bvid"`
	Dynamic         interface{} `json:"dynamic"`
	Pic             string      `json:"pic"`
	Title           string      `json:"title"`
	ID              int         `json:"id"`
	Summary         string      `json:"summary"`
	ImageUrls       []string    `json:"image_urls"`
	OriginImageUrls []string    `json:"origin_image_urls"`
	Sketch          struct {
		Title     string `json:"title"`
		DescText  string `json:"desc_text"`
		CoverURL  string `json:"cover_url"`
		TargetURL string `json:"target_url"`
	} `json:"sketch"`
	Stat struct {
		Aid      int `json:"aid"`
		View     int `json:"view"`
		Danmaku  int `json:"danmaku"`
		Reply    int `json:"reply"`
		Favorite int `json:"favorite"`
		Coin     int `json:"coin"`
		Share    int `json:"share"`
		Like     int `json:"like"`
	} `json:"stat"`
	Stats struct {
		Aid      int `json:"aid"`
		View     int `json:"view"`
		Danmaku  int `json:"danmaku"`
		Reply    int `json:"reply"`
		Favorite int `json:"favorite"`
		Coin     int `json:"coin"`
		Share    int `json:"share"`
		Like     int `json:"like"`
	} `json:"stats"`
	Owner struct {
		Name    string `json:"name"`
		Pubdate int    `json:"pubdate"`
		Mid     int    `json:"mid"`
	} `json:"owner"`
	Cover        string      `json:"cover"`
	ShortID      interface{} `json:"short_id"`
	LivePlayInfo struct {
		ParentAreaName string `json:"parent_area_name"`
		AreaName       string `json:"area_name"`
		Cover          string `json:"cover"`
		Link           string `json:"link"`
		Online         int    `json:"online"`
		RoomID         int    `json:"room_id"`
		LiveStatus     int    `json:"live_status"`
		WatchedShow    string `json:"watched_show"`
		Title          string `json:"title"`
	} `json:"live_play_info"`
	Intro      string      `json:"intro"`
	Schema     string      `json:"schema"`
	Author     interface{} `json:"author"`
	AuthorName string      `json:"author_name"`
	PlayCnt    int         `json:"play_cnt"`
	ReplyCnt   int         `json:"reply_cnt"`
	TypeInfo   string      `json:"type_info"`
	User       struct {
		Name  string `json:"name"`
		Uname string `json:"uname"`
	} `json:"user"`
	Desc          string `json:"desc"`
	ShareSubtitle string `json:"share_subtitle"`
	ShortLink     string `json:"short_link"`
	PublishTime   int    `json:"publish_time"`
	BannerURL     string `json:"banner_url"`
	Ctime         int    `json:"ctime"`
	Vest          struct {
		Content string `json:"content"`
	} `json:"vest"`
	Upper   string `json:"upper"`
	Origin  string `json:"origin"`
	Pubdate int    `json:"pubdate"`
	Rights  struct {
		IsCooperation int `json:"is_cooperation"`
	} `json:"rights"`
	Staff []struct {
		Title    string `json:"title"`
		Name     string `json:"name"`
		Follower int    `json:"follower"`
	} `json:"staff"`
}

// Desc 描述结构体
type Desc struct {
	Type         int    `json:"type"`
	DynamicIDStr string `json:"dynamic_id_str"`
	OrigType     int    `json:"orig_type"`
	Timestamp    int    `json:"timestamp"`
	Origin       struct {
		DynamicIDStr string `json:"dynamic_id_str"`
	} `json:"origin"`
	UserProfile struct {
		Info struct {
			Uname string `json:"uname"`
		} `json:"info"`
	} `json:"user_profile"`
}

// Vote 投票结构体
type Vote struct {
	ChoiceCnt int    `json:"choice_cnt"`
	Desc      string `json:"desc"`
	Endtime   int    `json:"endtime"`
	JoinNum   int    `json:"join_num"`
	Options   []struct {
		Idx    int    `json:"idx"`
		Desc   string `json:"desc"`
		ImgURL string `json:"img_url"`
	} `json:"options"`
}

// memberCard 个人信息卡片
type memberCard struct {
	Mid        string  `json:"mid"`
	Name       string  `json:"name"`
	Sex        string  `json:"sex"`
	Face       string  `json:"face"`
	Coins      float64 `json:"coins"`
	Regtime    int64   `json:"regtime"`
	Birthday   string  `json:"birthday"`
	Sign       string  `json:"sign"`
	Attentions []int64 `json:"attentions"`
	Fans       int     `json:"fans"`
	Friend     int     `json:"friend"`
	Attention  int     `json:"attention"`
	LevelInfo  struct {
		CurrentLevel int `json:"current_level"`
	} `json:"level_info"`
}

// roomCard 直播间卡片
type roomCard struct {
	RoomInfo struct {
		RoomID         int    `json:"room_id"`
		ShortID        int    `json:"short_id"`
		Title          string `json:"title"`
		LiveStatus     int    `json:"live_status"`
		AreaName       string `json:"area_name"`
		ParentAreaName string `json:"parent_area_name"`
		Keyframe       string `json:"keyframe"`
		Online         int    `json:"online"`
	} `json:"room_info"`
	AnchorInfo struct {
		BaseInfo struct {
			Uname string `json:"uname"`
		} `json:"base_info"`
	} `json:"anchor_info"`
}

// searchResult 查找b站用户结果
type searchResult struct {
	Mid    int64  `json:"mid"`
	Uname  string `json:"uname"`
	Gender int64  `json:"gender"`
	Usign  string `json:"usign"`
	Level  int64  `json:"level"`
}

// medalInfo b站牌子信息
type medalInfo struct {
	Mid              int64  `json:"target_id"`
	MedalName        string `json:"medal_name"`
	Level            int64  `json:"level"`
	MedalColorStart  int64  `json:"medal_color_start"`
	MedalColorEnd    int64  `json:"medal_color_end"`
	MedalColorBorder int64  `json:"medal_color_border"`
}

type medal struct {
	Uname     string `json:"target_name"`
	medalInfo `json:"medal_info"`
}

type medalSlice []medal

func (m medalSlice) Len() int {
	return len(m)
}
func (m medalSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func (m medalSlice) Less(i, j int) bool {
	return m[i].Level > m[j].Level
}

// vtb信息
type vtbDetail struct {
	Mid      int    `json:"mid"`
	Uname    string `json:"uname"`
	Video    int    `json:"video"`
	Roomid   int    `json:"roomid"`
	Rise     int    `json:"rise"`
	Follower int    `json:"follower"`
	GuardNum int    `json:"guardNum"`
	AreaRank int    `json:"areaRank"`
}
