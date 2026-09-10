// Package aifalse 暂时只有服务器监控
package aifalse

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/gg/factory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/sensors"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/FloatTech/ZeroBot-Plugin/kanban/banner"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// wmiGPUInfo 运行时初始化：
// - Windows: 通过 PowerShell Get-CimInstance / Get-PnpDevice 查询所有品牌显卡名称和显存
// - Linux/macOS: 通过 lspci 查询 VGA/3D 控制器
// 变量名沿用旧名（原本打算用 WMI 但该库在 Windows Server 上不可用，改用 PowerShell）
var (
	wmiGPUInfo func() []gpuInfo
)

type gpuInfo struct {
	Name     string
	Vendor   string  // "NVIDIA" / "Intel" / "AMD" / "Unknown"
	MemTotal float64 // MiB, 0 = unknown
}

// detectVendor 从显卡名称字符串中识别厂商（全平台通用）。
func detectVendor(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "nvidia"), strings.Contains(lower, "geforce"),
		strings.Contains(lower, "quadro"), strings.Contains(lower, "tesla"),
		strings.Contains(lower, "rtx"), strings.Contains(lower, "gtx"),
		strings.Contains(lower, "titan"):
		return "NVIDIA"
	case strings.Contains(lower, "amd"), strings.Contains(lower, "radeon"),
		strings.Contains(lower, "firepro"), strings.Contains(lower, "vega"),
		strings.Contains(lower, "ryzen"):
		return "AMD"
	case strings.Contains(lower, "intel"), strings.Contains(lower, "iris"),
		strings.Contains(lower, "uhd"), strings.Contains(lower, "arc"),
		strings.Contains(lower, "hd graphics"):
		return "Intel"
	}
	return "Unknown"
}

const (
	canvasW   = 1280
	cardw     = canvasW - 70 - 70 // 两侧各 70 边距
	topPad    = 70
	titleH    = 250
	basicH    = 380
	gap       = 40
	bottomPad = 70
)

var (
	boottime   = time.Now()
	isday      bool
	lightcolor = [3][4]uint8{{255, 70, 0, 255}, {255, 165, 0, 255}, {145, 240, 145, 255}}
	darkcolor  = [3][4]uint8{{215, 50, 0, 255}, {205, 135, 0, 255}, {115, 200, 115, 255}}

	// 后台 CPU 采样器：避免每次自检都阻塞 1 秒等 cpu.Percent
	cpuMu      sync.RWMutex
	cpuOverall float64
	cpuModel   string
	cpuMhz     float64
	cpuCore    int
	cpuThread  int

	// 资源缓存：头像按 uid 缓存，字体全局加载一次，背景读取本地文件
	avatarCache sync.Map // int64 -> []byte
	fontMu      sync.Mutex
	fontBytes   []byte
	bgMu        sync.Mutex
	bgImage     image.Image

	// 带超时的 HTTP 客户端：floatbox/web.NewDefaultClient 和 http.DefaultClient 均无超时，
	// 头像下载在网络异常时会挂死导致自检迟迟不出图，这里统一设置 10s 超时。
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	// 数据目录（data/aifalse/），背景图等本地资源存放于此
	dataFolder string
)

func init() { // 插件主体
	startCPUSampler()

	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "自检, 全局限速",
		PrivateDataFolder: "aifalse",
		Help: "- 查询计算机当前活跃度: [检查身体 | 自检 | 启动自检 | 系统状态]\n" +
			"- 设置默认限速为每 m [分钟 | 秒] n 次触发",
	})
	dataFolder = engine.DataFolder()
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
			now := time.Now().Hour()
			isday = now > 7 && now < 19

			botrunstatus := ctx.CallAction("get_status", zero.Params{}).Data
			botverisoninfo := ctx.GetVersionInfo()
			sb := &strings.Builder{}
			sb.WriteString("在线(")
			sb.WriteString(botverisoninfo.Get("app_name").String())
			sb.WriteString("-")
			sb.WriteString(botverisoninfo.Get("app_version").String())
			sb.WriteString(") | 收")
			sb.WriteString(botrunstatus.Get("stat").Get("message_received").String())
			sb.WriteString(" | 发")
			sb.WriteString(botrunstatus.Get("stat").Get("message_sent").String())
			sb.WriteString(" | 群")
			sb.WriteString(strconv.Itoa(len(ctx.GetGroupList().Array())))
			sb.WriteString(" | 好友")
			sb.WriteString(strconv.Itoa(len(ctx.GetFriendList().Array())))

			img, err := drawstatus(ctx.State["manager"].(*ctrl.Control[*zero.Ctx]), ctx.Event.SelfID, zero.BotConfig.NickName[0], sb.String())
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sendimg, err := factory.ToBytes(img)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.ImageBytes(sendimg)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
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

// startCPUSampler 启动一个常驻 goroutine 周期性采样 CPU 占用，
// 这样自检时无需再阻塞 1 秒等待 cpu.Percent 返回，显著加快生成速度。
func startCPUSampler() {
	info, err := cpu.Info()
	if err == nil && len(info) > 0 {
		core, _ := cpu.Counts(false)
		thread, _ := cpu.Counts(true)
		cpuMu.Lock()
		cpuModel = strings.TrimSpace(info[0].ModelName)
		cpuMhz = info[0].Mhz
		cpuCore = core
		cpuThread = thread
		cpuMu.Unlock()
	}
	go func() {
		for {
			p, err := cpu.Percent(time.Second, false)
			if err == nil && len(p) > 0 {
				cpuMu.Lock()
				cpuOverall = p[0]
				cpuMu.Unlock()
			}
			time.Sleep(time.Second)
		}
	}()
}

