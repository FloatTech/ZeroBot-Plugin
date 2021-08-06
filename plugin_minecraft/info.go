package minecraft

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type resultjson struct {
	IP    string `json:"ip"`
	Port  int    `json:"port"`
	Debug struct {
		Ping          bool `json:"ping"`
		Query         bool `json:"query"`
		Srv           bool `json:"srv"`
		Querymismatch bool `json:"querymismatch"`
		Ipinsrv       bool `json:"ipinsrv"`
		Cnameinsrv    bool `json:"cnameinsrv"`
		Animatedmotd  bool `json:"animatedmotd"`
		Cachetime     int  `json:"cachetime"`
		Apiversion    int  `json:"apiversion"`
	} `json:"debug"`
	Motd struct {
		Raw   []string `json:"raw"`
		Clean []string `json:"clean"`
		HTML  []string `json:"html"`
	} `json:"motd"`
	Players struct {
		Online int      `json:"online"`
		Max    int      `json:"max"`
		List   []string `json:"list"`
	} `json:"players"`
}

func init() {
	zero.OnRegex(`^/list (.*)$`).
		Handle(func(ctx *zero.Ctx) {
			// 支持多个服务器
			switch ctx.State["regex_matched"].([]string)[1] {
			case "ftbi": // 这里对应触发指令里的服务器名称
				ftbijson := infoapi("115.28.186.22:25710") //这里填对应mc服务器的登录地址
				var str = ftbijson.Players.List
				cs := strings.Join(str, "\n")
				ctx.SendChain(message.Text(
					"服务器名字: ", ftbijson.Motd.Raw[0], "\n",
					"在线人数: ", ftbijson.Players.Online, "/", ftbijson.Players.Max, "\n",
					"以下为玩家名字: ", "\n", cs,
				))
			case "ges": // 这里对应触发指令里的服务器名称
				gesjson := infoapi("115.28.186.22:25701") //这里填对应mc服务器的登录地址
				var str = gesjson.Players.List
				cs := strings.Join(str, "\n")
				ctx.SendChain(message.Text(
					"服务器名字: ", gesjson.Motd.Raw[0], "\n",
					"在线人数: ", gesjson.Players.Online, "/", gesjson.Players.Max, "\n",
					"以下为玩家名字: ", "\n", cs,
				))
			}
		})
}

// 开放api请求调用
func infoapi(addr string) *resultjson {
	url := "https://api.mcsrvstat.us/2/" + addr
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	result := &resultjson{}
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}
