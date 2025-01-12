package niuniu

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"

	"github.com/FloatTech/AnimeAPI/niu"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/rendercard"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func processRankingImg(allUsers niu.BaseInfos, ctx *zero.Ctx, t bool) ([]byte, error) {
	fontByte, err := file.GetLazyData(text.GlowSansFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	s := "牛牛长度"
	title := "牛牛长度排行"
	if !t {
		s = "牛牛深度"
		title = "牛牛深度排行"
	}
	ri := make([]*rendercard.RankInfo, len(allUsers))
	for i, user := range allUsers {
		resp, err := http.Get(fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=100", user.UID))
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
			TopLeftText:    ctx.CardOrNickName(user.UID),
			BottomLeftText: fmt.Sprintf("QQ:%d", user.UID),
			RightText:      fmt.Sprintf("%s:%.2fcm", s, user.Length),
		}
	}
	img, err := rendercard.DrawRankingCard(fontByte, title, ri)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	return buf.Bytes(), err
}
