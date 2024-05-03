package guessmusic

// config内容
type config struct {
	MusicPath   string    `json:"musicPath"`
	APIURL      string    `json:"apiURL"`
	Playlist    []listRaw `json:"playlist"`
	Defaultlist []dlist   `json:"defaultlist"`
	Local       bool      `json:"local"`
	API         bool      `json:"api"`
	Cookie      string    `json:"cookie"`
}

// 记录歌单绑定的网易云歌单ID
type listRaw struct {
	Name string `json:"name"` // 歌单名称
	ID   int64  `json:"id"`   // 歌单绑定的网易云ID
}

// 记录群默认猜歌
type dlist struct {
	GroupID int64  `json:"gid"`  // 群号
	Name    string `json:"name"` // 歌单名称
}

// 本地歌单列表信息
type listinfo struct {
	Name   string `json:"name"` // 歌单名称
	Number int    // 歌曲数量
	ID     int64  // 歌单绑定的歌曲ID
}

/*****************************************************************/
/***************NeteaseCloudMusicApi框架API************************/
/*****************************************************************/
// 获取登陆信息
type keyInfo struct {
	Data struct {
		Code   int    `json:"code"`
		Unikey string `json:"unikey"`
	} `json:"data"`
	Code int `json:"code"`
}
type cookyInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cookie  string `json:"cookie"`
}
type qrInfo struct {
	Code int `json:"code"`
	Data struct {
		Qrurl string `json:"qrurl"`
		Qrimg string `json:"qrimg"`
	} `json:"data"`
}

