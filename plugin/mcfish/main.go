// Package mcfish 钓鱼模拟器
package mcfish

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type fishdb struct {
	db *sql.Sqlite
	sync.RWMutex
}

// FishLimit 钓鱼次数上限
const FishLimit = 50

// 各物品信息
type jsonInfo struct {
	ZoneInfo    []zoneInfo    `json:"分类"` // 区域概率
	ArticleInfo []articleInfo `json:"物品"` // 物品信息
}
type zoneInfo struct {
	Name        string `json:"类型"`        // 类型
	Probability int    `json:"概率[0-100)"` // 概率
}
type articleInfo struct {
	Name        string `json:"名称"`                  // 名称
	Type        string `json:"类型"`                  // 类型
	Probability int    `json:"概率[0-100),omitempty"` // 概率
	Durable     int    `json:"耐久上限,omitempty"`      // 耐久
	Price       int    `json:"价格"`                  // 价格
}

type probabilityLimit struct {
	Min int
	Max int
}

type equip struct {
	ID          int64  // 用户
	Equip       string // 装备名称
	Durable     int    // 耐久
	Maintenance int    // 维修次数
	Induce      int    // 诱钓等级
	Favor       int    // 眷顾等级
}

type article struct {
	Duration int64
	Name     string
	Number   int
	Other    string // 耐久/维修次数/诱钓/眷顾
	Type     string
}

type store struct {
	Duration int64
	Name     string
	Number   int
	Price    int
	Other    string // 耐久/维修次数/诱钓/眷顾
	Type     string
}

type fishState struct {
	ID       int64
	Duration int64
	Fish     int
	Equip    int
}

type storeDiscount struct {
	Name     string
	Discount int
}

var (
	articlesInfo  = jsonInfo{}                            // 物品信息
	thingList     = make([]string, 0, 100)                // 竿列表
	poleList      = make([]string, 0, 10)                 // 竿列表
	fishList      = make([]string, 0, 10)                 // 鱼列表
	treasureList  = make([]string, 0, 10)                 // 鱼列表
	wasteList     = make([]string, 0, 10)                 // 垃圾列表
	probabilities = make(map[string]probabilityLimit, 50) // 概率分布
	priceList     = make(map[string]int, 50)              // 价格分布
	durationList  = make(map[string]int, 50)              // 装备耐久分布
	discountList  = make(map[string]int, 50)              // 价格波动信息
	enchantLevel  = []string{"0", "Ⅰ", "Ⅱ", "Ⅲ"}
	dbdata        = &fishdb{
		db: &sql.Sqlite{},
	}
)

