// Package chat 提供聊天记录管理和 AI 模型交互功能
package chat

import (
	_ "embed"
	"strings"

	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/chat"
	"github.com/fumiama/deepinfra/model"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// SystemPrompt 将 README.md 内容嵌入为默认系统提示词
//
//go:embed README.md
var SystemPrompt string

var (
	// AtPrefix @用户时给 LLM 的前缀提示, 需要和提示词一致
	AtPrefix = ">>"
	// NameL 包裹用户名的左括号, 需要和提示词一致
	NameL = "【"
	// NameR 包裹用户名的右括号, 需要和提示词一致
	NameR = "】"
)

// lst 全局聊天记录，每群/每用户独立保存最近 8 条
var lst = chat.NewLog[*item](16, 8, "\n\n", "自己随机开启新话题")

func init() {
	// 注册 ZeroBot 消息钩子，记录所有非空文本
	zero.OnMessage(func(ctx *zero.Ctx) bool {
		txt := ctx.ExtractPlainText()
		ctx.State["__zbputil_chat_txt__"] = txt
		return txt != ""
	}).FirstPriority().SetBlock(false).Handle(func(ctx *zero.Ctx) {
		// 计算群组 ID（私聊时使用负的 UserID）
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		// 将用户消息追加到对应群组的聊天记录
		lst.Add(gid, &item{
			isatme:   ctx.Event.IsToMe,
			usr:      ctx.Event.Sender.Name(),
			txt:      ctx.State["__zbputil_chat_txt__"].(string),
			atprefix: AtPrefix, namel: NameL, namer: NameR,
		}, false)
	})
}

type item struct {
	isatme                 bool
	usr, txt               string
	atprefix, namel, namer string
}

func (item *item) String() string {
	sb := strings.Builder{}
	if item.isatme {
		sb.WriteString(item.atprefix)
	}
	if item.usr != "" {
		sb.WriteString(item.namel)
		sb.WriteString(item.usr)
		sb.WriteString(item.namer)
	}
	sb.WriteString(item.txt)
	return sb.String()
}

// AddChatReply 将 AI 回复追加到指定群组的聊天记录
func AddChatReply(grp int64, txt string) {
	lst.Add(grp, &item{
		isatme: false, txt: txt,
		atprefix: AtPrefix, namel: NameL, namer: NameR,
	}, true)
}

// GetChatContext 根据聊天记录构造可执行的 deepinfra 模型请求
func GetChatContext(p model.Protocol, grp int64, sysp string, isusersys bool) deepinfra.Model {
	return lst.Modelize(p, grp, sysp, isusersys)
}

// ResetChat 清空全局聊天记录，重新开始
func ResetChat() {
	lst.Reset()
}

// ResetChatIn 清空 grps 的聊天记录，重新开始
func ResetChatIn(grps ...int64) {
	lst.ResetIn(grps...)
}
