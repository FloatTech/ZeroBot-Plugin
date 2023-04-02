// Package dish 程序员做饭指南zbp版，数据来源Anduin2017/HowToCook
package dish

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

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
	db      = &sql.Sqlite{}
	healthy = false
)

func init() {
	en := control.Register("dish", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "程序员做饭指南",
		Help:             "-怎么做[xxx]|[烹饪xxx]|[随机菜谱]|[随便做点菜]",
		PublicDataFolder: "Dish",
	})

	db.DBPath = en.DataFolder() + "dishes.db"

	if _, err := en.GetLazyData("dishes.db", true); err != nil {
		healthy = false
	} else if err = db.Open(time.Hour * 24); err != nil {
		healthy = false
	} else if err = db.Create("dishes", &dish{}); err != nil {
		healthy = false
	} else if count, countErr := db.Count("dishes"); countErr != nil {
		healthy = false
	} else {
		logrus.Infoln("[dish]加载", count, "条菜谱")
		healthy = true
	}

	if !healthy {
		logrus.Warnln("[dish]插件未能成功初始化")
	}

	en.OnPrefixGroup([]string{"怎么做", "烹饪"}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		if !healthy {
			ctx.SendChain(message.Text("客官，本店暂未开业"))
			return
		}

		name := ctx.NickName()
		dishName := ctx.State["args"].(string)
		var d dish
		if err := db.Find("dishes", &d, "WHERE name = '"+dishName+"'"); err != nil {
			ctx.SendChain(message.Text(fmt.Sprintf("未能为客官%s找到%s的做法qwq", name, dishName)))
		} else {
			ctx.SendChain(message.Text(fmt.Sprintf(
				"已为客官%s找到%s的做法辣！\n"+
					"原材料：%s\n"+
					"步骤：\n"+
					"%s",
				name, dishName, d.Materials, d.Steps),
			))
		}
	})

	en.OnPrefixGroup([]string{"随机菜谱", "随便做点菜"}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		if !healthy {
			ctx.SendChain(message.Text("客官，本店暂未开业"))
			return
		}

		name := ctx.NickName()
		var d dish
		if err := db.Pick("dishes", &d); err != nil {
			ctx.SendChain(message.Text(fmt.Sprintf("小店好像出错了，暂时端不出菜来惹")))
		} else {
			ctx.SendChain(message.Text(fmt.Sprintf(
				"已为客官%s送上%s的做法：\n"+
					"原材料：%s\n"+
					"步骤：\n"+
					"%s",
				name, d.Name, d.Materials, d.Steps),
			))
		}
	})
}
