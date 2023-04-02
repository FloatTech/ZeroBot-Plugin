// Package dish 程序员做饭指南zbp版，来源Anduin2017/HowToCook
//
// 使用前需要先git clone https://github.com/Anduin2017/HowToCook.git /path/to/data/Dish
package dish

import (
	"bufio"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type dishMenu struct {
	path    string
	name    string
	content dishContent
}

type dishContent struct {
	materials  []string
	operations []string
}

var (
	dishes   = map[string]dishMenu{}
	dishList []string
)

func scanDishes(dishesPath string) error {
	return filepath.Walk(dishesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ext := filepath.Ext(info.Name()); strings.ToLower(ext) == ".md" {
			dishName := strings.TrimSuffix(strings.ToLower(info.Name()), ".md")
			dishList = append(dishList, dishName)
			dishes[dishName] = dishMenu{path: path, name: dishName}
		}

		return nil
	})
}

func parseDish(dishPath string) (content dishContent, e error) {
	file, err := os.Open(dishPath)
	if err != nil {
		return dishContent{}, err
	}
	defer func() {
		closeErr := file.Close()
		if e == nil && closeErr != nil {
			e = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	var buffer []string
	for scanner.Scan() {
		text := scanner.Text()
		buffer = append(buffer, text)
	}

	if err := scanner.Err(); err != nil {
		return dishContent{}, err
	}

	var materialStartLine, materialEndLine, operationStartLine, operationEndLine = 0, 0, 0, 0
	for line, text := range buffer {
		buffer[line] = strings.Replace(text, " ", "", -1)
		buffer[line] = strings.Replace(text, "*", "", -1)
		buffer[line] = strings.Replace(text, "-", "", -1)
		switch text {
		case "## 必备原料和工具":
			materialStartLine = line + 2
		case "## 计算":
			materialEndLine = line - 1
		case "## 操作":
			operationStartLine = line + 2
		case "## 附加内容":
			operationEndLine = line - 1
		default:
		}
	}

	return dishContent{
		materials:  buffer[materialStartLine:materialEndLine],
		operations: buffer[operationStartLine:operationEndLine],
	}, nil
}

func formatDish(content dishContent) (material, operation string) {
	material = strings.Join(content.materials, "、")
	operation = ""
	for index, step := range content.operations {
		operation = fmt.Sprintf("%s%d. %s", operation, index+1, step)
		if index != len(content.operations)-1 {
			operation = fmt.Sprintf("%s\n", operation)
		}
	}

	return material, operation
}

func init() {
	en := control.Register("dish", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "程序员做饭指南",
		Help:             "- 怎么做[xxx]|[随机菜谱]",
		PublicDataFolder: "Dish",
	})

	rootDir := en.DataFolder() + "dishes"
	if err := scanDishes(rootDir); err != nil {
		panic(err)
	}
	for _, dish := range dishList {
		parsed, e := parseDish(dishes[dish].path)
		if e == nil {
			content := dishes[dish]
			content.content = parsed
			dishes[dish] = content
		}
	}

	en.OnPrefix("随机菜谱").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		name := ctx.NickName()
		index := rand.Intn(len(dishes))
		dish := dishes[dishList[index]]
		material, operation := formatDish(dish.content)
		ctx.SendChain(message.Text(fmt.Sprintf("客官%s这次的菜谱为%s\n原材料：%s\n做法：\n%s", name, dish.name, material, operation)))
	})

	en.OnPrefix("怎么做").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		dish := ctx.State["args"].(string)
		if content, has := dishes[dish]; has {
			material, operation := formatDish(content.content)
			ctx.SendChain(message.Text(fmt.Sprintf("已为客官呈上%s：\n原材料：%s\n做法：\n%s", dish, material, operation)))
		} else {
			ctx.SendChain(message.Text(fmt.Sprintf("没法为客官找到%s这道菜呢", dish)))
		}
	})
}
