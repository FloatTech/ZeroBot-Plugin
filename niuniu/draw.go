package niuniu

import (
	"fmt"
	"github.com/FloatTech/rendercard"
	"image"
	"net/http"
	"os"
)

var font, _ = os.ReadFile("./font/GlowSans.otf")

type drawUserRanking struct {
	Name string
	User *userInfo
}

func drawRanking(allUsers []drawUserRanking, t bool) (img image.Image, err error) {
	var (
		title string
		s     string
	)
	title = "牛牛深度排行"
	s = "牛牛深度"
	if t {
		title = "牛牛长度排行"
		s = "牛牛长度"
	}
	var ri []*rendercard.RankInfo
	for _, user := range allUsers {
		resp, err := http.Get(fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=100", user.User.UID))
		if err != nil {
			return nil, err
		}
		decode, _, err := image.Decode(resp.Body)
		if err != nil {
			return nil, err
		}
		ri = append(ri, &rendercard.RankInfo{
			Avatar:         decode,
			TopLeftText:    user.Name,
			BottomLeftText: fmt.Sprintf("QQ:%d", user.User.UID),
			RightText:      fmt.Sprintf("%s:%.2fcm", s, user.User.Length),
		})
	}
	img, err = rendercard.DrawRankingCard(font, title, ri)
	return
}
