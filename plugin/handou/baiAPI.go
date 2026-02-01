// Package handou 猜成语
package handou

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/FloatTech/floatbox/web"
	"github.com/sirupsen/logrus"
)

type baiduAPIData struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
	Data   struct {
		IdiomVersion int    `json:"idiomVersion"`
		Name         string `json:"name"`
		Sid          string `json:"sid"`
		Type         string `json:"type"`
		LessonInfo   any    `json:"lessonInfo"`
		RelationInfo struct {
			RelationName string `json:"relationName"`
			RelationList []struct {
				Name string   `json:"name"`
				Imgs []string `json:"imgs"`
			} `json:"relationList"`
		} `json:"relationInfo"`
		Imgs       []string `json:"imgs"`
		Definition []struct {
			Pinyin           string   `json:"pinyin"`
			Voice            string   `json:"voice"`
			Definition       []string `json:"definition"`
			DetailDefinition any      `json:"detailDefinition"`
		} `json:"definition"`
		DefinitionInfo struct {
			Definition        string `json:"definition"`
			SimilarDefinition string `json:"similarDefinition"`
			AncientDefinition string `json:"ancientDefinition"`
			ModernDefinition  string `json:"modernDefinition"`
			DetailMeans       []struct {
				Word       string `json:"word"`
				Definition string `json:"definition"`
			} `json:"detailMeans"`
			UsageTips     any    `json:"usageTips"`
			Yicuodian     any    `json:"yicuodian"`
			Baobian       string `json:"baobian"`
			WordFormation string `json:"wordFormation"`
		} `json:"definitionInfo"`
		Liju []struct {
			Name     string `json:"name"`
			ShowName string `json:"showName"`
		} `json:"liju"`
		Source  string `json:"source"`
		Story   any    `json:"story"`
		Antonym []struct {
			Name    string `json:"name"`
			IsClick bool   `json:"isClick"`
		} `json:"antonym"`
		Synonym  []string `json:"synonym"`
		Synonyms []struct {
			Name    string `json:"name"`
			IsClick bool   `json:"isClick"`
		} `json:"synonyms"`
		Tongyiyixing []struct {
			Name    string `json:"name"`
			IsClick bool   `json:"isClick"`
		} `json:"tongyiyixing"`
		ChuChu []struct {
			SourceChapter    string `json:"sourceChapter"`
			Source           string `json:"source"`
			Dynasty          string `json:"dynasty"`
			CiteOriginalText string `json:"citeOriginalText"`
			Author           string `json:"author"`
		} `json:"chuChu"`
		YinZheng []struct {
			SourceChapter    string `json:"sourceChapter"`
			Source           string `json:"source"`
			Dynasty          string `json:"dynasty"`
			CiteOriginalText string `json:"citeOriginalText"`
			Author           string `json:"author"`
		} `json:"yinZheng"`
		PictureList []any `json:"pictureList"`
		LessonTerms struct {
			TermList any `json:"termList"`
			HasTerms int `json:"hasTerms"`
		} `json:"lessonTerms"`
		LessonTermsNew struct {
			TermList any `json:"termList"`
			HasTerms int `json:"hasTerms"`
		} `json:"lessonTermsNew"`
		Baobian     string `json:"baobian"`
		Structure   string `json:"structure"`
		Pinyin      string `json:"pinyin"`
		Voice       string `json:"voice"`
		ZuowenQuery string `json:"zuowen_query"`
	} `json:"data"`
}

func geiAPIdata(s string) (*idiomJson, error) {
	url := "https://hanyuapp.baidu.com/dictapp/swan/termdetail?wd=" + url.QueryEscape(s) + "&client=pc&source_tag=2&lesson_from=xiaodu"
	logrus.Warningln(url)
	data, err := web.GetData(url)
	if err != nil {
		return nil, err
	}

	var apiData baiduAPIData
	err = json.Unmarshal(data, &apiData)
	if err != nil {
		return nil, err
	}
	if apiData.Data.Name == "" {
		return nil, fmt.Errorf("未找到该成语")
	}
	derivation := ""
	for _, v := range apiData.Data.ChuChu {
		if derivation != "" {
			derivation += "\n"
		}
		derivation += v.Dynasty + "·" + v.Author + " " + v.Source + "：" + v.CiteOriginalText
	}

	explanation := apiData.Data.DefinitionInfo.Definition + apiData.Data.DefinitionInfo.ModernDefinition
	if derivation == "" && explanation == "" {
		return nil, fmt.Errorf("无法获取成语词源和解释")
	}
	synonyms := make([]string, len(apiData.Data.Synonyms))
	for i, synonym := range apiData.Data.Synonyms {
		synonyms[i] = synonym.Name
	}
	for i, synonym := range apiData.Data.Synonym {
		if !slices.Contains(synonyms, synonym) {
			synonyms[i] = synonym
		}
	}
	liju := ""
	if len(apiData.Data.Liju) > 0 {
		liju = apiData.Data.Liju[0].Name
	}

	// 生成字符切片
	chars := make([]string, 0, len(s))
	for _, r := range s {
		chars = append(chars, string(r))
	}
	// 分割拼音
	pinyinSlice := strings.Split(apiData.Data.Pinyin, " ")
	if len(pinyinSlice) != len(chars) {
		pinyinSlice = strings.Split(apiData.Data.Definition[0].Pinyin, " ")
	}

	newIdiom := idiomJson{
		Word:         apiData.Data.Name,
		Chars:        chars,
		Pinyin:       pinyinSlice,
		Baobian:      apiData.Data.Baobian,
		Explanation:  explanation,
		Derivation:   derivation,
		Example:      liju,
		Abbreviation: apiData.Data.Structure,
		Synonyms:     synonyms,
	}
	return &newIdiom, nil
}

var mu sync.Mutex

func saveIdiomJson() error {
	mu.Lock()
	defer mu.Unlock()
	f, err := os.Create(idiomFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(&idiomInfoMap)
}
