package main

import (
	"fmt"

	m "gm/modules"
	"gm/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var Conf = &utils.YamlConfig{}

func main() {
	zero.Run(zero.Option{
		Host:          Conf.Host,
		Port:          Conf.Port,
		AccessToken:   Conf.AccessToken,
		NickName:      []string{"GroupManager"},
		CommandPrefix: "/",
		SuperUsers:    Conf.Master,
	})
	select {}
}

func init() {
	fmt.Println(`
====================[GroupManager]====================
* OneBot + ZeroBot + Golang
* Copyright © 2018-2020 Kanri, All Rights Reserved
* Project: https://github.com/Yiwen-Chan/GroupManager
=======================================================
`)
	Conf = utils.Load("./GroupManager.yml")
	m.Conf = Conf
	fmt.Println("[GroupManager] 有需要请按 GitHub 项目上描述的方法修改配置文件")
}
