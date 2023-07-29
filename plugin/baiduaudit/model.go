package baiduaudit

import (
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"

	"github.com/FloatTech/floatbox/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 服务网址:https://console.bce.baidu.com/ai/?_=1665977657185#/ai/antiporn/overview/index
// 返回参数说明：https://cloud.baidu.com/doc/ANTIPORN/s/Nk3h6xbb2
type baiduRes struct {
	mu sync.Mutex `json:"-"`
	// LogID          int          `json:"log_id"`         // 请求唯一id
	Conclusion     string       `json:"conclusion"`     // 审核结果, 可取值：合规、不合规、疑似、审核失败
	ConclusionType int          `json:"conclusionType"` // 审核结果类型, 可取值1.合规, 2.不合规, 3.疑似, 4.审核失败
	Data           []*auditData `json:"data"`
	ErrorCode      int          `json:"error_code"` // 错误提示码, 失败才返回, 成功不返回
	ErrorMsg       string       `json:"error_msg"`  // 错误提示信息, 失败才返回, 成功不返回
}

// 禁言检测
func (bdres *baiduRes) audit(ctx *zero.Ctx, configpath string) {
	bdres.mu.Lock()
	defer bdres.mu.Unlock()
	// 如果返回类型为2（不合规）, 0为合规, 3为疑似
	if bdres.ConclusionType != 2 {
		return
	}
	// 创建消息ID
	mid := message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64))
	// 获取群配置
	group := config.groupof(ctx.Event.GroupID)
	// 检测群配置里的不检测类型白名单, 忽略掉不检测的违规类型
	for i, b := range group.copyWhiteListType() {
		if i == bdres.Data[0].SubType && b {
			return
		}
	}
	// 生成回复文本
	res := group.reply(bdres)
	// 撤回消息
	ctx.DeleteMessage(mid)
	// 查看是否启用撤回后禁言
	if group.DMBAN {
		// 从历史违规记录中获取指定用户
		user := group.historyof(ctx.Event.UserID)
		// 用户违规次数自增
		atomic.AddInt64(&user.Count, 1)
		user.mu.Lock()
		// 用户违规原因记录
		user.ResList = append(user.ResList, bdres)
		user.mu.Unlock()
		// 保存到json
		err := config.saveto(configpath)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
		}
		var bantime int64
		// 查看是否开启禁言累加功能, 并计算禁言时间
		if group.BANTimeAddEnable {
			bantime = atomic.LoadInt64(&user.Count) * group.BANTimeAddTime * 60
		} else {
			bantime = group.BANTime * 60
		}
		// 执行禁言
		ctx.SetThisGroupBan(ctx.Event.UserID, bantime)
	}
	// 查看是否开启撤回提示
	if group.DMRemind {
		res = append(res, message.At(ctx.Event.Sender.ID))
		ctx.Send(res)
	}
}

type auditData struct {
	// Type           int    `json:"type"`           // 审核主类型, 11：百度官方违禁词库、12：文本反作弊、13:自定义文本黑名单、14:自定义文本白名单
	SubType int `json:"subType"` // 审核子类型, 0:含多种类型, 具体看官方链接, 1:违禁违规、2:文本色情、3:敏感信息、4:恶意推广、5:低俗辱骂 6:恶意推广-联系方式、7:恶意推广-软文推广
	// Conclusion     string `json:"conclusion"`     // 审核结果, 可取值：合规、不合规、疑似、审核失败
	// ConclusionType int    `json:"conclusionType"` // 审核结果类型, 可取值1.合规, 2.不合规, 3.疑似, 4.审核失败
	Msg  string `json:"msg"` // 不合规项描述信息
	Hits []*hit `json:"hits"`
} // 不合规/疑似/命中白名单项详细信息.响应成功并且conclusion为疑似或不合规或命中白名单时才返回, 响应失败或conclusion为合规且未命中白名单时不返回.

type auditHistory struct {
	mu      sync.Mutex  `json:"-"`
	Count   int64       `json:"key2"`    // 被禁次数
	ResList []*baiduRes `json:"reslist"` // 禁言原因
}

type hit struct {
	// DatasetName string   `json:"datasetName"`           // 违规项目所属数据集名称
	Words []string `json:"words"` // 送检文本命中词库的关键词（备注：建议参考新字段“wordHitPositions”, 包含信息更丰富：关键词以及对应的位置及标签信息）
	// Probability float64  `json:"probability,omitempty"` // 不合规项置信度
} // 送检文本违规原因的详细信息

