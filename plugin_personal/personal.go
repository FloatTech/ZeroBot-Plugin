package personal

import (
	"encoding/json"
	"fmt"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io/ioutil"
	"net/http"
)

func init() {

	zero.OnFullMatch("经典台词", zero.OnlyToMe,zero.SuperUserPermission).SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			k:=getDialogue()
			arr:= []int64{1197716421,1652997133,657595613,1481003793}
			for _,v:= range arr{
				ctx.SendPrivateMessage(v,message.Text(k.Newslist[0].Dialogue+"                      ——《"+k.Newslist[0].Source+"》"))
			}

		})
}

type TData struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Newslist []NData `json:"newslist"`
}
type NData struct {
	Dialogue string `json:"dialogue"`
	English string `json:"english"`
	Source string `json:"source"`
	Type int `json:"type"`
}
func getDialogue() TData{
	resp,_ := http.Get("http://api.tianapi.com/txapi/dialogue/index?key=4f56c7838d5b064147e71893d14ddeb1")
	bytes,_ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(bytes))

	var r TData
	err := json.Unmarshal(bytes, &r)
	if err != nil {
		fmt.Printf("err was %v", err)
	}
	return r


}