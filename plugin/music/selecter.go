// Package music 整合多平台音乐点播能力（基于 music-lib 重构）
package music

import (
	"fmt"

	ctrl "github.com/FloatTech/zbpctrl"     // 别名 zbpctrl 为 ctrl
	"github.com/FloatTech/zbputils/control" // 保持 control 原名
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/guohuiyuan/music-lib/kugou"
	"github.com/guohuiyuan/music-lib/kuwo"
	"github.com/guohuiyuan/music-lib/migu"
	"github.com/guohuiyuan/music-lib/netease"
	"github.com/guohuiyuan/music-lib/qq"
	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 平台映射：指令前缀 -> 点歌函数
var platformMap = map[string]func(string) (message.Segment, error){
	"咪咕": getMiguMusic,
	"酷我": getKuwoMusic,
	"酷狗": getKugouMusic,
	"网易": getNeteaseMusic,
	"qq": getQQMusic,
	"":   getKuwoMusic, // 默认点歌指向酷我
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

			// 获取目标平台处理函数
			processFunc, ok := platformMap[platformPrefix]
			if !ok {
				ctx.SendChain(message.Text("不支持的点播平台：", platformPrefix))
				return
			}

			// 执行点播并返回结果
			seg, err := processFunc(keyword)
			if err != nil {
				// 修改：直接传递 err，不需要 call .Error()
				ctx.SendChain(message.Text("点歌失败：", err))
				return
			}
			ctx.SendChain(seg)
		})
}

// 删除了 getMusicSegment 函数，因为已经通过 Map 直接分发

// --- 各平台适配层（基于 music-lib 实现） ---

// getMiguMusic 咪咕音乐点播
func getMiguMusic(keyword string) (message.Segment, error) {
	songs, err := migu.Search(keyword)
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "咪咕音乐搜索失败")
	}
	if len(songs) == 0 {
		return message.Segment{}, errors.New("咪咕音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 传入 &song (指针)
	playURL, err := migu.GetDownloadURL(&song)
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "获取咪咕播放链接失败")
	}
	if playURL == "" {
		return message.Segment{}, errors.New("获取咪咕播放链接失败：链接为空")
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
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "酷我音乐搜索失败")
	}
	if len(songs) == 0 {
		return message.Segment{}, errors.New("酷我音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 传入 &song (指针)
	playURL, err := kuwo.GetDownloadURL(&song)
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "获取酷我播放链接失败")
	}
	if playURL == "" {
		return message.Segment{}, errors.New("获取酷我播放链接失败：链接为空")
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
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "酷狗音乐搜索失败")
	}
	if len(songs) == 0 {
		return message.Segment{}, errors.New("酷狗音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 传入 &song (指针)
	playURL, err := kugou.GetDownloadURL(&song)
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "获取酷狗播放链接失败")
	}
	if playURL == "" {
		return message.Segment{}, errors.New("获取酷狗播放链接失败：链接为空")
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
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "网易云音乐搜索失败")
	}
	if len(songs) == 0 {
		return message.Segment{}, errors.New("网易云音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 获取播放直链
	playURL, err := netease.GetDownloadURL(&song)
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "获取网易云播放链接失败")
	}
	if playURL == "" {
		return message.Segment{}, errors.New("获取网易云播放链接失败：链接为空")
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
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "QQ音乐搜索失败")
	}
	if len(songs) == 0 {
		return message.Segment{}, errors.New("QQ音乐未找到相关歌曲：" + keyword)
	}
	song := songs[0]

	// 获取播放直链
	playURL, err := qq.GetDownloadURL(&song)
	if err != nil {
		return message.Segment{}, errors.Wrap(err, "获取QQ音乐播放链接失败")
	}
	if playURL == "" {
		return message.Segment{}, errors.New("获取QQ音乐播放链接失败：链接为空")
	}

	// 构造 CustomMusic
	return message.CustomMusic(
		fmt.Sprintf("https://y.qq.com/n/ryqq/songDetail/%s", song.ID),
		playURL,
		song.Name,
	).Add("content", song.Artist).Add("image", song.Cover).Add("subtype", "qq"), nil
}