// 获取歌单信息
type listInfoOfAPI struct {
	Code          int         `json:"code"`
	RelatedVideos interface{} `json:"relatedVideos"`
	Playlist      struct {
		ID                    int64       `json:"id"`
		Name                  string      `json:"name"`
		CoverImgID            int64       `json:"coverImgId"`
		CoverImgURL           string      `json:"coverImgUrl"`
		CoverImgIDStr         string      `json:"coverImgId_str"`
		AdType                int         `json:"adType"`
		UserID                int         `json:"userId"`
		CreateTime            int64       `json:"createTime"`
		Status                int         `json:"status"`
		OpRecommend           bool        `json:"opRecommend"`
		HighQuality           bool        `json:"highQuality"`
		NewImported           bool        `json:"newImported"`
		UpdateTime            int64       `json:"updateTime"`
		TrackCount            int         `json:"trackCount"`
		SpecialType           int         `json:"specialType"`
		Privacy               int         `json:"privacy"`
		TrackUpdateTime       int64       `json:"trackUpdateTime"`
		CommentThreadID       string      `json:"commentThreadId"`
		PlayCount             int         `json:"playCount"`
		TrackNumberUpdateTime int64       `json:"trackNumberUpdateTime"`
		SubscribedCount       int         `json:"subscribedCount"`
		CloudTrackCount       int         `json:"cloudTrackCount"`
		Ordered               bool        `json:"ordered"`
		Description           string      `json:"description"`
		Tags                  []string    `json:"tags"`
		UpdateFrequency       interface{} `json:"updateFrequency"`
		BackgroundCoverID     int         `json:"backgroundCoverId"`
		BackgroundCoverURL    interface{} `json:"backgroundCoverUrl"`
		TitleImage            int         `json:"titleImage"`
		TitleImageURL         interface{} `json:"titleImageUrl"`
		EnglishTitle          interface{} `json:"englishTitle"`
		OfficialPlaylistType  interface{} `json:"officialPlaylistType"`
		Subscribers           []struct {
			DefaultAvatar       bool        `json:"defaultAvatar"`
			Province            int         `json:"province"`
			AuthStatus          int         `json:"authStatus"`
			Followed            bool        `json:"followed"`
			AvatarURL           string      `json:"avatarUrl"`
			AccountStatus       int         `json:"accountStatus"`
			Gender              int         `json:"gender"`
			City                int         `json:"city"`
			Birthday            int         `json:"birthday"`
			UserID              int         `json:"userId"`
			UserType            int         `json:"userType"`
			Nickname            string      `json:"nickname"`
			Signature           string      `json:"signature"`
			Description         string      `json:"description"`
			DetailDescription   string      `json:"detailDescription"`
			AvatarImgID         int64       `json:"avatarImgId"`
			BackgroundImgID     int64       `json:"backgroundImgId"`
			BackgroundURL       string      `json:"backgroundUrl"`
			Authority           int         `json:"authority"`
			Mutual              bool        `json:"mutual"`
			ExpertTags          interface{} `json:"expertTags"`
			Experts             interface{} `json:"experts"`
			DjStatus            int         `json:"djStatus"`
			VipType             int         `json:"vipType"`
			RemarkName          interface{} `json:"remarkName"`
			AuthenticationTypes int         `json:"authenticationTypes"`
			AvatarDetail        interface{} `json:"avatarDetail"`
			Anchor              bool        `json:"anchor"`
			BackgroundImgIDStr  string      `json:"backgroundImgIdStr"`
			AvatarImgIDStr      string      `json:"avatarImgIdStr"`
			AvatarImgIDString   string      `json:"AvatarImgIDString"`
		} `json:"subscribers"`
		Subscribed interface{} `json:"subscribed"`
		Creator    struct {
			DefaultAvatar       bool        `json:"defaultAvatar"`
			Province            int         `json:"province"`
			AuthStatus          int         `json:"authStatus"`
			Followed            bool        `json:"followed"`
			AvatarURL           string      `json:"avatarUrl"`
			AccountStatus       int         `json:"accountStatus"`
			Gender              int         `json:"gender"`
			City                int         `json:"city"`
			Birthday            int         `json:"birthday"`
			UserID              int         `json:"userId"`
			UserType            int         `json:"userType"`
			Nickname            string      `json:"nickname"`
			Signature           string      `json:"signature"`
			Description         string      `json:"description"`
			DetailDescription   string      `json:"detailDescription"`
			AvatarImgID         int64       `json:"avatarImgId"`
			BackgroundImgID     int64       `json:"backgroundImgId"`
			BackgroundURL       string      `json:"backgroundUrl"`
			Authority           int         `json:"authority"`
			Mutual              bool        `json:"mutual"`
			ExpertTags          interface{} `json:"expertTags"`
			Experts             interface{} `json:"experts"`
			DjStatus            int         `json:"djStatus"`
			VipType             int         `json:"vipType"`
			RemarkName          interface{} `json:"remarkName"`
			AuthenticationTypes int         `json:"authenticationTypes"`
			AvatarDetail        struct {
				UserType        int    `json:"userType"`
				IdentityLevel   int    `json:"identityLevel"`
				IdentityIconURL string `json:"identityIconUrl"`
			} `json:"avatarDetail"`
			Anchor             bool   `json:"anchor"`
			BackgroundImgIDStr string `json:"backgroundImgIdStr"`
			AvatarImgIDStr     string `json:"avatarImgIdStr"`
			AvatarImgIDString  string `json:"AvatarImgIDString"`
		} `json:"creator"`
		Tracks []struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
			Pst  int    `json:"pst"`
			T    int    `json:"t"`
			Ar   []struct {
				ID    int           `json:"id"`
				Name  string        `json:"name"`
				Tns   []interface{} `json:"tns"`
				Alias []interface{} `json:"alias"`
			} `json:"ar"`
			Alia []string    `json:"alia"`
			Pop  int         `json:"pop"`
			St   int         `json:"st"`
			Rt   string      `json:"rt"`
			Fee  int         `json:"fee"`
			V    int         `json:"v"`
			Crbt interface{} `json:"crbt"`
			Cf   string      `json:"cf"`
			Al   struct {
				ID     int           `json:"id"`
				Name   string        `json:"name"`
				PicURL string        `json:"picUrl"`
				Tns    []interface{} `json:"tns"`
				PicStr string        `json:"pic_str"`
				Pic    int64         `json:"pic"`
			} `json:"al"`
			Dt int `json:"dt"`
			H  struct {
				Br   int     `json:"br"`
				Fid  int     `json:"fid"`
				Size int     `json:"size"`
				Vd   float64 `json:"vd"`
				Sr   int     `json:"sr"`
			} `json:"h"`
			M struct {
				Br   int     `json:"br"`
				Fid  int     `json:"fid"`
				Size int     `json:"size"`
				Vd   float64 `json:"vd"`
				Sr   int     `json:"sr"`
			} `json:"m"`
			L struct {
				Br   int     `json:"br"`
				Fid  int     `json:"fid"`
				Size int     `json:"size"`
				Vd   float64 `json:"vd"`
				Sr   int     `json:"sr"`
			} `json:"l"`
			Sq                   interface{}   `json:"sq"`
			Hr                   interface{}   `json:"hr"`
			A                    interface{}   `json:"a"`
			Cd                   string        `json:"cd"`
			No                   int           `json:"no"`
			RtURL                interface{}   `json:"rtUrl"`
			Ftype                int           `json:"ftype"`
			RtUrls               []interface{} `json:"rtUrls"`
			DjID                 int           `json:"djId"`
			Copyright            int           `json:"copyright"`
			SID                  int           `json:"s_id"`
			Mark                 int           `json:"mark"`
			OriginCoverType      int           `json:"originCoverType"`
			OriginSongSimpleData interface{}   `json:"originSongSimpleData"`
			TagPicList           interface{}   `json:"tagPicList"`
			ResourceState        bool          `json:"resourceState"`
			Version              int           `json:"version"`
			SongJumpInfo         interface{}   `json:"songJumpInfo"`
			EntertainmentTags    interface{}   `json:"entertainmentTags"`
			Single               int           `json:"single"`
			NoCopyrightRcmd      interface{}   `json:"noCopyrightRcmd"`
			Alg                  interface{}   `json:"alg"`
			Rtype                int           `json:"rtype"`
			Rurl                 interface{}   `json:"rurl"`
			Mst                  int           `json:"mst"`
			Cp                   int           `json:"cp"`
			Mv                   int           `json:"mv"`
			PublishTime          int64         `json:"publishTime"`
			Tns                  []string      `json:"tns,omitempty"`
		} `json:"tracks"`
		VideoIDs interface{} `json:"videoIds"`
		Videos   interface{} `json:"videos"`
		TrackIDs []struct {
			ID         int         `json:"id"`
			V          int         `json:"v"`
			T          int         `json:"t"`
			At         int64       `json:"at"`
			Alg        interface{} `json:"alg"`
			UID        int         `json:"uid"`
			RcmdReason string      `json:"rcmdReason"`
			Sc         interface{} `json:"sc"`
			Lr         int         `json:"lr,omitempty"`
		} `json:"trackIds"`
		ShareCount         int         `json:"shareCount"`
		CommentCount       int         `json:"commentCount"`
		RemixVideo         interface{} `json:"remixVideo"`
		SharedUsers        interface{} `json:"sharedUsers"`
		HistorySharedUsers interface{} `json:"historySharedUsers"`
		GradeStatus        string      `json:"gradeStatus"`
		Score              interface{} `json:"score"`
		AlgTags            interface{} `json:"algTags"`
	} `json:"playlist"`
	Urls       interface{} `json:"urls"`
	Privileges []struct {
		ID                 int         `json:"id"`
		Fee                int         `json:"fee"`
		Payed              int         `json:"payed"`
		RealPayed          int         `json:"realPayed"`
		St                 int         `json:"st"`
		Pl                 int         `json:"pl"`
		Dl                 int         `json:"dl"`
		Sp                 int         `json:"sp"`
		Cp                 int         `json:"cp"`
		Subp               int         `json:"subp"`
		Cs                 bool        `json:"cs"`
		Maxbr              int         `json:"maxbr"`
		Fl                 int         `json:"fl"`
		Pc                 interface{} `json:"pc"`
		Toast              bool        `json:"toast"`
		Flag               int         `json:"flag"`
		PaidBigBang        bool        `json:"paidBigBang"`
		PreSell            bool        `json:"preSell"`
		PlayMaxbr          int         `json:"playMaxbr"`
		DownloadMaxbr      int         `json:"downloadMaxbr"`
		MaxBrLevel         string      `json:"maxBrLevel"`
		PlayMaxBrLevel     string      `json:"playMaxBrLevel"`
		DownloadMaxBrLevel string      `json:"downloadMaxBrLevel"`
		PlLevel            string      `json:"plLevel"`
		DlLevel            string      `json:"dlLevel"`
		FlLevel            string      `json:"flLevel"`
		Rscl               int         `json:"rscl"`
		FreeTrialPrivilege struct {
			ResConsumable  bool        `json:"resConsumable"`
			UserConsumable bool        `json:"userConsumable"`
			ListenType     interface{} `json:"listenType"`
		} `json:"freeTrialPrivilege"`
		ChargeInfoList []struct {
			Rate          int         `json:"rate"`
			ChargeURL     interface{} `json:"chargeUrl"`
			ChargeMessage interface{} `json:"chargeMessage"`
			ChargeType    int         `json:"chargeType"`
		} `json:"chargeInfoList"`
	} `json:"privileges"`
	SharedPrivilege interface{} `json:"sharedPrivilege"`
	ResEntrance     interface{} `json:"resEntrance"`
}

