// Package coc coc插件
package coc

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/rendercard"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnPrefixGroup([]string{".loadcoc", "。loadcoc", ".LOADCOC"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		sampleFile := engine.DataFolder() + "面版填写示例.json"
		infoFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + DefaultJSONFile
		fileName := strings.TrimSpace(ctx.State["args"].(string))
		if fileName == "" {
			sourceFileStat, err := os.Stat(sampleFile)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}

			if !sourceFileStat.Mode().IsRegular() {
				ctx.SendChain(message.Text("[ERROR]:", sampleFile, " is not a regular file"))
				return
			}

			source, err := os.Open(sampleFile)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			defer source.Close()

			destination, err := os.Create(infoFile)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}

			defer destination.Close()
			_, err = io.Copy(destination, source)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Text("设置面板完成"))
			return
		}
		// 判断群文件是否存在
		fileSearchName, fileURL := getFileURLbyFileName(ctx, fileName)
		if fileSearchName == "" {
			ctx.SendChain(message.Text("请确认群文件文件名称是否正确或存在"))
			return
		}
		// 下载文件
		ctx.SendChain(message.Text("在群文件中找到了歌曲,信息如下:\n", fileSearchName, "\n确认正确后回复“是/否”进行设置"))
		next := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(是|否)`), ctx.CheckSession())
		recv, cancel := next.Repeat()
		defer cancel()
		wait := time.NewTimer(120 * time.Second)
		answer := ""
		for {
			select {
			case <-wait.C:
				wait.Stop()
				ctx.SendChain(message.Text("等待超时，取消设置"))
				return
			case c := <-recv:
				wait.Stop()
				answer = c.Event.Message.String()
			}
			if answer == "否" {
				ctx.SendChain(message.Text("设置已经取消"))
				return
			}
			if answer != "" {
				break
			}
		}
		err := file.DownloadTo(fileURL, infoFile)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Text("成功！"))
	})
	engine.OnPrefixGroup([]string{".coc", "。coc", ".COC"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		infoFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + DefaultJSONFile
		if file.IsNotExist(infoFile) {
			ctx.SendChain(message.Text("你群还没有布置coc,请相关人员后台布局coc.(详情看用法)"))
			return
		}
		var (
			cocInfo cocJSON
			err     error
		)
		if file.IsNotExist(engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".json") {
			cocInfo, err = loadPanel(gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			cocInfo.ID = uid
			baseMsg := strings.Split(ctx.State["args"].(string), "/")
			for _, msgInfo := range baseMsg {
				msgValue := strings.Split(msgInfo, "#")
				if msgValue[0] == "描述" {
					cocInfo.Other = append(cocInfo.Other, msgValue[1])
					continue
				}
				for i, info := range cocInfo.BaseInfo {
					if msgValue[0] == info.Name {
						munberValue, err := strconv.Atoi(msgValue[1])
						if err != nil {
							cocInfo.BaseInfo[i].Value = msgValue[1]
						} else {
							cocInfo.BaseInfo[i].Value = munberValue
						}
					}
				}
			}
			for i, info := range cocInfo.Attribute {
				max := info.MaxValue - info.MinValue
				negative := -1
				if info.MinValue < 0 {
					negative = 1
				}
				value := 0
				for i := 0; i < 3; i++ {
					value += rand.Intn(6) + 1
				}
				value = max*value*5/100 + negative*info.MinValue
				cocInfo.Attribute[i].Value = value
			}
			err = savePanel(cocInfo, gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
		} else {
			cocInfo, err = loadPanel(gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
		}
		pic, err := drawImage(cocInfo)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.ImageBytes(pic))
	})
}

// 遍历群文件
func getFileURLbyFileName(ctx *zero.Ctx, fileName string) (fileSearchName, fileURL string) {
	filesOfGroup := ctx.GetThisGroupRootFiles()
	files := filesOfGroup.Get("files").Array()
	folders := filesOfGroup.Get("folders").Array()
	// 遍历当前目录的文件名
	if len(files) != 0 {
		for _, fileNameOflist := range files {
			if strings.Contains(fileNameOflist.Get("file_name").String(), fileName) {
				fileSearchName = fileNameOflist.Get("file_name").String()
				fileURL = ctx.GetThisGroupFileUrl(fileNameOflist.Get("busid").Int(), fileNameOflist.Get("file_id").String())
				return
			}
		}
	}
	// 遍历子文件夹
	if len(folders) != 0 {
		for _, folderNameOflist := range folders {
			folderID := folderNameOflist.Get("folder_id").String()
			fileSearchName, fileURL = getFileURLbyfolderID(ctx, fileName, folderID)
			if fileSearchName != "" {
				return
			}
		}
	}
	return
}
func getFileURLbyfolderID(ctx *zero.Ctx, fileName, folderid string) (fileSearchName, fileURL string) {
	filesOfGroup := ctx.GetThisGroupFilesByFolder(folderid)
	files := filesOfGroup.Get("files").Array()
	folders := filesOfGroup.Get("folders").Array()
	// 遍历当前目录的文件名
	if len(files) != 0 {
		for _, fileNameOflist := range files {
			if strings.Contains(fileNameOflist.Get("file_name").String(), fileName) {
				fileSearchName = fileNameOflist.Get("file_name").String()
				fileURL = ctx.GetThisGroupFileUrl(fileNameOflist.Get("busid").Int(), fileNameOflist.Get("file_id").String())
				return
			}
		}
	}
	// 遍历子文件夹
	if len(folders) != 0 {
		for _, folderNameOflist := range folders {
			folderID := folderNameOflist.Get("folder_id").String()
			fileSearchName, fileURL = getFileURLbyfolderID(ctx, fileName, folderID)
			if fileSearchName != "" {
				return
			}
		}
	}
	return
}

func drawImage(userInfo cocJSON) (imagePicByte []byte, err error) {
	var (
		wg          sync.WaitGroup
		userIDBlock image.Image // 编号
		infoBlock   image.Image // 基本信息
		atrrBlock   image.Image // 属性信息
		otherBlock  image.Image // 其他信息
	)
	wg.Add(1)
	// 绘制ID
	go func() {
		defer wg.Done()
		userIDBlock, err = userInfo.drawIDBlock()
		if err != nil {
			return
		}
	}()
	wg.Add(1)
	// 绘制基本信息
	go func() {
		defer wg.Done()
		infoBlock, err = userInfo.drawInfoBlock()
		if err != nil {
			return
		}
	}()
	wg.Add(1)
	// 绘制属性框
	go func() {
		defer wg.Done()
		atrrBlock, err = userInfo.drawAttrBlock()
		if err != nil {
			return
		}
	}()
	wg.Add(1)
	// 绘制技能框
	go func() {
		defer wg.Done()
		otherBlock, err = userInfo.drawOtherBlock()
		if err != nil {
			return
		}
	}()
	wg.Wait()
	if userIDBlock == nil || infoBlock == nil || atrrBlock == nil || otherBlock == nil {
		return
	}
	// 计算图片高度
	backDX := 1020
	backDY := 15 + userIDBlock.Bounds().Dy() + 5 + infoBlock.Bounds().Dy() + 10 + atrrBlock.Bounds().Dy() + 20 + otherBlock.Bounds().Dy() + 10 + 10
	canvas := gg.NewContext(backDX, backDY)

	// 画底色
	canvas.DrawRectangle(0, 0, float64(backDX), float64(backDY))
	canvas.SetRGBA255(150, 150, 150, 255)
	canvas.Fill()
	canvas.DrawRectangle(10, 10, float64(backDX-20), float64(backDY-20))
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()

	// 放置头像
	getAvatar, err := web.GetData("http://q4.qlogo.cn/g?b=qq&nk=" + strconv.FormatInt(userInfo.ID, 10) + "&s=640")
	if err != nil {
		return
	}
	avatar, _, err := image.Decode(bytes.NewReader(getAvatar))
	if err != nil {
		return
	}
	avatarf := imgfactory.Size(avatar, 500, 500).Image()
	canvas.DrawImage(avatarf, 20, 30)
	// 头像框
	canvas.DrawRectangle(20, 30, 500, 500)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()

	// 编号
	tempDY := 15
	canvas.DrawImageAnchored(userIDBlock, backDX-10-20-userIDBlock.Bounds().Dx()/2, tempDY+userIDBlock.Bounds().Dy()/2, 0.5, 0.5)
	// 放入基本信息
	tempDY += +userIDBlock.Bounds().Dy() + 5
	canvas.DrawImage(infoBlock, 20+500+10, tempDY)
	// 放入属性信息
	tempDY += infoBlock.Bounds().Dy() + 10
	canvas.DrawImage(atrrBlock, 10, tempDY)
	// 放入其他信息
	tempDY += atrrBlock.Bounds().Dy() + 20
	canvas.DrawImage(otherBlock, 20, tempDY)

	return imgfactory.ToBytes(canvas.Image())
}

// 绘制ID区域
func (userInfo *cocJSON) drawIDBlock() (image.Image, error) {
	fontSize := 30.0
	userIDstr := "编号:" + strconv.FormatInt(userInfo.ID, 10)
	canvas := gg.NewContext(1, 1)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	textW, textH := canvas.MeasureString(userIDstr)
	canvas = gg.NewContext(int(textW*1.2), int(textH)*2+15)
	// 放字
	canvas.SetColor(color.Black)
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	canvas.DrawStringAnchored(userIDstr, textW*0.6, textH, 0.5, 0.5)
	// 画下划线
	canvas.DrawLine(0, textH*2, textW*1.2, textH*2)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	return canvas.Image(), nil
}

// 绘制基本信息区域
func (userInfo *cocJSON) drawInfoBlock() (image.Image, error) {
	fontSize := 50.0
	raw := len(userInfo.BaseInfo)
	maxStr := ""
	for _, info := range userInfo.BaseInfo {
		str := fmt.Sprintf("%v : %v", info.Name, info.Value)
		if len(maxStr) < len(str) {
			maxStr = str
		}
	}
	canvas := gg.NewContext(1, 1)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	textW, textH := canvas.MeasureString(maxStr)
	if textW > 500-20 {
		for ; textW >= 500-20; fontSize-- {
			if err = canvas.ParseFontFace(data, fontSize); err != nil {
				return nil, err
			}
			textW, textH = canvas.MeasureString(maxStr)
		}
	}
	backDY := 15 + int(textH*2)*raw
	if backDY < 500 {
		backDY = 500
	}
	canvas = gg.NewContext(500-40, backDY)
	// 画底色
	canvas.DrawRectangle(0, 0, 500-40, float64(backDY))
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 放字
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	for i, info := range userInfo.BaseInfo {
		str := fmt.Sprintf("%v : %v", info.Name, info.Value)
		textW, _ := canvas.MeasureString(str)
		textDY := 10 + textH*2*float64(i)
		canvas.SetColor(color.Black)
		canvas.DrawStringAnchored(str, 25+textW/2, textDY+textH/2, 0.5, 0.5)
		textDY += textH * 1.5
		// 画下划线
		canvas.DrawLine(25, textDY, 500-20-25, textDY)
		canvas.SetLineWidth(3)
		canvas.SetRGBA255(0, 0, 0, 255)
		canvas.Stroke()
	}
	return canvas.Image(), nil
}

// 绘制属性信息区域
func (userInfo *cocJSON) drawAttrBlock() (image.Image, error) {
	fontSize := 50.0
	raw := len(userInfo.Attribute)
	canvas := gg.NewContext(1, 1)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	_, textH := canvas.MeasureString("高度")
	nameW := 0.0
	valueW := 0.0
	for _, info := range userInfo.Attribute {
		textW, _ := canvas.MeasureString(info.Name)
		if nameW < textW {
			nameW = textW
		}
	}
	for _, info := range userInfo.Attribute {
		textW, _ := canvas.MeasureString(strconv.Itoa(info.MaxValue))
		if valueW < textW {
			valueW = textW
		}
	}
	barW := 500 - nameW - 10 - 10 - valueW - 10
	backX := 980
	backY := 20 + int(textH*2)*raw/2
	canvas = gg.NewContext(backX, backY)
	// 画底色
	canvas.DrawRectangle(0, 0, float64(backX), float64(backY))
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	/*/ 边框框
	canvas.DrawRectangle(0, 0, float64(backX), float64(raw))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	// 放字*/
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	r := -1.0
	for i, info := range userInfo.Attribute {
		infoNameW, _ := canvas.MeasureString(info.Name)
		valueStr := strconv.Itoa(info.Value)
		infoValueW, _ := canvas.MeasureString(valueStr)
		textX := 0.0
		if i%2 == 1 {
			textX = 510
		} else {
			r++
		}
		textY := 10.0 + textH*2*r
		canvas.SetColor(color.Black)
		// 名称
		canvas.DrawStringAnchored(info.Name, textX+nameW-infoNameW/2, textY+textH/2, 0.5, 0.5)
		// 数值
		canvas.DrawStringAnchored(valueStr, textX+nameW+10+barW+10+infoValueW/2, textY+textH/2, 0.5, 0.5)
		// 画属性条
		canvas.DrawRectangle(textX+nameW+10, textY, barW, textH*1.2)
		canvas.SetRGB255(150, 150, 150)
		canvas.Fill()
		canvas.SetRGB255(0, 0, 0)
		canvas.DrawRectangle(textX+nameW+10, textY, barW*((float64(info.Value)-float64(info.MinValue))/float64(info.MaxValue)-float64(info.MinValue)), textH*1.2)
		canvas.SetRGB255(102, 102, 102)
		canvas.Fill()
	}
	return canvas.Image(), nil
}

// 绘制其他信息区域
func (userInfo *cocJSON) drawOtherBlock() (pic image.Image, err error) {
	fontSize := 25.0
	glowsd, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	newplugininfo, err := rendercard.Truncate(glowsd, userInfo.Other, 970, 50)
	if err != nil {
		return
	}
	if len(newplugininfo) == 0 || (len(newplugininfo) < 2 && newplugininfo[0] == "") {
		newplugininfo = []string{"暂无其他资料"}
	}
	canvas := gg.NewContext(1, 1)
	if err = canvas.ParseFontFace(glowsd, fontSize); err != nil {
		return
	}
	_, textH := canvas.MeasureString("高度")
	raw := 50 + int(textH*2)*len(newplugininfo)
	if raw < 200 {
		raw = 200
	}
	canvas = gg.NewContext(980, raw)
	// 画底色
	canvas.DrawRectangle(0, 0, 980, float64(raw))
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 边框框
	canvas.DrawRectangle(0, 0, 980, float64(raw))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	// 放字
	if err = canvas.ParseFontFace(glowsd, fontSize); err != nil {
		return
	}
	canvas.SetColor(color.Black)
	for i, info := range newplugininfo {
		textW, _ := canvas.MeasureString(info)
		textDY := 25.0 + textH*2*float64(i)
		canvas.DrawStringAnchored(info, 15+textW/2, textDY+textH/2, 0.5, 0.5)
	}
	pic = canvas.Image()
	return
}
