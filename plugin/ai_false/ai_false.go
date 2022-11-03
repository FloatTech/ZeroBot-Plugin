// Package aifalse 暂时只有服务器监控
package aifalse

import (
	"fmt"
	"math"
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	engine := control.Register("aifalse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "自检, 全局限速",
		Help: "- 查询计算机当前活跃度: [检查身体 | 自检 | 启动自检 | 系统状态]\n" +
			"- 设置默认限速为每 m [分钟 | 秒] n 次触发",
	})
	c, ok := control.Lookup("aifalse")
	if !ok {
		panic("register aifalse error")
	}
	m := c.GetData(0)
	n := (m >> 16) & 0xffff
	m &= 0xffff
	if m != 0 || n != 0 {
		ctxext.SetDefaultLimiterManagerParam(time.Duration(m)*time.Second, int(n))
		logrus.Infoln("设置默认限速为每", m, "秒触发", n, "次")
	}
	engine.OnFullMatchGroup([]string{"检查身体", "自检", "启动自检", "系统状态"}, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"* CPU占用: ", cpuPercent(), "%\n",
				"* RAM占用: ", memPercent(), "%\n",
				"* 硬盘使用: ", diskPercent(),
			),
			)
		})
	engine.OnRegex(`^设置默认限速为每\s*(\d+)\s*(分钟|秒)\s*(\d+)\s*次触发$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				ctx.SendChain(message.Text("ERROR: no such plugin"))
				return
			}
			m, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if ctx.State["regex_matched"].([]string)[2] == "分钟" {
				m *= 60
			}
			if m >= 65536 || m <= 0 {
				ctx.SendChain(message.Text("ERROR: interval too big"))
				return
			}
			n, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[3], 10, 64)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if n >= 65536 || n <= 0 {
				ctx.SendChain(message.Text("ERROR: burst too big"))
				return
			}
			ctxext.SetDefaultLimiterManagerParam(time.Duration(m)*time.Second, int(n))
			err = c.SetData(0, (m&0xffff)|((n<<16)&0xffff0000))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("设置默认限速为每", m, "秒触发", n, "次"))
		})
}

func cpuPercent() float64 {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return -1
	}
	return math.Round(percent[0])
}

func memPercent() float64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return -1
	}
	return math.Round(memInfo.UsedPercent)
}

func diskPercent() string {
	parts, err := disk.Partitions(true)
	if err != nil {
		return err.Error()
	}
	msg := ""
	for _, p := range parts {
		diskInfo, err := disk.Usage(p.Mountpoint)
		if err != nil {
			msg += "\n  - " + err.Error()
			continue
		}
		pc := uint(math.Round(diskInfo.UsedPercent))
		if pc > 0 {
			msg += fmt.Sprintf("\n  - %s(%dM) %d%%", p.Mountpoint, diskInfo.Total/1024/1024, pc)
		}
	}
	return msg
}