// 获取歌单列表
type musicListOfApI struct {
	Songs []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
		Pst  int    `json:"pst"`
		T    int    `json:"t"`
		Ar   []struct {
			ID    int           `json:"id"`
			Name  string        `json:"name"`
			Tns   []interface{} `json:"tns"`
			Alias []interface{} `json:"alias"`
		} `json:"ar"`
		Alia []string    `json:"alia"`
		Pop  int         `json:"pop"`
		St   int         `json:"st"`
		Rt   string      `json:"rt"`
		Fee  int         `json:"fee"`
		V    int         `json:"v"`
		Crbt interface{} `json:"crbt"`
		Cf   string      `json:"cf"`
		Al   struct {
			ID     int           `json:"id"`
			Name   string        `json:"name"`
			PicURL string        `json:"picUrl"`
			Tns    []interface{} `json:"tns"`
			PicStr string        `json:"pic_str"`
			Pic    int64         `json:"pic"`
		} `json:"al"`
		Dt int `json:"dt"`
		H  struct {
			Br   int     `json:"br"`
			Fid  int     `json:"fid"`
			Size int     `json:"size"`
			Vd   float32 `json:"vd"`
			Sr   int     `json:"sr"`
		} `json:"h"`
		M struct {
			Br   int     `json:"br"`
			Fid  int     `json:"fid"`
			Size int     `json:"size"`
			Vd   float32 `json:"vd"`
			Sr   int     `json:"sr"`
		} `json:"m"`
		L struct {
			Br   int     `json:"br"`
			Fid  int     `json:"fid"`
			Size int     `json:"size"`
			Vd   float32 `json:"vd"`
			Sr   int     `json:"sr"`
		} `json:"l"`
		Sq                   interface{}   `json:"sq"`
		Hr                   interface{}   `json:"hr"`
		A                    interface{}   `json:"a"`
		Cd                   string        `json:"cd"`
		No                   int           `json:"no"`
		RtURL                interface{}   `json:"rtUrl"`
		Ftype                int           `json:"ftype"`
		RtUrls               []interface{} `json:"rtUrls"`
		DjID                 int           `json:"djId"`
		Copyright            int           `json:"copyright"`
		SID                  int           `json:"s_id"`
		Mark                 int           `json:"mark"`
		OriginCoverType      int           `json:"originCoverType"`
		OriginSongSimpleData interface{}   `json:"originSongSimpleData"`
		TagPicList           interface{}   `json:"tagPicList"`
		ResourceState        bool          `json:"resourceState"`
		Version              int           `json:"version"`
		SongJumpInfo         interface{}   `json:"songJumpInfo"`
		EntertainmentTags    interface{}   `json:"entertainmentTags"`
		AwardTags            interface{}   `json:"awardTags"`
		Single               int           `json:"single"`
		NoCopyrightRcmd      interface{}   `json:"noCopyrightRcmd"`
		Rtype                int           `json:"rtype"`
		Rurl                 interface{}   `json:"rurl"`
		Mst                  int           `json:"mst"`
		Cp                   int           `json:"cp"`
		Mv                   int           `json:"mv"`
		PublishTime          int64         `json:"publishTime"`
		Tns                  []string      `json:"tns,omitempty"`
	} `json:"songs"`
	Privileges []struct {
		ID                 int    `json:"id"`
		Fee                int    `json:"fee"`
		Payed              int    `json:"payed"`
		St                 int    `json:"st"`
		Pl                 int    `json:"pl"`
		Dl                 int    `json:"dl"`
		Sp                 int    `json:"sp"`
		Cp                 int    `json:"cp"`
		Subp               int    `json:"subp"`
		Cs                 bool   `json:"cs"`
		Maxbr              int    `json:"maxbr"`
		Fl                 int    `json:"fl"`
		Toast              bool   `json:"toast"`
		Flag               int    `json:"flag"`
		PreSell            bool   `json:"preSell"`
		PlayMaxbr          int    `json:"playMaxbr"`
		DownloadMaxbr      int    `json:"downloadMaxbr"`
		MaxBrLevel         string `json:"maxBrLevel"`
		PlayMaxBrLevel     string `json:"playMaxBrLevel"`
		DownloadMaxBrLevel string `json:"downloadMaxBrLevel"`
		PlLevel            string `json:"plLevel"`
		DlLevel            string `json:"dlLevel"`
		FlLevel            string `json:"flLevel"`
		Rscl               int    `json:"rscl"`
		FreeTrialPrivilege struct {
			ResConsumable  bool        `json:"resConsumable"`
			UserConsumable bool        `json:"userConsumable"`
			ListenType     interface{} `json:"listenType"`
		} `json:"freeTrialPrivilege"`
		ChargeInfoList []struct {
			Rate          int         `json:"rate"`
			ChargeURL     interface{} `json:"chargeUrl"`
			ChargeMessage interface{} `json:"chargeMessage"`
			ChargeType    int         `json:"chargeType"`
		} `json:"chargeInfoList"`
	} `json:"privileges"`
	Code int `json:"code"`
}

/*****************************************************************/
/*********************独角兽API随机抽歌信息**************************/
/*****************************************************************/
// 独角兽API随机抽歌信息
type ovooaData struct {
	Code int    `json:"code"`
	Text string `json:"text"`
	Data struct {
		Song   string `json:"song"`
		Singer string `json:"singer"`
		Cover  string `json:"cover"`
		Music  string `json:"Music"`
		ID     int    `json:"id"`
	} `json:"data"`
}
