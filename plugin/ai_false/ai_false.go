// Package aifalse 暂时只有服务器监控
package aifalse

import (
	"bytes"
	"image"
	"math"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/rendercard"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/disintegration/imaging"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	backgroundURL = "https://iw233.cn/api.php?sort=mp"
	referer       = "https://weibo.com/"
)

var boottime time.Time

func init() { // 插件主体
	boottime = time.Now()
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
			img, err := drawstatus(ctx.Event.SelfID, zero.BotConfig.NickName[0])
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if id := ctx.SendChain(message.ImageBytes(img)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
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

func drawstatus(uid int64, botname string) ([]byte, error) {

	disksinfo, err := disks()
	if err != nil {
		return nil, err
	}
	diskh := 40 + (20+50)*len(disksinfo) + 40 - 20

	minfo, err := moreinfo()
	if err != nil {
		return nil, err
	}
	minfoh := 30 + (20+32*72/96)*len(minfo) + 30 - 20

	canvas := gg.NewContext(1280, 70+250+40+380+diskh+40+minfoh+40+70)

	url, err := bilibili.GetRealURL(backgroundURL)
	if err != nil {
		return nil, err
	}
	data, err := web.RequestDataWith(web.NewDefaultClient(), url, "", referer, "", nil)
	if err != nil {
		return nil, err
	}

	back, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bh, bw, ch, cw := float64(back.Bounds().Dy()), float64(back.Bounds().Dx()), float64(canvas.H()), float64(canvas.W())

	title := gg.NewContext(canvas.W()-70-70, 250)

	textcard := gg.NewContext(canvas.W()-70-70, 380)

	diskcard := gg.NewContext(canvas.W()-70-70, diskh)

	moreinfocard := gg.NewContext(canvas.W()-70-70, minfoh)

	if bh/bw < ch/cw {
		back = img.Size(back, int(bw*ch/bh), int(bh*ch/bh)).Im
		canvas.DrawImageAnchored(back, canvas.W()/2, canvas.H()/2, 0.5, 0.5)
		title.DrawImageAnchored(imaging.Blur(back, 8), title.W()/2, canvas.H()/2-70, 0.5, 0.5)
		textcard.DrawImageAnchored(imaging.Blur(back, 8), textcard.W()/2, canvas.H()/2-title.H()-70-40, 0.5, 0.5)
		diskcard.DrawImageAnchored(imaging.Blur(back, 8), diskcard.W()/2, canvas.H()/2-title.H()-textcard.H()-70-40-40, 0.5, 0.5)
		moreinfocard.DrawImageAnchored(imaging.Blur(back, 8), moreinfocard.W()/2, canvas.H()/2-title.H()-textcard.H()-diskcard.H()-70-40-40-40, 0.5, 0.5)
	} else {
		back = img.Size(back, int(bw*cw/bw), int(bh*cw/bw)).Im
		canvas.DrawImage(back, 0, 0)
		title.DrawImage(imaging.Blur(back, 8), -70, -70)
		textcard.DrawImage(imaging.Blur(back, 8), -70, -70-title.H()-40)
		diskcard.DrawImage(imaging.Blur(back, 8), -70, -70-title.H()-40-textcard.H()-40)
		moreinfocard.DrawImage(imaging.Blur(back, 8), -70, -70-title.H()-40-textcard.H()-40-diskcard.H()-40)
	}

	title.DrawRoundedRectangle(1, 1, float64(title.W()-1*2), float64(title.H()-1*2), 16)
	title.SetLineWidth(3)
	title.SetRGBA255(255, 255, 255, 100)
	title.StrokePreserve()
	title.SetRGBA255(255, 255, 255, 100)
	title.Fill()

	data, err = web.GetData("http://q4.qlogo.cn/g?b=qq&nk=" + strconv.FormatInt(uid, 10) + "&s=640")
	if err != nil {
		return nil, err
	}

	avatarimg, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	avatar := img.Size(avatarimg, 200, 200)

	title.DrawImage(avatar.Circle(0).Im, (title.H()-avatar.H)/2, (title.H()-avatar.H)/2)

	err = title.LoadFontFace(text.GlowSansFontFile, 72)
	if err != nil {
		return nil, err
	}
	title.SetRGBA255(30, 30, 30, 255)
	fw, _ := title.MeasureString(botname)

	title.DrawStringAnchored(botname, float64(title.H())+fw/2, float64(title.H())*0.5/2, 0.5, 0.5)

	err = title.LoadFontFace(text.GlowSansFontFile, 24)
	if err != nil {
		return nil, err
	}
	title.SetRGBA255(30, 30, 30, 180)

	title.NewSubPath()
	title.MoveTo(float64(title.H()), float64(title.H())/2)
	title.LineTo(float64(title.W()-title.H()), float64(title.H())/2)
	title.Stroke()

	brt, err := botrunningtime()
	if err != nil {
		return nil, err
	}

	fw, _ = title.MeasureString(brt)

	title.DrawStringAnchored(brt, float64(title.H())+fw/2, float64(title.H())*(0.5+0.25/2), 0.5, 0.5)

	bs, err := botstatus()
	if err != nil {
		return nil, err
	}
	fw, _ = title.MeasureString(bs)

	title.DrawStringAnchored(bs, float64(title.H())+fw/2, float64(title.H())*(0.5+0.5/2), 0.5, 0.5)

	textcard.DrawRoundedRectangle(1, 1, float64(textcard.W()-1*2), float64(textcard.H()-1*2), 16)
	textcard.SetLineWidth(3)
	textcard.SetRGBA255(255, 255, 255, 100)
	textcard.StrokePreserve()
	textcard.SetRGBA255(255, 255, 255, 140)
	textcard.Fill()

	info, err := cpuramswap()
	if err != nil {
		return nil, err
	}
	for i, v := range info {
		offset := float64(i) * ((float64(textcard.W())-200*float64(len(info)))/float64(len(info)+1) + 200)

		textcard.SetRGBA255(235, 235, 235, 255)
		textcard.DrawCircle((float64(textcard.W())-200*float64(len(info)))/float64(len(info)+1)+200/2+offset, 20+200/2, 100)
		textcard.Fill()

		switch {
		case v.present > 90:
			textcard.SetRGBA255(255, 70, 0, 255)
		case v.present > 70:
			textcard.SetRGBA255(255, 165, 0, 255)
		default:
			textcard.SetRGBA255(145, 240, 145, 255)
		}

		textcard.NewSubPath()
		textcard.MoveTo((float64(textcard.W())-200*float64(len(info)))/float64(len(info)+1)+200/2+offset, 20+200/2)
		textcard.DrawEllipticalArc((float64(textcard.W())-200*float64(len(info)))/float64(len(info)+1)+200/2+offset, 20+200/2, 100, 100, -0.5*math.Pi, -0.5*math.Pi+2*v.present*0.01*math.Pi)
		textcard.Fill()

		textcard.SetRGBA255(255, 255, 255, 255)
		textcard.DrawCircle((float64(textcard.W())-200*float64(len(info)))/float64(len(info)+1)+200/2+offset, 20+200/2, 80)
		textcard.Fill()

		err = textcard.LoadFontFace(text.GlowSansFontFile, 42)
		if err != nil {
			return nil, err
		}

		textcard.SetRGBA255(213, 213, 213, 255)
		textcard.DrawStringAnchored(strconv.FormatFloat(v.present, 'f', 0, 64)+"%", (float64(textcard.W())-200*float64(len(info)))/float64(len(info)+1)+200/2+offset, 20+200/2, 0.5, 0.5)

		textcard.SetRGBA255(30, 30, 30, 255)
		_, fw := textcard.MeasureString(v.name)
		textcard.DrawStringAnchored(v.name, (float64(textcard.W())-200*float64(len(info)))/float64(len(info)+1)+200/2+offset, 20+200+15+textcard.FontHeight()/2, 0.5, 0.5)

		err = textcard.LoadFontFace(text.GlowSansFontFile, 20)
		if err != nil {
			return nil, err
		}
		textcard.SetRGBA255(30, 30, 30, 180)

		textoffsety := textcard.FontHeight() + 10
		for k, s := range v.text {
			textcard.DrawStringAnchored(s, (float64(textcard.W())-200*float64(len(info)))/float64(len(info)+1)+200/2+offset, 20+200+15+fw+15+textcard.FontHeight()/2+float64(k)*textoffsety, 0.5, 0.5)
		}
	}

	diskcard.DrawRoundedRectangle(1, 1, float64(diskcard.W()-1*2), float64(diskcard.H()-1*2), 16)
	diskcard.SetLineWidth(3)
	diskcard.SetRGBA255(255, 255, 255, 100)
	diskcard.StrokePreserve()
	diskcard.SetRGBA255(255, 255, 255, 140)
	diskcard.Fill()

	for i, v := range disksinfo {
		offset := float64(i)*(50+20) - 20

		diskcard.SetRGBA255(192, 192, 192, 255)
		diskcard.DrawRoundedRectangle(60, 40+(float64(diskh-40*2)-50*float64(len(disksinfo)))/float64(len(disksinfo)-1)+offset, float64(diskcard.W())-60-100, 50, 12)
		diskcard.Fill()

		switch {
		case v.present > 90:
			diskcard.SetRGBA255(255, 70, 0, 255)
		case v.present > 70:
			diskcard.SetRGBA255(255, 165, 0, 255)
		default:
			diskcard.SetRGBA255(145, 240, 145, 255)
		}

		diskcard.DrawRoundedRectangle(60, 40+(float64(diskh-40*2)-50*float64(len(disksinfo)))/float64(len(disksinfo)-1)+offset, (float64(diskcard.W())-60-100)*v.present*0.01, 50, 12)
		diskcard.Fill()

		err = diskcard.LoadFontFace(text.GlowSansFontFile, 32)
		if err != nil {
			return nil, err
		}
		diskcard.SetRGBA255(30, 30, 30, 255)
		diskcard.DrawStringAnchored(v.name, 60/2, 40+(float64(diskh-40*2)-50*float64(len(disksinfo)))/float64(len(disksinfo)-1)+50/2+offset, 0.5, 0.5)
		diskcard.DrawStringAnchored(v.text[0], (float64(diskcard.W())-60)/2, 40+(float64(diskh-40*2)-50*float64(len(disksinfo)))/float64(len(disksinfo)-1)+50/2+offset, 0.5, 0.5)
		diskcard.DrawStringAnchored(strconv.FormatFloat(v.present, 'f', 0, 64)+"%", float64(diskcard.W())-100/2, 40+(float64(diskh-40*2)-50*float64(len(disksinfo)))/float64(len(disksinfo)-1)+50/2+offset, 0.5, 0.5)
	}

	moreinfocard.DrawRoundedRectangle(1, 1, float64(moreinfocard.W()-1*2), float64(moreinfocard.H()-1*2), 16)
	moreinfocard.SetLineWidth(3)
	moreinfocard.SetRGBA255(255, 255, 255, 120)
	moreinfocard.StrokePreserve()
	moreinfocard.SetRGBA255(255, 255, 255, 160)
	moreinfocard.Fill()

	for i, v := range minfo {

		err = moreinfocard.LoadFontFace(text.GlowSansFontFile, 32)
		if err != nil {
			return nil, err
		}
		offset := float64(i)*(20+moreinfocard.FontHeight()) - 20

		moreinfocard.SetRGBA255(30, 30, 30, 255)

		fw, _ = moreinfocard.MeasureString(v.name)
		fw1, _ := moreinfocard.MeasureString(v.text[0])

		moreinfocard.DrawStringAnchored(v.name, 20+fw/2, 30+(float64(minfoh-30*2)-moreinfocard.FontHeight()*float64(len(minfo)))/float64(len(minfo)-1)+moreinfocard.FontHeight()/2+offset, 0.5, 0.5)
		moreinfocard.DrawStringAnchored(v.text[0], float64(moreinfocard.W())-20-fw1/2, 30+(float64(minfoh-30*2)-moreinfocard.FontHeight()*float64(len(minfo)))/float64(len(minfo)-1)+moreinfocard.FontHeight()/2+offset, 0.5, 0.5)
	}

	fullshadow := gg.NewContext(canvas.W(), canvas.H())

	fullshadow.SetRGBA255(0, 0, 0, 100)
	fullshadow.SetLineWidth(12)
	fullshadow.DrawRoundedRectangle(70, 70, float64(title.W()), float64(title.H()), 16)
	fullshadow.Stroke()
	fullshadow.DrawRoundedRectangle(70, float64(70+title.H()+40), float64(textcard.W()), float64(textcard.H()), 16)
	fullshadow.Stroke()
	fullshadow.DrawRoundedRectangle(70, float64(70+title.H()+40+textcard.H()+40), float64(diskcard.W()), float64(diskcard.H()), 16)
	fullshadow.Stroke()
	fullshadow.DrawRoundedRectangle(70, float64(70+title.H()+40+textcard.H()+40+diskcard.H()+40), float64(moreinfocard.W()), float64(moreinfocard.H()), 16)
	fullshadow.Stroke()

	canvas.DrawImage(imaging.Blur(fullshadow.Image(), 24), 0, 0)
	canvas.DrawImage(rendercard.Fillet(title.Image(), 16), 70, 70)
	canvas.DrawImage(rendercard.Fillet(textcard.Image(), 16), 70, 70+title.H()+40)
	canvas.DrawImage(rendercard.Fillet(diskcard.Image(), 16), 70, 70+title.H()+40+textcard.H()+40)
	canvas.DrawImage(rendercard.Fillet(moreinfocard.Image(), 16), 70, 70+title.H()+40+textcard.H()+40+diskcard.H()+40)

	sendimg, cl := writer.ToBytes(canvas.Image())
	defer cl()
	return sendimg, nil
}

func botrunningtime() (string, error) {
	hostinfo, err := host.Info()
	if err != nil {
		return "", err
	}
	t := &strings.Builder{}
	t.WriteString("ZeroBot-Plugin 已运行 ")
	t.WriteString(strconv.FormatInt((time.Now().Unix()-boottime.Unix())/86400, 10))
	t.WriteString(" 天 ")
	t.WriteString(time.Unix(time.Now().Unix()-boottime.Unix(), 0).UTC().Format("15:04:05"))
	t.WriteString(" | 系统运行 ")
	t.WriteString(strconv.FormatInt(int64(hostinfo.Uptime)/86400, 10))
	t.WriteString(" 天 ")
	t.WriteString(time.Unix(int64(hostinfo.Uptime), 0).UTC().Format("15:04:05"))
	return t.String(), nil
}

func botstatus() (string, error) {
	hostinfo, err := host.Info()
	if err != nil {
		return "", err
	}
	t := &strings.Builder{}
	t.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	t.WriteString(" | ")
	t.WriteString(runtime.Version())
	t.WriteString(" | ")
	t.WriteString(cases.Title(language.English).String(hostinfo.OS))
	return t.String(), nil
}

type status struct {
	present float64
	name    string
	text    []string
}

func cpuramswap() ([]*status, error) {
	info := make([]*status, 3)
	cpupresent, err := cpu.Percent(time.Second*3, false)
	if err != nil {
		return nil, err
	}
	cpuinfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	cpucores := strconv.Itoa(int(cpuinfo[0].Cores)) + " Core"
	cputimes := "最大 " + strconv.FormatFloat(cpuinfo[0].Mhz/1000, 'f', 1, 64) + "Ghz"

	info[0] = &status{
		present: math.Round(cpupresent[0]),
		name:    "CPU",
		text:    []string{cpucores, cputimes},
	}

	raminfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	ramtotal := "总共 " + storagesize(float64(raminfo.Total))
	ramused := "已用 " + storagesize(float64(raminfo.Used))
	ramfree := "剩余 " + storagesize(float64(raminfo.Free))

	info[1] = &status{
		present: math.Round(raminfo.UsedPercent),
		name:    "RAM",
		text:    []string{ramtotal, ramused, ramfree},
	}

	swapinfo, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}

	swaptotal := "总共 " + storagesize(float64(swapinfo.Total))
	swapused := "已用 " + storagesize(float64(swapinfo.Used))
	swapfree := "剩余 " + storagesize(float64(swapinfo.Free))

	info[2] = &status{
		present: math.Round(swapinfo.UsedPercent),
		name:    "SWAP",
		text:    []string{swaptotal, swapused, swapfree},
	}

	return info, nil
}

func storagesize(num float64) string {
	if num = num / 1024; num < 1 {
		return strconv.FormatFloat(num*1024, 'f', 2, 64) + "B"
	}
	if num = num / 1024; num < 1 {
		return strconv.FormatFloat(num*1024, 'f', 2, 64) + "KB"
	}
	if num = num / 1024; num < 1 {
		return strconv.FormatFloat(num*1024, 'f', 2, 64) + "MB"
	}
	if num = num / 1024; num < 1 {
		return strconv.FormatFloat(num*1024, 'f', 2, 64) + "GB"
	}
	return strconv.FormatFloat(num, 'f', 2, 64) + "TB"
}

func disks() ([]*status, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	diskinfo := make([]*status, len(parts))
	for i, v := range parts {
		mp := v.Mountpoint
		diskusage, err := disk.Usage(mp)
		usage := ""
		present := 0.0
		if err != nil {
			usage = err.Error()
		} else {
			usage = storagesize(float64(diskusage.Used)) + " / " + storagesize(float64(diskusage.Total))
			present = math.Round(diskusage.UsedPercent)
		}
		diskinfo[i] = &status{
			present: present,
			name:    mp,
			text:    []string{usage},
		}
	}

	return diskinfo, nil
}

func moreinfo() ([]*status, error) {
	minfo := make([]*status, 0, 8)

	hostinfo, err := host.Info()
	if err != nil {
		return nil, err
	}
	minfo = append(minfo, &status{name: "OS", text: []string{hostinfo.Platform}})

	cpuinfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	minfo = append(minfo, &status{name: "CPU", text: []string{cpuinfo[0].ModelName}})

	minfo = append(minfo, &status{name: "Version", text: []string{hostinfo.PlatformVersion}})

	plugincount := 0

	m, ok := control.Lookup("aifalse")
	if ok {
		m.Manager.ForEach(func(key string, manager *ctrl.Control[*zero.Ctx]) bool {
			plugincount++
			return true
		})
	}

	minfo = append(minfo, &status{name: "Plugin", text: []string{"共 " + strconv.Itoa(plugincount) + " 个"}})

	return minfo, nil
}
