package qingyunke

//定时早安,晚安
import (
	"github.com/robfig/cron"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	//所有群添加定时早安
	//zero.RangeBot(func(id int64, ctx *zero.Ctx) bool { // test the range bot function
	//	result := ctx.GetGroupList()
	//	log.Println(result)
	//	for _, v := range result.Array() {
	//		Daily(v.Get("group_id").Int())
	//	}
	//	return true
	//})
	zero.OnCommand("daily").SetBlock(false).FirstPriority().Handle(func(ctx *zero.Ctx) {
		log.Println(ctx.GetGroupList())
		result := ctx.GetGroupList()
		for _, v := range result.Array() {
			Daily(v.Get("group_id").Int())
		}

	})


}

func morningData(groupId int64) {
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		time.Sleep(time.Second * 1)
		ctx.SendGroupMessage(groupId, message.Image(getPicture()))
		ctx.SendGroupMessage(groupId, randText("啊......早上好...(哈欠)",
			"唔......吧唧...早上...哈啊啊~~~\n早上好......",
			"早上好......",
			"早上好呜......呼啊啊~~~~",
			"啊......早上好。\n昨晚也很激情呢！",
			"吧唧吧唧......怎么了...已经早上了么...",
			"早上好！",
			"......看起来像是傍晚，其实已经早上了吗？",
			"早上好......欸~~~脸好近呢"))
		return true
	})
}

func eveningData(groupId int64) {
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		time.Sleep(time.Second * 1)
		ctx.SendGroupMessage(groupId, message.Image(getPicture()))
		ctx.SendGroupMessage(groupId, randText("嗯哼哼~睡吧，就像平常一样安眠吧~o(≧▽≦)o",
			"......(打瞌睡)",
			"呼...呼...已经睡着了哦~...呼......",
			"......我、我会在这守着你的，请务必好好睡着"))
		return true
	})
}

func Daily(groupId int64) {
	log.Println("给" + strconv.FormatInt(groupId, 10) + "添加定时任务")
	c := cron.New()
	_ = c.AddFunc("0 30 7 * * ?", func() {
		morningData(groupId)
	})
	_ = c.AddFunc("0 30 22 * * ?", func() {
		eveningData(groupId)
	})
	c.Start()
}

func randText(text ...string) message.MessageSegment {
	length := len(text)
	return message.Text(text[rand.Intn(length)])
}
