// Package dish 程序员做饭指南zbp版，数据来源Anduin2017/HowToCook
package dish

import (
	"fmt"
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
	db          = &sql.Sqlite{}
	initialized = false
)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "程序员做饭指南",
		Help:             "-怎么做[xxx]|烹饪[xxx]|随机菜谱|随便做点菜",
		PublicDataFolder: "Dish",
	})

	db.DBPath = en.DataFolder() + "dishes.db"

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

		name := ctx.NickName()
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
		if err := db.Find("dish", &d, fmt.Sprintf("WHERE name like %%%s%%", dishName)); err != nil {
			return
		}

		ctx.SendChain(message.Text(fmt.Sprintf(
			"已为客官%s找到%s的做法辣！\n"+
				"原材料：%s\n"+
				"步骤：\n"+
				"%s",
			name, dishName, d.Materials, d.Steps),
		))
	})

	en.OnPrefixGroup([]string{"随机菜谱", "随便做点菜"}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		if !initialized {
			ctx.SendChain(message.Text("客官，本店暂未开业"))
			return
		}

		name := ctx.NickName()
		var d dish
		if err := db.Pick("dish", &d); err != nil {
			ctx.SendChain(message.Text("小店好像出错了，暂时端不出菜来惹"))
			logrus.Warnln("[dish]随机菜谱请求出错：" + err.Error())
			return
		}

		ctx.SendChain(message.Text(fmt.Sprintf(
			"已为客官%s送上%s的做法：\n"+
				"原材料：%s\n"+
				"步骤：\n"+
				"%s",
			name, d.Name, d.Materials, d.Steps),
		))
	})
}
