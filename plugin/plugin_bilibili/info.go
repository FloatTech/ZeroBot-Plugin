// Package bilibili 查询b站用户信息
package bilibili

import (
	"io/ioutil"
	"net/http"

	control "github.com/FloatTech/zbputils/control"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

var engine = control.Register("bilibili", order.AcquirePrio(), &control.Options{
	DisableOnDefault: false,
	Help: "bilibili\n" +
		"- >vup info [名字 | uid]\n" +
		"- >user info [名字 | uid]",
})

// 查成分的
func init() {
	engine.OnRegex(`^>(?:user|vup)\s?info\s?(.{1,25})$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			rest, err := uid(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			id := rest.Get("data.result.0.mid").String()
			url := "https://api.bilibili.com/x/relation/same/followings?vmid=" + id
			method := "GET"
			client := &http.Client{}
			req, err := http.NewRequest(method, url, nil)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			req.Header.Add("cookie", "CURRENT_FNVAL=80; _uuid=772B88E8-3ED1-D589-29BB-F6CB5214239A06137infoc; blackside_state=1; bfe_id=6f285c892d9d3c1f8f020adad8bed553; rpdid=|(umY~Jkl|kJ0J'uYkR|)lu|); fingerprint=0ec2b1140fb30b56d7b5e415bc3b5fb1; buvid_fp=C91F5265-3DF4-4D5A-9FF3-C546370B14C0143096infoc; buvid_fp_plain=C91F5265-3DF4-4D5A-9FF3-C546370B14C0143096infoc; SESSDATA=9e0266f6%2C1639637127%2Cb0172%2A61; bili_jct=96ddbd7e22d527abdc0501339a12d4d3; DedeUserID=695737880; DedeUserID__ckMd5=0117660e75db7b01; sid=5labuhaf; PVID=1; bfe_id=1e33d9ad1cb29251013800c68af42315")
			res, err := client.Do(req)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			data := string(body)
			ctx.SendChain(message.Text(
				"uid: ", rest.Get("data.result.0.mid").Int(), "\n",
				"name: ", rest.Get("data.result.0.uname").Str, "\n",
				"sex: ", []string{"", "", "女", "男"}[rest.Get("data.result.0.gender").Int()], "\n",
				"sign: ", rest.Get("data.result.0.usign").Str, "\n",
				"level: ", rest.Get("data.result.0.level").Int(), "\n",
				"follow: ", gjson.Get(data, "data.list.#.uname"),
			))
		})
}
