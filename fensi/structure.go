package fensi

// 搜索api的json结构体
type search struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Seid           string `json:"seid"`
		Page           int    `json:"page"`
		Pagesize       int    `json:"pagesize"`
		NumResults     int    `json:"numResults"`
		NumPages       int    `json:"numPages"`
		SuggestKeyword string `json:"suggest_keyword"`
		RqtType        string `json:"rqt_type"`
		CostTime       struct {
			ParamsCheck         string `json:"params_check"`
			GetUpuserLiveStatus string `json:"get upuser live status"`
			IllegalHandler      string `json:"illegal_handler"`
			AsResponseFormat    string `json:"as_response_format"`
			AsRequest           string `json:"as_request"`
			SaveCache           string `json:"save_cache"`
			DeserializeResponse string `json:"deserialize_response"`
			AsRequestFormat     string `json:"as_request_format"`
			Total               string `json:"total"`
			MainHandler         string `json:"main_handler"`
		} `json:"cost_time"`
		ExpList interface{} `json:"exp_list"`
		EggHit  int         `json:"egg_hit"`
		Result  []struct {
			Type       string `json:"type"`
			Mid        int    `json:"mid"`
			Uname      string `json:"uname"`
			Usign      string `json:"usign"`
			Fans       int    `json:"fans"`
			Videos     int    `json:"videos"`
			Upic       string `json:"upic"`
			VerifyInfo string `json:"verify_info"`
			Level      int    `json:"level"`
			Gender     int    `json:"gender"`
			IsUpuser   int    `json:"is_upuser"`
			IsLive     int    `json:"is_live"`
			RoomID     int    `json:"room_id"`
			Res        []struct {
				Aid          int    `json:"aid"`
				Bvid         string `json:"bvid"`
				Title        string `json:"title"`
				Pubdate      int    `json:"pubdate"`
				Arcurl       string `json:"arcurl"`
				Pic          string `json:"pic"`
				Play         string `json:"play"`
				Dm           int    `json:"dm"`
				Coin         int    `json:"coin"`
				Fav          int    `json:"fav"`
				Desc         string `json:"desc"`
				Duration     string `json:"duration"`
				IsPay        int    `json:"is_pay"`
				IsUnionVideo int    `json:"is_union_video"`
			} `json:"res"`
			OfficialVerify struct {
				Type int    `json:"type"`
				Desc string `json:"desc"`
			} `json:"official_verify"`
			HitColumns []interface{} `json:"hit_columns"`
		} `json:"result"`
		ShowColumn int `json:"show_column"`
	} `json:"data"`
}

// 账号信息api的json结构体
type accInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Mid       int    `json:"mid"`
		Name      string `json:"name"`
		Sex       string `json:"sex"`
		Face      string `json:"face"`
		Sign      string `json:"sign"`
		Rank      int    `json:"rank"`
		Level     int    `json:"level"`
		Jointime  int    `json:"jointime"`
		Moral     int    `json:"moral"`
		Silence   int    `json:"silence"`
		Birthday  string `json:"birthday"`
		Coins     int    `json:"coins"`
		FansBadge bool   `json:"fans_badge"`
		Official  struct {
			Role  int    `json:"role"`
			Title string `json:"title"`
			Desc  string `json:"desc"`
			Type  int    `json:"type"`
		} `json:"official"`
		Vip struct {
			Type       int   `json:"type"`
			Status     int   `json:"status"`
			DueDate    int64 `json:"due_date"`
			VipPayType int   `json:"vip_pay_type"`
			ThemeType  int   `json:"theme_type"`
			Label      struct {
				Path        string `json:"path"`
				Text        string `json:"text"`
				LabelTheme  string `json:"label_theme"`
				TextColor   string `json:"text_color"`
				BgStyle     int    `json:"bg_style"`
				BgColor     string `json:"bg_color"`
				BorderColor string `json:"border_color"`
			} `json:"label"`
			AvatarSubscript    int    `json:"avatar_subscript"`
			NicknameColor      string `json:"nickname_color"`
			Role               int    `json:"role"`
			AvatarSubscriptURL string `json:"avatar_subscript_url"`
		} `json:"vip"`
		Pendant struct {
			Pid               int    `json:"pid"`
			Name              string `json:"name"`
			Image             string `json:"image"`
			Expire            int    `json:"expire"`
			ImageEnhance      string `json:"image_enhance"`
			ImageEnhanceFrame string `json:"image_enhance_frame"`
		} `json:"pendant"`
		Nameplate struct {
			Nid        int    `json:"nid"`
			Name       string `json:"name"`
			Image      string `json:"image"`
			ImageSmall string `json:"image_small"`
			Level      string `json:"level"`
			Condition  string `json:"condition"`
		} `json:"nameplate"`
		UserHonourInfo struct {
			Mid    int         `json:"mid"`
			Colour interface{} `json:"colour"`
			Tags   interface{} `json:"tags"`
		} `json:"user_honour_info"`
		IsFollowed bool   `json:"is_followed"`
		TopPhoto   string `json:"top_photo"`
		Theme      struct {
		} `json:"theme"`
		SysNotice struct {
		} `json:"sys_notice"`
		LiveRoom struct {
			RoomStatus    int    `json:"roomStatus"`
			LiveStatus    int    `json:"liveStatus"`
			URL           string `json:"url"`
			Title         string `json:"title"`
			Cover         string `json:"cover"`
			Online        int    `json:"online"`
			Roomid        int    `json:"roomid"`
			RoundStatus   int    `json:"roundStatus"`
			BroadcastType int    `json:"broadcast_type"`
		} `json:"live_room"`
	} `json:"data"`
}