func drawstatus(m *ctrl.Control[*zero.Ctx], uid int64, botname string, botrunstatus string) (sendimg image.Image, err error) {
	// 并行采集较慢的硬件数据（磁盘、nvidia-smi、温度/风扇/功耗传感器），
	// 同时主流程读取已缓存/快速的资源（字体、头像、背景、内存）。
	gather := &sync.WaitGroup{}
	gather.Add(4)
	var (
		disks     []*status
		moreinfos []*status
		gpus      []*status
		temps     []*status
		fans      []*status
		powers    []*status
		eDisk     error
		eMore     error
	)
	go func() { defer gather.Done(); disks, eDisk = diskstate() }()
	go func() { defer gather.Done(); moreinfos, eMore = moreinfo(m) }()
	go func() { defer gather.Done(); gpus = gpustate() }()
	// tempstate 不返回 error（失败时返回 nil 切片），所以不需要 eTemp
	go func() { defer gather.Done(); temps, fans, powers = tempstate() }()

	basics, err := basicstate()
	if err != nil {
		return
	}
	back := loadBackground() // 自带兜底背景图，不会失败
	avatarbuf, _ := loadAvatar(uid)
	fontbyte, err := getFont()
	if err != nil {
		return
	}

	gather.Wait()
	if eDisk != nil {
		err = eDisk
		return
	}
	if eMore != nil {
		err = eMore
		return
	}
	// GPU / 温度为可选信息：取不到就跳过对应卡片
	if len(disks) == 0 {
		disks = []*status{{name: "/", text: []string{"无可用磁盘"}}}
	}

	var avatarf *factory.Factory
	if len(avatarbuf) > 0 {
		avatar, _, derr := image.Decode(bytes.NewReader(avatarbuf))
		if derr == nil {
			avatarf = factory.Size(avatar, 200, 200)
		}
	}

	// 计算各卡片高度与纵向位置
	diskH := barCardH(len(disks))
	moreH := infoCardH(len(moreinfos))
	var gpuH, tempH, fanH, powerH int
	if len(gpus) > 0 {
		// GPU 统一用 info 卡渲染（左名称右详情），
		// 比 bar 卡更适合展示 Intel/AMD 只有基础信息的场景，
		// 也避免了无利用率数据时显示假的 0% 进度条
		gpuH = infoCardH(len(gpus))
	}
	if len(temps) > 0 {
		tempH = infoCardH(len(temps))
	}
	if len(fans) > 0 {
		fanH = infoCardH(len(fans))
	}
	if len(powers) > 0 {
		powerH = infoCardH(len(powers))
	}

	titleTop := topPad
	basicTop := titleTop + titleH + gap
	diskTop := basicTop + basicH + gap
	y := diskTop + diskH
	advance := func(h int) {
		if h > 0 {
			y += gap + h
		}
	}
	advance(gpuH)
	gpuTop := y - gpuH
	advance(tempH)
	tempTop := y - tempH
	advance(fanH)
	fanTop := y - fanH
	advance(powerH)
	powerTop := y - powerH
	advance(moreH)
	moreTop := y - moreH
	footerTop := y + gap
	totalH := footerTop + bottomPad

	canvas := gg.NewContext(canvasW, totalH)
	cw, ch := float64(canvas.W()), float64(canvas.H())
	bh, bw := float64(back.Bounds().Dy()), float64(back.Bounds().Dx())
	if bh/bw < ch/cw {
		back = factory.Size(back, int(bw*ch/bh), int(bh*ch/bh)).Image()
		canvas.DrawImageAnchored(back, canvas.W()/2, canvas.H()/2, 0.5, 0.5)
	} else {
		back = factory.Size(back, int(bw*cw/bw), int(bh*cw/bw)).Image()
		canvas.DrawImage(back, 0, 0)
	}
	blurback := imaging.Blur(canvas.Image(), 8)
	if !isday {
		canvas.SetRGBA255(0, 0, 0, 50)
		canvas.DrawRectangle(0, 0, cw, ch)
		canvas.Fill()
	}

	// 标题卡片
	titleCard := newCard(titleTop, titleH, blurback)
	if avatarf != nil {
		titleCard.DrawImage(avatarf.Circle(0).Image(), (titleH-avatarf.H())/2, (titleH-avatarf.H())/2)
	}
	if err = titleCard.ParseFontFace(fontbyte, 72); err != nil {
		return
	}
	fw, _ := titleCard.MeasureString(botname)
	titleCard.SetColor(fontcolorswitch())
	titleCard.DrawStringAnchored(botname, float64(titleH)+fw/2, float64(titleH)*0.5/2, 0.5, 0.5)
	if err = titleCard.ParseFontFace(fontbyte, 24); err != nil {
		return
	}
	titleCard.SetColor(fontcolorswitch())
	titleCard.NewSubPath()
	titleCard.MoveTo(float64(titleH), float64(titleH)/2)
	titleCard.LineTo(float64(titleCard.W()-titleH), float64(titleH)/2)
	titleCard.Stroke()
	fw, _ = titleCard.MeasureString(botrunstatus)
	titleCard.DrawStringAnchored(botrunstatus, float64(titleH)+fw/2, float64(titleH)*(0.5+0.25/2), 0.5, 0.5)
	brt, err := botruntime()
	if err != nil {
		return
	}
	fw, _ = titleCard.MeasureString(brt)
	titleCard.DrawStringAnchored(brt, float64(titleH)+fw/2, float64(titleH)*(0.5+0.5/2), 0.5, 0.5)
	bs, err := botstatus()
	if err != nil {
		return
	}
	fw, _ = titleCard.MeasureString(bs)
	titleCard.DrawStringAnchored(bs, float64(titleH)+fw/2, float64(titleH)*(0.5+0.75/2), 0.5, 0.5)

	// 基础状态卡片（CPU / RAM / SWAP 圆环）
	basicCard := newCard(basicTop, basicH, blurback)
	if err = renderBasicCard(basicCard, basics[:], fontbyte); err != nil {
		return
	}

	// 磁盘卡片
	diskCard := newCard(diskTop, diskH, blurback)
	if err = renderBarCard(diskCard, disks, fontbyte); err != nil {
		return
	}

	// GPU 状态卡片（可选）—— info 卡样式（左名称右详情）
	var gpuCard *gg.Context
	if gpuH > 0 {
		gpuCard = newCard(gpuTop, gpuH, blurback)
		if err = renderInfoCard(gpuCard, gpus, fontbyte); err != nil {
			return
		}
	}

	// 硬件温度卡片（可选）
	var tempCard *gg.Context
	if tempH > 0 {
		tempCard = newCard(tempTop, tempH, blurback)
		if err = renderInfoCard(tempCard, temps, fontbyte); err != nil {
			return
		}
	}

	// 风扇转速卡片（可选，需要 LibreHardwareMonitor）
	var fanCard *gg.Context
	if fanH > 0 {
		fanCard = newCard(fanTop, fanH, blurback)
		if err = renderInfoCard(fanCard, fans, fontbyte); err != nil {
			return
		}
	}

	// 功耗卡片（可选，需要 LibreHardwareMonitor）
	var powerCard *gg.Context
	if powerH > 0 {
		powerCard = newCard(powerTop, powerH, blurback)
		if err = renderInfoCard(powerCard, powers, fontbyte); err != nil {
			return
		}
	}

	// 详细信息卡片
	moreCard := newCard(moreTop, moreH, blurback)
	if err = renderInfoCard(moreCard, moreinfos, fontbyte); err != nil {
		return
	}

	// 卡片阴影
	shadow := gg.NewContext(canvas.W(), canvas.H())
	shadow.SetRGBA255(0, 0, 0, 100)
	shadow.SetLineWidth(12)
	shadow.DrawRoundedRectangle(float64(70), float64(titleTop), float64(cardw), float64(titleH), 16)
	shadow.Stroke()
	shadow.DrawRoundedRectangle(float64(70), float64(basicTop), float64(cardw), float64(basicH), 16)
	shadow.Stroke()
	shadow.DrawRoundedRectangle(float64(70), float64(diskTop), float64(cardw), float64(diskH), 16)
	shadow.Stroke()
	if gpuH > 0 {
		shadow.DrawRoundedRectangle(float64(70), float64(gpuTop), float64(cardw), float64(gpuH), 16)
		shadow.Stroke()
	}
	if tempH > 0 {
		shadow.DrawRoundedRectangle(float64(70), float64(tempTop), float64(cardw), float64(tempH), 16)
		shadow.Stroke()
	}
	if fanH > 0 {
		shadow.DrawRoundedRectangle(float64(70), float64(fanTop), float64(cardw), float64(fanH), 16)
		shadow.Stroke()
	}
	if powerH > 0 {
		shadow.DrawRoundedRectangle(float64(70), float64(powerTop), float64(cardw), float64(powerH), 16)
		shadow.Stroke()
	}
	shadow.DrawRoundedRectangle(float64(70), float64(moreTop), float64(cardw), float64(moreH), 16)
	shadow.Stroke()
	canvas.DrawImage(imaging.Blur(shadow.Image(), 24), 0, 0)
	canvas.DrawImage(titleCard.Image(), 70, titleTop)
	canvas.DrawImage(basicCard.Image(), 70, basicTop)
	canvas.DrawImage(diskCard.Image(), 70, diskTop)
	if gpuH > 0 {
		canvas.DrawImage(gpuCard.Image(), 70, gpuTop)
	}
	if tempH > 0 {
		canvas.DrawImage(tempCard.Image(), 70, tempTop)
	}
	if fanH > 0 {
		canvas.DrawImage(fanCard.Image(), 70, fanTop)
	}
	if powerH > 0 {
		canvas.DrawImage(powerCard.Image(), 70, powerTop)
	}
	canvas.DrawImage(moreCard.Image(), 70, moreTop)

	if err = canvas.ParseFontFace(fontbyte, 28); err != nil {
		return
	}
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.DrawStringAnchored("Created By ZeroBot-Plugin "+banner.Version, float64(canvas.W())/2+3, float64(canvas.H())-float64(bottomPad)/2+3, 0.5, 0.5)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.DrawStringAnchored("Created By ZeroBot-Plugin "+banner.Version, float64(canvas.W())/2, float64(canvas.H())-float64(bottomPad)/2, 0.5, 0.5)

	sendimg = canvas.Image()
	return
}

