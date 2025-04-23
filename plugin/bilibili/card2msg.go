package bilibili

import (
	"encoding/json"
	"fmt"
	"time"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	msgType = map[int]string{
		1:    "转发了动态",
		2:    "有图营业",
		4:    "无图营业",
		8:    "投稿了视频",
		16:   "投稿了短视频",
		64:   "投稿了文章",
		256:  "投稿了音频",
		2048: "发布了简报",
		4200: "发布了直播",
		4308: "发布了直播",
	}
)

// dynamicCard2msg 处理DynCard
func dynamicCard2msg(dynamicCard *bz.DynamicCard) (msg []message.Segment, err error) {
	var (
		card  bz.Card
		vote  bz.Vote
		cType int
	)
	msg = make([]message.Segment, 0, 16)
	// 初始化结构体
	err = json.Unmarshal(binary.StringToBytes(dynamicCard.Card), &card)
	if err != nil {
		return
	}
	if dynamicCard.Extension.Vote != "" {
		err = json.Unmarshal(binary.StringToBytes(dynamicCard.Extension.Vote), &vote)
		if err != nil {
			return
		}
	}
	cType = dynamicCard.Desc.Type
	// 生成消息
	switch cType {
	case 1:
		msg = append(msg, message.Text(card.User.Uname, msgType[cType], "\n",
			card.Item.Content, "\n",
			"转发的内容: \n"))
		var originMsg []message.Segment
		var co bz.Card
		co, err = bz.LoadCardDetail(card.Origin)
		if err != nil {
			return
		}
		originMsg, err = card2msg(dynamicCard, &co, card.Item.OrigType)
		if err != nil {
			return
		}
		msg = append(msg, originMsg...)
	case 2:
		msg = append(msg, message.Text(card.User.Name, "在", time.Unix(int64(card.Item.UploadTime), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Item.Description))
		for i := 0; i < len(card.Item.Pictures); i++ {
			msg = append(msg, message.Image(card.Item.Pictures[i].ImgSrc))
		}
	case 4:
		msg = append(msg, message.Text(card.User.Uname, "在", time.Unix(int64(card.Item.Timestamp), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Item.Content, "\n"))
		if dynamicCard.Extension.Vote != "" {
			msg = append(msg, message.Text("【投票】", vote.Desc, "\n",
				"截止日期: ", time.Unix(int64(vote.Endtime), 0).Format("2006-01-02 15:04:05"), "\n",
				"参与人数: ", bz.HumanNum(vote.JoinNum), "\n",
				"投票选项( 最多选择", vote.ChoiceCnt, "项 )\n"))
			for i := 0; i < len(vote.Options); i++ {
				msg = append(msg, message.Text("- ", vote.Options[i].Idx, ". ", vote.Options[i].Desc, "\n"))
				if vote.Options[i].ImgURL != "" {
					msg = append(msg, message.Image(vote.Options[i].ImgURL))
				}
			}
		}
	case 8:
		msg = append(msg, message.Text(card.Owner.Name, "在", time.Unix(int64(card.Pubdate), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Title))
		msg = append(msg, message.Image(card.Pic))
		msg = append(msg, message.Text(card.Desc, "\n",
			card.ShareSubtitle, "\n",
			"视频链接: ", card.ShortLink, "\n"))
	case 16:
		msg = append(msg, message.Text(card.User.Name, "在", time.Unix(int64(card.Item.UploadTime), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Item.Description))
		msg = append(msg, message.Image(card.Item.Cover.Default))
	case 64:
		msg = append(msg, message.Text(card.Author.(map[string]any)["name"], "在", time.Unix(int64(card.PublishTime), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Title, "\n",
			card.Summary))
		for i := 0; i < len(card.ImageUrls); i++ {
			msg = append(msg, message.Image(card.ImageUrls[i]))
		}
		if card.ID != 0 {
			msg = append(msg, message.Text("文章链接: https://www.bilibili.com/read/cv", card.ID, "\n"))
		}
	case 256:
		msg = append(msg, message.Text(card.Upper, "在", time.Unix(int64(card.Ctime), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Title))
		msg = append(msg, message.Image(card.Cover))
		msg = append(msg, message.Text(card.Intro, "\n"))
		if card.ID != 0 {
			msg = append(msg, message.Text("音频链接: https://www.bilibili.com/audio/au", card.ID, "\n"))
		}

	case 2048:
		msg = append(msg, message.Text(card.User.Uname, msgType[cType], "\n",
			card.Vest.Content, "\n",
			card.Sketch.Title, "\n",
			card.Sketch.DescText, "\n"))
		msg = append(msg, message.Image(card.Sketch.CoverURL))
		msg = append(msg, message.Text("分享链接: ", card.Sketch.TargetURL, "\n"))
	case 4308:
		if dynamicCard.Desc.UserProfile.Info.Uname != "" {
			msg = append(msg, message.Text(dynamicCard.Desc.UserProfile.Info.Uname, msgType[cType], "\n"))
		}
		msg = append(msg, message.Image(card.LivePlayInfo.Cover))
		msg = append(msg, message.Text("\n", card.LivePlayInfo.Title, "\n",
			"房间号: ", card.LivePlayInfo.RoomID, "\n",
			"分区: ", card.LivePlayInfo.ParentAreaName))
		if card.LivePlayInfo.ParentAreaName != card.LivePlayInfo.AreaName {
			msg = append(msg, message.Text("-", card.LivePlayInfo.AreaName))
		}
		if card.LivePlayInfo.LiveStatus == 0 {
			msg = append(msg, message.Text("未开播 \n"))
		} else {
			msg = append(msg, message.Text("直播中 ", card.LivePlayInfo.WatchedShow, "\n"))
		}
		msg = append(msg, message.Text("直播链接: ", card.LivePlayInfo.Link))
	default:
		msg = append(msg, message.Text("动态id: ", dynamicCard.Desc.DynamicIDStr, "未知动态类型: ", cType, "\n"))
	}
	if dynamicCard.Desc.DynamicIDStr != "" {
		msg = append(msg, message.Text("动态链接: ", bz.TURL, dynamicCard.Desc.DynamicIDStr))
	}
	return
}

// card2msg cType=1, 2, 4, 8, 16, 64, 256, 2048, 4200, 4308时,处理Card字符串,cType为card类型
func card2msg(dynamicCard *bz.DynamicCard, card *bz.Card, cType int) (msg []message.Segment, err error) {
	var (
		vote bz.Vote
	)
	msg = make([]message.Segment, 0, 16)
	// 生成消息
	switch cType {
	case 1:
		msg = append(msg, message.Text(card.User.Uname, msgType[cType], "\n",
			card.Item.Content, "\n",
			"转发的内容: \n"))
		var originMsg []message.Segment
		var co bz.Card
		co, err = bz.LoadCardDetail(card.Origin)
		if err != nil {
			return
		}
		originMsg, err = card2msg(dynamicCard, &co, card.Item.OrigType)
		if err != nil {
			return
		}
		msg = append(msg, originMsg...)
	case 2:
		msg = append(msg, message.Text(card.User.Name, "在", time.Unix(int64(card.Item.UploadTime), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Item.Description))
		for i := 0; i < len(card.Item.Pictures); i++ {
			msg = append(msg, message.Image(card.Item.Pictures[i].ImgSrc))
		}
	case 4:
		msg = append(msg, message.Text(card.User.Uname, "在", time.Unix(int64(card.Item.Timestamp), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Item.Content, "\n"))
		if dynamicCard.Extension.Vote != "" {
			msg = append(msg, message.Text("【投票】", vote.Desc, "\n",
				"截止日期: ", time.Unix(int64(vote.Endtime), 0).Format("2006-01-02 15:04:05"), "\n",
				"参与人数: ", bz.HumanNum(vote.JoinNum), "\n",
				"投票选项( 最多选择", vote.ChoiceCnt, "项 )\n"))
			for i := 0; i < len(vote.Options); i++ {
				msg = append(msg, message.Text("- ", vote.Options[i].Idx, ". ", vote.Options[i].Desc, "\n"))
				if vote.Options[i].ImgURL != "" {
					msg = append(msg, message.Image(vote.Options[i].ImgURL))
				}
			}
		}
	case 8:
		msg = append(msg, message.Text(card.Owner.Name, "在", time.Unix(int64(card.Pubdate), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Title))
		msg = append(msg, message.Image(card.Pic))
		msg = append(msg, message.Text(card.Desc, "\n",
			card.ShareSubtitle, "\n",
			"视频链接: ", card.ShortLink, "\n"))
	case 16:
		msg = append(msg, message.Text(card.User.Name, "在", time.Unix(int64(card.Item.UploadTime), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Item.Description))
		msg = append(msg, message.Image(card.Item.Cover.Default))
	case 64:
		msg = append(msg, message.Text(card.Author.(map[string]any)["name"], "在", time.Unix(int64(card.PublishTime), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Title, "\n",
			card.Summary))
		for i := 0; i < len(card.ImageUrls); i++ {
			msg = append(msg, message.Image(card.ImageUrls[i]))
		}
		if card.ID != 0 {
			msg = append(msg, message.Text("文章链接: https://www.bilibili.com/read/cv", card.ID, "\n"))
		}
	case 256:
		msg = append(msg, message.Text(card.Upper, "在", time.Unix(int64(card.Ctime), 0).Format("2006-01-02 15:04:05"), msgType[cType], "\n",
			card.Title))
		msg = append(msg, message.Image(card.Cover))
		msg = append(msg, message.Text(card.Intro, "\n"))
		if card.ID != 0 {
			msg = append(msg, message.Text("音频链接: https://www.bilibili.com/audio/au", card.ID, "\n"))
		}

	case 2048:
		msg = append(msg, message.Text(card.User.Uname, msgType[cType], "\n",
			card.Vest.Content, "\n",
			card.Sketch.Title, "\n",
			card.Sketch.DescText, "\n"))
		msg = append(msg, message.Image(card.Sketch.CoverURL))
		msg = append(msg, message.Text("分享链接: ", card.Sketch.TargetURL, "\n"))
	case 4308:
		if dynamicCard.Desc.UserProfile.Info.Uname != "" {
			msg = append(msg, message.Text(dynamicCard.Desc.UserProfile.Info.Uname, msgType[cType], "\n"))
		}
		msg = append(msg, message.Image(card.LivePlayInfo.Cover))
		msg = append(msg, message.Text("\n", card.LivePlayInfo.Title, "\n",
			"房间号: ", card.LivePlayInfo.RoomID, "\n",
			"分区: ", card.LivePlayInfo.ParentAreaName))
		if card.LivePlayInfo.ParentAreaName != card.LivePlayInfo.AreaName {
			msg = append(msg, message.Text("-", card.LivePlayInfo.AreaName))
		}
		if card.LivePlayInfo.LiveStatus == 0 {
			msg = append(msg, message.Text("未开播 \n"))
		} else {
			msg = append(msg, message.Text("直播中 ", card.LivePlayInfo.WatchedShow, "\n"))
		}
		msg = append(msg, message.Text("直播链接: ", card.LivePlayInfo.Link))
	default:
		msg = append(msg, message.Text("动态id: ", dynamicCard.Desc.DynamicIDStr, "未知动态类型: ", cType, "\n"))
	}
	if dynamicCard.Desc.DynamicIDStr != "" {
		msg = append(msg, message.Text("动态链接: ", bz.TURL, dynamicCard.Desc.DynamicIDStr))
	}
	return
}

// dynamicDetail 用动态id查动态信息
func dynamicDetail(cookiecfg *bz.CookieConfig, dynamicIDStr string) (msg []message.Segment, err error) {
	dyc, err := bz.GetDynamicDetail(cookiecfg, dynamicIDStr)
	if err != nil {
		return
	}
	return dynamicCard2msg(&dyc)
}

// articleCard2msg 专栏转消息
func articleCard2msg(card bz.Card, defaultID string) (msg []message.Segment) {
	msg = make([]message.Segment, 0, 16)
	for i := 0; i < len(card.OriginImageUrls); i++ {
		msg = append(msg, message.Image(card.OriginImageUrls[i]))
	}
	msg = append(msg, message.Text("\n", card.Title, "\n", "UP主: ", card.AuthorName, "\n",
		"阅读: ", bz.HumanNum(card.Stats.View), " 评论: ", bz.HumanNum(card.Stats.Reply), "\n",
		bz.CVURL, defaultID))
	return
}

// liveCard2msg 直播卡片转消息
func liveCard2msg(card bz.RoomCard) (msg []message.Segment) {
	msg = make([]message.Segment, 0, 16)
	msg = append(msg, message.Image(card.RoomInfo.Keyframe))
	msg = append(msg, message.Text("\n", card.RoomInfo.Title, "\n",
		"主播: ", card.AnchorInfo.BaseInfo.Uname, "\n",
		"房间号: ", card.RoomInfo.RoomID, "\n"))
	if card.RoomInfo.ShortID != 0 {
		msg = append(msg, message.Text("短号: ", card.RoomInfo.ShortID, "\n"))
	}
	msg = append(msg, message.Text("分区: ", card.RoomInfo.ParentAreaName))
	if card.RoomInfo.ParentAreaName != card.RoomInfo.AreaName {
		msg = append(msg, message.Text("-", card.RoomInfo.AreaName))
	}
	if card.RoomInfo.LiveStatus == 0 {
		msg = append(msg, message.Text("未开播 \n"))
	} else {
		msg = append(msg, message.Text("直播中 ", bz.HumanNum(card.RoomInfo.Online), "人气\n"))
	}
	if card.RoomInfo.ShortID != 0 {
		msg = append(msg, message.Text("直播间链接: ", bz.LURL, card.RoomInfo.ShortID))
	} else {
		msg = append(msg, message.Text("直播间链接: ", bz.LURL, card.RoomInfo.RoomID))
	}

	return
}

// videoCard2msg 视频卡片转消息
func videoCard2msg(card bz.Card) (msg []message.Segment, err error) {
	var (
		mCard       bz.MemberCard
		onlineTotal bz.OnlineTotal
	)
	msg = make([]message.Segment, 0, 16)
	mCard, err = bz.GetMemberCard(card.Owner.Mid)
	msg = append(msg, message.Text("标题: ", card.Title, "\n"))
	if card.Rights.IsCooperation == 1 {
		for i := 0; i < len(card.Staff); i++ {
			msg = append(msg, message.Text(card.Staff[i].Title, ": ", card.Staff[i].Name, " 粉丝: ", bz.HumanNum(card.Staff[i].Follower), "\n"))
		}
	} else {
		if err != nil {
			msg = append(msg, message.Text("UP主: ", card.Owner.Name, "\n"))
		} else {
			msg = append(msg, message.Text("UP主: ", card.Owner.Name, " 粉丝: ", bz.HumanNum(mCard.Fans), "\n"))
		}
	}
	msg = append(msg, message.Image(card.Pic))
	data, err := web.GetData(fmt.Sprintf(bz.OnlineTotalURL, card.BvID, card.CID))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &onlineTotal)
	if err != nil {
		return
	}
	msg = append(msg, message.Text("👀播放: ", bz.HumanNum(card.Stat.View), " 💬弹幕: ", bz.HumanNum(card.Stat.Danmaku),
		"\n👍点赞: ", bz.HumanNum(card.Stat.Like), " 💰投币: ", bz.HumanNum(card.Stat.Coin),
		"\n📁收藏: ", bz.HumanNum(card.Stat.Favorite), " 🔗分享: ", bz.HumanNum(card.Stat.Share),
		"\n📝简介: ", card.Desc,
		"\n🏄‍♂️ 总共 ", onlineTotal.Data.Total, " 人在观看，", onlineTotal.Data.Count, " 人在网页端观看\n",
		bz.VURL, card.BvID, "\n\n"))
	return
}