var (
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "钓鱼",
		Help: "一款钓鱼模拟器\n----------指令----------\n" +
			"- 钓鱼看板/钓鱼商店\n- 购买xxx\n- 购买xxx [数量]\n- 出售xxx\n- 出售xxx [数量]\n" +
			"- 钓鱼背包\n- 装备[xx竿|三叉戟|美西螈]\n- 附魔[诱钓|海之眷顾]\n- 修复鱼竿\n- 合成[xx竿|三叉戟]\n" +
			"- 进行钓鱼\n- 进行n次钓鱼\n- 当前装备概率明细\n" +
			"规则:\n1.每日的商店价格是波动的!!如何最大化收益自己考虑一下喔\n" +
			"2.装备信息:\n-> 木竿 : 耐久上限:30 均价:100 上钩概率:0.7%\n-> 铁竿 : 耐久上限:50 均价:300 上钩概率:0.2%\n-> 金竿 : 耐久上限:70 均价700 上钩概率:0.06%\n" +
			"-> 钻石竿 : 耐久上限:100 均价1500 上钩概率:0.03%\n-> 下界合金竿 : 耐久上限:150 均价3100 上钩概率:0.01%\n-> 三叉戟 : 可使钓的鱼类物品数量变成3 耐久上限:300 均价4000 只能合成和交易\n" +
			"3.附魔书信息:\n-> 诱钓 : 减少上钩时间. 均价:1000, 上钩概率:0.59%\n-> 海之眷顾 : 增加宝藏上钩概率. 均价:2500, 上钩概率:0.39%\n" +
			"4.稀有物品:\n-> 唱片 : 出售物品时使用该物品使价格翻倍. 均价:3000, 上钩概率:0.01%\n-> 美西螈 : 可装备,获得隐形[钓鱼佬]buff,并让钓到除鱼竿和美西螈外的物品数量变成3,无耐久上限.不可修复/附魔,每次钓鱼消耗任意一鱼类物品. 均价:3000, 上钩概率:0.01%\n" +
			"5.鱼类信息:\n-> 鳕鱼 : 均价:10 上钩概率:0.69%\n-> 鲑鱼 : 均价:50 上钩概率:0.2%\n-> 热带鱼 : 均价:100 上钩概率:0.06%\n-> 河豚 : 均价:300 上钩概率:0.03%\n-> 鹦鹉螺 : 均价:500 上钩概率:0.01%\n-> 墨鱼 : 均价:500 上钩概率:0.01%\n" +
			"6.垃圾:\n-> 均价:10 上钩概率:30%\n" +
			"7.物品BUFF:\n-> 钓鱼佬 : 当背包名字含有'鱼'的物品数量超过100时激活,钓到物品概率提高至90%\n-> 修复大师 : 当背包鱼竿数量超过10时激活,修复物品时耐久百分百继承\n" +
			"8.合成:\n-> 铁竿 : 3x木竿\n-> 金竿 : 3x铁竿\n-> 钻石竿 : 3x金竿\n-> 下界合金竿 : 3x钻石竿\n-> 三叉戟 : 3x下界合金竿\n注:合成成功率90%,继承附魔等级合/3的等级\n" +
			"9.杂项:\n-> 无装备的情况下,每人最多可以购买3次100块钱的鱼竿\n-> 默认状态钓鱼上钩概率为60%(理论值!!!)\n-> 附魔的鱼竿会因附魔变得昂贵,每个附魔最高3级\n-> 三叉戟不算鱼竿",
		PublicDataFolder: "McFish",
	}).ApplySingle(ctxext.DefaultSingle)
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		dbdata.db.DBPath = engine.DataFolder() + "fishdata.db"
		err := dbdata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at main.go.1]:", err))
			return false
		}
		return true
	})
)

func init() {
	// go func() {
	_, err := engine.GetLazyData("articlesInfo.json", false)
	if err != nil {
		panic(err)
	}
	reader, err := os.Open(engine.DataFolder() + "articlesInfo.json")
	if err == nil {
		err = json.NewDecoder(reader).Decode(&articlesInfo)
	}
	if err == nil {
		err = reader.Close()
	}
	if err != nil {
		panic(err)
	}
	probableList := make([]int, 4)
	for _, info := range articlesInfo.ZoneInfo {
		switch info.Name {
		case "treasure":
			probableList[0] = info.Probability
		case "pole":
			probableList[1] = info.Probability
		case "fish":
			probableList[2] = info.Probability
		case "waste":
			probableList[3] = info.Probability
		}
	}
	probabilities["treasure"] = probabilityLimit{
		Min: 0,
		Max: probableList[0],
	}
	probabilities["pole"] = probabilityLimit{
		Min: probableList[0],
		Max: probableList[1],
	}
	probabilities["fish"] = probabilityLimit{
		Min: probableList[1],
		Max: probableList[2],
	}
	probabilities["waste"] = probabilityLimit{
		Min: probableList[2],
		Max: probableList[3],
	}
	min := make(map[string]int, 4)
	for _, info := range articlesInfo.ArticleInfo {
		switch {
		case info.Type == "pole" || info.Name == "美西螈":
			poleList = append(poleList, info.Name)
		case info.Type == "fish":
			fishList = append(fishList, info.Name)
		case info.Type == "waste":
			wasteList = append(wasteList, info.Name)
		case info.Type == "treasure":
			treasureList = append(treasureList, info.Name)
		}
		thingList = append(thingList, info.Name)
		priceList[info.Name] = info.Price
		if info.Durable != 0 {
			durationList[info.Name] = info.Durable
		}
		probabilities[info.Name] = probabilityLimit{
			Min: min[info.Type],
			Max: min[info.Type] + info.Probability,
		}
		min[info.Type] += info.Probability
	}
	// }()
}

