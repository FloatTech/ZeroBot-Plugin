/*
暂时只有服务器监控
*/
package plugin_ai_false

import (
	"math"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	zero.OnFullMatchGroup([]string{"检查身体", "自检", "启动自检", "系统状态"}).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"* CPU占用率: ", getCpuPercent(), "%\n",
				"* RAM占用率: ", getMemPercent(), "%\n",
				"* 硬盘活动率: ", getDiskPercent(), "%",
			),
			)
		})
}

func getCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return math.Round(percent[0])
}

func getMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return math.Round(memInfo.UsedPercent)
}

func getDiskPercent() float64 {
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	return math.Round(diskInfo.UsedPercent)
}
