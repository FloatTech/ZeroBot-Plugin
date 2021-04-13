package chat

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var poke = rate.NewManager(time.Minute*5, 8) // 戳一戳

func init() { // 插件主体
	myData := map[string][]string{}
	yaml.Unmarshal(FileRead("chat\\myData.yaml"),&myData)

	zero.OnRegex("^(.+?)$").Handle(func(ctx *zero.Ctx) {
		var nickname string = zero.BotConfig.NickName[0]
		text := ctx.State["regex_matched"].([]string)[1]

		if strings.Index(text,nickname+"跟我学 ")==0{
			kv := strings.Split(text[len(nickname+"跟我学 "):len(text)]," ")
			if len(kv)>0 && kv[1]!=""{
				myData[kv[0]] = kv[1:len(kv)]
				content,_ := yaml.Marshal(myData)
				FileWrite("chat\\myData.yaml",content)
				ctx.Send(nickname + "学会了有关["+kv[0]+"]的新知识呢：\n- "+strings.Join(kv[1:len(kv)],"\n- "))
			}else {
				ctx.Send("你想让"+nickname+"学些什么呀？")
			}
			return 
		}

		if strings.Index(text,nickname+"请忘掉 ")==0{
			keys := strings.Split(text[len(nickname+"请忘掉 "):len(text)]," ")
			if len(keys)>0 && keys[0]!=""{
				for _, key := range keys {
					if _,ok := myData[key];ok{
						delete(myData,key)
						ctx.Send("["+key+"]是什么呀？"+nickname+"已经忘光光了哦~")
					}else {
						ctx.Send(nickname+"的脑袋里可没有与["+key+"]有关的内容呢~")
					}
					content,_ := yaml.Marshal(myData)
					FileWrite("chat\\myData.yaml",content)
				}
			}else {
				ctx.Send("你想让"+nickname+"忘掉什么呀？")
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

	// 被喊名字
	zero.OnFullMatch("", zero.OnlyToMe).SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			time.Sleep(time.Second * 1)
			ctx.SendChain(message.Text(
				[]string{
					nickname + "在此，有何贵干~",
					"(っ●ω●)っ在~",
					"这里是" + nickname + "(っ●ω●)っ",
					nickname + "不在呢~",
				}[rand.Intn(4)],
			))
		})
	// 戳一戳
	zero.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			switch {
			case poke.Load(ctx.Event.UserID).AcquireN(3):
				// 5分钟共8块命令牌 一次消耗3块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("请不要戳", nickname, " >_<"))
			case poke.Load(ctx.Event.UserID).Acquire():
				// 5分钟共8块命令牌 一次消耗1块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("喂(#`O′) 戳", nickname, "干嘛！"))
			default:
				// 频繁触发，不回复
			}
			return
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