// 更新上限信息
func (sql *fishdb) updateFishInfo(uid int64, number int) (residue int, err error) {
	sql.Lock()
	defer sql.Unlock()
	userInfo := fishState{ID: uid}
	err = sql.db.Create("fishState", &userInfo)
	if err != nil {
		return 0, err
	}
	_ = sql.db.Find("fishState", &userInfo, "where ID = "+strconv.FormatInt(uid, 10))
	if time.Unix(userInfo.Duration, 0).Day() != time.Now().Day() {
		userInfo.Fish = 0
		userInfo.Duration = time.Now().Unix()
	}
	if userInfo.Fish >= FishLimit {
		return 0, nil
	}
	residue = number
	if userInfo.Fish+number > FishLimit {
		residue = FishLimit - userInfo.Fish
		number = residue
	}
	userInfo.Fish += number
	err = sql.db.Insert("fishState", &userInfo)
	return
}

/*********************************************************/
/************************装备相关函数***********************/
/*********************************************************/

func (sql *fishdb) checkEquipFor(uid int64) (ok bool, err error) {
	sql.Lock()
	defer sql.Unlock()
	userInfo := fishState{ID: uid}
	err = sql.db.Create("fishState", &userInfo)
	if err != nil {
		return false, err
	}
	if !sql.db.CanFind("fishState", "where ID = "+strconv.FormatInt(uid, 10)) {
		return true, nil
	}
	err = sql.db.Find("fishState", &userInfo, "where ID = "+strconv.FormatInt(uid, 10))
	if err != nil {
		return false, err
	}
	if userInfo.Equip > 3 {
		return false, nil
	}
	return true, nil
}

func (sql *fishdb) setEquipFor(uid int64) (err error) {
	sql.Lock()
	defer sql.Unlock()
	userInfo := fishState{ID: uid}
	err = sql.db.Create("fishState", &userInfo)
	if err != nil {
		return err
	}
	_ = sql.db.Find("fishState", &userInfo, "where ID = "+strconv.FormatInt(uid, 10))
	if err != nil {
		return err
	}
	userInfo.Equip++
	return sql.db.Insert("fishState", &userInfo)
}

// 获取装备信息
func (sql *fishdb) getUserEquip(uid int64) (userInfo equip, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("equips", &userInfo)
	if err != nil {
		return
	}
	if !sql.db.CanFind("equips", "where ID = "+strconv.FormatInt(uid, 10)) {
		return
	}
	err = sql.db.Find("equips", &userInfo, "where ID = "+strconv.FormatInt(uid, 10))
	return
}

// 更新装备信息
func (sql *fishdb) updateUserEquip(userInfo equip) (err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("equips", &userInfo)
	if err != nil {
		return
	}
	if userInfo.Durable == 0 {
		return sql.db.Del("equips", "where ID = "+strconv.FormatInt(userInfo.ID, 10))
	}
	return sql.db.Insert("equips", &userInfo)
}

