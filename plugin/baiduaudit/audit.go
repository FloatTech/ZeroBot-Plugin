// Package baiduaudit 百度内容审核
package baiduaudit

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Baidu-AIP/golang-sdk/aip/censor"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 服务网址:https://console.bce.baidu.com/ai/?_=1665977657185#/ai/antiporn/overview/index
// 返回参数说明：https://cloud.baidu.com/doc/ANTIPORN/s/Nk3h6xbb2
type baiduRes struct {
	LogID          int         `json:"log_id"`         // 请求唯一id
	Conclusion     string      `json:"conclusion"`     // 审核结果，可取值：合规、不合规、疑似、审核失败
	ConclusionType int         `json:"conclusionType"` // 审核结果类型，可取值1.合规，2.不合规，3.疑似，4.审核失败
	Data           []auditData `json:"data"`
	ErrorCode      int         `json:"error_code"` // 错误提示码，失败才返回，成功不返回
	ErrorMsg       string      `json:"error_msg"`  // 错误提示信息，失败才返回，成功不返回
}

type auditData struct {
	Type           int    `json:"type"`           // 审核主类型，11：百度官方违禁词库、12：文本反作弊、13:自定义文本黑名单、14:自定义文本白名单
	SubType        int    `json:"subType"`        // 审核子类型，0:含多种类型，具体看官方链接，1:违禁违规、2:文本色情、3:敏感信息、4:恶意推广、5:低俗辱骂 6:恶意推广-联系方式、7:恶意推广-软文推广
	Conclusion     string `json:"conclusion"`     // 审核结果，可取值：合规、不合规、疑似、审核失败
	ConclusionType int    `json:"conclusionType"` // 审核结果类型，可取值1.合规，2.不合规，3.疑似，4.审核失败
	Msg            string `json:"msg"`            // 不合规项描述信息
	Hits           []hit  `json:"hits"`
} // 不合规/疑似/命中白名单项详细信息。响应成功并且conclusion为疑似或不合规或命中白名单时才返回，响应失败或conclusion为合规且未命中白名单时不返回。

type hit struct {
	DatasetName string   `json:"datasetName"`           // 违规项目所属数据集名称
	Words       []string `json:"words"`                 // 送检文本命中词库的关键词（备注：建议参考新字段“wordHitPositions”，包含信息更丰富：关键词以及对应的位置及标签信息）
	Probability float64  `json:"probability,omitempty"` // 不合规项置信度
} //	送检文本违规原因的详细信息
type keyConfig struct {
	Key1   string          `json:"key1"`   // 百度云服务内容审核key存储
	Key2   string          `json:"key2"`   // 百度云服务内容审核key存储
	Groups map[int64]group `json:"groups"` // 群配置存储
}

type group struct {
	Enable             EnableMark             // 是否启用内容审核
	TextAudit          EnableMark             // 文本检测
	ImageAudit         EnableMark             // 图像检测
	DMRemind           EnableMark             // 撤回提示
	MoreRemind         EnableMark             // 详细违规提示
	DMBAN              EnableMark             // 撤回后禁言
	BANTimeAddEnable   EnableMark             // 禁言累加
	BANTime            int64                  // 标准禁言时间，禁用累加，但开启禁言的的情况下采用该值
	MaxBANTimeAddRange int64                  // 最大禁言时间累加范围，最高禁言时间
	BANTimeAddTime     int64                  // 禁言累加时间，该值是开启禁累加功能后，再次触发时，根据被禁次数X该值计算出的禁言时间
	WhiteListType      [8]bool                // 类型白名单，处于白名单类型的违规，不会被触发 0:含多种类型，具体看官方链接，1:违禁违规、2:文本色情、3:敏感信息、4:恶意推广、5:低俗辱骂 6:恶意推广-联系方式、7:恶意推广-软文推广
	AuditHistory       map[int64]auditHistory // 被封禁用户列表
}

// EnableMark 启用：●，禁用：○
type EnableMark bool

// String 打印启用状态
func (em EnableMark) String() string {
	if em {
		return "开启"
	}
	return "关闭"
}

type auditHistory struct {
	Count   int64      `json:"key2"`    // 被禁次数
	ResList []baiduRes `json:"reslist"` // 禁言原因
}

var bdcli *censor.ContentCensorClient // 百度云审核服务Client
var typetext = [8]string{
	0: "默认违禁词库",
	1: "违禁违规",
	2: "文本色情",
	3: "敏感信息",
	4: "恶意推广",
	5: "低俗辱骂",
	6: "恶意推广-联系方式",
	7: "恶意推广-软文推广",
} // 文本类型

