// Package gamesystem 基于zbp的猜歌插件
package gamesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"strings"
	"sync"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 图片输出
	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/rendercard"
	"github.com/FloatTech/zbputils/img/text"
)

const (
	serviceErr = "[gamesystem]error:"
	kanbanpath = "data/Control/icon.jpg"
)

type gameinfo struct {
	Command string `json:"游玩指令"` // 游玩指令
	Help    string `json:"游戏说明"` // 游戏说明
	Rewards string `json:"奖励说明"` // 奖励说明
}

type gameStatus struct {
	Name  string         `json:"游戏名称"`
	Info  gameinfo       `json:"游戏介绍"`
	Sales map[int64]bool `json:"上架情况"`
	Rooms []int64        `json:"房间列表"`
}

var (
	// 插件主体
	engine = control.Register("gamesystem", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "游戏系统",
		Help:             "- 游戏列表\n- 上架[游戏名]\n- 下架[游戏名]",
		PublicDataFolder: "GameSystem",
	})
	// 游戏控件
	cfgFile     = engine.DataFolder() + "gamesystem.json"
	mu          sync.RWMutex
	gamelist    = make(map[string]gameinfo, 30)
	gameManager = make(map[string]*gameStatus, 30)
)

func init() {
	engine.OnCommandGroup([]string{"上架", "下架"}, zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		gid := ctx.Event.GroupID
		if strings.Contains(model.Command, "上架") {
			err := whichGameSalesIn(model.Args, gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Text(model.Args, "游戏已上架"))
		} else {
			err := whichGameSalesOut(model.Args, gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Text(model.Args, "游戏已下架"))
		}
	})
	engine.OnFullMatch("游戏列表").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			i := 0
			var imgs []image.Image
			var yOfLine1 int // 第一列最大高度
			var yOfLine2 int // 第二列最大高度
			for gameName, info := range gamelist {
				img, err := rendercard.TextCardInfo{
					FontOfTitle:  text.SakuraFontFile,
					FontOfText:   text.SakuraFontFile,
					Title:        gameName,
					DisplayTitle: true,
					TitleSetting: "Center",
					Text: func() []string {
						var infoText []string
						if !whichGamePlayIn(gameName, gid) {
							infoText = append(infoText, []string{"游戏状态:", "     下架中"}...)
						}
						infoText = append(infoText, []string{
							"游戏指令:", "    " + info.Command,
							"游戏说明:", strings.ReplaceAll("    "+info.Help, "\n", "\n    "),
							"游戏奖励:", "    " + info.Rewards}...)
						return infoText
					}(),
					TextSetting: true,
				}.DrawTextCard()
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
				if i%2 == 0 { // 第一列
					yOfLine1 += img.Bounds().Max.Y + 20
				} else { // 第二列
					yOfLine2 += img.Bounds().Max.Y + 20
				}
				imgs = append(imgs, img)
				i++
			}
			lnperpg := math.Ceil(math.Max(yOfLine1, yOfLine2), (256 + 30))
			imgback, err := rendercard.Titleinfo{
				Line:          lnperpg,
				Lefttitle:     "游戏系统",
				Leftsubtitle:  "Game System",
				Righttitle:    "FloatTech",
				Rightsubtitle: "ZeroBot-Plugin",
				Fontpath:      text.SakuraFontFile,
				Imgpath:       kanbanpath,
			}.Drawtitle()
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			yOfLine := []int{0, 0}
			canvas := gg.NewContextForImage(imgback)
			// 插入游戏列表卡片
			for i, img := range imgs {
				canvas.DrawImage(img, 25+620*(i%2), 360+yOfLine[i%2])
				yOfLine[i%2] += img.Bounds().Max.Y + 20
			}
			data, cl := writer.ToBytes(canvas.Image())
			defer cl()
			if id := ctx.SendChain(message.ImageBytes(data)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}

// 载入游戏状态
func loadConfig(cfgFile string) error {
	if file.IsExist(cfgFile) {
		reader, err := os.Open(cfgFile)
		if err == nil {
			err = json.NewDecoder(reader).Decode(&gameManager)
		}
		if err != nil {
			return err
		}
		return reader.Close()
	} else {
		return saveConfig(cfgFile)
	}
}

// 保存游戏状态
func saveConfig(cfgFile string) error {
	reader, err := os.Create(cfgFile)
	if err == nil {
		err = json.NewEncoder(reader).Encode(&gameManager)
	}
	return err
}

// 注册游戏
func register(gameName string, gameinfo gameinfo) error {
	if len(gameManager) == 0 {
		err := loadConfig(cfgFile)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	defer mu.Unlock()
	gamelist[gameName] = gameinfo
	_, ok := gameManager[gameName]
	if !ok {
		gameManager[gameName] = &gameStatus{
			Name:  gameName,
			Info:  gameinfo,
			Sales: make(map[int64]bool),
			Rooms: make([]int64, 0),
		}
		return saveConfig(cfgFile)
	}
	return nil
}

// 判断游戏是否上架
func whichGamePlayIn(gameName string, groupID int64) bool {
	if len(gameManager) == 0 {
		err := loadConfig(cfgFile)
		if err != nil {
			panic(err)
		}
		fmt.Println("reaad congfig:", gameManager)
	}
	fmt.Println("before register:", gameManager)
	mu.Lock()
	defer mu.Unlock()
	status, ok := gameManager[gameName]
	if ok {
		sales, ok := status.Sales[groupID]
		if !ok {
			status.Sales[groupID] = true
			sales = true
			_ = saveConfig(cfgFile)
		}
		fmt.Println(gameManager)
		return sales
	}
	return false
}

// 上架游戏
func whichGameSalesIn(gameName string, groupID int64) error {
	if len(gameManager) == 0 {
		err := loadConfig(cfgFile)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	defer mu.Unlock()
	status, ok := gameManager[gameName]
	if ok {
		status.Sales[groupID] = true
		return saveConfig(cfgFile)
	}
	return errors.New("该游戏不存在或者未注册")
}

// 下架游戏
func whichGameSalesOut(gameName string, groupID int64) error {
	if len(gameManager) == 0 {
		err := loadConfig(cfgFile)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	defer mu.Unlock()
	status, ok := gameManager[gameName]
	if ok {
		status.Sales[groupID] = false
		return saveConfig(cfgFile)
	}
	return errors.New("该游戏不存在或者未注册")
}

// 创建房间
func whichGameRoomIn(gameName string, groupID int64) error {
	if len(gameManager) == 0 {
		err := loadConfig(cfgFile)
		if err != nil {
			return err
		}
	}
	if !whichGamePlayIn(gameName, groupID) {
		return errors.New("游戏已下架,无法游玩")
	}
	mu.Lock()
	defer mu.Unlock()
	status, ok := gameManager[gameName]
	if ok {
		for _, gid := range status.Rooms {
			if gid == groupID {
				return errors.New("游戏已创建房间,请等待结束后重试")
			}
		}
		// 创建房间
		status.Rooms = append(status.Rooms, groupID)
		return nil
	}
	return errors.New("游戏未完成注册")
}

// 关闭房间
func whichGameRoomOut(gameName string, groupID int64) error {
	if len(gameManager) == 0 {
		err := loadConfig(cfgFile)
		if err != nil {
			return err
		}
	}
	mu.Lock()
	defer mu.Unlock()
	status, ok := gameManager[gameName]
	if ok {
		index := 0
		for i, gid := range status.Rooms {
			if gid == groupID {
				index = i
			}
		}
		status.Rooms = append(status.Rooms[:index], status.Rooms[index+1:]...)
		return nil
	}
	return errors.New("游戏未完成注册")
}
