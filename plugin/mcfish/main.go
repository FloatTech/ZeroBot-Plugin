// Package mcfish 钓鱼模拟器
package mcfish

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/math"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type fishdb struct {
	sync.RWMutex
	db sql.Sqlite
}

// FishLimit 钓鱼次数上限
const FishLimit = 50

// version 规则版本号
const version = "5.6.1"

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
	Curse    int // 功德--(x)
	Bless    int // 功德++(x)
}

type storeDiscount struct {
	Name     string
	Discount int
}

// buff状态记录
// buff0: 优惠卷
type buffInfo struct {
	ID        int64
	Duration  int64
	BuyTimes  int `db:"Buff0"` // 购买次数
	Coupon    int `db:"Buff1"` // 优惠卷
	SalesPole int `db:"Buff2"` // 卖鱼竿上限
	BuyTing   int `db:"Buff3"` // 购买上限
	Buff4     int `db:"Buff4"` // 暂定
	Buff5     int `db:"Buff5"` // 暂定
	Buff6     int `db:"Buff6"` // 暂定
	Buff7     int `db:"Buff7"` // 暂定
	Buff8     int `db:"Buff8"` // 暂定
	Buff9     int `db:"Buff9"` // 暂定
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
	dbdata        fishdb
)

var (
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "钓鱼",
		Help: "一款钓鱼模拟器,规则:V" + version +
			"\n----------指令----------\n" +
			"- 钓鱼背包\n" +
			"- 进行钓鱼 / 进行n次钓鱼\n" +
			"- 修复鱼竿\n" +
			"- 钓鱼商店 / 钓鱼看板\n" +
			"- 购买xxx / 购买xxx [数量]\n- 出售xxx / 出售xxx [数量]\n" +
			"- 消除[绑定|宝藏]诅咒 / 消除[绑定|宝藏]诅咒 [数量]\n" +
			"- 装备[xx竿|三叉戟|美西螈]\n" +
			"- 附魔[诱钓|海之眷顾]\n" +
			"- 合成[xx竿|三叉戟]\n" +
			"- 出售所有垃圾\n" +
			"- 当前装备概率明细\n" +
			"- 查看钓鱼规则\n",
		PublicDataFolder: "McFish",
	}).ApplySingle(ctxext.DefaultSingle)
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		dbdata.db = sql.New(engine.DataFolder() + "fishdata.db")
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
	minMap := make(map[string]int, 4)
	for _, info := range articlesInfo.ArticleInfo {
		switch {
		case info.Type == "pole" || info.Name == "美西螈":
			poleList = append(poleList, info.Name)
		case info.Type == "fish" || info.Name == "海豚":
			fishList = append(fishList, info.Name)
		case info.Type == "waste":
			wasteList = append(wasteList, info.Name)
		case info.Type == "treasure":
			treasureList = append(treasureList, info.Name)
		}
		if info.Name != "宝藏诅咒" {
			thingList = append(thingList, info.Name)
			priceList[info.Name] = info.Price
		}
		if info.Durable != 0 {
			durationList[info.Name] = info.Durable
		}
		probabilities[info.Name] = probabilityLimit{
			Min: minMap[info.Type],
			Max: minMap[info.Type] + info.Probability,
		}
		minMap[info.Type] += info.Probability
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
	_ = sql.db.Find("fishState", &userInfo, "WHERE ID = ?", uid)
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

// 更新诅咒
func (sql *fishdb) updateCurseFor(uid int64, info string, number int) (err error) {
	if number < 1 {
		return
	}
	sql.Lock()
	defer sql.Unlock()
	userInfo := fishState{ID: uid}
	err = sql.db.Create("fishState", &userInfo)
	if err != nil {
		return err
	}
	changeCheck := false
	add := 0
	buffName := "宝藏诅咒"
	_ = sql.db.Find("fishState", &userInfo, "WHERE ID = ?", uid)
	if info == "fish" {
		userInfo.Bless += number
		for userInfo.Bless >= 75 {
			add++
			changeCheck = true
			buffName = "净化书"
			userInfo.Bless -= 75
		}
	} else {
		userInfo.Curse += number
		for userInfo.Curse >= 10 {
			add++
			changeCheck = true
			userInfo.Curse -= 10
		}
	}
	err = sql.db.Insert("fishState", &userInfo)
	if err != nil {
		return err
	}
	if changeCheck {
		table := strconv.FormatInt(uid, 10) + "Pack"
		thing := article{
			Duration: time.Now().Unix(),
			Name:     buffName,
			Type:     "treasure",
		}
		_ = sql.db.Find(table, &thing, "WHERE Name = ?", buffName)
		thing.Number += add
		return sql.db.Insert(table, &thing)
	}
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
	if !sql.db.CanFind("fishState", "WHERE ID = ?", uid) {
		return true, nil
	}
	err = sql.db.Find("fishState", &userInfo, "WHERE ID = ?", uid)
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
	_ = sql.db.Find("fishState", &userInfo, "WHERE ID = ?", uid)
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
	if !sql.db.CanFind("equips", "WHERE ID = ?", uid) {
		return
	}
	err = sql.db.Find("equips", &userInfo, "WHERE ID = ?", uid)
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
		return sql.db.Del("equips", "WHERE ID = ?", userInfo.ID)
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
	if !sql.db.CanFind(name, "WHERE Type = 'fish'") {
		return
	}
	fishInfo := article{}
	k := 0
	for i := number; i > 0 && k < len(fishList); {
		_ = sql.db.Find(name, &fishInfo, "WHERE Name = ?", fishList[k])
		if fishInfo.Number <= 0 {
			k++
			continue
		}
		if fishInfo.Number < i {
			k++
			i -= fishInfo.Number
			fishNames[fishInfo.Name] += fishInfo.Number
			fishInfo.Number = 0
		} else {
			fishNames[fishInfo.Name] += i
			fishInfo.Number -= i
			i = 0
		}
		if fishInfo.Number <= 0 {
			err = sql.db.Del(name, "WHERE Duration = ?", fishInfo.Duration)
		} else {
			err = sql.db.Insert(name, &fishInfo)
		}
		if err != nil {
			return
		}
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
	if !sql.db.CanFind(name, "WHERE Name = ?", thing) {
		return
	}
	err = sql.db.FindFor(name, &userInfo, "WHERE Name = ?", func() error {
		thingInfos = append(thingInfos, userInfo)
		return nil
	}, thing)
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
		return sql.db.Del(name, "WHERE Duration = ?", userInfo.Duration)
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
	if !sql.db.CanFind(name, "WHERE Name glob ?", "*"+thing+"*") {
		return
	}
	info := article{}
	err = sql.db.FindFor(name, &info, "WHERE Name glob ?", func() error {
		number += info.Number
		return nil
	}, "*"+thing+"*")
	return
}

// 获取用户的某类物品信息
func (sql *fishdb) getUserTypeInfo(uid int64, thingType string) (thingInfos []article, err error) {
	name := strconv.FormatInt(uid, 10) + "Pack"
	sql.Lock()
	defer sql.Unlock()
	userInfo := article{}
	err = sql.db.Create(name, &userInfo)
	if err != nil {
		return
	}
	if !sql.db.CanFind(name, "WHERE Type = ?", thingType) {
		return
	}
	err = sql.db.FindFor(name, &userInfo, "WHERE Type = ?", func() error {
		thingInfos = append(thingInfos, userInfo)
		return nil
	}, thingType)
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
	err = sql.db.Create("store", &store{})
	if err != nil {
		return false, err
	}
	lastTime := storeDiscount{}
	_ = sql.db.Find("stroeDiscount", &lastTime, "WHERE Name = 'lastTime'")
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
			thingInfo := store{}
			_ = sql.db.Find("store", &thingInfo, "WHERE Name = ?", name)
			if thingInfo.Number > 150 {
				// 控制价格浮动区间： -10%到10%
				thing.Discount = 90 + rand.Intn(20)
			}
			err = sql.db.Insert("stroeDiscount", &thing)
			if err != nil {
				return
			}
		default:
			_ = sql.db.Find("stroeDiscount", &thing, "WHERE Name = ?", name)
		}
		if thing.Discount != 0 {
			discountList[name] = thing.Discount
		} else {
			discountList[name] = 100
		}
	}
	thing := store{}
	var oldThing []store
	_ = sql.db.FindFor("stroeDiscount", &thing, "WHERE type = 'pole'", func() error {
		if time.Since(time.Unix(thing.Duration, 0)) > 24 {
			oldThing = append(oldThing, thing)
		}
		return nil
	})
	for _, info := range oldThing {
		_ = sql.db.Del("stroeDiscount", "WHERE Duration = ?", info.Duration)
	}
	if refresh {
		// 每天调控1种鱼
		fish := fishList[rand.Intn(len(fishList))]
		thingInfo := store{
			Duration: time.Now().Unix(),
			Name:     fish,
			Type:     "fish",
			Price:    priceList[fish] * discountList[fish] / 100,
		}
		_ = sql.db.Find("store", &thingInfo, "WHERE Name = ?", fish)
		thingInfo.Number += 100 - discountList[fish]
		if thingInfo.Number < 1 {
			thingInfo.Number = 100
		}
		_ = sql.db.Insert("store", &thingInfo)
		// 每天上架20本净化书
		thingInfo = store{
			Duration: time.Now().Unix(),
			Name:     "净化书",
			Type:     "article",
			Price:    priceList["净化书"] * discountList["净化书"] / 100,
		}
		_ = sql.db.Find("store", &thingInfo, "WHERE Name = '净化书'")
		thingInfo.Number = 20
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
	if !sql.db.CanFind("store", "WHERE Name = ?", thing) {
		return
	}
	err = sql.db.FindFor("store", &thingInfo, "WHERE Name = ?", func() error {
		thingInfos = append(thingInfos, thingInfo)
		return nil
	}, thing)
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
	if !sql.db.CanFind("store", "WHERE Duration = ?", thing.Duration) {
		return false, nil
	}
	err = sql.db.Find("store", &thing, "WHERE Duration = ?", thing.Duration)
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
		return sql.db.Del("store", "WHERE Duration = ?", thingInfo.Duration)
	}
	return sql.db.Insert("store", &thingInfo)
}

