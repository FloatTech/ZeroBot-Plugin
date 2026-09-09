package servicemenu

// 测试渲染：不走 bot，直接用真实渲染管线输出 PNG 供视觉验证
// 运行: go test -run TestRenderListPNG ./plugin/servicemenu
// 注意: chdir 到项目根目录，保证 data/Font、data/aifalse 相对路径可用

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"testing"

	"github.com/FloatTech/ZeroBot-Plugin/kanban/banner"
	"github.com/FloatTech/gg"
	"github.com/disintegration/imaging"
)

func TestMain(m *testing.M) {
	_ = os.Chdir("../..") // 项目根目录
	os.Exit(m.Run())
}

// newTestCanvas 构建带真实背景的画布 + 对应 blurback（与 renderServiceList 同流程）
func newTestCanvas(w, h int, tt theme) (*gg.Context, *image.RGBA) {
	c := gg.NewContext(w, h)
	bg := buildBackground(w, h, tt)
	c.DrawImage(bg, 0, 0)
	if tt.OverlayAlpha > 0 {
		oc := tt.OverlayColor
		c.SetRGBA255(int(oc.R), int(oc.G), int(oc.B), int(tt.OverlayAlpha))
		c.DrawRectangle(0, 0, float64(w), float64(h))
		c.Fill()
	}
	blurback := toRGBA(imaging.Blur(c.Image(), tt.Blur))
	return c, blurback
}

// writePNG 输出画布到项目根目录 PNG
func writePNG(t *testing.T, c *gg.Context, out string) {
	png, err := encodePNG(c.Image())
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	if err := os.WriteFile(out, png, 0o644); err != nil {
		t.Fatalf("写出失败: %v", err)
	}
	t.Logf("已输出 %s", out)
}

type fakePlugin struct{ name, brief string }

func renderTestList(t *testing.T, tt theme, out string) {
	plugins := []fakePlugin{
		{"job", "定时指令触发器"}, {"antiabuse", "违禁词检测"}, {"chat", "基础反应, 群空调"},
		{"chatcount", "聊天时长统计"}, {"sleepmanage", "睡眠小助手"}, {"airecord", "群应用: AI声聊"},
		{"atri", "atri人格文本回复"}, {"manager", "群管插件"}, {"aiwife", "ai随机生成老婆"},
		// 长简介：验证按像素宽度省略（应占满可用宽度而非提前截断）
		{"animetrace", "AnimeTrace 动画/Galgame 截图溯源，识别番剧出处与集数"},
		{"danbooru", "二次元图片标签查询，支持 endless 名片与同名搜索重定向"},
		{"event", "好友申请和群聊邀请审核事件处理，自动同意或拒绝入群请求"},
	}
	currentTheme = tt

	// 与 renderServiceList 相同的布局常量
	canvasW := 1200
	rows := 6 // 12 个测试插件 / 两列
	canvasH := cardPadding + headerH + rows*(itemH+cardMarginY) + footerH + cardPadding + 40
	c, blurback := newTestCanvas(canvasW, canvasH, tt)

	// header 卡
	drawNewCard(c, cardPadding, cardPadding, canvasW-cardPadding*2, headerH, blurback, tt)
	c.SetColor(tt.TextMain)
	c.LoadFontFace("data/Font/GlowSansSC-Normal-ExtraBold.ttf", 48)
	c.DrawString("ZeroBot-Plugin", float64(cardPadding+32), float64(cardPadding+60))
	drawTextOutlined(c, "OneBot + ZeroBot + Golang", "data/Font/regular-bold.ttf", 20, float64(cardPadding+32), float64(cardPadding+92), tt.TextSec)
	drawTextOutlined(c, banner.Version+" · FloatTech", "data/Font/regular-bold.ttf", 18, float64(cardPadding+32), float64(cardPadding+120), tt.TextSec)

	// 插件卡片
	cardAreaTop := cardPadding + headerH + cardMarginY
	cardW := (canvasW - cardPadding*2 - colGap*(itemsPerRow-1)) / itemsPerRow
	for j, p := range plugins {
		col := j % itemsPerRow
		row := j / itemsPerRow
		x := float64(cardPadding) + float64(col)*float64(cardW+colGap)
		cardTop := float64(cardAreaTop) + float64(row)*(itemH+cardMarginY)
		drawNewCard(c, int(x), int(cardTop), cardW, itemH, blurback, tt)
		drawPluginCardContent(c, int(x), int(cardTop), cardW, itemH, p.name, p.brief, j%4 != 3, tt)
	}

	writePNG(t, c, out)
}

