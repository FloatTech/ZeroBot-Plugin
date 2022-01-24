package main

import (
	_ "SkadiBot/plugins/arknights"
	_ "SkadiBot/plugins/bilibili"
	_ "SkadiBot/plugins/group"
	_ "SkadiBot/plugins/normal"
	"fmt"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	CommandPrefix string   `yaml:"CommandPrefix"`
	SuperUsers    []string `yaml:"SuperUsers"`
	WS            string   `yaml:"WS"`
	AccessToken   string   `yaml:"AccessToken"`
}

type Base64Hook struct {
}

func (hook *Base64Hook) Fire(entry *log.Entry) error {
	if strings.HasPrefix(entry.Message, "发送群消息") || strings.HasPrefix(entry.Message, "发送私聊消息") {
		reg := regexp.MustCompile(`"base64://([A-Za-z0-9+/]*)={0,2}`)
		entry.Message = reg.ReplaceAllString(entry.Message, "\"base64 file\"")
	}
	return nil
}

func (hook *Base64Hook) Levels() []log.Level {
	return log.AllLevels
}

func main() {
	err := os.Setenv("HTTP_PROXY", "")
	if err != nil {
		return
	}
	err = os.Setenv("HTTPS_PROXY", "")
	if err != nil {
		return
	}
	fmt.Println("交流群群号:258140966")
	fmt.Println("仓库地址：https://github.com/yuanyan3060/SkadiBot")
	rep, err := http.Get("https://cdn.jsdelivr.net/gh/yuanyan3060/SkadiBot/LICENSE")
	if err != nil {
		fmt.Println("获取项目状态失败")
		return
	}
	if rep.StatusCode != 200 {
		fmt.Println("本项目已失效")
		return
	}
	log.AddHook(&Base64Hook{})
	config := Config{
		CommandPrefix: "",
		SuperUsers:    nil,
		WS:            "ws://127.0.0.1:6700",
		AccessToken:   "",
	}
	configBytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		marshal, err := yaml.Marshal(&config)
		if err != nil {
			return
		}
		err = ioutil.WriteFile("config.yaml", marshal, 777)
		if err != nil {
			return
		}
		fmt.Println("已生成config.yaml文件，请修改配置后重启")
		for {
		}
	} else {
		err = yaml.Unmarshal(configBytes, &config)
		if err != nil {
			return
		}
		zero.Run(zero.Config{
			NickName:      []string{"bot"},
			CommandPrefix: config.CommandPrefix,
			SuperUsers:    config.SuperUsers,
			Driver: []zero.Driver{
				driver.NewWebSocketClient(config.WS, config.AccessToken),
			},
		})
		select {}
	}

}
