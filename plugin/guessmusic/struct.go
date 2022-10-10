package guessmusic

type listRaw struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type config struct {
	MusicPath string    `json:"musicPath"`
	Local     bool      `json:"local"`
	API       bool      `json:"api"`
	Cookie    string    `json:"cookie"`
	Playlist  []listRaw `json:"playlist"`
}

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
type topList struct {
	Code          int `json:"code"`
	RelatedVideos any `json:"relatedVideos"`
	Playlist      struct {
		ID                    int64    `json:"id"`
		Name                  string   `json:"name"`
		CoverImgID            int64    `json:"coverImgId"`
		CoverImgURL           string   `json:"coverImgUrl"`
		CoverImgIDStr         string   `json:"coverImgId_str"`
		AdType                int      `json:"adType"`
		UserID                int      `json:"userId"`
		CreateTime            int64    `json:"createTime"`
		Status                int      `json:"status"`
		OpRecommend           bool     `json:"opRecommend"`
		HighQuality           bool     `json:"highQuality"`
		NewImported           bool     `json:"newImported"`
		UpdateTime            int64    `json:"updateTime"`
		TrackCount            int      `json:"trackCount"`
		SpecialType           int      `json:"specialType"`
		Privacy               int      `json:"privacy"`
		TrackUpdateTime       int64    `json:"trackUpdateTime"`
		CommentThreadID       string   `json:"commentThreadId"`
		PlayCount             int      `json:"playCount"`
		TrackNumberUpdateTime int64    `json:"trackNumberUpdateTime"`
		SubscribedCount       int      `json:"subscribedCount"`
		CloudTrackCount       int      `json:"cloudTrackCount"`
		Ordered               bool     `json:"ordered"`
		Description           string   `json:"description"`
		Tags                  []string `json:"tags"`
		UpdateFrequency       any      `json:"updateFrequency"`
		BackgroundCoverID     int      `json:"backgroundCoverId"`
		BackgroundCoverURL    any      `json:"backgroundCoverUrl"`
		TitleImage            int      `json:"titleImage"`
		TitleImageURL         any      `json:"titleImageUrl"`
		EnglishTitle          any      `json:"englishTitle"`
		OfficialPlaylistType  any      `json:"officialPlaylistType"`
		Subscribers           []struct {
			DefaultAvatar       bool   `json:"defaultAvatar"`
			Province            int    `json:"province"`
			AuthStatus          int    `json:"authStatus"`
			Followed            bool   `json:"followed"`
			AvatarURL           string `json:"avatarUrl"`
			AccountStatus       int    `json:"accountStatus"`
			Gender              int    `json:"gender"`
			City                int    `json:"city"`
			Birthday            int    `json:"birthday"`
			UserID              int    `json:"userId"`
			UserType            int    `json:"userType"`
			Nickname            string `json:"nickname"`
			Signature           string `json:"signature"`
			Description         string `json:"description"`
			DetailDescription   string `json:"detailDescription"`
			AvatarImgID         int64  `json:"avatarImgId"`
			BackgroundImgID     int64  `json:"backgroundImgId"`
			BackgroundURL       string `json:"backgroundUrl"`
			Authority           int    `json:"authority"`
			Mutual              bool   `json:"mutual"`
			ExpertTags          any    `json:"expertTags"`
			Experts             any    `json:"experts"`
			DjStatus            int    `json:"djStatus"`
			VipType             int    `json:"vipType"`
			RemarkName          any    `json:"remarkName"`
			AuthenticationTypes int    `json:"authenticationTypes"`
			AvatarDetail        any    `json:"avatarDetail"`
			Anchor              bool   `json:"anchor"`
			BackgroundImgIDStr  string `json:"backgroundImgIdStr"`
			AvatarImgIDStr      string `json:"avatarImgIdStr"`
			AvatarImgIDString   string `json:"AvatarImgIDString"`
		} `json:"subscribers"`
		Subscribed any `json:"subscribed"`
		Creator    struct {
			DefaultAvatar       bool   `json:"defaultAvatar"`
			Province            int    `json:"province"`
			AuthStatus          int    `json:"authStatus"`
			Followed            bool   `json:"followed"`
			AvatarURL           string `json:"avatarUrl"`
			AccountStatus       int    `json:"accountStatus"`
			Gender              int    `json:"gender"`
			City                int    `json:"city"`
			Birthday            int    `json:"birthday"`
			UserID              int    `json:"userId"`
			UserType            int    `json:"userType"`
			Nickname            string `json:"nickname"`
			Signature           string `json:"signature"`
			Description         string `json:"description"`
			DetailDescription   string `json:"detailDescription"`
			AvatarImgID         int64  `json:"avatarImgId"`
			BackgroundImgID     int64  `json:"backgroundImgId"`
			BackgroundURL       string `json:"backgroundUrl"`
			Authority           int    `json:"authority"`
			Mutual              bool   `json:"mutual"`
			ExpertTags          any    `json:"expertTags"`
			Experts             any    `json:"experts"`
			DjStatus            int    `json:"djStatus"`
			VipType             int    `json:"vipType"`
			RemarkName          any    `json:"remarkName"`
			AuthenticationTypes int    `json:"authenticationTypes"`
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
				ID    int    `json:"id"`
				Name  string `json:"name"`
				Tns   []any  `json:"tns"`
				Alias []any  `json:"alias"`
			} `json:"ar"`
			Alia []string `json:"alia"`
			Pop  int      `json:"pop"`
			St   int      `json:"st"`
			Rt   string   `json:"rt"`
			Fee  int      `json:"fee"`
			V    int      `json:"v"`
			Crbt any      `json:"crbt"`
			Cf   string   `json:"cf"`
			Al   struct {
				ID     int    `json:"id"`
				Name   string `json:"name"`
				PicURL string `json:"picUrl"`
				Tns    []any  `json:"tns"`
				PicStr string `json:"pic_str"`
				Pic    int64  `json:"pic"`
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
			Sq                   any      `json:"sq"`
			Hr                   any      `json:"hr"`
			A                    any      `json:"a"`
			Cd                   string   `json:"cd"`
			No                   int      `json:"no"`
			RtURL                any      `json:"rtUrl"`
			Ftype                int      `json:"ftype"`
			RtUrls               []any    `json:"rtUrls"`
			DjID                 int      `json:"djId"`
			Copyright            int      `json:"copyright"`
			SID                  int      `json:"s_id"`
			Mark                 int      `json:"mark"`
			OriginCoverType      int      `json:"originCoverType"`
			OriginSongSimpleData any      `json:"originSongSimpleData"`
			TagPicList           any      `json:"tagPicList"`
			ResourceState        bool     `json:"resourceState"`
			Version              int      `json:"version"`
			SongJumpInfo         any      `json:"songJumpInfo"`
			EntertainmentTags    any      `json:"entertainmentTags"`
			Single               int      `json:"single"`
			NoCopyrightRcmd      any      `json:"noCopyrightRcmd"`
			Alg                  any      `json:"alg"`
			Rtype                int      `json:"rtype"`
			Rurl                 any      `json:"rurl"`
			Mst                  int      `json:"mst"`
			Cp                   int      `json:"cp"`
			Mv                   int      `json:"mv"`
			PublishTime          int64    `json:"publishTime"`
			Tns                  []string `json:"tns,omitempty"`
		} `json:"tracks"`
		VideoIds any `json:"videoIds"`
		Videos   any `json:"videos"`
		TrackIds []struct {
			ID         int    `json:"id"`
			V          int    `json:"v"`
			T          int    `json:"t"`
			At         int64  `json:"at"`
			Alg        any    `json:"alg"`
			UID        int    `json:"uid"`
			RcmdReason string `json:"rcmdReason"`
			Sc         any    `json:"sc"`
			Lr         int    `json:"lr,omitempty"`
		} `json:"trackIds"`
		ShareCount         int    `json:"shareCount"`
		CommentCount       int    `json:"commentCount"`
		RemixVideo         any    `json:"remixVideo"`
		SharedUsers        any    `json:"sharedUsers"`
		HistorySharedUsers any    `json:"historySharedUsers"`
		GradeStatus        string `json:"gradeStatus"`
		Score              any    `json:"score"`
		AlgTags            any    `json:"algTags"`
	} `json:"playlist"`
	Urls       any `json:"urls"`
	Privileges []struct {
		ID                 int    `json:"id"`
		Fee                int    `json:"fee"`
		Payed              int    `json:"payed"`
		RealPayed          int    `json:"realPayed"`
		St                 int    `json:"st"`
		Pl                 int    `json:"pl"`
		Dl                 int    `json:"dl"`
		Sp                 int    `json:"sp"`
		Cp                 int    `json:"cp"`
		Subp               int    `json:"subp"`
		Cs                 bool   `json:"cs"`
		Maxbr              int    `json:"maxbr"`
		Fl                 int    `json:"fl"`
		Pc                 any    `json:"pc"`
		Toast              bool   `json:"toast"`
		Flag               int    `json:"flag"`
		PaidBigBang        bool   `json:"paidBigBang"`
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
			ResConsumable  bool `json:"resConsumable"`
			UserConsumable bool `json:"userConsumable"`
			ListenType     any  `json:"listenType"`
		} `json:"freeTrialPrivilege"`
		ChargeInfoList []struct {
			Rate          int `json:"rate"`
			ChargeURL     any `json:"chargeUrl"`
			ChargeMessage any `json:"chargeMessage"`
			ChargeType    int `json:"chargeType"`
		} `json:"chargeInfoList"`
	} `json:"privileges"`
	SharedPrivilege any `json:"sharedPrivilege"`
	ResEntrance     any `json:"resEntrance"`
}

