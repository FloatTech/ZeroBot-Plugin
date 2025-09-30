// Package abineundo provides an explicit "from the beginning" (Latin: "ab ineundō")
// initialization anchor.
//
// Name origin:
//
//	Latin phrase "ab ineundō" meaning "from which is to be begun".
//
// Purpose:
//
//	Place this package at the very top of top-level main.go so its init (present
//	or future) executes before other plugin packages, filling in a predictable
//	plugin priority.
//
// Typical usage:
//
//	import (
//	    _ "github.com/your/module/abineundo" // priority anchor
//	    // ... other imports ...
//	)
//
// A blank identifier import preserves ordering side-effects without expanding the
// exported API surface.
//
// (No further code is required here; the package's presence alone defines ordering semantics.)
package abineundo

import (
	"bufio"
	_ "embed"
	"regexp"
	"strings"

	"github.com/FloatTech/zbputils/control"
	"github.com/sirupsen/logrus"
)

//go:embed ref/main/main.go
var maincode string

//go:embed ref/custom/register.go
var customcode string

const (
	statusnone = iota
	statushigh
	statushighend
	statusmid
	statusmidend
	statuslow
	statuslowend
)

var (
	priore       = regexp.MustCompile(`^\t// -{28}(高|中|低)优先级区-{28} //$`)
	mainpluginre = regexp.MustCompile(`^\t_ "github\.com/FloatTech/ZeroBot-Plugin/plugin/(\w+)"\s+// `)
	custpluginre = regexp.MustCompile(`^\t_ "github\.com/FloatTech/ZeroBot-Plugin/custom/plugin/(\w+)"\s+// `)
)

func init() {
	highprios := make([]string, 0, 64)
	midprios := make([]string, 0, 64)
	lowprios := make([]string, 0, 64)

	status := statusnone
	scanner := bufio.NewScanner(strings.NewReader(maincode))
	for scanner.Scan() {
		line := scanner.Text()

		prioword := ""
		match := priore.FindStringSubmatch(line)
		if len(match) > 1 {
			prioword = match[1]
		}
		switch prioword {
		case "高":
			switch status {
			case statusnone:
				status = statushigh
			case statushigh:
				status = statushighend
			default:
				panic("unexpected")
			}
		case "中":
			switch status {
			case statushighend:
				status = statusmid
			case statusmid:
				status = statusmidend
			default:
				panic("unexpected")
			}
		case "低":
			switch status {
			case statusmidend:
				status = statuslow
			case statuslow:
				status = statuslowend
			default:
				panic("unexpected")
			}
		default:
			switch status {
			case statusnone: // 还未开始匹配
				continue
			case statuslowend: // 匹配已结束
				break
			default: // 继续匹配插件
			}
		}

		// 在对应优先级区域内匹配插件
		if matches := mainpluginre.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			switch status {
			case statushigh:
				highprios = append(highprios, name)
			case statusmid:
				midprios = append(midprios, name)
			case statuslow:
				lowprios = append(lowprios, name)
			default: // 在不该匹配到插件的区域匹配到
				panic("unexpected")
			}
		}
	}

	custprios := make([]string, 0, 64)

	scanner = bufio.NewScanner(strings.NewReader(customcode))
	for scanner.Scan() {
		line := scanner.Text()

		if matches := custpluginre.FindStringSubmatch(line); len(matches) > 1 {
			custprios = append(custprios, matches[1])
		}
	}

	// 生成最终插件优先级表
	m := make(map[string]uint64, 4*(len(highprios)+len(midprios)+len(lowprios)+len(custprios)))
	i := 0
	for _, name := range highprios {
		m[name] = (uint64(i) + 1) * 10
		logrus.Debugln("[ab] set high plugin", name, "prio to", m[name])
		i++
	}
	for _, name := range custprios {
		m[name] = (uint64(i) + 1) * 10
		logrus.Debugln("[ab] set cust plugin", name, "prio to", m[name])
		i++
	}
	for _, name := range midprios {
		m[name] = (uint64(i) + 1) * 10
		logrus.Debugln("[ab] set mid plugin", name, "prio to", m[name])
		i++
	}
	for _, name := range lowprios {
		m[name] = (uint64(i) + 1) * 10
		logrus.Debugln("[ab] set low plugin", name, "prio to", m[name])
		i++
	}

	control.LoadCustomPriority(m)
}
