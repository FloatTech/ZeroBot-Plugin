package goba

import (
	_ "embed"
	"fmt"
	"time"
	"unicode/utf8"
)

//go:embed README.md
var sysp string

// Config stores mutable characteristics of the agent.
type Config struct {
	Nickname string // Nickname 昵称
	Sex      string // Sex 性别
	Chars    string // Chars 个性
}

func (ag *Agent) system(role PermRole, iter int, grp int64) (string, error) {
	tab, err := ag.perm.mdtable(role)
	if err != nil {
		return "", err
	}
	t := time.Now()
	typ := "群聊"
	if grp < 0 {
		typ = "私聊"
	}
	memory := ag.memoryof(grp)
	if len(memory) > 32768 {
		start := len(memory) - 32768
		for start < len(memory) && !utf8.ValidString(memory[start:]) {
			start++
		}
		memory = "（更早记忆未装载到本次请求，原文件保留）\n" + memory[start:]
	}
	return fmt.Sprintf(
		sysp, ag.cfg.Nickname, ag.cfg.Sex,
		ag.cfg.Chars, tab, memory,
		t.Format(time.RFC3339), t.Unix(), typ, iter,
	), nil
}