var (
	config     keyConfig // 插件配置
	configinit bool      // 配置初始化
	configpath string    // 配置路径
)

func init() {
	engine := control.Register("baiduaudit", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "百度内容审核",
		Help: "##该功能来自百度内容审核，需购买相关服务，并创建app##\n" +
			"- 获取BDAKey\n" +
			"- 配置BDAKey [API key] [Secret Key]\n" +
			"- 开启/关闭内容审核\n" +
			"- 开启/关闭撤回提示\n" +
			"- 开启/关闭详细提示\n" +
			"- 开启/关闭撤回禁言\n" +
			"##禁言时间设置## 禁言时间计算方式为：禁言次数*每次禁言累加时间,当达到最大禁言时间时，再次触发按最大禁言时间计算\n" +
			"- 开启/关闭禁言累加\n" +
			"- 设置撤回禁言时间[分钟，默认:1]\n" +
			"- 设置最大禁言时间[分钟，默认:60,最大43200]\n" +
			"- 设置每次累加时间[分钟，默认:1]\n" +
			"##检测类型设置## 类型编号列表:[1:违禁违规、2:文本色情、3:敏感信息、4:恶意推广、5:低俗辱骂 6:恶意推广-联系方式、7:恶意推广-软文推广]\n" +
			"- 查看检测类型\n" +
			"- 查看检测配置\n" +
			"- 设置检测类型[类型编号]\n" +
			"- 设置不检测类型[类型编号]\n" +
			"- 开启/关闭文本检测\n" +
			"- 开启/关闭图像检测\n" +
			"##测试功能##\n" +
			"- 测试文本检测[文本内容]\n" +
			"- 测试图像检测[图片]\n",
		PrivateDataFolder: "baiduaudit",
	})
	configpath = engine.DataFolder() + "config.json"
	loadConfig()
	if configinit {
		bdcli = censor.NewClient(config.Key1, config.Key2)
	}
	engine.OnFullMatch("获取BDAKey", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("接口key创建网址:\n" +
				"https://console.bce.baidu.com/ai/?_=1665977657185#/ai/antiporn/overview/index\n" +
				"免费8w次数领取地址:\n" +
				"https://console.bce.baidu.com/ai/?_=1665977657185#/ai/antiporn/overview/resource/getFree"))
		})

	engine.OnRegex("^查看检测(类型|配置)$", zero.AdminPermission, clientCheck).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取群配置
			group := getGroup(ctx.Event.GroupID)
			var msgs string
			k1 := ctx.State["regex_matched"].([]string)[1]
			if k1 == "类型" {
				msgs += "本群检测类型:"
				find := false
				// 遍历群检测类型名单
				for i, v := range group.WhiteListType {
					if !v {
						find = true
						msgs += fmt.Sprint("\n", i, ".", typetext[i])
					}
				}
				if !find {
					msgs += "无"
				}
			} else {
				// 生成配置文本
				msgs = fmt.Sprintf("本群配置:\n"+
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
			b, err := text.RenderToBase64(msgs, text.FontFile, 300, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("base64://" + binary.BytesToString(b)))
		})
	engine.OnRegex("^设置(不)?检测类型([01234567])$", zero.AdminPermission, clientCheck).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			defer jsonSave(config, configpath)
			k1 := ctx.State["regex_matched"].([]string)[1]
			k2 := ctx.State["regex_matched"].([]string)[2]
			group := getGroup(ctx.Event.GroupID)
			inputType, _ := strconv.Atoi(k2)
			if k1 == "不" {
				group.WhiteListType[inputType] = true // 不检测：则进入类型白名单
			} else {
				group.WhiteListType[inputType] = false // 检测：则退出白名单
			}
			config.Groups[ctx.Event.GroupID] = group
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("本群将%s检测%s类型内容", k1, typetext[inputType])))
		})
	engine.OnRegex("^设置(最大|每次|撤回)(累加|禁言)时间(\\d{1,5})$", zero.AdminPermission, clientCheck).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			defer jsonSave(config, configpath)
			k1 := ctx.State["regex_matched"].([]string)[1]
			k3 := ctx.State["regex_matched"].([]string)[3]
			group := getGroup(ctx.Event.GroupID)
			time, _ := strconv.ParseInt(k1, 10, 64)
			switch k1 {
			case "最大":
				group.MaxBANTimeAddRange = time
			case "每次":
				group.BANTimeAddTime = time
			case "撤回":
				group.BANTime = time
			}
			config.Groups[ctx.Event.GroupID] = group
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("本群%s禁言累加时间已设置为%s", k3, k1)))
		})
	engine.OnRegex("^(开启|关闭)(内容审核|撤回提示|撤回禁言|禁言累加|详细提示|文本检测|图像检测)$", zero.AdminPermission, clientCheck).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			defer jsonSave(config, configpath)
			k1 := ctx.State["regex_matched"].([]string)[1]
			k2 := ctx.State["regex_matched"].([]string)[2]
			isEnable := EnableMark(false)
			group := getGroup(ctx.Event.GroupID)
			if k1 == "开启" {
				isEnable = true
			}
			switch k2 {
			case "内容审核":
				group.Enable = isEnable
			case "撤回提示":
				group.DMRemind = isEnable
			case "撤回禁言":
				group.DMBAN = isEnable
			case "禁言累加":
				group.BANTimeAddEnable = isEnable
			case "详细提示":
				group.MoreRemind = isEnable
			case "文本检测":
				group.TextAudit = isEnable
			case "图像检测":
				group.ImageAudit = isEnable
			}
			config.Groups[ctx.Event.GroupID] = group
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("本群%s已%s", k2, k1)))
		})
	engine.OnRegex(`^配置BDAKey\s(.*)\s(.*)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			k1 := ctx.State["regex_matched"].([]string)[1]
			k2 := ctx.State["regex_matched"].([]string)[2]
			bdcli = censor.NewClient(k1, k2)
			config.Key1 = k1
			config.Key2 = k2
			if bdcli != nil {
				jsonSave(config, configpath)
				ctx.SendChain(message.Text("配置成功"))
			}
		})
	engine.OnMessage().SetBlock(false).Handle(func(ctx *zero.Ctx) {
		group, ok := config.Groups[ctx.Event.GroupID]
		// 如果没该配置，或者审核功能未开启直接跳过
		if !ok || !bool(group.Enable) {
			return
		}
		for _, elem := range ctx.Event.Message {
			switch elem.Type {
			case "image":
				if !group.ImageAudit {
					return
				}
				res := bdcli.ImgCensorUrl(elem.Data["url"], nil)
				bdres, err := jsonToBaiduRes(res)
				if err != nil {
					ctx.SendChain(message.Text("Error:", bdres.ErrorMsg, "(", bdres.ErrorCode, ")"))
					return
				}
				banCheck(ctx, bdres)

			case "text":
				if !group.TextAudit {
					return
				}
				res := bdcli.TextCensor(elem.Data["text"])
				bdres, err := jsonToBaiduRes(res)
				if err != nil {
					ctx.SendChain(message.Text("Error:", bdres.ErrorMsg, "(", bdres.ErrorCode, ")"))
					return
				}
				banCheck(ctx, bdres)
			}
		}
	})
	engine.OnPrefix("文本检测", clientCheck).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if bdcli == nil {
				ctx.SendChain(message.Text("Key未配置"))
				return
			}
			args := ctx.ExtractPlainText()
			res := bdcli.TextCensor(args)
			bdres, err := jsonToBaiduRes(res)
			if err != nil {
				ctx.SendChain(message.Text("Error:", bdres.ErrorMsg, "(", bdres.ErrorCode, ")"))
				return
			}
			group := getGroup(ctx.Event.GroupID)
			ctx.SendChain(buildResp(bdres, group)...)
		})
	engine.OnPrefix("^图像检测", clientCheck).SetBlock(false).
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
			bdres, err := jsonToBaiduRes(res)
			if err != nil {
				ctx.SendChain(message.Text("Error:", bdres.ErrorMsg, "(", bdres.ErrorCode, ")"))
				return
			}
			group := getGroup(ctx.Event.GroupID)
			ctx.SendChain(buildResp(bdres, group)...)
		})
}

// 禁言检测
func banCheck(ctx *zero.Ctx, bdres baiduRes) {
	// 如果返回类型为2（不合规），0为合规，3为疑似
	if bdres.ConclusionType == 2 {
		// 创建消息ID
		mid := message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64))
		// 获取群配置
		group := getGroup(ctx.Event.GroupID)
		// 检测群配置里的不检测类型白名单，忽略掉不检测的违规类型
		for i, b := range group.WhiteListType {
			if i == bdres.Data[0].SubType && b {
				return
			}
		}
		// 生成回复文本
		res := buildResp(bdres, group)
		// 撤回消息
		ctx.DeleteMessage(mid)
		// 查看是否启用撤回后禁言
		if group.DMBAN {
			// 从历史违规记录中获取指定用户
			user := group.getUser(ctx.Event.UserID)
			// 用户违规次数自增
			user.Count++
			// 用户违规原因记录
			user.ResList = append(user.ResList, bdres)
			// 覆写该用户到群违规记录中
			group.AuditHistory[ctx.Event.UserID] = user
			// 覆写该群信息
			config.Groups[ctx.Event.GroupID] = group
			// 保存到json
			jsonSave(config, configpath)
			var bantime int64
			// 查看是否开启禁言累加功能，并计算禁言时间
			if group.BANTimeAddEnable {
				bantime = user.Count * group.BANTimeAddTime * 60
			} else {
				bantime = group.BANTime * 60
			}
			// 执行禁言
			ctx.SetGroupBan(ctx.Event.GroupID, ctx.Event.UserID, bantime)
		}
		// 查看是否开启撤回提示
		if group.DMRemind {
			res = append(res, message.At(ctx.Event.Sender.ID))
			ctx.SendChain(res...)
		}
	}
}

// 获取群配置
func getGroup(groupID int64) group {
	g, ok := config.Groups[groupID]
	if ok {
		return g
	}
	if config.Groups == nil {
		config.Groups = make(map[int64]group)
	}
	g = group{
		TextAudit:          true,
		ImageAudit:         true,
		BANTime:            1,
		MaxBANTimeAddRange: 60,
		BANTimeAddTime:     1,
		WhiteListType:      [8]bool{},
		AuditHistory:       map[int64]auditHistory{},
	}
	config.Groups[groupID] = g
	return g
}

// 从群历史违规记录中获取用户
func (group *group) getUser(userID int64) auditHistory {
	audit, ok := group.AuditHistory[userID]
	if ok {
		return audit
	}
	// 如果没有用户，则创建一个并返回
	if group.AuditHistory == nil {
		group.AuditHistory = make(map[int64]auditHistory)
	}
	audit = auditHistory{0, []baiduRes{}}
	group.AuditHistory[userID] = audit
	return audit
}

// 客户端是否初始化检测
func clientCheck(ctx *zero.Ctx) bool {
	if bdcli == nil {
		ctx.SendChain(message.Text("Key未配置"))
		return false
	}
	return true
}

// 加载JSON配置文件
func loadConfig() {
	if file.IsExist(configpath) {
		data, err := os.OpenFile(configpath, os.O_RDONLY, 0755)
		if err != nil {
			panic(err)
		}
		err = json.NewDecoder(data).Decode(&config)
		if err != nil {
			panic(err)
		}
		configinit = true
	} else {
		config = keyConfig{}
		configinit = false
	}
}

// 保存配置文件
func jsonSave(v keyConfig, path string) {
	jsf, _ := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(jsf) // 结束时关闭句柄，释放资源
	err := json.NewEncoder(jsf).Encode(v)
	if err != nil {
		fmt.Println(err)
	}
}

// JSON反序列化
func jsonToBaiduRes(resjson string) (baiduRes, error) {
	var bdres baiduRes
	err := json.Unmarshal(binary.StringToBytes(resjson), &bdres)
	return bdres, err
}

// 生成回复文本
func buildResp(bdres baiduRes, group group) []message.MessageSegment {
	// 建立消息段
	msgs := make([]message.MessageSegment, 0, 8)
	// 生成简略审核结果回复
	msgs = append(msgs, message.Text(bdres.Conclusion, "\n"))
	// 查看是否开启详细审核内容提示，并确定审核内容值为疑似，或者不合规
	if !group.MoreRemind {
		return msgs
	}
	// 遍历返回的不合规数据，生成详细违规内容
	for i, datum := range bdres.Data {
		msgs = append(msgs, message.Text("[", i, "]:", datum.Msg, "\n"))
		// 检查命中词条是否大于0
		if len(datum.Hits) == 0 {
			return msgs
		}
		// 遍历打印命中的违规词条
		for _, hit := range datum.Hits {
			if len(datum.Hits) == 0 {
				return msgs
			}
			msgs = append(msgs, message.Text("("))
			for i4, i3 := range hit.Words {
				// 检查是否是最后一个要打印的词条，如果是则不加上逗号
				if i4 != len(hit.Words)-1 {
					msgs = append(msgs, message.Text(i3, ","))
				} else {
					msgs = append(msgs, message.Text(i3))
				}
			}
			msgs = append(msgs, message.Text(")"))
		}
	}
	return msgs
}