func TestRenderListPNG(t *testing.T) {
	for _, tt := range themes {
		renderTestList(t, tt, fmt.Sprintf("glass_real_%s.png", tt.Name))
	}
}

// TestRenderCloseupPNG 单卡放大特写：检查边缘折射 / 四角 / 颗粒感
func TestRenderCloseupPNG(t *testing.T) {
	tt := themes[0]
	currentTheme = tt

	c, blurback := newTestCanvas(760, 560, tt)

	// 一大一小两块玻璃
	drawNewCard(c, 60, 60, 640, 220, blurback, tt)
	drawPluginCardContent(c, 60, 60, 640, 220, "liquid glass", "凸透镜放大 + 边缘折射 + 色散", true, tt)

	drawNewCard(c, 60, 330, 300, 120, blurback, tt)
	drawPluginCardContent(c, 60, 330, 300, 120, "job", "定时指令触发器", true, tt)

	drawNewCard(c, 400, 330, 300, 120, blurback, tt)
	drawPluginCardContent(c, 400, 330, 300, 120, "manager", "群管插件", false, tt)

	writePNG(t, c, "glass_closeup.png")
}

// TestCropDebug 裁剪放大整表的卡片角部，定位黑刺瑕疵
func TestCropDebug(t *testing.T) {
	src, err := imaging.Open("glass_real_glass-dark.png")
	if err != nil {
		t.Skip("先跑 TestRenderListPNG")
	}
	// job 卡所在区域（含左右角 + 上下缘）
	crop := imaging.Crop(src, image.Rect(50, 265, 470, 400))
	zoom := imaging.Resize(crop, 420*3, 135*3, imaging.NearestNeighbor)
	_ = imaging.Save(zoom, "glass_crop.png")
	t.Log("已输出 glass_crop.png")

	// 6x 特写：job 卡左上角 (70,290) 及左缘
	crop2 := imaging.Crop(src, image.Rect(40, 262, 150, 372))
	zoom2 := imaging.Resize(crop2, 110*6, 110*6, imaging.NearestNeighbor)
	_ = imaging.Save(zoom2, "glass_corner6x.png")
	t.Log("已输出 glass_corner6x.png")

	// 像素扫描：横穿 job 卡左缘（y=320），定位暗线位置
	for _, py := range []int{300, 320, 340} {
		row := ""
		for px := 60; px <= 84; px += 2 {
			r, g, b, _ := src.At(px, py).RGBA()
			row += fmt.Sprintf("x%d:%d,%d,%d ", px, r>>8, g>>8, b>>8)
		}
		t.Logf("y=%d %s", py, row)
	}
}

