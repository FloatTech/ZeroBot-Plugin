package niuniu

import (
	"fmt"
	"image"
	"net/http"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/rendercard"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
)

type drawUserRanking struct {
	name string
	user *userInfo
}

type drawer []drawUserRanking

func (allUsers drawer) draw(t bool) (img image.Image, err error) {
	fontbyte, err := file.GetLazyData(text.GlowSansFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
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
	ri := make([]*rendercard.RankInfo, len(allUsers))
	for i, user := range allUsers {
		resp, err := http.Get(fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=100", user.user.UID))
		if err != nil {
			return nil, err
		}
		decode, _, err := image.Decode(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		ri[i] = &rendercard.RankInfo{
			Avatar:         decode,
			TopLeftText:    user.name,
			BottomLeftText: fmt.Sprintf("QQ:%d", user.user.UID),
			RightText:      fmt.Sprintf("%s:%.2fcm", s, user.user.Length),
		}
	}
	img, err = rendercard.DrawRankingCard(fontbyte, title, ri)
	return
}
