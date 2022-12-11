// Package base 基础指令
package base

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/FloatTech/floatbox/process"
	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func init() {
	go func() {
		process.SleepAbout1sTo2s()
		ctx := zero.GetBot(1015464740)
		m, ok := control.Lookup("yaner")
		if ok {
			gid := m.GetData(-2504407110)
			if gid != 0 {
				ctx.SendGroupMessage(gid, message.Text("我回来了😊"))
			} else {
				ctx.SendPrivateMessage(2504407110, message.Text("我回来了😊"))
			}
		}
		err := m.SetData(-2504407110, 0)
		if err != nil {
			ctx.SendPrivateMessage(2504407110, message.Text(err))
		}
	}()
	// 电脑状态
	zero.OnFullMatchGroup([]string{"检查身体", "自检", "启动自检", "系统状态"}, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"* CPU占用: ", cpuPercent(), "%\n",
				"* RAM占用: ", memPercent(), "%\n",
				"* 硬盘使用: ", diskPercent(),
			),
			)
		})
	// 重启
	zero.OnFullMatchGroup([]string{"重启", "洗手手"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			m, ok := control.Lookup("yaner")
			if ok {
				err := m.SetData(-2504407110, ctx.Event.GroupID)
				if err == nil {
					ctx.SendChain(message.Text("好的"))
				} else {
					ctx.SendPrivateMessage(2504407110, message.Text(err))
				}
			}
			os.Exit(0)
		})
	// 运行 CQ 码
	zero.OnPrefix("run", zero.SuperUserPermission).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			// 可注入，权限为主人
			ctx.Send(message.UnescapeCQCodeText(ctx.State["args"].(string)))
		})
	// 撤回最后的发言
	zero.OnRegex(`^\[CQ:reply,id=(.*)].*`, zero.KeywordRule("多嘴")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取消息id
			mid := ctx.State["regex_matched"].([]string)[1]
			// 撤回消息
			if ctx.Event.Message[1].Data["qq"] != "" {
				var nickname = zero.BotConfig.NickName[0]
				ctx.SendChain(message.Text("9494，要像", nickname, "一样乖乖的才行哟~"))
			} else {
				ctx.SendChain(message.Text("呜呜呜呜"))
			}
			ctx.DeleteMessage(message.NewMessageIDFromString(mid))
			ctx.DeleteMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64)))
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
