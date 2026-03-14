// Package pigpig 猪猪表情包
package pigpig

import (
	"encoding/json"
	"errors"
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

// pigResponse 内部结构体
type pigResponse struct {
	Total  int        `json:"total"`
	Images []pigImage `json:"images"`
}

// pigImage 内部结构体
type pigImage struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
}

var (
	pigCache       []pigImage
	pigMap         = make(map[string]*pigImage)
	pigMutex       sync.RWMutex
	lastUpdateTime = time.Now()

	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "来份猪猪",
		Help:              "- 随机猪猪：随机发送一张猪猪表情\n- 搜索猪猪 [关键词]：搜索相关猪猪\n- 猪猪id [id]：精确查找",
		PrivateDataFolder: "Pig",
	})
)

func init() {
	_ = checkAndUpdateData()
	// 1. 随机猪猪
	engine.OnRegex(`^(随机猪猪|来份猪猪|抽个猪猪)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if err := checkAndUpdateData(); err != nil {
			ctx.SendChain(message.Text("[Pig] ERROR: ", err, "\nEXP: 随机猪猪失败，获取数据错误"))
			return
		}

		pigMutex.RLock()
		defer pigMutex.RUnlock()

		if len(pigCache) == 0 {
			ctx.SendChain(message.Text("[Pig] ERROR: 暂无猪猪数据，请联系管理员"))
			return
		}

		target := pigCache[rand.Intn(len(pigCache))]
		imgData, err := target.fetch()
		if err != nil {
			ctx.SendChain(message.Text("[Pig] ERROR: ", err, "\nEXP: 图片加载失败"))
			return
		}

		ctx.SendChain(
			message.Text(fmt.Sprintf("🐷 ID: %s | %s", target.ID, target.Title)),
			message.ImageBytes(imgData),
		)
	})

	// 2. 搜索猪猪
	engine.OnRegex(`^搜索猪猪\s+(.+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		keyword := strings.TrimSpace(ctx.State["regex_matched"].([]string)[1])

		if err := checkAndUpdateData(); err != nil {
			ctx.SendChain(message.Text("[Pig] ERROR: ", err, "\nEXP: 搜索猪猪失败，获取数据错误"))
			return
		}

		pigMutex.RLock()
		defer pigMutex.RUnlock()

		var results []pigImage
		for _, p := range pigCache {
			if strings.Contains(p.Title, keyword) {
				results = append(results, p)
			}
		}

		if len(results) == 0 {
			ctx.SendChain(message.Text("[Pig] ERROR: 未找到包含“", keyword, "”的猪猪"))
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

		sb.WriteString("\n为您返回第一个猪猪：\n💡 提示：输入“猪猪id [id]”可精确获取")

		imgData, err := results[0].fetch()
		if err != nil {
			ctx.SendChain(message.Text(sb.String(), "\n\n[Pig] ERROR: ", err, "\nEXP: 图片加载失败"))
			return
		}

		ctx.SendChain(
			message.Text(sb.String()),
			message.ImageBytes(imgData), // 直接使用 ImageBytes
		)
	})

	// 3. 猪猪id精确查找
	engine.OnRegex(`^猪猪id\s+(\d+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		targetID := ctx.State["regex_matched"].([]string)[1]

		if err := checkAndUpdateData(); err != nil {
			ctx.SendChain(message.Text("[Pig] ERROR: ", err, "\nEXP: 精确查找失败，获取数据错误"))
			return
		}

		pigMutex.RLock()
		defer pigMutex.RUnlock()

		target, exists := pigMap[targetID]
		if !exists {
			ctx.SendChain(message.Text("[Pig] ERROR: 未找到 ID 为 ", targetID, " 的猪猪"))
			return
		}

		imgData, err := target.fetch()
		if err != nil {
			ctx.SendChain(message.Text("[Pig] ERROR: ", err, "\nEXP: 图片加载失败"))
			return
		}

		ctx.SendChain(
			message.Text(fmt.Sprintf("🐷 ID: %s | %s", target.ID, target.Title)),
			message.ImageBytes(imgData),
		)
	})
}

// checkAndUpdateData
func checkAndUpdateData() error {
	pigMutex.Lock()
	defer pigMutex.Unlock()

	// 如果有缓存且距上次更新不足 24 小时，直接返回
	if len(pigCache) > 0 && time.Since(lastUpdateTime) < 24*time.Hour {
		return nil
	}

	dataBytes, err := engine.GetLazyData("pig_data.json", true)
	if err != nil {
		return errors.New("读取数据文件失败: " + err.Error())
	}

	var data pigResponse
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return errors.New("解析JSON失败: " + err.Error())
	}

	if len(data.Images) == 0 {
		return errors.New("数据文件为空")
	}

	pigCache = data.Images

	// 更新缓存时，顺便重构一份查询 Map
	newMap := make(map[string]*pigImage, len(pigCache))
	for i := range pigCache {
		newMap[pigCache[i].ID] = &pigCache[i]
	}
	pigMap = newMap

	lastUpdateTime = time.Now()
	return nil
}

func (img *pigImage) fetch() ([]byte, error) {
	if img.Filename == "" {
		return nil, errors.New("图片数据异常，缺少文件名")
	}

	targetPath := filepath.Join("assets", img.Filename)

	imgData, err := engine.GetLazyData(targetPath, true)
	if err != nil {
		return nil, errors.New("图片资源缺失 (" + targetPath + "): " + err.Error())
	}

	return imgData, nil
}
