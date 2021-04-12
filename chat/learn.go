package chat

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

func init()  {
	botName := "小沫" //填写机器人的名称
	myData := map[string][]string{}
	yaml.Unmarshal(FileRead("chat\\myData.yaml"),&myData)

	zero.OnRegex("^(.+?)$").Handle(func(ctx *zero.Ctx) {

		text := ctx.State["regex_matched"].([]string)[1]
		if strings.Index(text,botName+"跟我学")==0{
			kv := strings.Split(text[len(botName+"跟我学 "):len(text)]," ")
			if len(kv)>1 && kv[1]!=""{
				myData[kv[0]] = kv[1:len(kv)]
				content,_ := yaml.Marshal(myData)
				FileWrite("chat\\myData.yaml",content)
				ctx.Send(botName + "学会了奇怪的新知识["+kv[0]+"]：\n- "+strings.Join(kv[1:len(kv)],"\n- "))
			}else {
				ctx.Send("你想让"+botName+"学些什么呀？")
			}
			return
		}

		for k,vs := range myData{
			if strings.Index(text,k) != -1{
				rand.Seed(time.Now().Unix())
				ctx.Send(vs[rand.Intn(len(vs))])
				return
			}
		}
	})
}

func FileRead(path string) []byte {
	//读取文件数据
	file,_ := os.Open(path)
	defer file.Close()
	data,_ := ioutil.ReadAll(file)
	return data
}

func FileWrite(path string,content []byte) int {
	//写入文件数据
	ioutil.WriteFile(path,content,0644)
	return len(content)
}
