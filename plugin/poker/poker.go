package poker

import (
	"math/rand"
	"os"
	"path"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 图片来源 https://www.bilibili.com/opus/834601953403076633

var cardImgPathList []string

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "抽扑克牌",
		Help:             "- 抽扑克\n- poker",
		PublicDataFolder: "Poker",
	}).ApplySingle(ctxext.DefaultSingle)

	// 初始化扑克牌图片文件路径
	entries, _ := os.ReadDir(engine.DataFolder())
	for _, entry := range entries {
		imgPath := path.Join(engine.DataFolder(), entry.Name())
		cardImgPathList = append(cardImgPathList, imgPath)
	}

	engine.OnFullMatchGroup([]string{"抽扑克", "poker"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			imgMsg, err := drawPoker()
			if err != nil {
				ctx.Send("[poker]读取扑克图片失败")
				return
			}
			ctx.Send(imgMsg)
		})
}

func drawPoker() (msg message.MessageSegment, err error) {
	randomIndex := rand.Intn(len(cardImgPathList))
	randomImgPath := cardImgPathList[randomIndex]
	imgData, err := os.ReadFile(randomImgPath)
	if err != nil {
		return
	}
	msg = message.ImageBytes(imgData)
	return msg, nil
}
