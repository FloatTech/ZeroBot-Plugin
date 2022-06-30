package guessmusic

type config struct {
	MusicPath string `json:"musicPath"`
	Local     bool   `json:"local"`
	Api       bool   `json:"api"`
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
