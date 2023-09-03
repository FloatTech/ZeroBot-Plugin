// Package baiduaudit 百度内容审核
package baiduaudit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Baidu-AIP/golang-sdk/aip/censor"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/binary"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
)

var (
	bdcli  *censor.ContentCensorClient // 百度云审核服务Client
	txttyp = [...]string{
		0: "默认违禁词库",
		1: "违禁违规",
		2: "文本色情",
		3: "敏感信息",
		4: "恶意推广",
		5: "低俗辱骂",
		6: "恶意推广-联系方式",
		7: "恶意推广-软文推广",
	} // 文本类型
	config = newconfig() // 插件配置
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "百度内容审核",
		Help: "##该功能来自百度内容审核, 需购买相关服务, 并创建app##\n" +
			"- 获取BDAKey\n" +
			"- 配置BDAKey [API key] [Secret Key]\n" +
			"- 开启/关闭内容审核\n" +
			"- 开启/关闭撤回提示\n" +
			"- 开启/关闭详细提示\n" +
			"- 开启/关闭撤回禁言\n" +
			"##禁言时间设置## 禁言时间计算方式为：禁言次数*每次禁言累加时间,当达到最大禁言时间时, 再次触发按最大禁言时间计算\n" +
			"- 开启/关闭禁言累加\n" +
			"- 设置撤回禁言时间[分钟, 默认:1]\n" +
			"- 设置最大禁言时间[分钟, 默认:60,最大43200]\n" +
			"- 设置每次累加时间[分钟, 默认:1]\n" +
			"##检测类型设置## 类型编号列表:[1:违禁违规、2:文本色情、3:敏感信息、4:恶意推广、5:低俗辱骂 6:恶意推广-联系方式、7:恶意推广-软文推广]\n" +
			"- 查看检测类型\n" +
			"- 查看检测配置\n" +
			"- 设置检测类型[类型编号]\n" +
			"- 设置不检测类型[类型编号]\n" +
			"- 开启/关闭文本检测\n" +
			"- 开启/关闭图像检测\n" +
			"##测试功能##\n" +
			"- ^文本检测[文本内容]\n" +
			"- ^图像检测[图片]\n",
		PrivateDataFolder: "baiduaudit",
	})

	configpath := engine.DataFolder() + "config.json"
	err := config.load(configpath)
	if err != nil {
		logrus.Warnln("[baiduaudit] 加载配置错误:", err)
	} else if config.Key1 != "" && config.Key2 != "" {
		bdcli = censor.NewClient(config.Key1, config.Key2)
	}

	engine.OnFullMatch("获取BDAKey", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("接口key创建网址:\n" +
				"https://console.bce.baidu.com/ai/?_=1665977657185#/ai/antiporn/overview/index\n" +
				"免费8w次数领取地址:\n" +
				"https://console.bce.baidu.com/ai/?_=1665977657185#/ai/antiporn/overview/resource/getFree"))
		})

	engine.OnRegex("^查看检测(类型|配置)$", zero.AdminPermission, hasinit).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取群配置
			group := config.groupof(ctx.Event.GroupID)
			msg := ""
			k1 := ctx.State["regex_matched"].([]string)[1]
			if k1 == "类型" {
				sb := strings.Builder{}
				sb.WriteString("本群检测类型:")
				found := false
				// 遍历群检测类型名单
				for i, v := range group.copyWhiteListType() {
					if !v {
						found = true
						sb.WriteByte('\n')
						sb.WriteString(strconv.Itoa(i))
						sb.WriteByte('.')
						sb.WriteString(txttyp[i])
					}
				}
				if !found {
					sb.WriteString("无")
				}
				msg = sb.String()
			} else {
				// 生成配置文本
				msg = fmt.Sprintf("本群配置:\n"+
					"内容审核:%s\n"+
					"-文本:%s\n"+
					"-图像:%s\n"+
					"撤回提示:%s\n"+
					"-详细提示:%s\n"+
					"撤回禁言:%s\n"+
					"-禁言累加:%s\n"+
					"-撤回禁言时间:%v分钟\n"+
					"-每次累加时间:%v分钟\n"+
					"-最大禁言时间:%v分钟", group.Enable, group.TextAudit, group.ImageAudit, group.DMRemind, group.MoreRemind, group.DMBAN, group.BANTimeAddEnable, group.BANTime, group.BANTimeAddTime, group.MaxBANTimeAddRange)
			}
			b, err := text.RenderToBase64(msg, text.FontFile, 300, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("base64://" + binary.BytesToString(b)))
		})

	engine.OnRegex("^设置(不)?检测类型([0-7])$", zero.AdminPermission, hasinit).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			k1 := ctx.State["regex_matched"].([]string)[1]
			k2 := ctx.State["regex_matched"].([]string)[2]
			group := config.groupof(ctx.Event.GroupID)
			inputType, _ := strconv.Atoi(k2)
			group.setWhiteListType(inputType, k1 == "不")
			err := config.saveto(configpath)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("本群将%s检测%s类型内容", k1, txttyp[inputType])))
		})

	engine.OnRegex("^设置(最大|每次|撤回)(累加|禁言)时间(\\d{1,5})$", zero.AdminPermission, hasinit).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			k1 := ctx.State["regex_matched"].([]string)[1]
			k3 := ctx.State["regex_matched"].([]string)[3]
			time, _ := strconv.ParseInt(k1, 10, 64)
			config.groupof(ctx.Event.GroupID).set(func(g *group) {
				switch k1 {
				case "最大":
					g.MaxBANTimeAddRange = time
				case "每次":
					g.BANTimeAddTime = time
				case "撤回":
					g.BANTime = time
				}
			})
			err := config.saveto(configpath)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("本群%s禁言累加时间已设置为%s", k3, k1)))
		})

	engine.OnRegex("^(开启|关闭)(内容审核|撤回提示|撤回禁言|禁言累加|详细提示|文本检测|图像检测)$", zero.AdminPermission, hasinit).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			k1 := ctx.State["regex_matched"].([]string)[1]
			k2 := ctx.State["regex_matched"].([]string)[2]
			isEnable := mark(k1 == "开启")
			config.groupof(ctx.Event.GroupID).set(func(g *group) {
				switch k2 {
				case "内容审核":
					g.Enable = isEnable
				case "撤回提示":
					g.DMRemind = isEnable
				case "撤回禁言":
					g.DMBAN = isEnable
				case "禁言累加":
					g.BANTimeAddEnable = isEnable
				case "详细提示":
					g.MoreRemind = isEnable
				case "文本检测":
					g.TextAudit = isEnable
				case "图像检测":
					g.ImageAudit = isEnable
				}
			})
			err := config.saveto(configpath)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("本群%s已%s", k2, k1)))
		})

	engine.OnRegex(`^配置BDAKey\s(.*)\s(.*)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			k1 := ctx.State["regex_matched"].([]string)[1]
			k2 := ctx.State["regex_matched"].([]string)[2]
			bdcli = censor.NewClient(k1, k2)
			config.setkey(k1, k2)
			if bdcli != nil {
				err := config.saveto(configpath)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text("配置成功"))
			}
		})

	engine.OnMessage(config.isgroupexist).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		group := config.groupof(ctx.Event.GroupID)
		if !bool(group.Enable) {
			return
		}
		var bdres baiduRes
		var err error
		for _, elem := range ctx.Event.Message {
			switch elem.Type {
			case "image":
				if !group.ImageAudit || elem.Data["url"] == "" {
					continue
				}
				res := bdcli.ImgCensorUrl(elem.Data["url"], nil)
				bdres, err = parse2BaiduRes(res)
				if err != nil {
					continue
				}
			case "text":
				if !group.TextAudit || elem.Data["text"] == "" {
					continue
				}
				bdres, err = parse2BaiduRes(bdcli.TextCensor(elem.Data["text"]))
				if err != nil {
					continue
				}
			default:
				continue
			}
		}
		bdres.audit(ctx, configpath)
	})

	engine.OnPrefix("^文本检测", hasinit).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.ExtractPlainText()
			res := bdcli.TextCensor(args)
			bdres, err := parse2BaiduRes(res)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", bdres.ErrorMsg, "(", bdres.ErrorCode, ")"))
				return
			}
			ctx.Send(config.groupof(ctx.Event.GroupID).reply(&bdres))
		})

	engine.OnPrefix("^图像检测", hasinit).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			var urls []string
			for _, elem := range ctx.Event.Message {
				if elem.Type == "image" {
					if elem.Data["url"] != "" {
						urls = append(urls, elem.Data["url"])
					}
				}
			}
			if len(urls) == 0 {
				return
			}
			res := bdcli.ImgCensorUrl(urls[0], nil)
			bdres, err := parse2BaiduRes(res)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", bdres.ErrorMsg, "(", bdres.ErrorCode, ")"))
				return
			}
			ctx.Send(config.groupof(ctx.Event.GroupID).reply(&bdres))
		})
}

// 客户端是否初始化检测
func hasinit(ctx *zero.Ctx) bool {
	if bdcli == nil {
		ctx.SendChain(message.Text("Key未配置"))
		return false
	}
	return true
}

func parse2BaiduRes(resjson string) (bdres baiduRes, err error) {
	err = json.Unmarshal(binary.StringToBytes(resjson), &bdres)
	return
}