// 更新购买次数
func (sql *fishdb) updateBuyTimeFor(uid int64, add int) (err error) {
	sql.Lock()
	defer sql.Unlock()
	userInfo := buffInfo{ID: uid}
	err = sql.db.Create("buff", &userInfo)
	if err != nil {
		return err
	}
	_ = sql.db.Find("buff", &userInfo, "WHERE ID = ?", uid)
	userInfo.BuyTimes += add
	if userInfo.BuyTimes > 20 {
		userInfo.BuyTimes -= 20
		userInfo.Coupon = 3
	}
	return sql.db.Insert("buff", &userInfo)
}

// 使用优惠卷
func (sql *fishdb) useCouponAt(uid int64, times int) (int, error) {
	useTimes := -1
	sql.Lock()
	defer sql.Unlock()
	userInfo := buffInfo{ID: uid}
	err := sql.db.Create("buff", &userInfo)
	if err != nil {
		return useTimes, err
	}
	_ = sql.db.Find("buff", &userInfo, "WHERE ID = ?", uid)
	if userInfo.Coupon > 0 {
		useTimes = math.Min(userInfo.Coupon, times)
		userInfo.Coupon -= useTimes
	}
	return useTimes, sql.db.Insert("buff", &userInfo)
}

// 买卖上限检测
func (sql *fishdb) checkCanSalesFor(uid int64, saleName string, salesNum int) (int, error) {
	sql.Lock()
	defer sql.Unlock()
	userInfo := buffInfo{ID: uid}
	err := sql.db.Create("buff", &userInfo)
	if err != nil {
		return salesNum, err
	}
	_ = sql.db.Find("buff", &userInfo, "WHERE ID = ?", uid)
	if time.Now().Day() != time.Unix(userInfo.Duration, 0).Day() {
		userInfo.Duration = time.Now().Unix()
		userInfo.SalesPole = 0
		userInfo.BuyTing = 0
		err := sql.db.Insert("buff", &userInfo)
		if err != nil {
			return salesNum, err
		}
	}
	if strings.Contains(saleName, "竿") {
		if userInfo.SalesPole >= 10 {
			salesNum = -1
		}
	} else if !checkIsWaste(saleName) {
		maxSales := 30 - userInfo.BuyTing
		if maxSales < 0 {
			salesNum = 0
		}
		if salesNum > maxSales {
			salesNum = maxSales
		}
	}

	return salesNum, err
}

// 更新买卖鱼上限，假定sales变量已经在 checkCanSalesFor 进行了防护
func (sql *fishdb) updateCanSalesFor(uid int64, saleName string, sales int) error {
	sql.Lock()
	defer sql.Unlock()
	userInfo := buffInfo{ID: uid}
	err := sql.db.Create("buff", &userInfo)
	if err != nil {
		return err
	}
	_ = sql.db.Find("buff", &userInfo, "WHERE ID = ?", uid)
	if strings.Contains(saleName, "竿") {
		userInfo.SalesPole++
	} else if !checkIsWaste(saleName) {
		userInfo.BuyTing += sales
	}
	return sql.db.Insert("buff", &userInfo)
}

// 检测物品是否是垃圾
func checkIsWaste(thing string) bool {
	for _, v := range wasteList {
		if v == thing {
			return true
		}
	}
	return false
}
