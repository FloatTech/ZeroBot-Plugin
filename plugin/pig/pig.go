// Package pigpig 猪猪表情包
package pigpig

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// PigResponse 对应精简后的 JSON 结构
type PigResponse struct {
	Total  int        `json:"total"`
	Images []PigImage `json:"images"`
}

// PigImage 图片信息结构
type PigImage struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
}

var (
	// 数据缓存
	pigCache []PigImage
	// 读写锁
	pigMutex sync.RWMutex
	// 上次更新时间
	lastUpdateTime time.Time
	// 引擎实例
	engine *control.Engine
)

func init() {
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "随机/搜索猪猪表情包",
		Help:             "- 随机猪猪：随机发送一张猪猪表情\n- 搜索猪猪 [关键词]：搜索相关猪猪\n- 猪猪id [id]：精确查找",
		PublicDataFolder: "Pig",
	})

	engine.OnRegex(`^随机猪猪$`).SetBlock(true).Handle(handleRandomPig)
	engine.OnRegex(`^搜索猪猪\s+(.+)$`).SetBlock(true).Handle(handleSearchPig)
	engine.OnRegex(`^猪猪id\s+(\d+)$`).SetBlock(true).Handle(handlePigByID)
}

// checkAndUpdateData 检查并更新数据 (读取 pig_data.json)
func checkAndUpdateData(ctx *zero.Ctx) error {
	pigMutex.Lock()
	defer pigMutex.Unlock()

	now := time.Now()
	// 如果缓存为空，或者今天是新的一天，则尝试更新
	shouldUpdate := len(pigCache) == 0 || now.Format("2006-01-02") != lastUpdateTime.Format("2006-01-02")

	if shouldUpdate {
		if ctx != nil {
			ctx.SendChain(message.Text("🐷 正在同步今日猪猪数据，请稍候..."))
		}

		// 读取根目录下的 pig_data.json
		dataBytes, err := engine.GetLazyData("pig_data.json", true)
		if err != nil {
			return fmt.Errorf("读取数据文件失败: %v", err)
		}

		var data PigResponse
		if err := json.Unmarshal(dataBytes, &data); err != nil {
			return fmt.Errorf("解析JSON失败: %v", err)
		}

		if len(data.Images) == 0 {
			return fmt.Errorf("数据文件为空")
		}

		pigCache = data.Images
		lastUpdateTime = now

		if ctx != nil {
			ctx.SendChain(message.Text(fmt.Sprintf("✅ 同步完成！当前共有 %d 只猪猪。", len(pigCache))))
		}
	}
	return nil
}

// fetchImageLazy 按需从 assets 文件夹获取图片并转为 Base64
func fetchImageLazy(img PigImage) (string, error) {
	if img.Filename == "" {
		return "", fmt.Errorf("图片数据异常，缺少文件名")
	}

	// 【核心逻辑】拼接 assets 子目录
	// JSON 里的 Filename 是 "2072.jpg"
	// 实际在仓库里的路径是 "Pig/assets/2072.jpg"
	// GetLazyData 会自动在前加上 "Pig/"，传 "assets/2072.jpg"
	targetPath := filepath.Join("assets", img.Filename)

	// false 表示优先使用本地文件，不强制从网络拉取
	imgData, err := engine.GetLazyData(targetPath, false)
	if err != nil {
		return "", fmt.Errorf("图片资源缺失 (%s): %v", targetPath, err)
	}

	return "base64://" + base64.StdEncoding.EncodeToString(imgData), nil
}

// handleRandomPig 处理随机猪猪
func handleRandomPig(ctx *zero.Ctx) {
	if err := checkAndUpdateData(ctx); err != nil {
		ctx.SendChain(message.Text("❌ 获取数据失败: ", err))
		return
	}

	pigMutex.RLock()
	defer pigMutex.RUnlock()

	if len(pigCache) == 0 {
		ctx.SendChain(message.Text("❌ 暂无猪猪数据"))
		return
	}

	idx := rand.Intn(len(pigCache))
	target := pigCache[idx]

	b64Image, err := fetchImageLazy(target)
	if err != nil {
		ctx.SendChain(message.Text("❌ 图片加载失败: ", err))
		return
	}

	ctx.SendChain(
		message.Text(fmt.Sprintf("🐷 ID: %s | %s", target.ID, target.Title)),
		message.Image(b64Image),
	)
}

// handleSearchPig 处理搜索猪猪
func handleSearchPig(ctx *zero.Ctx) {
	keyword := ctx.State["regex_matched"].([]string)[1]
	keyword = strings.TrimSpace(keyword)

	if err := checkAndUpdateData(ctx); err != nil {
		ctx.SendChain(message.Text("❌ 获取数据失败: ", err))
		return
	}

	pigMutex.RLock()
	defer pigMutex.RUnlock()

	var results []PigImage
	for _, p := range pigCache {
		// 模糊匹配标题
		if strings.Contains(p.Title, keyword) {
			results = append(results, p)
		}
	}

	if len(results) == 0 {
		ctx.SendChain(message.Text(fmt.Sprintf("❌ 未找到包含“%s”的猪猪。", keyword)))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔎 根据关键词“%s”找到 %d 只猪猪：\n", keyword, len(results)))

	maxShow := 10
	for i, p := range results {
		if i >= maxShow {
			sb.WriteString(fmt.Sprintf("\n...等共 %d 条结果", len(results)))
			break
		}
		sb.WriteString(fmt.Sprintf("%d: %s (ID: %s)\n", i+1, p.Title, p.ID))
	}

	sb.WriteString("\n为您返回第一个猪猪：\n")
	sb.WriteString("💡 提示：输入“猪猪id [id]”可精确获取")

	b64Image, err := fetchImageLazy(results[0])
	if err != nil {
		ctx.SendChain(message.Text(sb.String() + "\n❌ 图片加载失败: " + err.Error()))
		return
	}

	ctx.SendChain(
		message.Text(sb.String()),
		message.Image(b64Image),
	)
}

// handlePigByID 处理ID精确查找
func handlePigByID(ctx *zero.Ctx) {
	targetID := ctx.State["regex_matched"].([]string)[1]

	if err := checkAndUpdateData(ctx); err != nil {
		ctx.SendChain(message.Text("❌ 获取数据失败: ", err))
		return
	}

	pigMutex.RLock()
	defer pigMutex.RUnlock()

	var target *PigImage
	for _, p := range pigCache {
		if p.ID == targetID {
			val := p
			target = &val
			break
		}
	}

	if target == nil {
		ctx.SendChain(message.Text(fmt.Sprintf("❌ 未找到 ID 为 %s 的猪猪。", targetID)))
		return
	}

	b64Image, err := fetchImageLazy(*target)
	if err != nil {
		ctx.SendChain(message.Text("❌ 图片加载失败: ", err))
		return
	}

	ctx.SendChain(
		message.Text(fmt.Sprintf("🐷 ID: %s | %s", target.ID, target.Title)),
		message.Image(b64Image),
	)
}