//共同关注api的json结构体
type followings struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		List []struct {
			Mid          int         `json:"mid"`
			Attribute    int         `json:"attribute"`
			Mtime        int         `json:"mtime"`
			Tag          interface{} `json:"tag"`
			Special      int         `json:"special"`
			ContractInfo struct {
				IsContractor bool `json:"is_contractor"`
				Ts           int  `json:"ts"`
				IsContract   bool `json:"is_contract"`
			} `json:"contract_info"`
			Uname          string `json:"uname"`
			Face           string `json:"face"`
			Sign           string `json:"sign"`
			OfficialVerify struct {
				Type int    `json:"type"`
				Desc string `json:"desc"`
			} `json:"official_verify"`
			Vip struct {
				VipType       int    `json:"vipType"`
				VipDueDate    int64  `json:"vipDueDate"`
				DueRemark     string `json:"dueRemark"`
				AccessStatus  int    `json:"accessStatus"`
				VipStatus     int    `json:"vipStatus"`
				VipStatusWarn string `json:"vipStatusWarn"`
				ThemeType     int    `json:"themeType"`
				Label         struct {
					Path        string `json:"path"`
					Text        string `json:"text"`
					LabelTheme  string `json:"label_theme"`
					TextColor   string `json:"text_color"`
					BgStyle     int    `json:"bg_style"`
					BgColor     string `json:"bg_color"`
					BorderColor string `json:"border_color"`
				} `json:"label"`
				AvatarSubscript    int    `json:"avatar_subscript"`
				NicknameColor      string `json:"nickname_color"`
				AvatarSubscriptURL string `json:"avatar_subscript_url"`
			} `json:"vip"`
		} `json:"list"`
		ReVersion int64 `json:"re_version"`
		Total     int   `json:"total"`
	} `json:"data"`
}

// 粉丝信息api的json结构体
type follower struct {
	Mid         int    `json:"mid"`
	UUID        string `json:"uuid"`
	Uname       string `json:"uname"`
	Video       int    `json:"video"`
	Roomid      int    `json:"roomid"`
	Sign        string `json:"sign"`
	Notice      string `json:"notice"`
	Face        string `json:"face"`
	Rise        int    `json:"rise"`
	TopPhoto    string `json:"topPhoto"`
	ArchiveView int    `json:"archiveView"`
	Follower    int    `json:"follower"`
	LiveStatus  int    `json:"liveStatus"`
	RecordNum   int    `json:"recordNum"`
	GuardNum    int    `json:"guardNum"`
	LastLive    struct {
		Online int   `json:"online"`
		Time   int64 `json:"time"`
	} `json:"lastLive"`
	GuardChange   int    `json:"guardChange"`
	GuardType     []int  `json:"guardType"`
	AreaRank      int    `json:"areaRank"`
	Online        int    `json:"online"`
	Title         string `json:"title"`
	Time          int64  `json:"time"`
	LiveStartTime int    `json:"liveStartTime"`
}