func (sql *fishdb) pickFishFor(uid int64, number int) (fishNames map[string]int, err error) {
	fishNames = make(map[string]int, 6)
	name := strconv.FormatInt(uid, 10) + "Pack"
	sql.Lock()
	defer sql.Unlock()
	userInfo := article{}
	err = sql.db.Create(name, &userInfo)
	if err != nil {
		return
	}
	count, err := sql.db.Count(name)
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	if !sql.db.CanFind(name, "where Type is 'fish'") {
		return
	}
	fishTypes := make([]article, 0, count)
	fishInfo := article{}
	err = sql.db.FindFor(name, &fishInfo, "where Type is 'fish'", func() error {
		fishTypes = append(fishTypes, fishInfo)
		return nil
	})
	if err != nil {
		return
	}
	if len(fishTypes) == 0 {
		return
	}
	max := 0
	for _, info := range fishTypes {
		max += info.Number
	}
	if max < number {
		number = max
	}
	for i := number; i > 0; {
		randNumber := rand.Intn(len(fishTypes))
		if fishTypes[randNumber].Number <= 0 {
			continue
		}
		fishTypes[randNumber].Number--
		if fishTypes[randNumber].Number <= 0 {
			err = sql.db.Del(name, "where Duration = "+strconv.FormatInt(fishTypes[randNumber].Duration, 10))
		} else {
			err = sql.db.Insert(name, &fishTypes[randNumber])
		}
		if err != nil {
			return
		}
		fishNames[fishTypes[randNumber].Name]++
		i--
	}
	return
}

/*********************************************************/
/************************背包相关函数***********************/
/*********************************************************/

// 获取用户背包信息
func (sql *fishdb) getUserPack(uid int64) (thingInfos []article, err error) {
	sql.Lock()
	defer sql.Unlock()
	userInfo := article{}
	err = sql.db.Create(strconv.FormatInt(uid, 10)+"Pack", &userInfo)
	if err != nil {
		return
	}
	count, err := sql.db.Count(strconv.FormatInt(uid, 10) + "Pack")
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	err = sql.db.FindFor(strconv.FormatInt(uid, 10)+"Pack", &userInfo, "ORDER by Type, Name, Other ASC", func() error {
		thingInfos = append(thingInfos, userInfo)
		return nil
	})
	return
}

// 获取用户物品信息
func (sql *fishdb) getUserThingInfo(uid int64, thing string) (thingInfos []article, err error) {
	name := strconv.FormatInt(uid, 10) + "Pack"
	sql.Lock()
	defer sql.Unlock()
	userInfo := article{}
	err = sql.db.Create(name, &userInfo)
	if err != nil {
		return
	}
	count, err := sql.db.Count(name)
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	if !sql.db.CanFind(name, "where Name = '"+thing+"'") {
		return
	}
	err = sql.db.FindFor(name, &userInfo, "where Name = '"+thing+"'", func() error {
		thingInfos = append(thingInfos, userInfo)
		return nil
	})
	return
}

// 更新用户物品信息
func (sql *fishdb) updateUserThingInfo(uid int64, userInfo article) (err error) {
	name := strconv.FormatInt(uid, 10) + "Pack"
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create(name, &userInfo)
	if err != nil {
		return
	}
	if userInfo.Number == 0 {
		return sql.db.Del(name, "where Duration = "+strconv.FormatInt(userInfo.Duration, 10))
	}
	return sql.db.Insert(name, &userInfo)
}

// 获取某关键字的数量
func (sql *fishdb) getNumberFor(uid int64, thing string) (number int, err error) {
	name := strconv.FormatInt(uid, 10) + "Pack"
	sql.Lock()
	defer sql.Unlock()
	userInfo := article{}
	err = sql.db.Create(name, &userInfo)
	if err != nil {
		return
	}
	count, err := sql.db.Count(name)
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	if !sql.db.CanFind(name, "where Name glob '*"+thing+"*'") {
		return
	}
	info := article{}
	err = sql.db.FindFor(name, &info, "where Name glob '*"+thing+"*'", func() error {
		number += info.Number
		return nil
	})
	return
}

/*********************************************************/
/************************商店相关函数***********************/
/*********************************************************/

