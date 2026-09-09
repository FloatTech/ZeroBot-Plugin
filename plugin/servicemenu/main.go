// Package servicemenu 服务菜单：自检图同款毛玻璃 UI + 主题切换 + 优雅重载
package servicemenu

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"

	// 注册jpg/gif解码器，背景图解码不依赖其他插件是否加载
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/kanban/banner"
	"github.com/FloatTech/gg"
	zbpctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// theme 主题配色（iOS Liquid Glass：凸透镜放大 + 重磨砂 + 边缘折射 + 色散 + 奶白洗白）
type theme struct {
	Name           string
	DisplayName    string
	BackgroundFrom [3]uint8
	BackgroundTo   [3]uint8
	OverlayColor   color.RGBA // 整图覆盖色（暗主题黑压暗、亮主题白提亮）
	OverlayAlpha   uint8      // 整图覆盖层 alpha (0-255)
	Blur           float64    // 背景模糊半径（磨砂强度，目标风格 26-34）
	Magnify        float64    // 凸透镜放大倍率（1.2-1.4，核心特征）
	Refraction     float64    // 边缘折射强度（像素）
	Dispersion     float64    // 色散强度（0-1），RGB 通道分离
	RimAlpha       uint8      // 顶部轮廓光 alpha
	BottomDark     uint8      // 底部内阴影 alpha
	Tint           color.RGBA // 玻璃色调（奶白）
	TintAlpha      uint8      // 奶白洗白强度（55-76）
	Saturate       float64    // 饱和度提升（1.15-1.3）
	TopLight       uint8      // 顶部高光带 alpha
	StrokeAlpha    uint8      // 边缘细描边 alpha
	Brightness     float64    // 玻璃亮度
	TextMain       color.RGBA
	TextSec        color.RGBA
}

var themes = []theme{
	// 1. 暗夜玻璃 - 冷紫调（苹果风格：中心通透 + 边缘折射）
	{
		Name: "glass-dark", DisplayName: "暗夜玻璃",
		BackgroundFrom: [3]uint8{55, 48, 72},
		BackgroundTo:   [3]uint8{72, 60, 95},
		OverlayColor:   color.RGBA{R: 0, G: 0, B: 0, A: 255},
		OverlayAlpha:   55,
		Blur:           10, Magnify: 1.22, Refraction: 18, Dispersion: 0.12,
		RimAlpha: 140, BottomDark: 65,
		Tint: color.RGBA{R: 250, G: 248, B: 255, A: 255}, TintAlpha: 38,
		Saturate: 1.3, TopLight: 55, StrokeAlpha: 60, Brightness: 1.1,
		TextMain: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextSec:  color.RGBA{R: 255, G: 255, B: 255, A: 215},
	},
	// 2. 柔光玻璃 - 纯白奶感（目标风格）
	{
		Name: "glass-light", DisplayName: "柔光玻璃",
		BackgroundFrom: [3]uint8{232, 228, 242},
		BackgroundTo:   [3]uint8{218, 222, 238},
		OverlayColor:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
		OverlayAlpha:   25,
		Blur:           10, Magnify: 1.22, Refraction: 18, Dispersion: 0.12,
		RimAlpha: 140, BottomDark: 65,
		Tint: color.RGBA{R: 255, G: 255, B: 255, A: 255}, TintAlpha: 46,
		Saturate: 1.25, TopLight: 60, StrokeAlpha: 60, Brightness: 1.06,
		TextMain: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextSec:  color.RGBA{R: 255, G: 255, B: 255, A: 215},
	},
	// 3. 霓虹赛博 - 青调奶白
	{
		Name: "neon", DisplayName: "霓虹赛博",
		BackgroundFrom: [3]uint8{20, 8, 35},
		BackgroundTo:   [3]uint8{42, 14, 58},
		OverlayColor:   color.RGBA{R: 0, G: 0, B: 0, A: 255},
		OverlayAlpha:   75,
		Blur:           12, Magnify: 1.26, Refraction: 22, Dispersion: 0.15,
		RimAlpha: 145, BottomDark: 70,
		Tint: color.RGBA{R: 244, G: 255, B: 252, A: 255}, TintAlpha: 44,
		Saturate: 1.38, TopLight: 65, StrokeAlpha: 70, Brightness: 1.12,
		TextMain: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextSec:  color.RGBA{R: 240, G: 255, B: 250, A: 215},
	},
	// 4. 极简暖灰 - 暖米奶白
	{
		Name: "minimal", DisplayName: "极简暖灰",
		BackgroundFrom: [3]uint8{238, 236, 232},
		BackgroundTo:   [3]uint8{228, 226, 220},
		OverlayColor:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
		OverlayAlpha:   35,
		Blur:           9, Magnify: 1.2, Refraction: 16, Dispersion: 0.1,
		RimAlpha: 135, BottomDark: 60,
		Tint: color.RGBA{R: 255, G: 252, B: 246, A: 255}, TintAlpha: 42,
		Saturate: 1.2, TopLight: 55, StrokeAlpha: 55, Brightness: 1.05,
		TextMain: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextSec:  color.RGBA{R: 255, G: 253, B: 248, A: 215},
	},
}