// TestDarkPixelForensics 逐像素取证：step=1 扫描左缘全线 + 分层隔离（玻璃/阴影/描边）
// 定位 1px 深色竖线与角部斑点来自哪一层
func TestDarkPixelForensics(t *testing.T) {
	tt := themes[0]
	currentTheme = tt

	// 复现 renderTestList 的 job 卡布局
	canvasW := 1200
	rows := 3
	canvasH := cardPadding + headerH + rows*(itemH+cardMarginY) + footerH + cardPadding + 40
	c, blurback := newTestCanvas(canvasW, canvasH, tt)
	_ = c
	cardAreaTop := cardPadding + headerH + cardMarginY
	cardW := (canvasW - cardPadding*2 - colGap*(itemsPerRow-1)) / itemsPerRow
	x, y, w, h := cardPadding, cardAreaTop, cardW, itemH
	t.Logf("job 卡: x=%d y=%d w=%d h=%d cardRadius=%d edgeW=%.1f", x, y, w, h, cardRadius, clampF(float64(h)*0.30, 14, 36))

	// === 分层渲染 ===
	// L1: 仅玻璃
	glass := renderLiquidGlass(blurback, x, y, w, h, tt)
	// L2: 仅阴影
	shadow := getBlurredShadow(w, h, cardRadius)
	// L3: 完整卡片（画在副本 canvas 上）
	c2, blurback2 := newTestCanvas(canvasW, canvasH, tt)
	drawNewCard(c2, x, y, w, h, blurback2, tt)

	scanDark := func(label string, img *image.RGBA, ox, oy int) {
		found := 0
		for py := y; py < y+60 && py-oy < img.Bounds().Dy(); py++ {
			for px := x - 6; px <= x+8; px++ {
				lx, ly := px-ox, py-oy
				if lx < 0 || ly < 0 || lx >= img.Bounds().Dx() || ly >= img.Bounds().Dy() {
					continue
				}
				r, g, b, a := img.At(lx, ly).RGBA()
				r8, g8, b8, a8 := int(r>>8), int(g>>8), int(b>>8), int(a>>8)
				if a8 > 100 && r8 < 90 && g8 < 90 {
					t.Logf("[%s] 暗点 (%d,%d) RGB=%d,%d,%d A=%d", label, px, py, r8, g8, b8, a8)
					found++
					if found > 12 {
						return
					}
				}
			}
		}
		if found == 0 {
			t.Logf("[%s] 左缘 y+0..60 无暗点", label)
		}
	}

	// blurback 本身（对照组：证明背景没有暗点）
	scanDark("blurback", blurback, 0, 0)
	// 仅玻璃层
	scanDark("glass-only", glass, x, y)
	// 阴影层（在其自身坐标系里扫）
	scanDark("shadow-only", shadow, x-shadowPad, y-shadowPad)
	// 完整卡片
	scanDark("full-card", toRGBA(c2.Image()), 0, 0)

	// === step=1 全线扫描：完整卡片左缘 x-6..x+8, y..y+60 全部输出 ===
	full := toRGBA(c2.Image())
	line := ""
	for py := y; py < y+40; py++ {
		r, g, b, _ := full.At(x+2, py).RGBA()
		v := (int(r>>8) + int(g>>8) + int(b>>8)) / 3
		if v < 120 {
			line += fmt.Sprintf(" (%d,%d:%d)", x+2, py, v)
		}
	}
	t.Logf("x=%d 列上 40 行内暗像素(<120):%s", x+2, line)

	// === 玻璃层内部取证：对暗点反算采样坐标 ===
	cx, cy := float64(w)/2, float64(h)/2
	rad := float64(cardRadius)
	edgeW := clampF(float64(h)*0.30, 14, 36)
	sdf := func(fx, fy float64) float64 {
		qx := math.Abs(fx-cx) - (float64(w)/2 - rad)
		qy := math.Abs(fy-cy) - (float64(h)/2 - rad)
		return math.Min(math.Max(qx, qy), 0) + math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) - rad
	}
	for py := 0; py < 60; py++ {
		for px := 0; px < 12; px++ {
			r, g, b, a := glass.At(px, py).RGBA()
			if a>>8 <= 100 || int(r>>8) >= 90 {
				continue
			}
			fx, fy := float64(px)+0.5, float64(py)+0.5
			dE := sdf(fx, fy)
			nx := sdf(fx+1, fy) - sdf(fx-1, fy)
			ny := sdf(fx, fy+1) - sdf(fx, fy-1)
			if nl := math.Hypot(nx, ny); nl > 1e-6 {
				nx, ny = nx/nl, ny/nl
			}
			mag := tt.Refraction * smoothStep(-edgeW*1.5, -edgeW*0.2, dE) * smoothStep(2.0, -3.0, dE)
			sx := (fx+nx*mag-cx)/tt.Magnify + cx
			sy := (fy+ny*mag-cy)/tt.Magnify + cy
			sr, sg, sb, _ := blurback.At(int(sx)+x, int(sy)+y).RGBA()
			t.Logf("glass 暗点 local(%d,%d) dEdge=%.2f mag=%.1f n=(%.2f,%.2f) sample=(%.1f,%.1f)->canvas(%d,%d) 采样色=%d,%d,%d 输出=%d,%d,%d",
				px, py, dE, mag, nx, ny, sx, sy, int(sx)+x, int(sy)+y, sr>>8, sg>>8, sb>>8, r>>8, g>>8, b>>8)
		}
	}

	// 输出放大图目视
	crop := imaging.Crop(full, image.Rect(x-12, y-12, x+60, y+72))
	z6 := imaging.Resize(crop, 72*6, 84*6, imaging.NearestNeighbor)
	_ = imaging.Save(z6, "glass_forensic6x.png")
	t.Log("已输出 glass_forensic6x.png")
}