// 刷新商店信息
func (sql *fishdb) refreshStroeInfo() (ok bool, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("stroeDiscount", &storeDiscount{})
	if err != nil {
		return false, err
	}
	lastTime := storeDiscount{}
	_ = sql.db.Find("stroeDiscount", &lastTime, "where Name = 'lastTime'")
	refresh := false
	timeNow := time.Now().Day()
	if timeNow != lastTime.Discount {
		lastTime = storeDiscount{
			Name:     "lastTime",
			Discount: timeNow,
		}
		err = sql.db.Insert("stroeDiscount", &lastTime)
		if err != nil {
			return false, err
		}
		refresh = true
	}
	for _, name := range thingList {
		thing := storeDiscount{}
		switch refresh {
		case true:
			thingDiscount := 50 + rand.Intn(150)
			thing = storeDiscount{
				Name:     name,
				Discount: thingDiscount,
			}
			err = sql.db.Insert("stroeDiscount", &thing)
			if err != nil {
				return
			}
		default:
			_ = sql.db.Find("stroeDiscount", &thing, "where Name = '"+name+"'")
		}
		if thing.Discount != 0 {
			discountList[name] = thing.Discount
		} else {
			discountList[name] = 100
		}
	}
	if refresh { // 每天调控1种鱼
		thingInfo := store{}
		err = sql.db.Create("store", &thingInfo)
		if err != nil {
			return
		}
		fish := fishList[rand.Intn(len(fishList))]
		_ = sql.db.Find("store", &thingInfo, "where Name = '"+fish+"'")
		if thingInfo == (store{}) {
			thingInfo.Duration = time.Now().Unix()
			thingInfo.Type = "fish"
			thingInfo.Name = fish
			thingInfo.Price = priceList[fish] * discountList[fish] / 100
		}
		thingInfo.Number += (100 - discountList[fish])
		if thingInfo.Number < 1 {
			thingInfo.Number = 1
		}
		_ = sql.db.Insert("store", &thingInfo)
	}
	return true, nil
}

// 获取商店信息
func (sql *fishdb) getStoreInfo() (thingInfos []store, err error) {
	sql.Lock()
	defer sql.Unlock()
	thingInfo := store{}
	err = sql.db.Create("store", &thingInfo)
	if err != nil {
		return
	}
	count, err := sql.db.Count("store")
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	err = sql.db.FindFor("store", &thingInfo, "ORDER by Type, Name, Price ASC", func() error {
		thingInfos = append(thingInfos, thingInfo)
		return nil
	})
	return
}

// 获取商店物品信息
func (sql *fishdb) getStoreThingInfo(thing string) (thingInfos []store, err error) {
	sql.Lock()
	defer sql.Unlock()
	thingInfo := store{}
	err = sql.db.Create("store", &thingInfo)
	if err != nil {
		return
	}
	count, err := sql.db.Count("store")
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	if !sql.db.CanFind("store", "where Name = '"+thing+"'") {
		return
	}
	err = sql.db.FindFor("store", &thingInfo, "where Name = '"+thing+"'", func() error {
		thingInfos = append(thingInfos, thingInfo)
		return nil
	})
	return
}

// 获取商店物品信息
func (sql *fishdb) checkStoreFor(thing store, number int) (ok bool, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("store", &thing)
	if err != nil {
		return
	}
	count, err := sql.db.Count("store")
	if err != nil {
		return
	}
	if count == 0 {
		return false, nil
	}
	if !sql.db.CanFind("store", "where Duration = "+strconv.FormatInt(thing.Duration, 10)) {
		return false, nil
	}
	err = sql.db.Find("store", &thing, "where Duration = "+strconv.FormatInt(thing.Duration, 10))
	if err != nil {
		return
	}
	if thing.Number < number {
		return false, nil
	}
	return true, nil
}

// 更新商店信息
func (sql *fishdb) updateStoreInfo(thingInfo store) (err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("store", &thingInfo)
	if err != nil {
		return
	}
	if thingInfo.Number == 0 {
		return sql.db.Del("store", "where Duration = "+strconv.FormatInt(thingInfo.Duration, 10))
	}
	return sql.db.Insert("store", &thingInfo)
}