type keyConfig struct {
	mu     sync.Mutex       `json:"-"`
	Key1   string           `json:"key1"`   // 百度云服务内容审核key存储
	Key2   string           `json:"key2"`   // 百度云服务内容审核key存储
	Groups map[int64]*group `json:"groups"` // 群配置存储
}

func newconfig() (kc keyConfig) {
	kc.Groups = make(map[int64]*group, 64)
	return
}

func (kc *keyConfig) setkey(k1, k2 string) {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	kc.Key1 = k1
	kc.Key2 = k2
}

// 加载JSON配置文件
func (kc *keyConfig) load(filename string) error {
	if file.IsNotExist(filename) {
		return nil
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	kc.mu.Lock()
	defer kc.mu.Unlock()
	return json.NewDecoder(f).Decode(kc)
}

func (kc *keyConfig) isgroupexist(ctx *zero.Ctx) (ok bool) {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	_, ok = kc.Groups[ctx.Event.GroupID]
	return
}

// 获取群配置
func (kc *keyConfig) groupof(groupID int64) *group {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	g, ok := kc.Groups[groupID]
	if ok {
		return g
	}
	g = &group{
		TextAudit:          true,
		ImageAudit:         true,
		BANTime:            1,
		MaxBANTimeAddRange: 60,
		BANTimeAddTime:     1,
		AuditHistory:       map[int64]*auditHistory{},
	}
	kc.Groups[groupID] = g
	return g
}

// 保存配置文件
func (kc *keyConfig) saveto(filename string) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(kc)
}

type group struct {
	mu                 sync.Mutex
	Enable             mark                    // 是否启用内容审核
	TextAudit          mark                    // 文本检测
	ImageAudit         mark                    // 图像检测
	DMRemind           mark                    // 撤回提示
	MoreRemind         mark                    // 详细违规提示
	DMBAN              mark                    // 撤回后禁言
	BANTimeAddEnable   mark                    // 禁言累加
	BANTime            int64                   // 标准禁言时间, 禁用累加, 但开启禁言的的情况下采用该值
	MaxBANTimeAddRange int64                   // 最大禁言时间累加范围, 最高禁言时间
	BANTimeAddTime     int64                   // 禁言累加时间, 该值是开启禁累加功能后, 再次触发时, 根据被禁次数X该值计算出的禁言时间
	WhiteListType      [8]bool                 // 类型白名单, 处于白名单类型的违规, 不会被触发 0:含多种类型, 具体看官方链接, 1:违禁违规、2:文本色情、3:敏感信息、4:恶意推广、5:低俗辱骂 6:恶意推广-联系方式、7:恶意推广-软文推广
	AuditHistory       map[int64]*auditHistory // 被封禁用户列表
}

func (g *group) set(f func(g *group)) {
	g.mu.Lock()
	f(g)
	g.mu.Unlock()
}

func (g *group) setWhiteListType(typ int, ok bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.WhiteListType[typ] = ok
}

func (g *group) copyWhiteListType() [8]bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.WhiteListType
}

// 从群历史违规记录中获取用户
func (g *group) historyof(userID int64) *auditHistory {
	g.mu.Lock()
	defer g.mu.Unlock()
	audit, ok := g.AuditHistory[userID]
	if ok {
		return audit
	}
	// 如果没有用户, 则创建一个并返回
	if g.AuditHistory == nil {
		g.AuditHistory = make(map[int64]*auditHistory)
	}
	audit = &auditHistory{}
	g.AuditHistory[userID] = audit
	return audit
}

// 生成回复文本
func (g *group) reply(bdres *baiduRes) message.Message {
	g.mu.Lock()
	defer g.mu.Unlock()
	// 建立消息段
	msgs := make([]message.MessageSegment, 0, 8)
	// 生成简略审核结果回复
	msgs = append(msgs, message.Text(bdres.Conclusion, "\n"))
	// 查看是否开启详细审核内容提示, 并确定审核内容值为疑似, 或者不合规
	if !g.MoreRemind {
		return msgs
	}
	// 遍历返回的不合规数据, 生成详细违规内容
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
				// 检查是否是最后一个要打印的词条, 如果是则不加上逗号
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

type mark bool

// String 打印启用状态
func (em mark) String() string {
	if em {
		return "开启"
	}
	return "关闭"
}