// newCard 创建一张带毛玻璃背景与描边的卡片画布。
func newCard(topY, h int, blurback image.Image) *gg.Context {
	c := gg.NewContext(cardw, h)
	c.DrawRoundedRectangle(1, 1, float64(c.W()-2), float64(h-2), 16)
	c.ClipPreserve()
	c.DrawImage(blurback, -70, -topY)
	c.SetColor(colorswitch(140))
	c.FillPreserve()
	c.SetLineWidth(3)
	c.SetColor(colorswitch(100))
	c.ResetClip()
	c.Stroke()
	return c
}

// renderBasicCard 渲染 CPU / RAM / SWAP 圆环卡片。
func renderBasicCard(card *gg.Context, state []*status, fontbyte []byte) error {
	bslen := len(state)
	for i, v := range state {
		offset := float64(i) * ((float64(card.W())-200*float64(bslen))/float64(bslen+1) + 200)
		cx := (float64(card.W())-200*float64(bslen))/float64(bslen+1) + 200/2 + offset
		cy := float64(20 + 200/2)

		card.SetRGBA255(57, 57, 57, 255)
		if isday {
			card.SetRGBA255(235, 235, 235, 255)
		}
		card.DrawCircle(cx, cy, 100)
		card.Fill()

		colors := darkcolor
		if isday {
			colors = lightcolor
		}
		switch {
		case v.precent > 90:
			card.SetColor(slice2color(colors[0]))
		case v.precent > 70:
			card.SetColor(slice2color(colors[1]))
		default:
			card.SetColor(slice2color(colors[2]))
		}
		card.NewSubPath()
		card.MoveTo(cx, cy)
		card.DrawEllipticalArc(cx, cy, 100, 100, -0.5*math.Pi, -0.5*math.Pi+2*v.precent*0.01*math.Pi)
		card.Fill()

		card.SetColor(colorswitch(255))
		card.DrawCircle(cx, cy, 80)
		card.Fill()

		if err := card.ParseFontFace(fontbyte, 42); err != nil {
			return err
		}
		card.SetRGBA255(213, 213, 213, 255)
		card.DrawStringAnchored(strconv.FormatFloat(v.precent, 'f', 0, 64)+"%", cx, cy, 0.5, 0.5)

		card.SetColor(fontcolorswitch())
		_, fw := card.MeasureString(v.name)
		card.DrawStringAnchored(v.name, cx, 20+200+15+card.FontHeight()/2, 0.5, 0.5)

		if err := card.ParseFontFace(fontbyte, 20); err != nil {
			return err
		}
		card.SetColor(fontcolorswitch())
		textoffsety := card.FontHeight() + 10
		for k, s := range v.text {
			card.DrawStringAnchored(s, cx, 20+200+15+fw+15+card.FontHeight()/2+float64(k)*textoffsety, 0.5, 0.5)
		}
	}
	return nil
}

