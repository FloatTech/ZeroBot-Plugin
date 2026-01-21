// Package music 整合多平台音乐点播能力（基于 music-lib 重构）
package music

import (
	"errors"
	"fmt"

	ctrl "github.com/FloatTech/zbpctrl"     // 别名 zbpctrl 为 ctrl
	"github.com/FloatTech/zbputils/control" // 保持 control 原名
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/guohuiyuan/music-lib/kugou"
	"github.com/guohuiyuan/music-lib/kuwo"
	"github.com/guohuiyuan/music-lib/migu"
	"github.com/guohuiyuan/music-lib/netease"
	"github.com/guohuiyuan/music-lib/qq"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 平台映射：指令前缀 -> 平台名称
var platformMap = map[string]string{
	"咪咕": "migu",
	"酷我": "kuwo",
	"酷狗": "kugou",
	"网易": "netease",
	"qq": "qq",
	"":   "kuwo", // 默认点歌指向酷我
}

func init() {
	// 注册指令处理器
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "点歌",
		Help: "- 点歌[xxx] (默认酷我)\n" +
			"- 网易点歌[xxx]\n" +
			"- 酷我点歌[xxx]\n" +
			"- 酷狗点歌[xxx]\n" +
			"- 咪咕点歌[xxx]\n" +
			"- qq点歌[xxx]\n",
	}).OnRegex(`^(.{0,2})点歌\s?(.{1,25})$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			matches := ctx.State["regex_matched"].([]string)
			platformPrefix := matches[1]
			keyword := matches[2]

			// 获取目标平台
			targetPlatform, ok := platformMap[platformPrefix]
			if !ok {
				ctx.SendChain(message.Text("不支持的点播平台：", platformPrefix))
				return
			}

			// 执行点播并返回结果
			seg, err := getMusicSegment(targetPlatform, keyword)
			if err != nil {
				ctx.SendChain(message.Text("点歌失败：", err.Error()))
				return
			}
			ctx.SendChain(seg)
		})
}

// getMusicSegment 根据平台和关键词获取音乐消息段
func getMusicSegment(platform, keyword string) (message.Segment, error) {
	switch platform {
	case "migu":
		return getMiguMusic(keyword)
	case "kuwo":
		return getKuwoMusic(keyword)
	case "kugou":
		return getKugouMusic(keyword)
	case "netease":
		return getNeteaseMusic(keyword)
	case "qq":
		return getQQMusic(keyword)
	default:
		return message.Segment{}, errors.New("未知的音乐平台：" + platform)
	}
}

// --- 各平台适配层（基于 music-lib 实现） ---

// getMiguMusic 咪咕音乐点播
func getMiguMusic(keyword string) (message.Segment, error) {
	songs, err := migu.Search(keyword)
	if err != nil || len(songs) == 0 {
		return message.Segment{}, errors.New("咪咕音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 传入 &song (指针)
	playURL, err := migu.GetDownloadURL(&song)
	if err != nil || playURL == "" {
		return message.Segment{}, errors.New("获取咪咕播放链接失败：" + err.Error())
	}

	return message.CustomMusic(
		fmt.Sprintf("https://music.migu.cn/v3/music/song/%s", song.ID),
		playURL,
		song.Name,
	).Add("content", song.Artist).Add("image", song.Cover).Add("subtype", "migu"), nil
}

// getKuwoMusic 酷我音乐点播
func getKuwoMusic(keyword string) (message.Segment, error) {
	songs, err := kuwo.Search(keyword)
	if err != nil || len(songs) == 0 {
		return message.Segment{}, errors.New("酷我音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 传入 &song (指针)
	playURL, err := kuwo.GetDownloadURL(&song)
	if err != nil || playURL == "" {
		return message.Segment{}, errors.New("获取酷我播放链接失败：" + err.Error())
	}

	return message.CustomMusic(
		fmt.Sprintf("https://www.kuwo.cn/play_detail/%s", song.ID),
		playURL,
		song.Name,
	).Add("content", song.Artist).Add("image", song.Cover).Add("subtype", "kuwo"), nil
}

// getKugouMusic 酷狗音乐点播
func getKugouMusic(keyword string) (message.Segment, error) {
	songs, err := kugou.Search(keyword)
	if err != nil || len(songs) == 0 {
		return message.Segment{}, errors.New("酷狗音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 传入 &song (指针)
	playURL, err := kugou.GetDownloadURL(&song)
	if err != nil || playURL == "" {
		return message.Segment{}, errors.New("获取酷狗播放链接失败：" + err.Error())
	}

	return message.CustomMusic(
		"https://www.kugou.com/",
		playURL,
		song.Name,
	).Add("content", song.Artist).Add("image", song.Cover).Add("subtype", "kugou"), nil
}

// getNeteaseMusic 网易云音乐点播
func getNeteaseMusic(keyword string) (message.Segment, error) {
	songs, err := netease.Search(keyword)
	if err != nil || len(songs) == 0 {
		return message.Segment{}, errors.New("网易云音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 获取播放直链
	playURL, err := netease.GetDownloadURL(&song)
	if err != nil || playURL == "" {
		return message.Segment{}, errors.New("获取网易云播放链接失败：" + err.Error())
	}

	// 构造 CustomMusic
	return message.CustomMusic(
		fmt.Sprintf("https://music.163.com/#/song?id=%s", song.ID),
		playURL,
		song.Name,
	).Add("content", song.Artist).Add("image", song.Cover).Add("subtype", "163"), nil
}

// getQQMusic QQ音乐点播
func getQQMusic(keyword string) (message.Segment, error) {
	songs, err := qq.Search(keyword)
	if err != nil || len(songs) == 0 {
		return message.Segment{}, errors.New("QQ音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 获取播放直链
	playURL, err := qq.GetDownloadURL(&song)
	if err != nil || playURL == "" {
		return message.Segment{}, errors.New("获取QQ音乐播放链接失败：" + err.Error())
	}

	// 构造 CustomMusic
	return message.CustomMusic(
		fmt.Sprintf("https://y.qq.com/n/ryqq/songDetail/%s", song.ID),
		playURL,
		song.Name,
	).Add("content", song.Artist).Add("image", song.Cover).Add("subtype", "qq"), nil
}