// TestGGBugProbe 最小复现：分别测试 gg 的描边（带 alpha 的 SolidPattern）与
// 圆角矩形 Clip+渐变填充，是否会在路径附近写入垃圾像素（黑点/蓝点）
func TestGGBugProbe(t *testing.T) {
	probe := func(label string, fn func(c *gg.Context)) {
		c := gg.NewContext(240, 240)
		c.SetRGBA255(200, 200, 200, 255)
		c.DrawRectangle(0, 0, 240, 240)
		c.Fill()
		fn(c)
		bad := 0
		for py := 0; py < 240; py++ {
			for px := 0; px < 240; px++ {
				r, g, b, a := c.Image().At(px, py).RGBA()
				r8, g8, b8, a8 := int(r>>8), int(g>>8), int(b>>8), int(a>>8)
				// 灰底 200 与白色系之外的颜色都算异常
				if a8 < 250 || (r8 < 150 && g8 < 150) || b8 > 240 && r8 < 100 {
					if bad < 8 {
						t.Logf("[%s] 异常像素 (%d,%d) RGB=%d,%d,%d A=%d", label, px, py, r8, g8, b8, a8)
					}
					bad++
				}
			}
		}
		t.Logf("[%s] 异常像素总数: %d", label, bad)
	}

	// A: 描边 - SetStrokeStyle(白色 A=60) + Stroke
	probe("stroke-alpha60", func(c *gg.Context) {
		c.SetLineWidth(1.2)
		c.SetStrokeStyle(gg.NewSolidPattern(color.RGBA{R: 255, G: 255, B: 255, A: 60}))
		c.DrawRoundedRectangle(20.5, 20.5, 199, 199, 16)
		c.Stroke()
	})

	// B: 描边 - 不设样式（默认黑色描边，预期全黑一圈作为对照）
	probe("stroke-default", func(c *gg.Context) {
		c.SetLineWidth(1.2)
		c.DrawRoundedRectangle(20.5, 20.5, 199, 199, 16)
		c.Stroke()
	})

	// C: Clip 圆角矩形 + 线性渐变填充（复刻高光带，radius 16 vs 高度 96）
	probe("clip-gradient", func(c *gg.Context) {
		c.DrawRoundedRectangle(22, 21, 196, 96, 16)
		c.Clip()
		hl := gg.NewLinearGradient(0, 20, 0, 116)
		hl.AddColorStop(0, color.RGBA{R: 255, G: 255, B: 255, A: 55})
		hl.AddColorStop(0.5, color.RGBA{R: 255, G: 255, B: 255, A: 18})
		hl.AddColorStop(1, color.RGBA{R: 255, G: 255, B: 255, A: 0})
		c.SetFillStyle(hl)
		c.DrawRectangle(20, 20, 200, 96)
		c.Fill()
		c.ResetClip()
	})

	// D: Clip 圆角矩形 + 纯色填充（隔离渐变因素）
	probe("clip-solid", func(c *gg.Context) {
		c.DrawRoundedRectangle(22, 21, 196, 96, 16)
		c.Clip()
		c.SetRGBA255(255, 255, 255, 40)
		c.DrawRectangle(20, 20, 200, 96)
		c.Fill()
		c.ResetClip()
	})
}

