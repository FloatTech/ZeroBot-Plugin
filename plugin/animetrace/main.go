// Package animetrace AnimeTrace 动画/Galgame识别
package animetrace

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"
	"strings"

	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/disintegration/imaging"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("animetrace", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "AnimeTrace 动画/Galgame识别插件",
		Help:             "- Gal识图\n- 动漫识图\n- 动漫识图 2\n- 动漫识图 [模型名]\n- Gal识图 [模型名]",
	})

	engine.OnPrefix("gal识图", zero.OnlyGroup, zero.MustProvidePicture).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := ctx.State["args"].(string)
		var model string
		switch strings.TrimSpace(args) {
		case "":
			model = "full_game_model_kira" // 默认使用的模型
		default:
			model = args // 自定义设置模型
		}
		processImageRecognition(ctx, model)
	})

	engine.OnPrefix("动漫识图", zero.OnlyGroup, zero.MustProvidePicture).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := ctx.State["args"].(string)
		var model string
		switch strings.TrimSpace(args) {
		case "":
			model = "anime_model_lovelive"
		case "2":
			model = "pre_stable"
		default:
			model = args
		}
		processImageRecognition(ctx, model)
	})
}

// 处理图片识别
func processImageRecognition(ctx *zero.Ctx, model string) {
	urls := ctx.State["image_url"].([]string)
	if len(urls) == 0 {
		return
	}
	imageData, err := imgfactory.Load(urls[0])
	if err != nil {
		ctx.Send(message.Text("下载图片失败: ", err))
		return
	}
	// ctx.Send(message.Text(model))
	respBody, err := createAndSendMultipartRequest("https://api.animetrace.com/v1/search", imageData, map[string]string{
		"is_multi":  "0",
		"model":     model,
		"ai_detect": "0",
	})
	if err != nil {
		ctx.Send(message.Text("识别请求失败: ", err))
		return
	}
	code := gjson.Get(string(respBody), "code").Int()
	if code != 0 {
		ctx.Send(message.Text("错误: ", gjson.Get(string(respBody), "zh_message").String()))
		return
	}
	dataArray := gjson.Get(string(respBody), "data").Array()
	if len(dataArray) == 0 {
		ctx.Send(message.Text("未识别到任何角色"))
		return
	}
	var sk message.Message
	sk = append(sk, ctxext.FakeSenderForwardNode(ctx, message.Text("共识别到 ", len(dataArray), " 个角色，可能是以下来源")))
	for _, value := range dataArray {
		boxArray := value.Get("box").Array()
		imgWidth, imgHeight := imageData.Bounds().Dx(), imageData.Bounds().Dy() // 你可以从 `imageData.Bounds()` 获取
		box := []int{
			int(boxArray[0].Float() * float64(imgWidth)),
			int(boxArray[1].Float() * float64(imgHeight)),
			int(boxArray[2].Float() * float64(imgWidth)),
			int(boxArray[3].Float() * float64(imgHeight)),
		}
		croppedImg := imaging.Crop(imageData, image.Rect(box[0], box[1], box[2], box[3]))
		var buf bytes.Buffer
		if err := imaging.Encode(&buf, croppedImg, imaging.JPEG, imaging.JPEGQuality(80)); err != nil {
			ctx.Send(message.Text("图片编码失败: ", err))
			continue
		}

		base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
		var sb strings.Builder
		value.Get("character").ForEach(func(_, character gjson.Result) bool {
			sb.WriteString(fmt.Sprintf("《%s》的角色 %s\n", character.Get("work").String(), character.Get("character").String()))
			return true
		})
		sk = append(sk, ctxext.FakeSenderForwardNode(ctx, message.Image("base64://"+base64Str), message.Text(sb.String())))
	}
	ctx.SendGroupForwardMessage(ctx.Event.GroupID, sk)
}

// 发送图片识别请求
func createAndSendMultipartRequest(url string, img image.Image, formFields map[string]string) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 直接编码图片
	part, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		return nil, errors.New("创建文件字段失败: " + err.Error())
	}
	if err := jpeg.Encode(part, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, errors.New("图片编码失败: " + err.Error())
	}

	// 写入其他字段
	for key, value := range formFields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, errors.New("写入表单字段失败 (" + key + "): " + err.Error())
		}
	}

	if err := writer.Close(); err != nil {
		return nil, errors.New("关闭 multipart writer 失败: " + err.Error())
	}

	return web.PostData(url, writer.FormDataContentType(), body)
}