// renderBarCard 渲染横向进度条卡片（磁盘 / GPU 复用）。
func renderBarCard(card *gg.Context, state []*status, fontbyte []byte) error {
	if err := card.ParseFontFace(fontbyte, 32); err != nil {
		return err
	}
	dslen := len(state)
	if dslen == 1 {
		v := state[0]
		card.SetRGBA255(57, 57, 57, 255)
		if isday {
			card.SetRGBA255(192, 192, 192, 255)
		}
		card.DrawRoundedRectangle(40, 40, float64(card.W())-40-100, 50, 12)
		card.ClipPreserve()
		card.Fill()

		colors := darkcolor
		if isday {
			colors = lightcolor
		}
		switch {
		case v.precent > 90:
			card.SetColor(slice2color(colors[0]))
		case v.precent > 70:
			card.SetColor(slice2color(colors[1]))
		default:
			card.SetColor(slice2color(colors[2]))
		}
		card.DrawRoundedRectangle(40, 40, (float64(card.W())-40-100)*v.precent*0.01, 50, 12)
		card.Fill()
		card.ResetClip()

		card.SetColor(fontcolorswitch())
		fw, _ := card.MeasureString(v.name)
		fw1, _ := card.MeasureString(v.text[0])
		card.DrawStringAnchored(v.name, 40+10+fw/2, 40+50/2, 0.5, 0.5)
		card.DrawStringAnchored(v.text[0], (float64(card.W())-100-10)-fw1/2, 40+50/2, 0.5, 0.5)
		card.DrawStringAnchored(strconv.FormatFloat(v.precent, 'f', 0, 64)+"%", float64(card.W())-100/2, 40+50/2, 0.5, 0.5)
	} else {
		for i, v := range state {
			offset := float64(i)*(50+20) - 20
			barY := 40 + (float64(card.H()-40*2)-50*float64(dslen))/float64(dslen-1) + offset

			card.SetRGBA255(57, 57, 57, 255)
			if isday {
				card.SetRGBA255(192, 192, 192, 255)
			}
			card.DrawRoundedRectangle(40, barY, float64(card.W())-40-100, 50, 12)
			card.ClipPreserve()
			card.Fill()

			colors := darkcolor
			if isday {
				colors = lightcolor
			}
			switch {
			case v.precent > 90:
				card.SetColor(slice2color(colors[0]))
			case v.precent > 70:
				card.SetColor(slice2color(colors[1]))
			default:
				card.SetColor(slice2color(colors[2]))
			}
			card.DrawRoundedRectangle(40, barY, (float64(card.W())-40-100)*v.precent*0.01, 50, 12)
			card.Fill()
			card.ResetClip()

			card.SetColor(fontcolorswitch())
			fw, _ := card.MeasureString(v.name)
			fw1, _ := card.MeasureString(v.text[0])
			card.DrawStringAnchored(v.name, 40+10+fw/2, barY+50/2, 0.5, 0.5)
			card.DrawStringAnchored(v.text[0], (float64(card.W())-100-10)-fw1/2, barY+50/2, 0.5, 0.5)
			card.DrawStringAnchored(strconv.FormatFloat(v.precent, 'f', 0, 64)+"%", float64(card.W())-100/2, barY+50/2, 0.5, 0.5)
		}
	}
	return nil
}

// renderInfoCard 渲染左侧名称 / 右侧数值的信息卡片（详细信息 / 温度复用）。
func renderInfoCard(card *gg.Context, state []*status, fontbyte []byte) error {
	if err := card.ParseFontFace(fontbyte, 32); err != nil {
		return err
	}
	milen := len(state)
	for i, v := range state {
		offset := float64(i)*(20+card.FontHeight()) - 20
		card.SetColor(fontcolorswitch())
		fw, _ := card.MeasureString(v.name)
		fw1, _ := card.MeasureString(v.text[0])
		var y float64
		if milen == 1 {
			y = float64(card.H()) / 2
		} else {
			y = 30 + (float64(card.H()-30*2)-card.FontHeight()*float64(milen))/float64(milen-1) + card.FontHeight()/2 + offset
		}
		card.DrawStringAnchored(v.name, 20+fw/2, y, 0.5, 0.5)
		card.DrawStringAnchored(v.text[0], float64(card.W())-20-fw1/2, y, 0.5, 0.5)
	}
	return nil
}

func botruntime() (string, error) {
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
	t.WriteString(" | Compiled by ")
	t.WriteString(runtime.Version())
	t.WriteString(" | ")
	t.WriteString(cases.Title(language.English).String(hostinfo.OS))
	t.WriteString(" ")
	t.WriteString(runtime.GOARCH)
	return t.String(), nil
}

