// Package dish 程序员做饭指南zbp版，数据来源Anduin2017/HowToCook
package dish

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type dish struct {
	ID        uint32 `db:"id"`
	Name      string `db:"name"`
	Materials string `db:"materials"`
	Steps     string `db:"steps"`
}

var (
	db          sql.Sqlite
	initialized = false
)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "程序员做饭指南",
		Help:             "-怎么做[xxx]|烹饪[xxx]|随机菜谱|随便做点菜",
		PublicDataFolder: "Dish",
	})

	db = sql.New(en.DataFolder() + "dishes.db")

	if _, err := en.GetLazyData("dishes.db", true); err != nil {
		logrus.Warnln("[dish]获取菜谱数据库文件失败")
	} else if err = db.Open(time.Hour); err != nil {
		logrus.Warnln("[dish]连接菜谱数据库失败")
	} else if err = db.Create("dish", &dish{}); err != nil {
		logrus.Warnln("[dish]同步菜谱数据表失败")
	} else if count, err := db.Count("dish"); err != nil {
		logrus.Warnln("[dish]统计菜谱数据失败")
	} else {
		logrus.Infoln("[dish]加载", count, "条菜谱")
		initialized = true
	}

	if !initialized {
		logrus.Warnln("[dish]插件未能成功初始化")
	}

	en.OnPrefixGroup([]string{"怎么做", "烹饪"}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		if !initialized {
			ctx.SendChain(message.Text("客官，本店暂未开业"))
			return
		}

		name := ctx.CardOrNickName(ctx.Event.UserID)
		dishName := ctx.State["args"].(string)

		if dishName == "" {
			return
		}

		if strings.Contains(dishName, "'") ||
			strings.Contains(dishName, "\"") ||
			strings.Contains(dishName, "\\") ||
			strings.Contains(dishName, ";") {
			return
		}

		var d dish
		if err := db.Find("dish", &d, "WHERE name LIKE ?", "%"+dishName+"%"); err != nil {
			ctx.SendChain(message.Text("客官，本店没有" + dishName))
			return
		}

		ctx.SendChain(message.Text(
			"已为客官", name, "找到", d.Name, "的做法辣！\n",
			"原材料：", d.Materials, "\n",
			"步骤：", d.Steps,
		))
	})

	en.OnFullMatchGroup([]string{"随机菜谱", "随便做点菜"}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		if !initialized {
			ctx.SendChain(message.Text("客官，本店暂未开业"))
			return
		}

		name := ctx.CardOrNickName(ctx.Event.UserID)
		var d dish
		if err := db.Pick("dish", &d); err != nil {
			ctx.SendChain(message.Text("小店好像出错了，暂时端不出菜来惹"))
			logrus.Warnln("[dish]随机菜谱请求出错：" + err.Error())
			return
		}

		ctx.SendChain(message.Text(
			"已为客官", name, "送上", d.Name, "的做法：\n",
			"原材料：", d.Materials, "\n",
			"步骤：", d.Steps,
		))
	})
}
