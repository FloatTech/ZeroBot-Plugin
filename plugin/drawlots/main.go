// Package drawlots 多功能抽签插件
package drawlots

import (
	"errors"
	"image"
	"image/gif"
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
		Brief:             "多功能抽签",
		Help:              "支持图包文件夹和gif抽签\n-------------\n- (刷新)抽签列表\n- 抽签-[签名]\n- 查看抽签[gif签名]\n- 添加抽签[签名][gif图片]\n- 删除抽签[gif签名]",
		PrivateDataFolder: "drawlots",
	}).ApplySingle(ctxext.DefaultSingle)
	datapath = file.BOTPATH + "/" + en.DataFolder()
)

func init() {
	var err error
	go func() {
		lotsList, err = getList()
		if err != nil {
			logrus.Infoln("[drawlots]加载失败:", err)
		} else {
			logrus.Infoln("[drawlots]加载", len(lotsList), "个抽签")
		}
	}()
	en.OnFullMatchGroup([]string{"抽签列表", "刷新抽签列表"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		lotsList, err = getList() // 刷新列表
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		messageText := make([]string, 0, len(lotsList))
		messageText = append(messageText, []string{
			" 签 名 [ 类 型 ] ", "—————————————————",
		}...)
		for name, fileType := range lotsList {
			messageText = append(messageText, []string{
				name + "[" + fileType + "]", "----------",
			}...)
		}
		textPic, err := text.RenderToBase64(strings.Join(messageText, "\n"), text.BoldFontFile, 400, 50)
		if err != nil {
			return
		}
		ctx.SendChain(message.Image("base64://" + helper.BytesToString(textPic)))
	})
	en.OnPrefix("抽签-").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		lotsType := strings.TrimSpace(ctx.State["args"].(string))
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
		lotsImg, err := randGif(lotsType + "." + fileType)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		// 生成图片
		data, cl := writer.ToBytes(lotsImg)
		defer cl()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.ImageBytes(data))
	})
	en.OnPrefix("查看抽签", zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ID := ctx.Event.MessageID
		lotsName := strings.TrimSpace(ctx.State["args"].(string))
		fileType, ok := lotsList[lotsName]
		if !ok {
			ctx.SendChain(message.Reply(ID), message.Text("该签不存在,请确认是否存在"))
			return
		}
		if fileType == "file" {
			ctx.SendChain(message.Reply(ID), message.Text("仅支持查看gif抽签"))
			return
		}
		ctx.SendChain(message.Reply(ID), message.Image("file:///"+datapath+lotsName+"."+fileType))
	})
	en.OnPrefix("添加抽签", func(ctx *zero.Ctx) bool {
		if zero.SuperUserPermission(ctx) {
			return zero.MustProvidePicture(ctx)
		}
		return false
	}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ID := ctx.Event.MessageID
		lotsName := strings.TrimSpace(ctx.State["args"].(string))
		if lotsName == "" {
			ctx.SendChain(message.Reply(ID), message.Text("请使用正确的指令形式"))
			return
		}
		Picurl := ctx.State["image_url"].([]string)[0]
		err := file.DownloadTo(Picurl, datapath+"/"+lotsName+".gif")
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		lotsList[lotsName] = "gif"
		ctx.SendChain(message.Reply(ID), message.Text("成功"))
	})
	en.OnPrefix("删除抽签", zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ID := ctx.Event.MessageID
		lotsName := strings.TrimSpace(ctx.State["args"].(string))
		fileType, ok := lotsList[lotsName]
		if !ok {
			ctx.SendChain(message.Reply(ID), message.Text("该签不存在,请确认是否存在"))
			return
		}
		if fileType == "file" {
			ctx.SendChain(message.Reply(ID), message.Text("图包请手动移除(保护图源误删),谢谢"))
			return
		}
		err := os.Remove(datapath + lotsName + "." + fileType)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		delete(lotsList, lotsName)
		ctx.SendChain(message.Reply(ID), message.Text("成功"))
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
	for _, lots := range files {
		if lots.IsDir() {
			list[lots.Name()] = "file"
			continue
		}
		before, after, ok := strings.Cut(lots.Name(), ".")
		if !ok || before == "" {
			continue
		}
		list[before] = after
	}
	return
}

func randFile(path string, indexMax int) (string, error) {
	picPath := datapath + path
	files, err := os.ReadDir(picPath)
	if err != nil {
		return "", err
	}
	if len(files) > 0 {
		music := files[rand.Intn(len(files))]
		// 如果是文件夹就递归
		if music.IsDir() {
			indexMax--
			if indexMax <= 0 {
				return "", errors.New("该文件夹存在太多非图片文件,请清理")
			}
			return randFile(path, indexMax)
		}
		return picPath + "/" + music.Name(), err
	}
	return "", errors.New("该抽签不存在签内容")
}

func randGif(gifName string) (image.Image, error) {
	name := datapath + gifName
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	im, err := gif.DecodeAll(file)
	_ = file.Close()
	if err != nil {
		return nil, err
	}
	/*
		firstImg, err := img.Load(name)
		if err != nil {
			return nil, err
		}
		ims := make([]*image.NRGBA, len(im.Image))
		for i, v := range im.Image {
			ims[i] = img.Size(firstImg, firstImg.Bounds().Max.X, firstImg.Bounds().Max.Y).InsertUpC(v, 0, 0, firstImg.Bounds().Max.X/2, firstImg.Bounds().Max.Y/2).Clone().Im
		}
		/*/
	// 如果gif图片出现信息缺失请使用上面注释掉的代码，把下面注释了(上面的存在bug)
	ims := make([]*image.NRGBA, len(im.Image))
	for i, v := range im.Image {
		ims[i] = img.Size(v, v.Bounds().Max.X, v.Bounds().Max.Y).Im
	} // */
	return ims[rand.Intn(len(ims))], err
}
