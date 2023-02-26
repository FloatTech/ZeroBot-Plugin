// Package ygo 一些关于ygo的插件
package ygo

import (
	"errors"
	"fmt"
	"image"
	"math/rand"
	"os"
	"strings"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/fumiama/jieba/util/helper"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	lotsList = make(map[string]string, 100)
	en       = control.Register("drawlots", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "gif抽签",
		Help:              "多功能抽签\n支持图片文件夹和gif\n-------------\n- 签列表\n- 抽xxx\n- 添加抽签xxx[gif图片]",
		PrivateDataFolder: "drawlots",
	}).ApplySingle(ctxext.DefaultSingle)
	datapath = file.BOTPATH + "/" + en.DataFolder()
)

func init() {
	var err error
	go func() {
		lotsList, err = getList()
		if err != nil {
			service, ok := control.Lookup("drawlots")
			if ok {
				service.Disable(0)
			}
			logrus.Infoln("[drawlots]发生错误,已尝试主动全局禁用该插件.错误信息:", err)
		} else {
			logrus.Infoln("[drawlots]加载", len(lotsList), "个抽签")
		}
	}()
	en.OnFullMatch("签列表").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		messageText := make([]string, 0, len(lotsList))
		messageText = append(messageText, []string{
			" 抽 签 签 名 [ 类 型 ] ", "----------",
		}...)
		for name, fileType := range lotsList {
			messageText = append(messageText, []string{
				name + "[" + fileType + "]", "----------",
			}...)
		}
		textPic, err := text.RenderToBase64(strings.Join(messageText, "\n"), text.BoldFontFile, 1080, 50)
		if err != nil {
			return
		}
		ctx.SendChain(message.Image("base64://" + helper.BytesToString(textPic)))
	})
	en.OnPrefix("抽").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		lotsType := ctx.State["args"].(string)
		fileType, ok := lotsList[lotsType]
		if !ok {
			ctx.SendChain(message.Text("该签不存在,无法抽签"))
			return
		}
		if fileType == "file" {
			picPath, err := randFile(lotsType, 3)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("file:///"+picPath))
			return
		}
		lotsImg, err := randGif(lotsType)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		// 生成图片
		data, cl := writer.ToBytes(lotsImg)
		defer cl()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.ImageBytes(data))
	})
	en.OnPrefix("添加抽签", zero.MustProvidePicture, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		id := ctx.Event.MessageID
		lotsName := ctx.State["args"].(string)
		if lotsName == "" {
			ctx.SendChain(message.Reply(id), message.Text("请使用正确的指令形式"))
			return
		}
		fmt.Println(lotsName)
		Picurl := ctx.State["image_url"].([]string)[0]
		err := file.DownloadTo(Picurl, datapath+"/"+lotsName+".jpg")
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		lotsList[lotsName] = "gif"
		ctx.SendChain(message.Reply(id), message.Text("成功"))
	})
}

func getList() (list map[string]string, err error) {
	list = make(map[string]string, 100)
	files, err := os.ReadDir(datapath)
	if err != nil {
		return
	}
	if len(files) == 0 {
		err = errors.New("不存在任何抽签")
		return
	}
	for _, name := range files {
		if name.IsDir() {
			list[name.Name()] = "file"
			continue
		}
		before, after, _ := strings.Cut(name.Name(), ".")
		if before == "" {
			continue
		}
		list[before] = after
	}
	return
}

func randFile(path string, indexMax int) (string, error) {
	files, err := os.ReadDir(datapath + "/" + path)
	if err != nil {
		return "", err
	}
	if len(files) > 1 {
		music := files[rand.Intn(len(files))]
		// 如果是文件夹就递归
		if music.IsDir() {
			indexMax--
			if indexMax <= 0 {
				return "", errors.New("该文件夹存在太多非图片文件,请清理")
			}
			return randFile(path, indexMax)
		} else {
			return path + "/" + music.Name(), err
		}
	}
	return "", errors.New("该抽签不存在")
}

func randGif(gifName string) (image.Image, error) {
	var err error
	var face []*image.NRGBA
	name := datapath + gifName + ".gif"
	face, err = img.LoadAllFrames(name, 500, 500)
	if err != nil {
		face = make([]*image.NRGBA, 0)
		first, err := img.LoadFirstFrame(name, 500, 500)
		if err != nil {
			return nil, err
		}
		face = append(face, first.Im)
	}
	return face[rand.Intn(len(face))], err
}