type status struct {
	precent float64
	name    string
	text    []string
}

func basicstate() (stateinfo [3]*status, err error) {
	cpuMu.RLock()
	overall := cpuOverall
	mhz := cpuMhz
	core := cpuCore
	thread := cpuThread
	cpuMu.RUnlock()

	cores := strconv.Itoa(core) + "C" + strconv.Itoa(thread) + "T"
	times := "未知"
	if mhz > 0 {
		times = "最大 " + strconv.FormatFloat(mhz/1000, 'f', 1, 64) + "Ghz"
	}
	stateinfo[0] = &status{
		precent: math.Round(overall),
		name:    "CPU",
		text:    []string{cores, times},
	}

	raminfo, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	stateinfo[1] = &status{
		precent: math.Round(raminfo.UsedPercent),
		name:    "RAM",
		text: []string{
			"总共 " + storagefmt(float64(raminfo.Total)),
			"已用 " + storagefmt(float64(raminfo.Used)),
			"剩余 " + storagefmt(float64(raminfo.Free)),
		},
	}

	swapinfo, err := mem.SwapMemory()
	if err != nil {
		return
	}
	stateinfo[2] = &status{
		precent: math.Round(swapinfo.UsedPercent),
		name:    "SWAP",
		text: []string{
			"总共 " + storagefmt(float64(swapinfo.Total)),
			"已用 " + storagefmt(float64(swapinfo.Used)),
			"剩余 " + storagefmt(float64(swapinfo.Free)),
		},
	}
	return
}

func storagefmt(num float64) string {
	if num /= 1024; num < 1 {
		return strconv.FormatFloat(num*1024, 'f', 2, 64) + "B"
	}
	if num /= 1024; num < 1 {
		return strconv.FormatFloat(num*1024, 'f', 2, 64) + "KB"
	}
	if num /= 1024; num < 1 {
		return strconv.FormatFloat(num*1024, 'f', 2, 64) + "MB"
	}
	if num /= 1024; num < 1 {
		return strconv.FormatFloat(num*1024, 'f', 2, 64) + "GB"
	}
	return strconv.FormatFloat(num, 'f', 2, 64) + "TB"
}

func diskstate() (stateinfo []*status, err error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return
	}
	stateinfo = make([]*status, 0, len(parts))
	for _, v := range parts {
		mp := v.Mountpoint
		if strings.HasPrefix(mp, "/snap/") || strings.HasPrefix(mp, "/apex/") {
			continue
		}
		diskusage, err := disk.Usage(mp)
		if err != nil {
			continue
		}
		stateinfo = append(stateinfo, &status{
			precent: math.Round(diskusage.UsedPercent),
			name:    mp,
			text:    []string{storagefmt(float64(diskusage.Used)) + " / " + storagefmt(float64(diskusage.Total))},
		})
	}
	return stateinfo, nil
}

// gpustate 采集 GPU 状态，合并 nvidia-smi（NVIDIA 详细数据）和全量显卡列表（所有品牌）。
// 优先级：先拿全量列表（含 Intel/AMD/NVIDIA），再用 nvidia-smi 的详细数据覆盖匹配的 NVIDIA 卡。
// 没有可用 GPU 时返回 nil，调用方跳过 GPU 卡片。
func gpustate() []*status {
	// 1. 先拿 nvidia-smi 的 NVIDIA 详细数据（利用率/温度/功耗）
	nvidiaList, _ := nvidiaGPU()
	logrus.Debugf("[aifalse] nvidia-smi 返回 %d 块", len(nvidiaList))

	// 2. 再拿全量显卡列表（所有品牌，跨平台）
	var allGPUs []gpuInfo
	if wmiGPUInfo != nil {
		allGPUs = wmiGPUInfo()
		logrus.Debugf("[aifalse] 全量 GPU 列表 %d 个", len(allGPUs))
	}
	stateinfo := make([]*status, 0, len(nvidiaList)+len(allGPUs))

	// 3. 构建 NVIDIA 名称 → status 的映射，用于合并
	nvidiaByName := make(map[string]*status, len(nvidiaList))
	for _, s := range nvidiaList {
		nvidiaByName[s.name] = s
	}

	// 4. 按全量列表顺序输出，匹配到 NVIDIA 的用 nvidia-smi 详细数据覆盖
	usedNvidia := make(map[string]bool, len(nvidiaList))
	for _, g := range allGPUs {
		vendor := g.Vendor
		name := g.Name

		if vendor == "NVIDIA" {
			// 尝试用 nvidia-smi 匹配（名称精确匹配或模糊包含）
			var matched *status
			if s, ok := nvidiaByName[name]; ok {
				matched = s
			} else {
				// 模糊匹配：全量列表可能带 "NVIDIA Corporation" 前缀，nvidia-smi 返回纯净名
				for nName, nS := range nvidiaByName {
					if strings.Contains(strings.ToLower(name), strings.ToLower(nName)) ||
						strings.Contains(strings.ToLower(nName), strings.ToLower(name)) {
						matched = nS
						break
					}
				}
			}
			if matched != nil {
				stateinfo = append(stateinfo, matched)
				usedNvidia[matched.name] = true
				continue
			}
			// NVIDIA 但 nvidia-smi 没数据 —— 驱动未装或 WSL 环境
			stateinfo = append(stateinfo, basicGPUStatus(vendor, name, g.MemTotal, "NVIDIA 驱动未就绪，无详细监控数据"))
			continue
		}

		// Intel / AMD / Unknown —— 展示基础信息
		stateinfo = append(stateinfo, basicGPUStatus(vendor, name, g.MemTotal, ""))
	}

	// 5. 全量列表可能漏掉某些 NVIDIA 卡（PowerShell 过滤等），确保 nvidia-smi 找到的也加上
	for nName, nS := range nvidiaByName {
		if !usedNvidia[nName] {
			stateinfo = append(stateinfo, nS)
		}
	}

	logrus.Infof("[aifalse] 最终 GPU 列表 %d 个 (nvidia-smi=%d, 全量=%d)",
		len(stateinfo), len(nvidiaList), len(allGPUs))
	return stateinfo
}

