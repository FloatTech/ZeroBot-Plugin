// Package mcfish 钓鱼模拟器
package mcfish

import (
	"math/rand"
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

/*
type userInfo struct {
	UserID int64 // QQ
	Bait   int   // 鱼饵
	// 鱼竿 10% 装备/耐久/维修次数/诱钓/眷顾
	WoodenPole  string // 木竿属性 70%
	IronPole    string // 铁竿属性 20%
	GoldenPole  string // 金竿属性 6%
	DiamondPole string // 钻石竿属性 3%
	NetherPole  string // 下界合金竿属性 1%
	// 	鱼 30%
	Cod      int // 鳕鱼数量 70%
	Salmon   int // 鲑鱼数量 20%
	Tropical int // 热带鱼 6%
	Globe    int // 河豚 3%
	Nautilus int // 鹦鹉螺 1%
	// 宝藏 1%
	Induce int // 诱钓 50%
	Favor  int // 眷顾 50%
	// 垃圾 59%
	Leaf    int // 荷叶 10%
	Rod     int // 木棍 10%
	bamboo  int // 竹子 10%
	Shoe    int // 鞋子 10%
	Bottle  int // 瓶子 10%
	Hanger  int // 拌线钩 10%
	Bone    int // 骨头 10%
	Leather int // 皮革 10%
	Carrion int // 腐肉 10%
	Bowl    int // 碗 10%
}
*/

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
}

type store struct {
	Duration int64
	Name     string
	Number   int
	Price    int
	Other    string // 耐久/维修次数/诱钓/眷顾
}

type fishState struct {
	ID       int64
	Duration int64
	Fish     int
}

type storeDiscount struct {
	Name     string
	Discount int
}

var (
	equipAttribute = map[string]int{
		"木竿": 30, "铁竿": 50, "金竿": 70, "钻石竿": 100, "下界合金竿": 150,
	}
	thingPice = map[string]int{
		"鳕鱼": 10, "鲑鱼": 50, "热带鱼": 100, "河豚": 300, "鹦鹉螺": 500, "木竿": 100, "铁竿": 300, "金竿": 700, "钻石竿": 1500, "下界合金竿": 3100, "诱钓": 1000, "海之眷顾": 2500,
	}
	discount     = make(map[string]int, 10)
	wasteList    = []string{"唱片", "木棍", "帽子", "鞋子", "瓶子", "拌线钩", "骨头", "皮革", "腐肉", "碗"}
	enchantLevel = []string{"0", "Ⅰ", "Ⅱ", "Ⅲ"}
	dbdata       = &fishdb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("mcfish", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "钓鱼",
		Help: "一款钓鱼模拟器\n----------指令----------\n" +
			"- 钓鱼看板\n- 购买xxx\n- 购买xxx [数量]\n- 出售xxx\n- 出售xxx [数量]\n" +
			"- 钓鱼背包\n- 装备xx竿\n- 附魔鱼竿[诱钓|海之眷顾]\n- 修复鱼竿\n" +
			"- 进行钓鱼\n" +
			"注:\n1.每日的商店价格是波动的!!如何最大化收益自己考虑一下喔\n" +
			"2.装备信息:\n-> 木竿 : 耐久上限:30 均价:100 上钩概率:7%\n-> 铁竿 : 耐久上限:50 均价:300 上钩概率:2%\n" +
			"-> 金竿 : 耐久上限:70 均价700 上钩概率:0.6%\n-> 钻石竿 : 耐久上限:100 均价1500 上钩概率:0.3%\n-> 下界合金竿 : 耐久上限:150 均价3100 上钩概率:0.1%\n" +
			"3.附魔书信息:\n-> 诱钓 : 减少上钩时间.均价:1000, 上钩概率:0.5%\n-> 海之眷顾 : 增加宝藏上钩概率.均价:2500, 上钩概率:0.5%\n" +
			"3.鱼类信息:\n-> 鳕鱼 : 均价:10 上钩概率:21%\n-> 鲑鱼 : 均价:50 上钩概率:6%\n-> 热带鱼 : 均价:100 上钩概率:1.8%\n-> 河豚 : 均价:300 上钩概率:0.9%\n-> 鹦鹉螺 : 均价:500 上钩概率:0.3%",
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
	fishLimit = 50 // 钓鱼次数上限
)

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
	err = sql.db.FindFor(strconv.FormatInt(uid, 10)+"Pack", &userInfo, "group by Duration", func() error {
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
	err = sql.db.FindFor("store", &thingInfo, "group by Duration", func() error {
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

// 更新上限信息
func (sql *fishdb) updateFishInfo(uid int64) (ok bool, err error) {
	ok = true
	sql.Lock()
	defer sql.Unlock()
	userInfo := fishState{ID: uid}
	err = sql.db.Create("fishState", &userInfo)
	if err != nil {
		return false, err
	}
	_ = sql.db.Find("fishState", &userInfo, "where ID = "+strconv.FormatInt(uid, 10))
	if time.Unix(userInfo.Duration, 0).Day() != time.Now().Day() {
		userInfo.Fish = 0
		userInfo.Duration = time.Now().Unix()
	}
	if userInfo.Fish > fishLimit {
		ok = false
	}
	userInfo.Fish++
	return ok, sql.db.Insert("fishState", &userInfo)
}

// 更新上限信息
func (sql *fishdb) refreshStroeInfo() (ok bool, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("stroeDiscount", &storeDiscount{})
	if err != nil {
		return false, err
	}
	lastTime := storeDiscount{}
	_ = sql.db.Find("stroeDiscount", &lastTime, "where Nme = 'lastTime'")
	refresh := false
	if time.Now().Day() != lastTime.Discount {
		lastTime = storeDiscount{
			Name:     "lastTime",
			Discount: time.Now().Day(),
		}
		err = sql.db.Insert("stroeDiscount", &lastTime)
		if err != nil {
			return false, err
		}
		refresh = true
	}
	for name := range thingPice {
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
			// 每日上架1鱼竿
			durable := rand.Intn(30) + 1
			maintenance := rand.Intn(10)
			induceLevel := rand.Intn(2)
			favorLevel := rand.Intn(2)
			pice := thingPice["木竿"] - (equipAttribute["木竿"] - durable) - maintenance*2 + induceLevel*1000 + favorLevel*2500
			pole := store{
				Duration: time.Now().Unix(),
				Name:     "木竿",
				Number:   1,
				Price:    pice,
				Other:    strconv.Itoa(durable) + "/" + strconv.Itoa(maintenance) + "/" + strconv.Itoa(induceLevel) + "/" + strconv.Itoa(favorLevel),
			}
			err = sql.db.Insert("stroeDiscount", &pole)
			if err != nil {
				return
			}
		default:
			err = sql.db.Find("stroeDiscount", &thing, "where Name = '"+name+"'")
			if err != nil {
				return
			}
		}
		if thing.Discount != 0 {
			discount[name] = thing.Discount
		} else {
			discount[name] = 100
		}
	}
	return true, nil
}
