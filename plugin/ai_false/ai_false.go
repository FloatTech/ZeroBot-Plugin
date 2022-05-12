// Package aifalse 暂时只有服务器监控
package aifalse

import (
	"fmt"
	"math"
	"os"
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func init() { // 插件主体
	engine := control.Register("aifalse", &control.Options{
		DisableOnDefault: false,
		Help: "AIfalse\n" +
			"- 查询计算机当前活跃度: [检查身体 | 自检 | 启动自检 | 系统状态]",
	})
	engine.OnFullMatchGroup([]string{"检查身体", "自检", "启动自检", "系统状态"}, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			temp := fmt.Sprint(
				"* CPU占用: ", cpuPercent(), "%\n",
				"* RAM占用: ", memPercent(), "%\n",
				"* 硬盘使用: ", diskPercent())
			txt, err := text.RenderToBase64(temp, text.FontFile, 230, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + helper.BytesToString(txt))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
		})
	engine.OnFullMatch("清理缓存", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := os.RemoveAll("data/cache/*")
			if err != nil {
				ctx.SendChain(message.Text("错误: ", err.Error()))
			} else {
				ctx.SendChain(message.Text("成功!"))
			}
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