// TestLayerDebug 分层隔离实验：定位黑刺来自哪一层（阴影/玻璃/描边高光）
func TestLayerDebug(t *testing.T) {
	tt := themes[0]
	currentTheme = tt

	const panelW, panelH = 480, 150
	c := gg.NewContext(panelW, panelH*4)
	// 浅色平底背景（模拟浅色二次元图）
	c.SetRGBA255(235, 220, 215, 255)
	c.DrawRectangle(0, 0, float64(panelW), float64(panelH*4))
	c.Fill()

	// 先构建干净的 blurback（平底 + overlay + blur），避免把面板内容采进去
	flat := gg.NewContext(panelW, panelH*4)
	flat.SetRGBA255(235, 220, 215, 255)
	flat.DrawRectangle(0, 0, float64(panelW), float64(panelH*4))
	flat.Fill()
	if tt.OverlayAlpha > 0 {
		oc := tt.OverlayColor
		flat.SetRGBA255(int(oc.R), int(oc.G), int(oc.B), int(tt.OverlayAlpha))
		flat.DrawRectangle(0, 0, float64(panelW), float64(panelH*4))
		flat.Fill()
	}
	blurback := toRGBA(imaging.Blur(flat.Image(), tt.Blur))

	x, y, w, h := 70, 35, 337, 80

	// 面板1: 完整 drawNewCard
	drawNewCard(c, x, y, w, h, blurback, tt)

	// 面板2: 仅阴影
	shadow := getBlurredShadow(w, h, cardRadius)
	c.DrawImage(shadow, x-shadowPad, panelH+y-shadowPad+8)

	// 面板3: 仅玻璃（无阴影无描边）
	glass := renderLiquidGlass(blurback, x, y, w, h, tt)
	c.DrawImage(glass, x, panelH*2+y)

	// 面板4: 仅描边 + 顶部高光（无阴影无玻璃）
	r := float64(cardRadius)
	py3 := float64(panelH*3 + y)
	c.DrawRoundedRectangle(float64(x)+2, py3+1, float64(w)-4, float64(h)*0.4, r)
	c.Clip()
	hl := gg.NewLinearGradient(0, py3, 0, py3+float64(h)*0.4)
	hl.AddColorStop(0, color.RGBA{R: 255, G: 255, B: 255, A: tt.TopLight})
	hl.AddColorStop(0.5, color.RGBA{R: 255, G: 255, B: 255, A: uint8(int(tt.TopLight) / 3)})
	hl.AddColorStop(1, color.RGBA{R: 255, G: 255, B: 255, A: 0})
	c.SetFillStyle(hl)
	c.DrawRectangle(float64(x), py3, float64(w), float64(h)*0.4)
	c.Fill()
	c.ResetClip()
	c.SetLineWidth(1.2)
	c.SetRGBA255(255, 255, 255, int(tt.StrokeAlpha))
	c.DrawRoundedRectangle(float64(x)+0.5, py3+0.5, float64(w)-1, float64(h)-1, r)
	c.Stroke()

	// 放大 2.5x 便于查看
	zoom := imaging.Resize(c.Image(), panelW*5/2, panelH*10, imaging.NearestNeighbor)
	_ = imaging.Save(zoom, "glass_layers.png")
	t.Log("已输出 glass_layers.png")

	// 像素取证：p4 顶部左角外围 (卡片左上角在 x=70, y=panelH*3+35=485)
	img := c.Image()
	for _, py := range []int{474, 478, 482, 486, 490} {
		row := ""
		for px := 56; px <= 88; px += 4 {
			r, g, b, a := img.At(px, py).RGBA()
			row += fmt.Sprintf("[%3d,%3d,%3d,%3d] ", r>>8, g>>8, b>>8, a>>8)
		}
		t.Logf("y=%d x=56..88: %s", py, row)
	}
}
