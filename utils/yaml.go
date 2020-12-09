package utils

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type YamlConfig struct {
	Version     string   `yaml:"插件版本"`
	Host        string   `yaml:"监听地址"`
	Port        string   `yaml:"监听端口"`
	AccessToken string   `yaml:"Token"`
	Master      []string `yaml:"主人QQ"`
}

func DefaultConfig() *YamlConfig {
	return &YamlConfig{
		Version:     "1",
		Host:        "127.0.0.1",
		Port:        "8080",
		AccessToken: "",
		Master: []string{
			"66666666",
			"88888888",
		},
	}
}

func Load(p string) *YamlConfig {
	if !PathExists(p) {
		DefaultConfig().Save(p)
	}
	c := YamlConfig{}
	err := yaml.Unmarshal([]byte(ReadAllText(p)), &c)
	if err != nil {
		fmt.Println("[GroupManager] 尝试加载配置文件失败: 读取文件失败")
		fmt.Println("[GroupManager] 原配置文件已备份")
		os.Rename(p, p+".backup"+strconv.FormatInt(time.Now().Unix(), 10))
		DefaultConfig().Save(p)
	}
	c = YamlConfig{}
	yaml.Unmarshal([]byte(ReadAllText(p)), &c)
	return &c
}

func (c *YamlConfig) Save(p string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		fmt.Println("[GroupManager] 写入新的配置文件失败")
		return err
	}
	WriteAllText(p, string(data))
	return nil
}
