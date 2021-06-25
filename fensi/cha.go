package fensi

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io/ioutil"
	"net/http"
	"strconv"
)


func init() {
	zero.OnRegex(`^/查 (.{1,25})$`).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			uid := searchapi(keyword).Data.Result[0].Mid

			if searchapi(keyword).Data.NumResults == 0 {
				ctx.Send("名字没搜到")
				return
			}

			accinfojson := accinfo(strconv.Itoa(uid))
			hanjian := follow(strconv.Itoa(uid))
			ctx.SendChain(message.Text(
				"uid:   ", accinfojson.Data.Mid, "\n",
				"name:  ", accinfojson.Data.Name, "\n",
				"sex:   ", accinfojson.Data.Sex, "\n",
				"sign:  ", accinfojson.Data.Sign, "\n",
				"level: ", accinfojson.Data.Level, "\n",
				"birthday: ", accinfojson.Data.Birthday, "\n",
				"follow: ", gjson.Get(hanjian, "data.list.#.uname"),
			))
		})
}

func init() {
	zero.OnRegex(`^/查UID:(.{1,25})$`).
		Handle(func(ctx *zero.Ctx) {
		uid := ctx.State["regex_matched"].([]string)[1]
		accinfojson := accinfo(uid)
		hanjian := follow(uid)
		status := accinfojson.Code

		if status != 0 {
			ctx.Send("UID非法")
			return
		}

		ctx.SendChain(message.Text(
			"uid:   ", accinfojson.Data.Mid, "\n",
			"name:  ", accinfojson.Data.Name, "\n",
			"sex:   ", accinfojson.Data.Sex, "\n",
			"sign:  ", accinfojson.Data.Sign, "\n",
			"level: ", accinfojson.Data.Level, "\n",
			"birthday: ", accinfojson.Data.Birthday, "\n",
			"follow: ", gjson.Get(hanjian, "data.list.#.uname"),
			))
	})
}


func accinfo(uid string) *accInfo {

	url := "http://api.bilibili.com/x/space/acc/info?mid=" + uid
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	result := &accInfo{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}


func follow(uid string) string {

	url := "https://api.bilibili.com/x/relation/same/followings?vmid=" + uid
	method := "GET"

	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("cookie", "CURRENT_FNVAL=80; _uuid=772B88E8-3ED1-D589-29BB-F6CB5214239A06137infoc; blackside_state=1; bfe_id=6f285c892d9d3c1f8f020adad8bed553; rpdid=|(umY~Jkl|kJ0J'uYkR|)lu|); fingerprint=0ec2b1140fb30b56d7b5e415bc3b5fb1; buvid_fp=C91F5265-3DF4-4D5A-9FF3-C546370B14C0143096infoc; buvid_fp_plain=C91F5265-3DF4-4D5A-9FF3-C546370B14C0143096infoc; SESSDATA=9e0266f6%2C1639637127%2Cb0172%2A61; bili_jct=96ddbd7e22d527abdc0501339a12d4d3; DedeUserID=695737880; DedeUserID__ckMd5=0117660e75db7b01; sid=5labuhaf; PVID=1; bfe_id=1e33d9ad1cb29251013800c68af42315")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	return string(body)
}
