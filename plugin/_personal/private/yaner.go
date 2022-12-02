package yaner

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"archive/zip"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

const zbpPath = "/Users/liuyu.fang/Documents/ZeroBot-Plug/"

var (
	poke = rate.NewManager[int64](time.Minute*5, 6) // 戳一戳
	// Axis表情的 codechina 镜像
	res = "https://gitcode.net/weixin_49234624/zbpdata/-/raw/main/faceimg-liuyu/"
)

func init() { // 插件主体
	// 更新zbp
	zero.OnFullMatch("检查更新", zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var msg []string
			var img []byte
			err := fileZipTo(zbpPath+"ZeroBot-Plugin", zbpPath+"ZeroBot-Plugin"+time.Now().Format("_2006_01_02_15_04_05")+".zip")
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			msg = append(msg, "已经对旧版zbp压缩备份\n\n开始检查更新")
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			/*/ remote add
			cmd := exec.Command("git", "remote add upstream git@github.com:FloatTech/ZeroBot-Plugin.git")
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = zbpPath + "ZeroBot-Plugin"
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", stderr.String())
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
			// remote -v
			cmd = exec.Command("git", "remote -v")
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = zbpPath + "ZeroBot-Plugin"
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", stderr.String())
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
			// fetch*/
			cmd := exec.Command("git", "fetch", "upstream", "master")
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = zbpPath + "ZeroBot-Plugin"
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", stderr.String(), cmd.Dir)
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
			// merge
			cmd = exec.Command("git", "merge", "upstream/master")
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = zbpPath + "ZeroBot-Plugin"
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", stderr.String())
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
			// push
			cmd = exec.Command("git", "push", "-u", "origin", "master")
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = zbpPath + "ZeroBot-Plugin"
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", stderr.String())
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
			// pull
			cmd = exec.Command("git", "pull", "--tags", "-r", "origin", "master")
			msg = append(msg, "Command:", strings.Join(cmd.Args, " "))
			cmd.Dir = zbpPath + "ZeroBot-Plugin"
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				msg = append(msg, "StdErr:", stderr.String())
				// 输出图片
				img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
				return
			}
			msg = append(msg, "StdOut:", stdout.String())
			// 输出图片
			img, err = text.RenderToBase64(strings.Join(msg, "\n"), text.BoldFontFile, 1280, 50)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Image("base64://" + binary.BytesToString(img)))
		})
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
	zero.OnFullMatchGroup([]string{"重启", "restart", "kill", "洗手手"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
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
	engine := control.Register("yaner", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "基础指令",
		Help:             "柳如娮的基础指令",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"检测到唤醒环境:\n",
				"* CPU占用: ", cpuPercent(), "%\n",
				"* RAM占用: ", memPercent(), "%\n",
				"* 硬盘使用: ", diskPercent(), "\n确认ok。\n",
			))
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("嘿嘿，娮儿闪亮登场！锵↘锵↗~"))
		},
		OnDisable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("宝↗生↘永↗梦↘！！！！"))
		},
	})
	// 被喊名字
	engine.OnKeywordGroup([]string{"自我介绍", "你是谁", "你谁"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("你好，我叫柳如娮。\n你可以叫我娮儿、小娮，当然你叫我机器人也可以ಠಿ_ಠ"))
		})
	engine.OnFullMatch("", zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			time.Sleep(time.Second * 1)
			switch rand.Intn(3) {
			case 1:
				ctx.SendChain(randImage("WZ.jpg", "ZZZZ.gif"))
			default:
				ctx.SendChain(message.Text(
					[]string{
						nickname + "在窥屏哦",
						"我在听",
						"请问找" + nickname + "有什么事吗",
						"？怎么了",
					}[rand.Intn(4)],
				))
			}
		})
	// 戳一戳
	engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			if !poke.Load(ctx.Event.GroupID).AcquireN(1) {
				return // 最多戳6次
			}
			nickname := zero.BotConfig.NickName[0]
			switch rand.Intn(7) {
			case 1:
				time.Sleep(time.Second * 1)
				ctx.SendChain(randText("哼！（打手）"))
				ctx.SendChain(message.Poke(ctx.Event.UserID))
			default:
				time.Sleep(time.Second * 1)
				ctx.SendChain(randText(
					"哼！",
					"（打手）",
					nickname+"的脸不是拿来捏的！",
					nickname+"要生气了哦",
					"?",
				))
			}
		})
	engine.OnKeywordGroup([]string{"好吗", "行不行", "能不能", "可不可以"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			switch rand.Intn(4) {
			case 0:
				nickname := zero.BotConfig.NickName[0]
				if rand.Intn(2) == 0 {
					ctx.SendChain(message.Text(nickname + "..." + nickname + "觉得不行"))
				} else {
					ctx.SendChain(message.Text(nickname + "..." + nickname + "觉得可以！"))
				}
			case 1:
				ctx.SendChain(randImage("Ask-YES.jpg", "Ask-NO.jpg", "Ask-YES.jpg"))
			}
		})
}

// 打包成zip文件
func fileZipTo(src_dir string, zip_file_name string) error {
	// 创建：zip文件
	zipfile, err := os.Create(zip_file_name)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// 打开：zip文件
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// 遍历路径信息
	filepath.Walk(src_dir, func(path string, info os.FileInfo, _ error) error {

		// 如果是源路径，提前进行下一个遍历
		if path == src_dir {
			return nil
		}

		// 获取：文件头信息
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(path, src_dir+`/`)

		// 判断：文件是不是文件夹
		if info.IsDir() {
			header.Name += `/`
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}

		// 创建：压缩包头部信息
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			io.Copy(writer, file)
		}
		return nil
	})
	return nil
}

func randText(text ...string) message.MessageSegment {
	return message.Text(text[rand.Intn(len(text))])
}

func randImage(file ...string) message.MessageSegment {
	return message.Image(res + file[rand.Intn(len(file))])
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
