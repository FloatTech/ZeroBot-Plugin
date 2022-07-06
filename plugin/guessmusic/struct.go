package guessmusic

type config struct {
	MusicPath string `json:"musicPath"`
	Local     bool   `json:"local"`
	API       bool   `json:"api"`
}

type paugramData struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Cover    string `json:"cover"`
	Lyric    string `json:"lyric"`
	SubLyric string `json:"sub_lyric"`
	Link     string `json:"link"`
	Cached   bool   `json:"cached"`
}

type animeData struct {
	Msg string `json:"msg"`
	Res struct {
		ID        string `json:"id"`
		AnimeInfo struct {
			Desc  string `json:"desc"`
			ID    string `json:"id"`
			Atime int    `json:"atime"`
			Logo  string `json:"logo"`
			Year  int    `json:"year"`
			Bg    string `json:"bg"`
			Title string `json:"title"`
			Month int    `json:"month"`
		} `json:"anime_info"`
		PlayURL   string `json:"play_url"`
		Atime     int    `json:"atime"`
		Title     string `json:"title"`
		Author    string `json:"author"`
		Type      string `json:"type"`
		Recommend bool   `json:"recommend"`
	} `json:"res"`
	Code int `json:"code"`
}

type netEaseData struct {
	Code int `json:"code"`
	Data struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Picurl      string `json:"picurl"`
		Artistsname string `json:"artistsname"`
	} `json:"data"`
}

type autumnfishData struct {
	Result struct {
		Songs []struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Artists []struct {
				ID        int           `json:"id"`
				Name      string        `json:"name"`
				PicURL    interface{}   `json:"picUrl"`
				Alias     []interface{} `json:"alias"`
				AlbumSize int           `json:"albumSize"`
				PicID     int           `json:"picId"`
				Img1V1URL string        `json:"img1v1Url"`
				Img1V1    int           `json:"img1v1"`
				Trans     interface{}   `json:"trans"`
			} `json:"artists"`
			Album struct {
				ID     int    `json:"id"`
				Name   string `json:"name"`
				Artist struct {
					ID        int           `json:"id"`
					Name      string        `json:"name"`
					PicURL    interface{}   `json:"picUrl"`
					Alias     []interface{} `json:"alias"`
					AlbumSize int           `json:"albumSize"`
					PicID     int           `json:"picId"`
					Img1V1URL string        `json:"img1v1Url"`
					Img1V1    int           `json:"img1v1"`
					Trans     interface{}   `json:"trans"`
				} `json:"artist"`
				PublishTime int64 `json:"publishTime"`
				Size        int   `json:"size"`
				CopyrightID int   `json:"copyrightId"`
				Status      int   `json:"status"`
				PicID       int64 `json:"picId"`
				Mark        int   `json:"mark"`
			} `json:"album"`
			Duration    int           `json:"duration"`
			CopyrightID int           `json:"copyrightId"`
			Status      int           `json:"status"`
			Alias       []interface{} `json:"alias"`
			Rtype       int           `json:"rtype"`
			Ftype       int           `json:"ftype"`
			TransNames  []string      `json:"transNames"`
			Mvid        int           `json:"mvid"`
			Fee         int           `json:"fee"`
			RURL        interface{}   `json:"rUrl"`
			Mark        int           `json:"mark"`
		} `json:"songs"`
		HasMore   bool `json:"hasMore"`
		SongCount int  `json:"songCount"`
	} `json:"result"`
	Code int `json:"code"`
}