// basicGPUStatus 构造一张基础 GPU 状态卡（无利用率/温度等详细指标时使用）。
func basicGPUStatus(vendor, name string, memMiB float64, extraNote string) *status {
	displayName := name
	if vendor != "Unknown" {
		displayName = vendor + " " + name
	}
	var detail string
	switch {
	case memMiB > 0 && extraNote != "":
		detail = "显存 " + strconv.FormatFloat(memMiB, 'f', 0, 64) + " MiB · " + extraNote
	case memMiB > 0:
		detail = "显存 " + strconv.FormatFloat(memMiB, 'f', 0, 64) + " MiB"
	case extraNote != "":
		detail = extraNote
	default:
		detail = "共享显存 · 温度见下方传感器"
	}
	return &status{
		precent: 0, // 无利用率数据
		name:    displayName,
		text:    []string{detail},
	}
}

// nvidiaGPU 通过 nvidia-smi 采集 NVIDIA GPU 状态。
func nvidiaGPU() ([]*status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "nvidia-smi",
		"--query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw",
		"--format=csv,noheader,nounits").Output()
	if err != nil {
		logrus.Warnln("[aifalse] nvidia-smi 执行失败:", err, "输出:", string(out))
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	stateinfo := make([]*status, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			logrus.Warnf("[aifalse] nvidia-smi 字段数不足: %d, 行: %s", len(fields), line)
			continue
		}
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		util, _ := strconv.ParseFloat(fields[1], 64)
		memUsed, _ := strconv.ParseFloat(fields[2], 64)
		memTotal, _ := strconv.ParseFloat(fields[3], 64)
		temp, _ := strconv.ParseFloat(fields[4], 64)
		power, _ := strconv.ParseFloat(fields[5], 64)
		// info 卡样式下 precent 不单独显示，把利用率放进详情文本
		detail := fmt.Sprintf("%.0f%% · %s / %s MiB · %.0f°C · %.1f W",
			util,
			strconv.FormatFloat(memUsed, 'f', 0, 64),
			strconv.FormatFloat(memTotal, 'f', 0, 64),
			temp, power)
		stateinfo = append(stateinfo, &status{
			precent: util,
			name:    fields[0],
			text:    []string{detail},
		})
	}
	logrus.Infof("[aifalse] nvidiaGPU 解析完成: %d 块 GPU", len(stateinfo))
	return stateinfo, nil
}

// tempstate 采集各硬件温度、风扇转速、CPU功耗。
// 并行尝试三种途径（都给 3s 超时）：
//  1. HWiNFO64 注册表 HKLM\SOFTWARE\HWiNFO64\VSB —— Windows 专属，最优先
//     你服务器已经有 HWiNFO 后台在跑（HWInfoSever 进程），直接读注册表就行
//  2. LibreHardwareMonitor WMI root/LibreHardwareMonitor/Sensor —— 开源免费备选
//  3. gopsutil.SensorsTemperatures() —— 跨平台通用，Windows 上只能拿到 ACPI 热区
//
// 如果 HWiNFO 或 LHM 提供了真实的 CPU 温度/风扇/功耗数据，会自动过滤掉 gopsutil 返回的
// ACPI 主板热区（那些 25-30°C 的值很容易误导）。
func tempstate() (temps []*status, fans []*status, powers []*status) {
	// Windows 上异步触发 LHM 自动配置（下载 + 启动），不阻塞当前采集
	go lhmAutoSetup()

	type gopsutilResult struct {
		ts []sensors.TemperatureStat
	}
	gopsCh := make(chan gopsutilResult, 1)
	go func() {
		ts, _ := sensors.SensorsTemperatures()
		gopsCh <- gopsutilResult{ts}
	}()

	lhmCh := make(chan []lhmSensor, 1)
	go func() {
		// WMI 和 HTTP REST API 并行跑，谁先返回数据就用谁，两个都有就合并
		wmiCh := make(chan []lhmSensor, 1)
		httpCh := make(chan []lhmSensor, 1)
		go func() { wmiCh <- queryLibreHardwareMonitor() }()
		go func() { httpCh <- queryLibreHardwareMonitorHTTP() }()
		wmi := <-wmiCh
		http := <-httpCh
		lhmCh <- mergeLHMSensors(wmi, http)
	}()

	type hwiResult struct {
		temps, fans, powers []*status
	}
	hwiCh := make(chan hwiResult, 1)
	go func() {
		t, f, p := queryHWiNFO()
		hwiCh <- hwiResult{t, f, p}
	}()

	// 并行收集，给 3s 总超时
	var gpTS []sensors.TemperatureStat
	var lhmSensors []lhmSensor
	var hwiTemps, hwiFans, hwiPowers []*status
	timeout := time.After(3 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case r := <-gopsCh:
			gpTS = r.ts
		case r := <-lhmCh:
			lhmSensors = r
		case r := <-hwiCh:
			hwiTemps, hwiFans, hwiPowers = r.temps, r.fans, r.powers
		case <-timeout:
			logrus.Debugln("[aifalse] 温度采集超时")
		}
	}
	// 兜底收剩余
	select {
	case r := <-gopsCh:
		if gpTS == nil {
			gpTS = r.ts
		}
	default:
	}
	select {
	case r := <-lhmCh:
		if lhmSensors == nil {
			lhmSensors = r
		}
	default:
	}
	select {
	case r := <-hwiCh:
		if hwiTemps == nil && hwiFans == nil && hwiPowers == nil {
			hwiTemps, hwiFans, hwiPowers = r.temps, r.fans, r.powers
		}
	default:
	}

	// === HWiNFO 优先 ===
	realSensorSource := false
	if len(hwiTemps)+len(hwiFans)+len(hwiPowers) > 0 {
		temps = append(temps, hwiTemps...)
		fans = append(fans, hwiFans...)
		powers = append(powers, hwiPowers...)
		realSensorSource = true
		logrus.Infof("[aifalse] HWiNFO: 温度 %d, 风扇 %d, 功耗 %d",
			len(hwiTemps), len(hwiFans), len(hwiPowers))
	}

	// === LibreHardwareMonitor 次优先 ===
	if len(lhmSensors) > 0 {
		logrus.Infof("[aifalse] LibreHardwareMonitor 返回 %d 个传感器", len(lhmSensors))
		for _, s := range lhmSensors {
			switch s.SensorType {
			case "Temperature":
				temps = append(temps, &status{
					name:    lhmSensorDisplayName(s),
					text:    []string{fmt.Sprintf("%.1f°C", s.Value)},
					precent: 0,
				})
				realSensorSource = true
			case "Fan":
				fans = append(fans, &status{
					name:    lhmSensorDisplayName(s),
					text:    []string{fmt.Sprintf("%.0f RPM", s.Value)},
					precent: 0,
				})
				realSensorSource = true
			case "Power":
				powers = append(powers, &status{
					name:    lhmSensorDisplayName(s),
					text:    []string{fmt.Sprintf("%.1f W", s.Value)},
					precent: 0,
				})
				realSensorSource = true
			}
		}
	}

	// === gopsutil 最后 ===
	if len(gpTS) > 0 {
		if realSensorSource {
			// 有真实传感器数据了，过滤掉 ACPI 主板热区（25-30°C 误导人）
			gpTS = filterACPIThermalZones(gpTS)
		}
		temps = append(temps, cleanGopsutilTemps(gpTS)...)
	}

	// 简化温度卡片：只保留 CPU Core 相关温度，去掉 Package/Average/Max/主板热区等
	temps = simplifyTemps(temps)

	logrus.Infof("[aifalse] 最终: 温度 %d, 风扇 %d, 功耗 %d (realSensorSource=%v)",
		len(temps), len(fans), len(powers), realSensorSource)
	return temps, fans, powers
}

