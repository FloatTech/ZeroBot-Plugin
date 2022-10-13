package guessmusic

// config内容
type config struct {
	MusicPath   string    `json:"musicPath"`
	Local       bool      `json:"local"`
	API         bool      `json:"api"`
	Cookie      string    `json:"cookie"`
	Playlist    []listRaw `json:"playlist"`
	Defaultlist []dlist   `json:"defaultlist"`
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