var currentTheme = themes[0]

// ===== 随机背景图（复用自检图 data/aifalse/ 目录）=====
const bgDataDir = "data/aifalse"

var (
	bgListOnce sync.Once
	bgFiles    []string
)

func listBgFiles() []string {
	bgListOnce.Do(func() {
		entries, err := os.ReadDir(bgDataDir)
		if err != nil {
			return
		}
		exts := map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true,
			".gif": true, ".webp": true, ".bmp": true,
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if exts[strings.ToLower(filepath.Ext(e.Name()))] {
				bgFiles = append(bgFiles, filepath.Join(bgDataDir, e.Name()))
			}
		}
	})
	return bgFiles
}

func loadRandomBg() image.Image {
	files := listBgFiles()
	if len(files) == 0 {
		return nil
	}
	path := files[rand.Intn(len(files))]
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		logrus.Warn("[servicemenu] 解码背景图失败:", path, err)
		return nil
	}
	return img
}

func SetTheme(name string) bool {
	for _, t := range themes {
		if strings.EqualFold(t.Name, name) {
			currentTheme = t
			return true
		}
	}
	return false
}

func getPlugins() []*zbpctrl.Control[*zero.Ctx] {
	var plugins []*zbpctrl.Control[*zero.Ctx]
	control.ForEachByPrio(func(i int, m *zbpctrl.Control[*zero.Ctx]) bool {
		plugins = append(plugins, m)
		return true
	})
	return plugins
}

