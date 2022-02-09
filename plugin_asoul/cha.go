package asoul

import (
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io/ioutil"
	"net/http"
	"strings"
)

// 查成分的
func init() {
	// 插件初始化，获取vtbs数据入库
	//go func() {
	//	process.SleepAbout1sTo2s()
	//	_ = os.MkdirAll(datapath, 0755)
	//	db := &sql.Sqlite{DBPath: dbfile}
	//	db.Create("vtbs", &vtbs{})
	//	log.Infof("[vtbs]创建数据库和表完成")
	//	vtbData := getVtbs().Array()
	//	fmt.Println(vtbData)
	//	for _, i := range vtbData {
	//		fmt.Println(i)
	//		db.Insert("vtbs", &vtbs{
	//			Mid: i.Get("mid").Int(),
	//			Un:  i.Get("uname").Str,
	//			Rid: i.Get("roomid").Int(),
	//		})
	//	}
	//	log.Infof("[vtbs]获取%v条vtbs数据，插入数据库成功", len(vtbData))
	//}()
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
	engine.OnRegex(`^/查 UID:(.{1,25})$`).SetBlock(true).
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
	query1, args, err := sqlx.In("select uname from vtbs where mid in (?)", follows)
	res := []string{}
	err = db.Select(&res, query1, args...)
	if err != nil {
		panic(err)
	}
	return res
}

// 获取vtbs数据返回
func getVtbs() gjson.Result {
	url := "https://api.vtbs.moe/v1/short"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Error("请求api失败", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Error("请求api失败", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	json := gjson.ParseBytes(body)

	return json
}
