package asoul

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// 查成分的
func init() {
	// 插件主体,匹配用户名字
	engine.OnRegex(`^/查\s(.{1,25})$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			rest, err := getMid(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			mid := rest.Get("data.result.0.mid").String()
			info := getinfo(mid)
			attentions := compared(info.Card.Attentions)
			ctx.SendChain(message.Image(info.Card.Face),
				message.Text(
					"id: ", info.Card.Name, "\n",
					"uid: ", info.Card.Mid, "\n",
					"性别: ", info.Card.Sex, "\n",
					"等级: ", info.Card.LevelInfo.CurrentLevel, "\n",
					"关注数: ", info.Card.Attention, "\n",
					"粉丝数: ", info.Card.Fans, "\n",
					"使用装扮: ", info.Card.Pendant.Name, "\n",
					"关注的vtb（共", len(attentions), "个）: ", "\n",
					strings.Join(attentions, "，"),
				))
		})
	// 插件主体,匹配用户uid
	engine.OnRegex(`^/查UID:(.{1,25})$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			mid := keyword
			info := getinfo(mid)
			attentions := compared(info.Card.Attentions)
			ctx.SendChain(message.Image(info.Card.Face),
				message.Text(
					"id: ", info.Card.Name, "\n",
					"uid: ", info.Card.Mid, "\n",
					"性别: ", info.Card.Sex, "\n",
					"等级: ", info.Card.LevelInfo.CurrentLevel, "\n",
					"关注数: ", info.Card.Attention, "\n",
					"粉丝数: ", info.Card.Fans, "\n",
					"使用装扮: ", info.Card.Pendant.Name, "\n",
					"关注的vtb（共", len(attentions), "个）: ", "\n",
					strings.Join(attentions, "，"),
				))
		})
}

// 通过触发指令的名字获取用户的uid
func getMid(keyword string) (gjson.Result, error) {
	api := "http://api.bilibili.com/x/web-interface/search/type?search_type=bili_user&&user_type=1&keyword=" + keyword
	resp, err := http.Get(api)
	if err != nil {
		return gjson.Result{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return gjson.Result{}, errors.New("code not 200")
	}
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json := gjson.ParseBytes(data)
	if json.Get("data.numResults").Int() == 0 {
		return gjson.Result{}, errors.New("查无此人")
	}
	return json, nil
}

// 获取被查用户信息
func getinfo(mid string) *follows {
	url := "https://account.bilibili.com/api/member/getCardByMid?mid=" + mid
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	result := &follows{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}

// 对比数据库获取关注用户的名字
func compared(follows []int) []string {
	var db *sqlx.DB
	db, _ = sqlx.Open("sqlite3", dbfile)
	defer db.Close()
	query1, args, err := sqlx.In("select uname from vtbs where mid in (?)", follows)
	if err != nil {
		log.Errorln("[element]查找失败", err)
	}
	res := []string{}
	err = db.Select(&res, query1, args...)
	if err != nil {
		log.Errorln("[element]查找失败", err)
	}
	return res
}

//首次启动初始化插件, 异步处理！！
func init() {
	go func() {
		getVtbs()
	}()
}

// 获取vtbs数据返回
func getVtbs() {
	url := "https://api.vtbs.moe/v1/short"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Error("[Element]请求api失败", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Error("[Element]请求api失败")
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorln(err)
	}
	json := gjson.ParseBytes(body)

	if _, err = os.Stat(dbfile); err != nil || os.IsNotExist(err) {
		// 生成文件
		var err error
		_ = os.MkdirAll(datapath, 0755)
		f, err := os.Create(dbfile)
		if err != nil {
			log.Error("[Element]", err)
		}
		log.Infof("[Element]数据库文件(%v)创建成功", dbfile)
		time.Sleep(1 * time.Second)
		defer f.Close()
		// 打开数据库制表
		db, err := gorm.Open("sqlite3", dbfile)
		if err != nil {
			log.Errorln("[Element]打开数据库失败：", err)
		}
		db.AutoMigrate(vtbs{})
		time.Sleep(1 * time.Second)
		// 插入数据
		for _, i := range json.Array() {
			db.Create(&vtbs{
				Mid:   i.Get("mid").Int(),
				Uname: i.Get("uname").Str,
				Rid:   i.Get("roomid").Int(),
			})
		}
		log.Infof("[Element]vtbs更新完成，插入（%v）条数据", len(json.Array()))
		defer db.Close()
	}
}
