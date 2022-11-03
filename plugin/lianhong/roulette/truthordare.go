// Package partygame 真心话大冒险
package partygame

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type tdata struct {
	Version string `json:"version"`
	Data    []struct {
		Name string   `json:"name"`
		Des  string   `json:"des"`
		Tags []string `json:"tags"`
	} `json:"data"`
}

var (
	action    tdata
	question  tdata
	punishmap sync.Map
)

func init() { // 插件主体
	engine := control.Register("truthordare", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "真心话大冒险\n- 来点乐子[@xxx]\n- 饶恕[@xxx]\n- 惩罚[@xxx]\n- 反省[@xxx]",
		PublicDataFolder: "Truthordare",
	})
	actionData, err := engine.GetLazyData("action.json", false)
	if err != nil {
		panic(err)
	}
	questionData, err := engine.GetLazyData("question.json", false)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(actionData, &action)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(questionData, &question)
	if err != nil {
		panic(err)
	}
	logrus.Infoln("[Truthordare]加载", len(question.Data), "条真心话")
	logrus.Infoln("[Truthordare]加载", len(action.Data), "条大冒险")
	engine.OnRegex(`^(真心话大冒险|来点刺激|来点乐子)`).Handle(func(ctx *zero.Ctx) {
		key := fmt.Sprintf("%v-%v", ctx.Event.GroupID, ctx.Event.UserID)
		v, ok := punishmap.Load(key)
		if ok {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("罪行尚未被饶恕, 赎罪方式是", v))
			return
		}
		puid := ctx.Event.UserID
		if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
			puid, _ = strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		}
		ctx.Event.UserID = puid
		getTruthOrDare(ctx)
	})
	engine.OnRegex(`^(饶恕|阿门|释放|原谅|赦免)`, zero.AdminPermission, zero.OnlyGroup).Handle(func(ctx *zero.Ctx) {
		puid := ctx.Event.UserID
		if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
			puid, _ = strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		}
		key := fmt.Sprintf("%v-%v", ctx.Event.GroupID, puid)
		punishmap.Delete(key)
		ctx.SendChain(message.At(puid), message.Text("恭喜你恢复自由之身"))
	})
	engine.OnRegex(`^(惩罚|降下神罚)`, zero.AdminPermission, zero.OnlyGroup).Handle(func(ctx *zero.Ctx) {
		puid := ctx.Event.UserID
		if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
			puid, _ = strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		}
		ctx.Event.UserID = puid
		getTruthOrDare(ctx)
	})
	engine.OnRegex(`^(反省|检查罪行)`, zero.OnlyGroup).Handle(func(ctx *zero.Ctx) {
		puid := ctx.Event.UserID
		if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
			puid, _ = strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		}
		key := fmt.Sprintf("%v-%v", ctx.Event.GroupID, puid)
		v, ok := punishmap.Load(key)
		if ok {
			ctx.SendChain(message.At(puid), message.Text("你是罪人, 赎罪方式是", v))
		}
		role := ctx.GetGroupMemberInfo(ctx.Event.GroupID, puid, true).Get("role").String()
		ctx.Event.UserID = puid
		if zero.SuperUserPermission(ctx) || role != "member" {
			ctx.SendChain(message.At(puid), message.Text("你是上帝"))
		}
		if !ok && !zero.SuperUserPermission(ctx) && role == "member" {
			ctx.SendChain(message.At(puid), message.Text("你是平民"))
		}
	})
}

func getAction() string {
	return action.Data[rand.Intn(len(action.Data))].Name
}

func getQuestion() string {
	return question.Data[rand.Intn(len(question.Data))].Name
}

func getActionOrQuestion() string {
	if time.Now().UnixNano()%2 == 0 {
		return getAction()
	}
	return getQuestion()
}

func getTruthOrDare(ctx *zero.Ctx) {
	next, cancel := zero.NewFutureEvent("message", 999, false, ctx.CheckSession(), zero.FullMatchRule("真心话", "大冒险")).Repeat()
	defer cancel()
	key := fmt.Sprintf("%v-%v", ctx.Event.GroupID, ctx.Event.UserID)
	ctx.SendChain(message.At(ctx.Event.UserID), message.Text("你将受到严峻的惩罚, 请选择惩罚, 真心话还是大冒险?"))
	for {
		select {
		case <-time.After(time.Second * 20):
			ctx.SendChain(message.Text("时间太久啦！", zero.BotConfig.NickName[0], "帮你选择"))
			p := getActionOrQuestion()
			punishmap.Store(key, p)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("恭喜你获得\"", p, "\"的惩罚"))
			return
		case c := <-next:
			msg := c.Event.Message.ExtractPlainText()
			var p string
			if msg == "真心话" {
				p = getQuestion()
			} else if msg == "大冒险" {
				p = getAction()
			}
			punishmap.Store(key, p)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("恭喜你获得\"", p, "\"的惩罚"))
			return
		}
	}
}