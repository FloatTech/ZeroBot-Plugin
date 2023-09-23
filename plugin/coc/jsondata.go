// Package coc coc插件
package coc

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/FloatTech/floatbox/file"
)

func init() {
	go func() {
		// 新建默认coc面板
		cfgFile := engine.DataFolder() + DefaultYamlFile
		if file.IsNotExist(cfgFile) {
			// 配置默认 config
			baseAttr := []baseInfo{
				{
					Name:  "名称1",
					Value: "",
				},
				{
					Name:  "名称2",
					Value: "",
				},
			}
			attributes := []attribute{
				{
					Name:     "属性1",
					MaxValue: 100,
					MinValue: 0,
					Value:    50,
				},
				{
					Name:     "属性2",
					MaxValue: 100,
					MinValue: 0,
					Value:    50,
				},
				{
					Name:     "属性3",
					MaxValue: 100,
					MinValue: 0,
					Value:    50,
				},
			}
			defaultYal := cocYaml{
				BaseInfo:  baseAttr,
				Attribute: attributes,
			}
			err := savePanel(defaultYal)
			if err != nil {
				panic(err.Error())
			}
		}
		sampleFile := engine.DataFolder() + "面版填写示例.yml"
		if file.IsNotExist(sampleFile) {
			baseAttr := []baseInfo{
				{
					Name:  "昵称",
					Value: "",
				},
				{
					Name:  "职业",
					Value: "",
				},
			}
			attributes := []attribute{
				{
					Name:     "外貌",
					MaxValue: 100,
					MinValue: 0,
					Value:    100,
				},
				{
					Name:     "体型",
					MaxValue: 100,
					MinValue: 0,
					Value:    80,
				},
				{
					Name:     "体质",
					MaxValue: 100,
					MinValue: 0,
					Value:    25,
				},
				{
					Name:     "力量",
					MaxValue: 100,
					MinValue: 0,
					Value:    25,
				},
				{
					Name:     "敏捷",
					MaxValue: 100,
					MinValue: 0,
					Value:    25,
				},
				{
					Name:     "教育",
					MaxValue: 100,
					MinValue: 0,
					Value:    25,
				},
				{
					Name:     "智力",
					MaxValue: 100,
					MinValue: 0,
					Value:    25,
				},
				{
					Name:     "意志",
					MaxValue: 100,
					MinValue: 0,
					Value:    25,
				},
				{
					Name:     "幸运",
					MaxValue: 100,
					MinValue: 0,
					Value:    25,
				},
				{
					Name:     "信誉",
					MaxValue: 100,
					MinValue: 0,
					Value:    25,
				},
			}
			defaultYaml := cocYaml{
				BaseInfo:  baseAttr,
				Attribute: attributes,
			}
			reader, err := os.Create(sampleFile)
			if err == nil {
				err = yaml.NewEncoder(reader).Encode(&defaultYaml)
			}
			if err != nil {
				panic(err.Error())
			}
		}
	}()
}

// 加载数据(2个参数：群号，用户)
func loadPanel(gid int64, uid ...int64) (info cocYaml, err error) {
	cfgFile := strconv.FormatInt(gid, 10)
	if uid != nil {
		cfgFile += "/" + strconv.FormatInt(uid[0], 10) + ".yml"
	} else {
		cfgFile += "/" + DefaultYamlFile
	}
	reader, err := os.Open(engine.DataFolder() + cfgFile)
	if err == nil {
		err = yaml.NewDecoder(reader).Decode(&info)
	}
	if err == nil {
		err = reader.Close()
	}
	return
}

// 保存数据(3个参数：数据，群号，用户)
func savePanel(cfg cocYaml, infoID ...int64) error {
	cfgFile := engine.DataFolder() + DefaultYamlFile
	if infoID != nil {
		str := ""
		for i, ID := range infoID {
			str += strconv.FormatInt(ID, 10)
			if i != len(infoID)-1 {
				str += "/"
				// if file.IsNotExist(engine.DataFolder() + str) {
				//	_, err := os.Create(engine.DataFolder() + str)
				//	return err
				// }
			}
		}
		cfgFile = engine.DataFolder() + str + ".yml"
	}
	reader, err := os.Create(cfgFile)
	if err == nil {
		err = yaml.NewEncoder(reader).Encode(&cfg)
	}
	return err
}

// 加载设置(2个参数：群号，用户)
func loadSetting(gid int64) (info settingInfo, err error) {
	mu.Lock()
	defer mu.Unlock()
	cfgFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + SettingYamlFile
	if file.IsNotExist(cfgFile) {
		// info.DefaultDice = 100
		return
	}
	reader, err := os.Open(cfgFile)
	if err == nil {
		err = yaml.NewDecoder(reader).Decode(&info)
	}
	if err == nil {
		err = reader.Close()
	}
	settingGoup[gid] = info
	return
}

// 保存数据(3个参数：数据，群号，用户)
func saveSetting(info settingInfo, gid int64) error {
	mu.Lock()
	defer mu.Unlock()
	cfgFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + SettingYamlFile
	reader, err := os.Create(cfgFile)
	if err == nil {
		err = yaml.NewEncoder(reader).Encode(&info)
	}
	settingGoup[gid] = info
	return err
}