// simplifyTemps 只保留 CPU Core 相关温度，过滤掉：
//   - CPU Package / CPU 封装温度（通常和 Core Max 接近，冗余）
//   - Core Average / Core Max（Core #1/2/3/4 已经覆盖）
//   - 主板温度 / ACPI 热区（没有真实传感器时的兜底，已经被上面过滤掉了）
//   - GPU 温度（GPU 卡片自己已经显示了，冗余）
//
// 保留：CPU Core #1, CPU Core #2, CPU Core #3, CPU Core #4 ...（各核心温度）
// 如果没有 Core 温度（比如老 CPU 或 Linux），则退而保留 Package 或其他 CPU 温度。
func simplifyTemps(temps []*status) []*status {
	if len(temps) <= 1 {
		return temps // 只有一个，不管它是什么都留着
	}

	// 第一遍：优先挑 "Core" 相关
	var coreOnly []*status
	var cpuFallback []*status
	for _, t := range temps {
		name := strings.ToLower(strings.TrimSpace(t.name))
		switch {
		case strings.Contains(name, "core") && !strings.Contains(name, "average") && !strings.Contains(name, "max"):
			coreOnly = append(coreOnly, t)
		case strings.Contains(name, "package"):
			cpuFallback = append(cpuFallback, t)
		case strings.Contains(name, "cpu"):
			cpuFallback = append(cpuFallback, t)
		}
	}

	if len(coreOnly) > 0 {
		return coreOnly
	}
	if len(cpuFallback) > 0 {
		return cpuFallback
	}
	return temps // 什么都没匹配上，原样返回
}

// filterACPIThermalZones 过滤掉 gopsutil 返回的 ACPI ThermalZone 条目。
// 当 HWiNFO / LHM 提供了真实的 CPU/GPU 温度时，ACPI 主板热区（通常 25-30°C）是多余的。
func filterACPIThermalZones(ts []sensors.TemperatureStat) []sensors.TemperatureStat {
	filtered := make([]sensors.TemperatureStat, 0, len(ts))
	for _, t := range ts {
		key := strings.ToLower(strings.TrimSpace(t.SensorKey))
		if strings.Contains(key, "acpi") || strings.Contains(key, "thermalzone") {
			continue // 跳过 ACPI 主板热区
		}
		filtered = append(filtered, t)
	}
	if len(filtered) < len(ts) {
		logrus.Debugf("[aifalse] 过滤掉 %d 个 ACPI 热区", len(ts)-len(filtered))
	}
	return filtered
}

// cleanGopsutilTemps 清洗 gopsutil 温度数据，ACPI ThermalZone 条目重命名。
// 清洗后若出现同名条目（如多个 ACPI 热区都叫"主板温度"），自动加编号后缀。
func cleanGopsutilTemps(ts []sensors.TemperatureStat) []*status {
	var result []*status
	// 第一步：清洗名称
	type rawEntry struct {
		displayName string
		temp        float64
	}
	cleaned := make([]rawEntry, 0, len(ts))
	seen := make(map[string]bool, len(ts))
	for _, t := range ts {
		if t.Temperature <= 0 {
			continue
		}
		key := strings.TrimSpace(t.SensorKey)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		cleaned = append(cleaned, rawEntry{
			displayName: sanitizeSensorName(key, t.Temperature),
			temp:        t.Temperature,
		})
	}

	// 第二步：处理同名冲突 → 加编号后缀
	nameCount := make(map[string]int)
	for _, e := range cleaned {
		nameCount[e.displayName]++
	}
	nameIdx := make(map[string]int)
	for _, e := range cleaned {
		if nameCount[e.displayName] > 1 {
			nameIdx[e.displayName]++
			result = append(result, &status{
				name:    fmt.Sprintf("%s #%d", e.displayName, nameIdx[e.displayName]),
				text:    []string{fmt.Sprintf("%.1f°C", e.temp)},
				precent: 0,
			})
		} else {
			result = append(result, &status{
				name:    e.displayName,
				text:    []string{fmt.Sprintf("%.1f°C", e.temp)},
				precent: 0,
			})
		}
	}
	return result
}

