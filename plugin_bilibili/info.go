package plugin_bilibili

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	zero.OnRegex(`^>user info\s(.{1,25})$`).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			res, err := uid(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			id := res.Get("data.result.0.mid").Int()
			url := fmt.Sprintf("https://api.bilibili.com/x/relation/same/followings?vmid=%d", id)
			client := &http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			req.Header.Add("cookie", "CURRENT_FNVAL=80; _uuid=772B88E8-3ED1-D589-29BB-F6CB5214239A06137infoc; blackside_state=1; bfe_id=6f285c892d9d3c1f8f020adad8bed553; rpdid=|(umY~Jkl|kJ0J'uYkR|)lu|); fingerprint=0ec2b1140fb30b56d7b5e415bc3b5fb1; buvid_fp=C91F5265-3DF4-4D5A-9FF3-C546370B14C0143096infoc; buvid_fp_plain=C91F5265-3DF4-4D5A-9FF3-C546370B14C0143096infoc; SESSDATA=9e0266f6%2C1639637127%2Cb0172%2A61; bili_jct=96ddbd7e22d527abdc0501339a12d4d3; DedeUserID=695737880; DedeUserID__ckMd5=0117660e75db7b01; sid=5labuhaf; PVID=1; bfe_id=1e33d9ad1cb29251013800c68af42315")
			resp, err := client.Do(req)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				ctx.SendChain(message.Text("ERROR: code ", resp.StatusCode))
				return
			}
			data, _ := ioutil.ReadAll(resp.Body)
			json := gjson.ParseBytes(data)
			ctx.SendChain(message.Text(
				"uid: ", res.Get("data.result.0.mid").Int(), "\n",
				"name: ", res.Get("data.result.0.uname").Str, "\n",
				"sex: ", []string{"", "", "女", "男"}[res.Get("data.result.0.gender").Int()], "\n",
				"sign: ", res.Get("data.result.0.usign").Str, "\n",
				"level: ", res.Get("data.result.0.level").Int(), "\n",
				"follow: ", json.Get("data.list.#.uname"),
			))
		})
}
