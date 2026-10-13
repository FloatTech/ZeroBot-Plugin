package control

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/wdvxdr1123/ZeroBot/message"
)

func TestServiceListPagesAndNodePayload(t *testing.T) {
	entries := make([]serviceListEntry, 45)
	for i := range entries {
		entries[i] = serviceListEntry{Name: fmt.Sprintf("plugin-%02d", i), Brief: "功能说明", Help: "/命令 [参数]\n第二行用法", Enabled: i%2 == 0}
	}
	seen := map[string]bool{}
	for page := 1; page <= 3; page++ {
		got, pages, err := serviceListPage(fmt.Sprint(page), len(entries))
		if err != nil || got != page || pages != 3 {
			t.Fatal("pagination failed")
		}
		nodes := serviceListMessage(entries, got, pages, "测试机器人", 87654321, page != 1)
		if len(nodes) != 1+min(20, len(entries)-(page-1)*20) {
			t.Fatal("page node bounds")
		}
		for i, node := range nodes {
			if node.Type != "node" || node.Data["uin"] != "87654321" || node.Data["name"] != "测试机器人" {
				t.Fatal("sender must be bot")
			}
			text := message.ParseMessageFromString(node.Data["content"]).ExtractPlainText()
			if i == 0 {
				if !strings.Contains(text, fmt.Sprintf("第 %d/3 页", page)) || !strings.Contains(text, "共 45 个") {
					t.Fatal("page header missing")
				}
				continue
			}
			entry := entries[(page-1)*20+i-1]
			if seen[entry.Name] || !strings.Contains(text, entry.Name) || !strings.Contains(text, entry.Help) || !strings.Contains(text, "使用方式") {
				t.Fatal("lost/duplicate plugin usage")
			}
			if strings.Contains(text, "当前会话：启用") != entry.Enabled {
				t.Fatal("effective state not represented")
			}
			seen[entry.Name] = true
		}
		wire, err := json.Marshal(nodes)
		if err != nil || !utf8.Valid(wire) {
			t.Fatal("invalid node serialization")
		}
	}
	if len(seen) != 45 {
		t.Fatal("catalog pagination lost plugins")
	}
	for _, arg := range []string{"0", "-1", "+1", "01", "2 3", "abc", "99999999999999999999999", "4"} {
		if _, _, err := serviceListPage(arg, 45); err == nil {
			t.Fatal("invalid page accepted", arg)
		}
	}
	if page, pages, err := serviceListPage(" \t", 0); page != 1 || pages != 1 || err != nil {
		t.Fatal("empty list page invalid")
	}
	if nodes := serviceListMessage(nil, 1, 1, "", 87654321, false); len(nodes) != 2 || !strings.Contains(nodes[1].Data["content"], "未加载插件") {
		t.Fatal("empty catalog lacks explanation")
	}
}

func TestServiceListMissingHelpUnicodeAndLiteralCQ(t *testing.T) {
	entry := serviceListEntry{Name: "llm", Brief: "说明", Help: strings.Repeat("汉🙂\n", 700) + "尾部", Enabled: true}
	text := serviceListNodeText(entry)
	if !utf8.ValidString(text) || utf8.RuneCountInString(text) > 1200 || !strings.HasSuffix(text, "完整说明请发送 /用法llm") {
		t.Fatal("long help corrupted or silently truncated")
	}
	if text := serviceListNodeText(serviceListEntry{Name: "no-help"}); !strings.Contains(text, "未提供使用说明") || !strings.Contains(text, "暂无功能说明") {
		t.Fatal("missing usage falsely supplied")
	}
	nodes := serviceListMessage([]serviceListEntry{{Name: "literal", Help: "用法：[CQ:image,file=bad] <script>text</script>"}}, 1, 1, "机器人", 87654321, false)
	content := message.ParseMessageFromString(nodes[1].Data["content"])
	if len(content) != 1 || content[0].Type != "text" || !strings.Contains(content.ExtractPlainText(), "[CQ:image,file=bad]") {
		t.Fatal("usage became executable CQ segment")
	}
}