// sanitizeSensorName 把 gopsutil 的传感器 key（如 "ACPI\ThermalZone\TZ00_0"）改成对用户友好的名字。
func sanitizeSensorName(key string, temp float64) string {
	// ACPI ThermalZone：主板环境温度，温度通常 25-40°C
	if strings.Contains(key, "ACPI") || strings.Contains(key, "ThermalZone") ||
		strings.Contains(key, "TZ00") || strings.Contains(key, "TZ01") ||
		strings.Contains(key, "TZ02") {
		if temp < 45 {
			return "主板温度"
		}
		return "主板热区"
	}
	// hwmon 下的常见命名
	if strings.Contains(key, "coretemp") {
		return "CPU " + key
	}
	if strings.Contains(key, "k10temp") {
		return "AMD " + key
	}
	if strings.Contains(key, "nouveau") || strings.Contains(key, "nvrm") {
		return "GPU " + key
	}
	return key
}

func moreinfo(m *ctrl.Control[*zero.Ctx]) (stateinfo []*status, err error) {
	var mems runtime.MemStats
	runtime.ReadMemStats(&mems)
	fmtmem := storagefmt(float64(mems.Alloc))

	hostinfo, err := host.Info()
	if err != nil {
		return
	}
	cpuMu.RLock()
	modelName := cpuModel
	cpuMu.RUnlock()
	if modelName == "" {
		modelName = "未知"
	}

	count := len(m.Manager.M)
	stateinfo = []*status{
		{name: "OS", text: []string{hostinfo.Platform}},
		{name: "CPU", text: []string{modelName}},
		{name: "Version", text: []string{hostinfo.PlatformVersion}},
		{name: "Plugin", text: []string{"共 " + strconv.Itoa(count) + " 个"}},
		{name: "Memory", text: []string{"已用 " + fmtmem}},
	}
	return
}

// barCardH 返回横向进度条卡片的高度（磁盘 / GPU）。
func barCardH(n int) int {
	if n < 1 {
		n = 1
	}
	return 60 + 70*n
}

// infoCardH 返回信息行卡片的高度（详细信息 / 温度）。
func infoCardH(n int) int {
	if n < 1 {
		n = 1
	}
	return 40 + 44*n
}

// loadBackground 从本地 data/aifalse/ 目录随机挑选一张背景图。
// 不再依赖网络下载，避免因图片服务不可用导致自检迟迟不出图。
// 若目录下没有可用图片，返回兜底纯色背景。
func loadBackground() image.Image {
	bgMu.Lock()
	defer bgMu.Unlock()

	imgs := listBackgroundImages()
	if len(imgs) == 0 {
		logrus.Warnln("[aifalse] data/aifalse/ 下没有找到背景图，使用兜底背景")
		if bgImage != nil {
			return bgImage
		}
		return fallbackBackground()
	}
	// 随机选一张
	path := imgs[rand.Intn(len(imgs))]
	data, err := os.ReadFile(path)
	if err != nil {
		logrus.Warnln("[aifalse] 读取背景图失败:", err)
		if bgImage != nil {
			return bgImage
		}
		return fallbackBackground()
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		logrus.Warnln("[aifalse] 解码背景图失败:", path, err)
		if bgImage != nil {
			return bgImage
		}
		return fallbackBackground()
	}
	bgImage = img
	return img
}

// listBackgroundImages 列出 data/aifalse/ 下所有支持的图片文件。
func listBackgroundImages() []string {
	entries, err := os.ReadDir(dataFolder)
	if err != nil {
		return nil
	}
	exts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
	}
	var imgs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if exts[strings.ToLower(filepath.Ext(e.Name()))] {
			imgs = append(imgs, filepath.Join(dataFolder, e.Name()))
		}
	}
	return imgs
}

// fallbackBackground 生成一张兜底的深色背景图。
func fallbackBackground() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, 1280, 720))
	for y := 0; y < 720; y++ {
		for x := 0; x < 1280; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 25, G: 25, B: 35, A: 255})
		}
	}
	return img
}

// loadAvatar 按 uid 缓存 QQ 头像。下载失败时返回空（调用方跳过头像绘制）。
func loadAvatar(uid int64) ([]byte, error) {
	if v, ok := avatarCache.Load(uid); ok {
		if b, ok := v.([]byte); ok && len(b) > 0 {
			return b, nil
		}
	}
	url := "https://q4.qlogo.cn/g?b=qq&nk=" + strconv.FormatInt(uid, 10) + "&s=640"
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("头像状态码: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	avatarCache.Store(uid, data)
	return data, nil
}

// getFont 全局只加载一次字体。
func getFont() ([]byte, error) {
	fontMu.Lock()
	defer fontMu.Unlock()
	if len(fontBytes) > 0 {
		return fontBytes, nil
	}
	b, err := file.GetLazyData(text.GlowSansFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	fontBytes = b
	return b, nil
}

func colorswitch(a uint8) color.Color {
	if isday {
		return color.NRGBA{255, 255, 255, a}
	}
	return color.NRGBA{0, 0, 0, a}
}

func fontcolorswitch() color.Color {
	if isday {
		return color.NRGBA{30, 30, 30, 255}
	}
	return color.NRGBA{235, 235, 235, 255}
}

func slice2color(c [4]uint8) color.Color {
	return color.NRGBA{c[0], c[1], c[2], c[3]}
}
