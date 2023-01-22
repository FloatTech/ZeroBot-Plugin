// Package jiami 兽语加密与解密
package jiami

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	jiami1 = "http://ovooa.com/API/sho_u/?msg=%v"          // 加密api地址
	jiami2 = "http://ovooa.com/API/sho_u/?format=1&msg=%v" // 解密api地址

)

type nmd struct { // struct解析格式大概是
	Data struct {
		Message string
	} `json:"data"`
}

func init() { // 主函数
	en := control.Register("jiami", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "兽语|喵喵加解密",
		Help: "兽语|喵喵加解密\n" +
			"- 兽语|喵喵加密xxx\n- 兽语|喵喵解密xxx",
	})
	en.OnRegex(`^(兽语|喵喵)加密\s*(.+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		str := ctx.State["regex_matched"].([]string)[2]
		es, err := web.GetData(fmt.Sprintf(jiami1, str)) // 将网站返回结果赋值
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		var r nmd                    // r数组
		err = json.Unmarshal(es, &r) // 填api返回结果，struct地址
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		if ctx.State["regex_matched"].([]string)[1] == "喵喵" {
			r.jsontos(">")
		}
		ctx.SendChain(message.Text(r.Data.Message)) // 输出提取后的结果
	})

	en.OnRegex(`^(兽语|喵喵)解密\s*(.+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		str := ctx.State["regex_matched"].([]string)[2]
		if ctx.State["regex_matched"].([]string)[1] == "喵喵" {
			str = replace(str, "<")
		}
		es, err := web.GetData(fmt.Sprintf(jiami2, str)) // 将网站返回结果赋值
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		var n nmd                    // r数组
		err = json.Unmarshal(es, &n) // 填api返回结果，struct地址
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		ctx.SendChain(message.Text(n.Data.Message)) // 输出提取后的结果
	})
}

// 匹配转码解码
func (r *nmd) jsontos(k string) {
	if k == ">" || k == "<" {
		r.Data.Message = replace(r.Data.Message, k)
	}
}

// 转字符
func replace(s, k string) string {
	switch k {
	case ">":
		s = strings.ReplaceAll(s, "嗷", "喵")
		s = strings.ReplaceAll(s, "啊", "苗")
		s = strings.ReplaceAll(s, "呜", "瞄")
	case "<":
		s = strings.ReplaceAll(s, "喵", "嗷")
		s = strings.ReplaceAll(s, "苗", "啊")
		s = strings.ReplaceAll(s, "瞄", "呜")
	}
	return s
}