func init() {
	// ===== 命令注册 =====

	zero.OnCommandGroup([]string{"服务列表", "service_list"}, zero.OnlyToMe).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			page := 1
			// 从命令参数里取页码
			if raw := ctx.State["args"]; raw != nil {
				if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
					page = atoi(strings.TrimSpace(s))
				}
			}
			img, err := renderServiceList(gid, page)
			if err != nil {
				ctx.SendChain(message.Text("渲染失败: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(img))
		})

	zero.OnRegex(`^服务列表\s*(\d*)\s*$`).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			page := 1
			if m := ctx.State["regex_matched"].([]string); len(m) > 1 && m[1] != "" {
				page = atoi(m[1])
			}
			img, err := renderServiceList(gid, page)
			if err != nil {
				ctx.SendChain(message.Text("渲染失败: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(img))
		})

	zero.OnRegex(`^(菜单用法|menuusage)\s+(\S+)$`).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			name := strings.ToLower(ctx.State["regex_matched"].([]string)[2])
			m, ok := control.Lookup(name)
			if !ok {
				ctx.SendChain(message.Text("没有找到插件: ", name))
				return
			}
			img, err := renderUsageCard(m)
			if err != nil {
				ctx.SendChain(message.Text("渲染失败: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(img))
		})

	zero.OnRegex(`^主题\s+(\S+)$`).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			name := strings.ToLower(ctx.State["regex_matched"].([]string)[1])
			if SetTheme(name) {
				ctx.SendChain(message.Text("主题已切换为: ", currentTheme.DisplayName))
			} else {
				var names []string
				for _, t := range themes {
					names = append(names, t.Name+"("+t.DisplayName+")")
				}
				ctx.SendChain(message.Text("没有这个主题。可用: ", strings.Join(names, ", ")))
			}
		})

	zero.OnFullMatchGroup([]string{"主题列表", "listtheme"}).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			var sb strings.Builder
			sb.WriteString("可用主题:\n")
			for i, t := range themes {
				marker := " "
				if t.Name == currentTheme.Name {
					marker = "★"
				}
				sb.WriteString(marker + " " + itoa(i+1) + ". " + t.Name + " - " + t.DisplayName + "\n")
			}
			sb.WriteString("\n用法: menu <编号> 切换  |  主题 <name> 切换")
			ctx.SendChain(message.Text(sb.String()))
		})

	zero.OnRegex(`^menu\s+(\d+)$`).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			n := atoi(ctx.State["regex_matched"].([]string)[1])
			if n < 1 || n > len(themes) {
				var sb strings.Builder
				sb.WriteString("主题编号无效。可用:\n")
				for i, t := range themes {
					sb.WriteString(itoa(i+1) + ". " + t.DisplayName)
					if t.Name == currentTheme.Name {
						sb.WriteString(" ★")
					}
					sb.WriteString("\n")
				}
				ctx.SendChain(message.Text(sb.String()))
				return
			}
			currentTheme = themes[n-1]
			ctx.SendChain(message.Text("🎨 UI 已切换为 [", itoa(n), "] ", currentTheme.DisplayName))
		})

	zero.OnFullMatchGroup([]string{"重载全部", "reloadall", "重载bot"}, zero.SuperUserPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("收到重载命令，正在重启进程..."))
			go gracefulRestart()
		})

	zero.OnRegex(`^重载\s+(\S+)$`, zero.SuperUserPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			name := strings.ToLower(ctx.State["regex_matched"].([]string)[1])
			_, ok := control.Lookup(name)
			if !ok {
				ctx.SendChain(message.Text("没找到插件: ", name))
				return
			}
			ctx.SendChain(message.Text("注意: Go 编译后的插件无法单独热重载，将重启整个进程..."))
			go gracefulRestart()
		})
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func gracefulRestart() {
	logrus.Infoln("[servicemenu] 准备重启进程...")
	time.Sleep(1 * time.Second)
	exePath, err := os.Executable()
	if err != nil {
		logrus.Errorf("[servicemenu] 获取 exe 路径失败: %v", err)
		os.Exit(0)
		return
	}
	if runtime.GOOS == "windows" {
		dir := exePath[:strings.LastIndex(exePath, `\`)]
		cmd := exec.Command("cmd", "/c", "start", "/min", exePath)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			logrus.Errorf("[servicemenu] 启动新进程失败: %v", err)
		}
	} else {
		cmd := exec.Command(exePath, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		go cmd.Run()
	}
	os.Exit(0)
}

// ================ 图片渲染（和自检图同款模式）================

const (
	cardPadding = 70
	cardMarginY = 20
	colGap      = 32
	itemH       = 120
	headerH     = 200
	footerH     = 60
	cardRadius  = 18
	itemsPerRow = 2 // 两列：卡片更宽，字号可整体放大，无需放大图片即可看清
	pageCount   = 2 // 固定分成 2 页
)

// toRGBA 把任意 image.Image 转成 *image.RGBA（imaging.Blur 返回 NRGBA，SubImage 需要 RGBA）
func toRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}

// renderServiceList 渲染服务列表（固定分 2 页，page 从 1 开始）
func renderServiceList(gid int64, page int) ([]byte, error) {
	t := currentTheme
	allPlugins := getPlugins()
	total := len(allPlugins)
	if total == 0 {
		total = 1
		allPlugins = append(allPlugins, nil) // 占位，防止除 0
	}

	perPage := (total + pageCount - 1) / pageCount // 向上取整
	if page < 1 {
		page = 1
	}
	if page > pageCount {
		page = pageCount
	}

	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}
	pagePlugins := allPlugins[start:end]

	canvasW := 1200
	rows := int(math.Ceil(float64(len(pagePlugins)) / float64(itemsPerRow)))
	if rows < 1 {
		rows = 1
	}
	canvasH := cardPadding + headerH + rows*(itemH+cardMarginY) + footerH + cardPadding + 40

	c := gg.NewContext(canvasW, canvasH)
	bg := buildBackground(canvasW, canvasH, t)
	c.DrawImage(bg, 0, 0)

	// 整图统一覆盖层（暗主题压暗、亮主题提亮）
	if t.OverlayAlpha > 0 {
		oc := t.OverlayColor
		c.SetRGBA255(int(oc.R), int(oc.G), int(oc.B), int(t.OverlayAlpha))
		c.DrawRectangle(0, 0, float64(canvasW), float64(canvasH))
		c.Fill()
	}

	// blurback 必须对"已画完 Overlay 的完整 canvas"做 blur
	// 这样卡片内外色调 100% 一致，毛玻璃才真实（和自检图 aifalse 同做法）
	blurback := toRGBA(imaging.Blur(c.Image(), t.Blur))

	// 标题卡
	headerX, headerY := cardPadding, cardPadding
	headerW := canvasW - cardPadding*2
	drawNewCard(c, headerX, headerY, headerW, headerH, blurback, t)

	c.SetColor(t.TextMain)
	c.LoadFontFace("data/Font/GlowSansSC-Normal-ExtraBold.ttf", 48)
	c.DrawString("ZeroBot-Plugin", float64(headerX+32), float64(headerY+60))

	const secFont = "data/Font/regular-bold.ttf"
	drawTextOutlined(c, "OneBot + ZeroBot + Golang", secFont, 20, float64(headerX+32), float64(headerY+92), t.TextSec)
	drawTextOutlined(c, banner.Version+" · FloatTech", secFont, 18, float64(headerX+32), float64(headerY+120), t.TextSec)

	c.SetColor(t.TextMain)
	c.LoadFontFace("data/Font/GlowSansSC-Normal-ExtraBold.ttf", 40)
	rightText := "服务列表"
	fw, _ := c.MeasureString(rightText)
	c.DrawString(rightText, float64(canvasW-cardPadding-32)-fw, float64(headerY+60))

	subRight := "Server List"
	c.LoadFontFace(secFont, 24)
	fw2, _ := c.MeasureString(subRight)
	drawTextOutlined(c, subRight, secFont, 24, float64(canvasW-cardPadding-32)-fw2, float64(headerY+96), t.TextSec)

	// 分隔线用矩形填充（gg 的 Stroke 会写入垃圾像素，已弃用）
	c.SetRGBA255(128, 128, 128, 60)
	c.DrawRectangle(float64(headerX+32), float64(headerY+headerH-41), float64(canvasW-cardPadding-32-(headerX+32)), 2)
	c.Fill()

	infoText := "总插件 " + itoa(total) + " · 第 " + itoa(page) + "/" + itoa(pageCount) + " 页 · 主题 " + t.DisplayName
	drawTextOutlined(c, infoText, secFont, 16, float64(headerX+32), float64(headerY+headerH-18), t.TextSec)

	// 插件卡片（当前页）
	cardAreaTop := headerY + headerH + cardMarginY
	for i, m := range pagePlugins {
		if m == nil {
			continue
		}
		col := i % itemsPerRow
		row := i / itemsPerRow
		cardW := (canvasW - cardPadding*2 - colGap*(itemsPerRow-1)) / itemsPerRow
		x := float64(cardPadding) + float64(col)*float64(cardW+colGap)
		cardTop := float64(cardAreaTop) + float64(row)*(itemH+cardMarginY)
		enabled := m.IsEnabledIn(gid)
		drawNewCard(c, int(x), int(cardTop), cardW, itemH, blurback, t)
		drawPluginCardContent(c, int(x), int(cardTop), cardW, itemH, m.Service, m.Options.Brief, enabled, t)
	}

	// Footer
	footerY := canvasH - cardPadding + 10
	drawTextOutlined(c, "服务列表 [1/2]  |  menu <编号>  |  主题 <名>  |  重载全部  |  用法 <英文名>", secFont, 16, float64(cardPadding), float64(footerY), t.TextSec)

	return encodePNG(c.Image())
}

// ============ liquid-glass 核心算法移植（from huangj17/liquid-glass-react & shuding/liquid-glass）========

// smoothStep 平滑插值，对应 TS 版 smoothStep(a, b, t)
// 注意：必须用 clamp 实现（原 TS 版支持 a > b 的反向区间，
// 例如 smoothStep(1.5, -0.5, d) 表示 d<-0.5 时为 1、d>1.5 时为 0。
// 若用 "t<a return 0" 的分支写法，反向调用时蒙版会完全反转——
// 卡片内部全透明、只有四角漏出一点错位内容，产生颗粒磨损感）
func smoothStep(a, b, t float64) float64 {
	x := (t - a) / (b - a)
	if x < 0 {
		x = 0
	}
	if x > 1 {
		x = 1
	}
	return x * x * (3 - 2*x)
}

// clampF 浮点 clamp
func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// sampleBilinear 双线性插值采样（消除位移采样的颗粒感）
func sampleBilinear(bg *image.RGBA, rect image.Rectangle, sx, sy float64) color.RGBA {
	gx := sx + float64(rect.Min.X)
	gy := sy + float64(rect.Min.Y)
	x0 := int(math.Floor(gx))
	y0 := int(math.Floor(gy))
	tx := gx - float64(x0)
	ty := gy - float64(y0)

	x1, y1 := x0+1, y0+1
	W := bg.Bounds().Dx()
	H := bg.Bounds().Dy()
	clamp := func(v, lo, hi int) int {
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	}
	x0 = clamp(x0, 0, W-1)
	x1 = clamp(x1, 0, W-1)
	y0 = clamp(y0, 0, H-1)
	y1 = clamp(y1, 0, H-1)

	c00 := bg.At(x0, y0).(color.RGBA)
	c10 := bg.At(x1, y0).(color.RGBA)
	c01 := bg.At(x0, y1).(color.RGBA)
	c11 := bg.At(x1, y1).(color.RGBA)

	lerp := func(a, b, tt float64) float64 { return a + (b-a)*tt }
	rr := lerp(lerp(float64(c00.R), float64(c10.R), tx), lerp(float64(c01.R), float64(c11.R), tx), ty)
	g := lerp(lerp(float64(c00.G), float64(c10.G), tx), lerp(float64(c01.G), float64(c11.G), tx), ty)
	bb := lerp(lerp(float64(c00.B), float64(c10.B), tx), lerp(float64(c01.B), float64(c11.B), tx), ty)
	return color.RGBA{R: uint8(rr + 0.5), G: uint8(g + 0.5), B: uint8(bb + 0.5), A: 255}
}

// renderLiquidGlass 按 iOS Liquid Glass 原理渲染卡片区域（from forum.cocos.org/t/topic/171941）：
//
//	凸透镜放大(核心) → 边缘折射 → RGB 色散 → 轮廓光 → 亮度/饱和 → 奶白洗白
func renderLiquidGlass(blurback *image.RGBA, x, y, w, h int, t theme) *image.RGBA {
	rect := image.Rect(x, y, x+w, y+h)
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	bw := float64(w)
	bh := float64(h)
	cx := bw / 2
	cy := bh / 2
	rad := float64(cardRadius)
	// 边缘折射带宽度：随卡片高度缩放（苹果液态玻璃的折射集中在外环 15%-30%）
	edgeW := clampF(bh*0.30, 14, 36)

	// 圆角矩形 SDF（内部为负、边缘为 0、外部为正）
	sdf := func(fx, fy float64) float64 {
		qx := math.Abs(fx-cx) - (bw/2 - rad)
		qy := math.Abs(fy-cy) - (bh/2 - rad)
		return math.Min(math.Max(qx, qy), 0) + math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) - rad
	}

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			fx := float64(px) + 0.5
			fy := float64(py) + 0.5
			dEdge := sdf(fx, fy)

			// 圆角矩形 alpha：以名义边缘为中心的 1.6px 对称渐隐（经典 SDF AA）。
			// 过窄（1px 偏外）会锯齿，过宽（3.5px）半透明边缘带压在阴影上呈暗环毛边
			alphaF := smoothStep(0.8, -0.8, dEdge)
			if alphaF <= 0 {
				out.Set(px, py, color.RGBA{0, 0, 0, 0})
				continue
			}

			// SDF 外法线（数值梯度）：折射方向垂直于最近边缘，随形状/尺寸自适应
			nx := sdf(fx+1, fy) - sdf(fx-1, fy)
			ny := sdf(fx, fy+1) - sdf(fx, fy-1)
			if nl := math.Hypot(nx, ny); nl > 1e-6 {
				nx /= nl
				ny /= nl
			} else {
				nx, ny = 0, 0
			}

			// === 1. 边缘折射：越靠近边缘越向外取样（外围背景被"压"进边缘环）
			// 最外 5px 收敛到 0：边界处与卡外背景无错位，避免重影拖痕接缝 ===
			mag := t.Refraction * smoothStep(-edgeW*1.5, -edgeW*0.2, dEdge) * smoothStep(2.0, -3.0, dEdge)

			// === 2. 凸透镜放大（核心）：采样区缩小 = 背景被放大 ===
			sx := (fx+nx*mag-cx)/t.Magnify + cx
			sy := (fy+ny*mag-cy)/t.Magnify + cy

			// === 3. RGB 色散：R 前移，B 后移 ===
			ab := t.Dispersion
			r := float64(sampleBilinear(blurback, rect, sx+nx*mag*ab, sy+ny*mag*ab).R)
			g := float64(sampleBilinear(blurback, rect, sx, sy).G)
			b := float64(sampleBilinear(blurback, rect, sx-nx*mag*ab, sy-ny*mag*ab).B)

			// === 4. 轮廓光：SDF 细边缘带（~5px），顶部最亮、侧边次之（环境光反射）===
			band := smoothStep(-7, -1.5, dEdge)
			wr := float64(t.RimAlpha) / 255.0
			spec := band * (0.35 + 0.65*clampF(-ny, 0, 1)) * wr
			r = r*(1-spec) + 255*spec
			g = g*(1-spec) + 255*spec
			b = b*(1-spec) + 255*spec

			// === 5. 底部内阴影（玻璃厚度）===
			bd := band * clampF(ny, 0, 1) * float64(t.BottomDark) / 255.0
			r *= 1 - bd
			g *= 1 - bd
			b *= 1 - bd

			// === 6. 亮度 + 饱和度 ===
			r *= t.Brightness
			g *= t.Brightness
			b *= t.Brightness
			gray := 0.2126*r + 0.7152*g + 0.0722*b
			r = gray + (r-gray)*t.Saturate
			g = gray + (g-gray)*t.Saturate
			b = gray + (b-gray)*t.Saturate

			// === 7. 奶白洗白（液态玻璃 milk 质感）===
			ta := float64(t.TintAlpha) / 255.0
			r = r*(1-ta) + float64(t.Tint.R)*ta
			g = g*(1-ta) + float64(t.Tint.G)*ta
			b = b*(1-ta) + float64(t.Tint.B)*ta

			// === 8. 顶部高光带（SDF 实现，替代 gg Clip+LinearGradient——
			// 该库的渐变填充在圆弧段会写入黑色垃圾像素）===
			hh := bh * 0.4
			if fy < hh {
				ttH := fy / hh
				ha := float64(t.TopLight)
				if ttH < 0.5 {
					ha = ha + (ha/3-ha)*(ttH*2)
				} else {
					ha = ha / 3 * (1 - (ttH-0.5)*2)
				}
				inside := smoothStep(0, -2.5, dEdge) // 贴圆角轮廓，向内 2.5px 过渡
				a := ha / 255.0 * inside
				r = r*(1-a) + 255*a
				g = g*(1-a) + 255*a
				b = b*(1-a) + 255*a
			}

			// === 9. 边缘细描边（SDF 带，替代 gg Stroke——该库的矢量描边
			// 在圆弧段会忽略样式写入黑色垃圾像素，即角部黑刺的元凶）===
			sa := float64(t.StrokeAlpha) / 255.0 *
				smoothStep(-1.8, -1.1, dEdge) * smoothStep(0.5, -0.3, dEdge)
			if sa > 0 {
				r = r*(1-sa) + 255*sa
				g = g*(1-sa) + 255*sa
				b = b*(1-sa) + 255*sa
			}

			// 输出必须预乘 alpha（image.RGBA 是预乘格式，否则边缘渐隐像素会被二次相乘变暗）
			aF := 255 * alphaF
			out.Set(px, py, color.RGBA{
				R: uint8(clampF(r*alphaF, 0, 255) + 0.5),
				G: uint8(clampF(g*alphaF, 0, 255) + 0.5),
				B: uint8(clampF(b*alphaF, 0, 255) + 0.5),
				A: uint8(aF + 0.5),
			})
		}
	}
	return out
}

// shadowCache 阴影图缓存（按尺寸），避免每张卡片重复模糊
var (
	shadowCacheMu sync.Mutex
	shadowCache   = map[image.Point]*image.RGBA{}
)

// getBlurredShadow 生成圆角矩形高斯模糊阴影图（含 pad 边距）
// 阴影矩形相对卡片左右内缩 3px、整体只向下偏移：避免模糊 halo 贴着卡片
// 左右/圆角形成深色环带（浅色背景上呈黑刺/磨损感）
func getBlurredShadow(w, h, radius int) (img *image.RGBA, pad int) {
	pad = 28
	key := image.Point{w, h}
	shadowCacheMu.Lock()
	cached, ok := shadowCache[key]
	shadowCacheMu.Unlock()
	if ok {
		return cached, pad
	}

	sc := gg.NewContext(w+pad*2, h+pad*2)
	sc.SetColor(color.RGBA{R: 15, G: 15, B: 35, A: 50})
	sc.DrawRoundedRectangle(float64(pad)+3, float64(pad), float64(w)-6, float64(h), float64(radius))
	sc.Fill()
	blurred := toRGBA(imaging.Blur(sc.Image(), 12))

	shadowCacheMu.Lock()
	shadowCache[key] = blurred
	shadowCacheMu.Unlock()
	return blurred, pad
}

// drawNewCard 液态玻璃卡片（iOS Liquid Glass）
// 分层（从下到上）：
//  1. 高斯模糊阴影（干净无角部残影）
//  2. 液态玻璃底（凸透镜放大 + 边缘折射 + 色散 + 轮廓光 + 顶部高光 + 细描边 + 奶白洗白）
//
// 注意：不要用 gg 的 Stroke/Clip+Gradient 画描边和高光——FloatTech/gg 的矢量
// 描边与渐变填充在圆角弧段会忽略样式写入黑色垃圾像素（角部黑刺/黑弧的元凶），
// 两者均已改为 renderLiquidGlass 内部的 SDF 距离场实现。
func drawNewCard(c *gg.Context, x, y, w, h int, blurback *image.RGBA, t theme) {
	// === 1. 高斯模糊阴影（内缩+下移落影，无左右 halo）===
	shadow, pad := getBlurredShadow(w, h, cardRadius)
	c.DrawImage(shadow, x-pad, y-pad+8)

	// === 2. 液态玻璃底（含高光带与描边）===
	glass := renderLiquidGlass(blurback, x, y, w, h, t)
	c.DrawImage(glass, x, y)
}

// ellipsizeByWidth 按像素宽度截断文本（rune 安全），超宽以 "..." 结尾。
// 旧的按字节 len(brief)>24 截断有两个问题：中文每字 3 字节导致约 8 个汉字就
// 被省略（远没占满卡片可用宽度），且 brief[:23] 可能切断 UTF-8 多字节字符
// 产生乱码。现改为实测渲染宽度，二分找最长前缀。
func ellipsizeByWidth(c *gg.Context, text, fontPath string, size, maxW float64) string {
	c.LoadFontFace(fontPath, size)
	if tw, _ := c.MeasureString(text); tw <= maxW {
		return text
	}
	const ell = "..."
	runes := []rune(text)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if w, _ := c.MeasureString(string(runes[:mid]) + ell); w <= maxW {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return string(runes[:lo]) + ell
}

// drawTextOutlined 带黑色描边绘制文字。FloatTech/gg 的 Stroke 在曲线
// 字形上会写入垃圾像素（圆角黑弧同源 bug），因此用多向偏移暗色底绘
// 模拟描边再绘主体：1px 实描边 + 1.8px 淡晕两层，保证白字在奶白玻璃上远看清晰。
func drawTextOutlined(c *gg.Context, text, fontPath string, size float64, x, y float64, fill color.RGBA) {
	c.LoadFontFace(fontPath, size)
	c.SetRGBA255(25, 28, 40, 170)
	for _, d := range [...][2]float64{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
		{-0.71, -0.71}, {0.71, -0.71}, {-0.71, 0.71}, {0.71, 0.71},
	} {
		c.DrawString(text, x+d[0], y+d[1])
	}
	c.SetRGBA255(25, 28, 40, 70)
	for _, d := range [...][2]float64{
		{-1.8, 0}, {1.8, 0}, {0, -1.8}, {0, 1.8},
		{-1.3, -1.3}, {1.3, -1.3}, {-1.3, 1.3}, {1.3, 1.3},
	} {
		c.DrawString(text, x+d[0], y+d[1])
	}
	c.SetColor(fill)
	c.DrawString(text, x, y)
}

func drawPluginCardContent(c *gg.Context, x, y, w, h int, name, brief string, enabled bool, t theme) {
	// 状态色条
	if enabled {
		c.SetRGBA255(136, 178, 0, 255)
	} else {
		c.SetRGBA255(204, 51, 51, 255)
	}
	c.DrawRoundedRectangle(float64(x)+7, float64(y+24), float64(6), float64(h-48), 3)
	c.Fill()

	const (
		nameFont      = "data/Font/GlowSansSC-Normal-ExtraBold.ttf"
		briefFont     = "data/Font/regular-bold.ttf"
		nameBaseline  = 55 // 按 itemH=120 视觉居中：文本块上下各留约 30px
		briefBaseline = 89
	)

	// 插件名（黑色描边，远看清晰）
	drawTextOutlined(c, name, nameFont, 32, float64(x)+30, float64(y+nameBaseline), t.TextMain)

	// Brief：按实际像素宽度省略，右侧留出状态徽章区（圆徽 28 + 边距 16 + 间隙 10）
	availW := float64(w) - 30 - 54
	brief = ellipsizeByWidth(c, brief, briefFont, 22, availW)
	drawTextOutlined(c, brief, briefFont, 22, float64(x)+30, float64(y+briefBaseline), t.TextSec)

	// 状态徽章：右侧圆点 + 矢量勾/叉（✓/✗ 在 GlowSansSC 无字形，DrawString
	// 永远渲染不出来，改用直线段绘制图标，必然渲染且远看清晰）
	const iconD = 28.0
	cx := float64(x+w) - 16 - iconD/2
	cy := float64(y) + float64(h)/2
	if enabled {
		c.SetRGBA255(104, 166, 0, 240)
	} else {
		c.SetRGBA255(204, 51, 51, 240)
	}
	c.DrawCircle(cx, cy, iconD/2)
	c.Fill()

	r := iconD / 2
	c.SetStrokeStyle(gg.NewSolidPattern(color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	c.SetLineWidth(3)
	c.SetLineCap(gg.LineCapRound)
	if enabled {
		// 勾：短臂下探到中低点，再长臂扬到右上
		c.MoveTo(cx-0.42*r, cy+0.05*r)
		c.LineTo(cx-0.10*r, cy+0.38*r)
		c.LineTo(cx+0.45*r, cy-0.30*r)
		c.Stroke()
	} else {
		// 叉：两条对角直线
		c.MoveTo(cx-0.30*r, cy-0.30*r)
		c.LineTo(cx+0.30*r, cy+0.30*r)
		c.MoveTo(cx+0.30*r, cy-0.30*r)
		c.LineTo(cx-0.30*r, cy+0.30*r)
		c.Stroke()
	}
}

// buildBackground 构建主题渐变 + 随机本地图（cover 模式）
func buildBackground(w, h int, t theme) *image.RGBA {
	bg := gg.NewContext(w, h)

	fr, fg, fb := float64(t.BackgroundFrom[0]), float64(t.BackgroundFrom[1]), float64(t.BackgroundFrom[2])
	tr, tg, tb := float64(t.BackgroundTo[0]), float64(t.BackgroundTo[1]), float64(t.BackgroundTo[2])
	grad := gg.NewLinearGradient(0, 0, 0, float64(h))
	grad.AddColorStop(0, color.RGBA{R: uint8(fr), G: uint8(fg), B: uint8(fb), A: 255})
	grad.AddColorStop(1, color.RGBA{R: uint8(tr), G: uint8(tg), B: uint8(tb), A: 255})
	bg.SetFillStyle(grad)
	bg.DrawRectangle(0, 0, float64(w), float64(h))
	bg.Fill()

	randomImg := loadRandomBg()
	if randomImg != nil {
		srcW := randomImg.Bounds().Dx()
		srcH := randomImg.Bounds().Dy()
		scale := math.Max(float64(w)/float64(srcW), float64(h)/float64(srcH))
		newW := int(float64(srcW) * scale)
		newH := int(float64(srcH) * scale)
		resized := imaging.Resize(randomImg, newW, newH, imaging.Box)
		final := imaging.CropCenter(resized, w, h)
		bg.DrawImage(final, 0, 0)
	}

	return bg.Image().(*image.RGBA)
}

func renderUsageCard(m *zbpctrl.Control[*zero.Ctx]) ([]byte, error) {
	t := currentTheme
	cardW := 900
	headerH := 110

	help := m.Options.Help
	if help == "" {
		help = "该插件无帮助文档。"
	}
	lines := strings.Count(help, "\n") + 1
	lineH := 28
	bodyH := lines*lineH + 60
	cardH := cardPadding + headerH + bodyH + cardPadding

	c := gg.NewContext(cardW, cardH)
	bg := buildBackground(cardW, cardH, t)
	c.DrawImage(bg, 0, 0)

	if t.OverlayAlpha > 0 {
		oc := t.OverlayColor
		c.SetRGBA255(int(oc.R), int(oc.G), int(oc.B), int(t.OverlayAlpha))
		c.DrawRectangle(0, 0, float64(cardW), float64(cardH))
		c.Fill()
	}

	blurback := toRGBA(imaging.Blur(c.Image(), t.Blur))

	drawNewCard(c, cardPadding, cardPadding, cardW-cardPadding*2, headerH, blurback, t)
	c.SetColor(t.TextMain)
	c.LoadFontFace("data/Font/GlowSansSC-Normal-ExtraBold.ttf", 48)
	c.DrawString(m.Service, float64(cardPadding+32), float64(cardPadding+60))

	c.SetColor(t.TextSec)
	c.LoadFontFace("data/Font/regular-bold.ttf", 18)
	c.DrawString(m.Options.Brief, float64(cardPadding+32), float64(cardPadding+90))

	bodyY := cardPadding + headerH + cardMarginY
	drawNewCard(c, cardPadding, bodyY, cardW-cardPadding*2, bodyH, blurback, t)
	c.SetColor(t.TextMain)
	c.LoadFontFace("data/Font/regular-bold.ttf", 16)
	y := float64(bodyY) + 36
	for _, line := range strings.Split(help, "\n") {
		c.DrawString(line, float64(cardPadding+28), y)
		y += float64(lineH)
	}
	return encodePNG(c.Image())
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