type topMusicInfo struct {
	Songs []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
		Pst  int    `json:"pst"`
		T    int    `json:"t"`
		Ar   []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Tns   []any  `json:"tns"`
			Alias []any  `json:"alias"`
		} `json:"ar"`
		Alia []string `json:"alia"`
		Pop  int      `json:"pop"`
		St   int      `json:"st"`
		Rt   string   `json:"rt"`
		Fee  int      `json:"fee"`
		V    int      `json:"v"`
		Crbt any      `json:"crbt"`
		Cf   string   `json:"cf"`
		Al   struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			PicURL string `json:"picUrl"`
			Tns    []any  `json:"tns"`
			PicStr string `json:"pic_str"`
			Pic    int64  `json:"pic"`
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
		Sq                   any      `json:"sq"`
		Hr                   any      `json:"hr"`
		A                    any      `json:"a"`
		Cd                   string   `json:"cd"`
		No                   int      `json:"no"`
		RtURL                any      `json:"rtUrl"`
		Ftype                int      `json:"ftype"`
		RtUrls               []any    `json:"rtUrls"`
		DjID                 int      `json:"djId"`
		Copyright            int      `json:"copyright"`
		SID                  int      `json:"s_id"`
		Mark                 int      `json:"mark"`
		OriginCoverType      int      `json:"originCoverType"`
		OriginSongSimpleData any      `json:"originSongSimpleData"`
		TagPicList           any      `json:"tagPicList"`
		ResourceState        bool     `json:"resourceState"`
		Version              int      `json:"version"`
		SongJumpInfo         any      `json:"songJumpInfo"`
		EntertainmentTags    any      `json:"entertainmentTags"`
		AwardTags            any      `json:"awardTags"`
		Single               int      `json:"single"`
		NoCopyrightRcmd      any      `json:"noCopyrightRcmd"`
		Rtype                int      `json:"rtype"`
		Rurl                 any      `json:"rurl"`
		Mst                  int      `json:"mst"`
		Cp                   int      `json:"cp"`
		Mv                   int      `json:"mv"`
		PublishTime          int64    `json:"publishTime"`
		Tns                  []string `json:"tns,omitempty"`
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
			ResConsumable  bool `json:"resConsumable"`
			UserConsumable bool `json:"userConsumable"`
			ListenType     any  `json:"listenType"`
		} `json:"freeTrialPrivilege"`
		ChargeInfoList []struct {
			Rate          int `json:"rate"`
			ChargeURL     any `json:"chargeUrl"`
			ChargeMessage any `json:"chargeMessage"`
			ChargeType    int `json:"chargeType"`
		} `json:"chargeInfoList"`
	} `json:"privileges"`
	Code int `json:"code"`
}
