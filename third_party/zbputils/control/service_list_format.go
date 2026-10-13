package control

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wdvxdr1123/ZeroBot/message"
)

const serviceListPageSize = 20
const serviceListNodeLimit = 1200

type serviceListEntry struct {
	Name, Brief, Help string
	Enabled           bool
}

func serviceListPage(arg string, total int) (int, int, error) {
	pages := max(1, (total+serviceListPageSize-1)/serviceListPageSize)
	arg = strings.TrimSpace(arg)
	page := 1
	if arg != "" {
		value, err := strconv.Atoi(arg)
		if err != nil || value < 1 || strconv.Itoa(value) != arg {
			return 0, pages, fmt.Errorf("页码须为正整数，例如 /服务列表 2")
		}
		page = value
	}
	if page > pages {
		return 0, pages, fmt.Errorf("共 %d 页，请发送 /服务列表 1 到 /服务列表 %d", pages, pages)
	}
	return page, pages, nil
}

func serviceListNodeText(entry serviceListEntry) string {
	state := "禁用"
	if entry.Enabled {
		state = "启用"
	}
	brief, help := strings.TrimSpace(entry.Brief), strings.TrimSpace(entry.Help)
	if brief == "" {
		brief = "暂无功能说明"
	}
	if help == "" {
		help = "该插件未提供使用说明，请查看项目文档。"
	}
	text := fmt.Sprintf("插件：%s\n说明：%s\n当前会话：%s\n\n使用方式：\n%s", entry.Name, brief, state, help)
	runes := []rune(text)
	if len(runes) <= serviceListNodeLimit {
		return text
	}
	footer := []rune("\n…用法过长，完整说明请发送 /用法" + entry.Name)
	if len(footer) > 200 {
		footer = footer[:200]
	}
	return string(runes[:serviceListNodeLimit-len(footer)]) + string(footer)
}

func serviceListMessage(entries []serviceListEntry, page, pages int, nickname string, selfID int64, group bool) message.Message {
	scope := "当前私聊"
	if group {
		scope = "当前群"
	}
	header := fmt.Sprintf("服务列表与用法 · 第 %d/%d 页\n共 %d 个已加载插件，每页最多 %d 个。\n状态范围：%s；插件启用后仍需满足各自权限和触发条件。\n翻页：/服务列表 页码；完整用法：/用法插件名。群内请 @ 机器人。", page, pages, len(entries), serviceListPageSize, scope)
	if page < pages {
		header += fmt.Sprintf("\n下一页：/服务列表 %d", page+1)
	}
	if page > 1 {
		header += fmt.Sprintf("\n上一页：/服务列表 %d", page-1)
	}
	if nickname == "" {
		nickname = "ZeroBot 服务列表"
	}
	result := message.Message{message.CustomNode(nickname, selfID, message.Message{message.Text(header)})}
	start := (page - 1) * serviceListPageSize
	if len(entries) == 0 {
		return append(result, message.CustomNode(nickname, selfID, message.Message{message.Text("当前程序未加载插件。")}))
	}
	for _, entry := range entries[start:min(start+serviceListPageSize, len(entries))] {
		result = append(result, message.CustomNode(nickname, selfID, message.Message{message.Text(serviceListNodeText(entry))}))
	}
	return result
}
