package ASfalse

import (
	"math/rand"
    "time"	
	"github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/disk"  
    "github.com/shirou/gopsutil/mem"
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	PRIO   = -1
	RES    = "file:///E:/Picture data/"
	ENABLE = true
)


func GetCpuPercent() float64 {
	percent, _:= cpu.Percent(time.Second, false)
	return percent[0]
}
 
func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}
 
func GetDiskPercent() float64 {
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	return diskInfo.UsedPercent
}

func init() { // 插件主体
	zero.OnKeywordGroup([]string{"身体检查","自检","启动自检","系统状态"}, zero.AdminPermission).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"人家当前CPU占用率是:  ",GetCpuPercent(),"%\n",
				"人家当前RAM占用率是:  ",GetMemPercent(),"%\n",
				"人家当前硬盘活动率是:  ", GetDiskPercent(), "%\n",
			),
			)
		})
}

func randText(text ...string) message.MessageSegment {
	length := len(text)
	return message.Text(text[rand.Intn(length)])
}

func randImage(file ...string) message.MessageSegment {
	length := len(file)
	return message.Image(RES + file[rand.Intn(length)])
}

func randRecord(file ...string) message.MessageSegment {
	length := len(file)
	return message.Record(RES + file[rand.Intn(length)])
}
func strToInt(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}
