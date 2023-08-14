// Package coc coc插件
package coc

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/FloatTech/floatbox/file"
)

func init() {
	go func() {
		// 新建默认coc面板
		cfgFile := engine.DataFolder() + DefaultJsonFile
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
			defaultJson := cocJson{
				BaseInfo:  baseAttr,
				Attribute: attributes,
			}
			err := savePanel(defaultJson)
			if err != nil {
				panic(err.Error())
			}
		}
		sampleFile := engine.DataFolder() + "面版填写示例.json"
		if file.IsNotExist(sampleFile) {
			baseAttr := []baseInfo{
				{
					Name:  "昵称",
					Value: "",
				},
				{
					Name:  "身份",
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
			}
			defaultJson := cocJson{
				BaseInfo:  baseAttr,
				Attribute: attributes,
			}
			reader, err := os.Create(sampleFile)
			if err == nil {
				err = json.NewEncoder(reader).Encode(&defaultJson)
			}
			if err != nil {
				panic(err.Error())
			}
		}
	}()
}

// 加载数据(2个参数：群号，用户)
func loadPanel(gid int64, uid ...int64) (info cocJson, err error) {
	cfgFile := strconv.FormatInt(gid, 10)
	if uid != nil {
		cfgFile += "/" + strconv.FormatInt(uid[0], 10) + ".json"
	} else {
		cfgFile += "/" + DefaultJsonFile
	}
	reader, err := os.Open(engine.DataFolder() + cfgFile)
	if err == nil {
		err = json.NewDecoder(reader).Decode(&info)
	}
	if err == nil {
		err = reader.Close()
	}
	return
}

// 保存数据(3个参数：数据，群号，用户)
func savePanel(cfg cocJson, infoID ...int64) error {
	cfgFile := engine.DataFolder() + DefaultJsonFile
	if infoID != nil {
		str := ""
		for i, ID := range infoID {
			str += strconv.FormatInt(ID, 10)
			if i != len(infoID)-1 {
				str += "/"
				//if file.IsNotExist(engine.DataFolder() + str) {
				//	_, err := os.Create(engine.DataFolder() + str)
				//	return err
				//}
			}
		}
		cfgFile = engine.DataFolder() + str + ".json"
	}
	reader, err := os.Create(cfgFile)
	if err == nil {
		err = json.NewEncoder(reader).Encode(&cfg)
	}
	return err
}

// 加载设置(2个参数：群号，用户)
func loadSetting(gid int64) (info settingInfo, err error) {
	cfgFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + SettingJsonFile
	if file.IsNotExist(cfgFile) {
		//info.DefaultDice = 100
		return
	}
	reader, err := os.Open(cfgFile)
	if err == nil {
		err = json.NewDecoder(reader).Decode(&info)
	}
	if err == nil {
		err = reader.Close()
	}
	settingGoup[gid] = info
	return
}

// 保存数据(3个参数：数据，群号，用户)
func saveSetting(info settingInfo, gid int64) error {
	cfgFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + SettingJsonFile
	reader, err := os.Create(cfgFile)
	if err == nil {
		err = json.NewEncoder(reader).Encode(&info)
	}
	settingGoup[gid] = info
	return err
}
