// Package aiimage AI画图
package aiimage

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/web"
	sql "github.com/FloatTech/sqlite"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

func init() {
	var sdb = &storage{}

	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Extra:            control.ExtraFromString("aiimage"),
		Brief:            "AI画图",
		Help: "- 设置AI画图密钥xxx\n" +
			"- 设置AI画图接口地址https://api.siliconflow.cn/v1/images/generations\n" +
			"- 设置AI画图模型名Kwai-Kolors/Kolors\n" +
			"- 查看AI画图配置\n" +
			"- AI画图 [描述]",
		PrivateDataFolder: "aiimage",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		sdb.db = sql.New(en.DataFolder() + "aiimage.db")
		err := sdb.db.Open(time.Hour)
		if err == nil {
			// 创建配置表
			err = sdb.db.Create("config", &imageConfig{})
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return false
			}
			return true
		}
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	})

	en.OnPrefix("设置AI画图密钥", getdb, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			apiKey := strings.TrimSpace(ctx.State["args"].(string))
			cfg := sdb.getConfig()
			err := sdb.setConfig(apiKey, cfg.APIURL, cfg.ModelName)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: 设置API密钥失败: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置API密钥"))
		})

	en.OnPrefix("设置AI画图接口地址", getdb, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			apiURL := strings.TrimSpace(ctx.State["args"].(string))
			cfg := sdb.getConfig()
			err := sdb.setConfig(cfg.APIKey, apiURL, cfg.ModelName)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: 设置API地址失败: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置API地址"))
		})

	en.OnPrefix("设置AI画图模型名", getdb, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			modelName := strings.TrimSpace(ctx.State["args"].(string))
			cfg := sdb.getConfig()
			err := sdb.setConfig(cfg.APIKey, cfg.APIURL, modelName)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: 设置模型失败: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置模型: ", modelName))
		})

	en.OnFullMatch("查看AI画图配置", getdb, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(sdb.PrintConfig()))
		})

	en.OnPrefix("AI画图", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女思考中..."))
			prompt := strings.TrimSpace(ctx.State["args"].(string))
			if prompt == "" {
				ctx.SendChain(message.Text("请输入图片描述"))
				return
			}

			cfg := sdb.getConfig()
			if cfg.APIKey == "" || cfg.APIURL == "" || cfg.ModelName == "" {
				ctx.SendChain(message.Text("请先配置API密钥、地址和模型"))
				return
			}

			// 准备请求数据
			reqBytes, _ := json.Marshal(map[string]interface{}{
				"model":               cfg.ModelName,
				"prompt":              prompt,
				"image_size":          "1024x1024",
				"batch_size":          4,
				"num_inference_steps": 20,
				"guidance_scale":      7.5,
			})

			// 发送API请求
			data, err := web.RequestDataWithHeaders(
				web.NewDefaultClient(),
				cfg.APIURL,
				"POST",
				func(req *http.Request) error {
					req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
					req.Header.Set("Content-Type", "application/json")
					return nil
				},
				bytes.NewReader(reqBytes),
			)
			if err != nil {
				ctx.SendChain(message.Text("API请求失败: ", err))
				return
			}

			// 解析API响应
			jsonData := gjson.ParseBytes(data)
			images := jsonData.Get("images")
			if !images.Exists() {
				images = jsonData.Get("data")
				if !images.Exists() {
					ctx.SendChain(message.Text("未获取到图片URL"))
					return
				}
			}

			// 发送生成的图片和相关信息
			inferenceTime := jsonData.Get("timings.inference").Float()
			seed := jsonData.Get("seed").Int()
			msg := make(message.Message, 0, 1)
			msg = append(msg, ctxext.FakeSenderForwardNode(ctx, message.Text("图片生成成功!\n",
				"提示词: ", prompt, "\n",
				"模型: ", cfg.ModelName, "\n",
				"推理时间: ", inferenceTime, "秒\n",
				"种子: ", seed)))

			// 添加所有图片
			images.ForEach(func(_, value gjson.Result) bool {
				url := value.Get("url").String()
				if url != "" {
					msg = append(msg, ctxext.FakeSenderForwardNode(ctx, message.Image(url)))
				}
				return true
			})

			if len(msg) > 0 {
				ctx.Send(msg)
			}
		})
}
