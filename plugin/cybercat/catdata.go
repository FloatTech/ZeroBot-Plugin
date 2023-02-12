// Package cybercat 云养猫
package cybercat

import (
	"sort"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/web"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var catType = map[string]string{
	"Abyssinian": "阿比西尼亚猫", "American Shorthairs": "美洲短毛猫", "American Wirehair": "美国硬毛猫", "Bengal": "孟加拉虎", "Bombay": "孟买猫", "Cyprus": "塞浦路斯猫",
	"British Shorthair": "英国短毛猫", "British Longhair": "英国长毛猫", "European Burmese": "缅甸猫", "Chartreux": "夏特鲁斯猫", "Cornish Rex": "康沃尔-雷克斯猫",
	"Egyptian Mau": "埃及猫", "Japanese Bobtail": "日本短尾猫", "Korat": "呵叻猫", "Manx": "马恩岛猫", "Ocicat": "欧西猫", "Nebelung": "内华达猫", "Cymric": "威尔士猫",
	"Oriental Shorthair": "东方短毛猫", "Pixie-bob": "北美洲短毛猫", "Ragamuffin": "褴褛猫", "Russian Blue": "俄罗斯蓝猫", "Scottish Fold": "苏格兰折耳猫", "Siamese": "暹罗猫",
	"Tonkinese": "东京猫", "Balinese": "巴厘岛猫", "Birman": "比尔曼猫", "Cymic": "金力克长毛猫", "Himalayan": "喜马拉雅猫", "Munchkin": "曼基康猫", "Colorpoint Shorthair": "重点色短毛猫",
	"Javanese": "爪哇猫", "Maine Coon": "缅因猫", "Norwegian Forest Cat": "挪威森林猫", "Persian": "波斯猫", "Ragdoll": "布偶猫", "Malayan": "马来猫", "Cheetoh": "奇多猫",
	"Somali": "索马里猫", "Turkish Angora": "土耳其安哥拉猫", "Savannah": "沙凡那猫", "Selkirk Rex": "塞尔凯克卷毛猫", "Siberian": "西伯利亚猫", "Donskoy": "顿斯科伊猫", "Burmilla": "博美拉猫",
	"Singapura": "新加坡猫", "Snowshoe": "雪鞋猫", "Toyger": "玩具虎猫", "Turkish Van": "土耳其梵猫", "York Chocolate": "约克巧克力猫", "Dragon Li": "中国狸花猫",
	"Exotic Shorthair": "异国短毛猫", "Havana Brown": "哈瓦那褐猫", "Khao Manee": "泰国御猫", "Kurilian": "千岛短尾猫", "LaPerm": "拉邦猫", "Arabian Mau": "美英猫",
	"Bambino": "班比诺猫", "Devon Rex": "德文狸猫", "California Spangled": "加州闪亮猫", "Sphynx": "斯芬克斯猫", "Chantilly-Tiffany": "查达利/蒂法尼猫", "Chausie": "非洲狮子猫",
	"American Curl": "美国卷耳猫", "Aegean": "爱琴猫", "American Bobtail": "美国短尾猫", "Australian Mist": "澳大利亚雾猫", "Nekomusume": "猫娘", "": "未知品种"}

type catdb struct {
	db *sql.Sqlite
	sync.RWMutex
}

type catInfo struct {
	User      int64   // 主人
	Name      string  // 喵喵名称
	Type      string  // 品种
	Satiety   float64 // 饱食度
	Mood      int     // 心情
	Weight    float64 // 体重
	LastTime  int64   // 上次喂养时间
	Work      int64   // 打工时间
	ArenaTime int64   // 上次PK时间
	Food      float64 // 食物数量
	Picurl    string  // 猫猫图片
}

var (
	catdata = &catdb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("cybercat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "云养猫",
		Help: "一款既能能赚钱(?)又能看猫的养成类插件\n\n" +
			"- 吸猫\n(随机返回一只猫)\n- 买猫\n- 买猫粮\n- 买n袋猫粮\n- 喂猫\n- 喂猫n斤猫粮\n" +
			"- 猫猫打工\n- 猫猫打工[1-9]小时\n- 猫猫状态\n- 喵喵改名叫xxx\n" +
			"- 喵喵pk@对方QQ\n- 猫猫排行榜\n" +
			"Tips: 打工期间的猫猫无法喂养哦",
		PrivateDataFolder: "cybercat",
	}).ApplySingle(ctxext.DefaultSingle)
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		catdata.db.DBPath = engine.DataFolder() + "catdata.db"
		err := catdata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	})
)

func init() {
	engine.OnFullMatch("吸猫").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		typeOfcat, url := getPicURL()
		if url == "" {
			ctx.SendChain(message.Text("[ERROR]: 404"))
			return
		}
		ctx.SendChain(message.Text(catType[typeOfcat]), message.Image(url))
	})
}

func getPicURL() (catType, url string) {
	data, _ := web.GetData("https://api.thecatapi.com/v1/images/search")
	if data == nil {
		return
	}
	picID := gjson.ParseBytes(data).Get("0.id").String()
	picdata, _ := web.GetData("https://api.thecatapi.com/v1/images/" + picID)
	if picdata == nil {
		return
	}
	return gjson.ParseBytes(picdata).Get("breeds.0.name").String(), gjson.ParseBytes(picdata).Get("url").String()
}

func (sql *catdb) insert(gid string, dbInfo catInfo) error {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create(gid, &catInfo{})
	if err != nil {
		return err
	}
	return sql.db.Insert(gid, &dbInfo)
}

func (sql *catdb) find(gid, uid string) (dbInfo catInfo, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create(gid, &catInfo{})
	if err != nil {
		return
	}
	if !sql.db.CanFind(gid, "where user = "+uid) {
		return catInfo{}, nil // 规避没有该用户数据的报错
	}
	err = sql.db.Find(gid, &dbInfo, "where user = "+uid)
	return
}

func (sql *catdb) del(gid, uid string) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Del(gid, "where user = "+uid)
}

func (sql *catdb) delcat(gid, uid string) error {
	sql.Lock()
	defer sql.Unlock()
	dbInfo := catInfo{}
	_ = sql.db.Find(gid, &dbInfo, "where user = "+uid)
	newInfo := catInfo{
		User: dbInfo.User,
		Food: dbInfo.Food,
	}
	return sql.db.Insert(gid, &newInfo)
}

type catDataList []catInfo

func (s catDataList) Len() int {
	return len(s)
}
func (s catDataList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s catDataList) Less(i, j int) bool {
	return s[i].Weight > s[j].Weight
}
func (sql *catdb) getGroupdata(gid string) (list catDataList, err error) {
	sql.RLock()
	defer sql.RUnlock()
	info := catInfo{}
	err = sql.db.FindFor(gid, &info, "group by Weight", func() error {
		if info.Name != "" {
			list = append(list, info)
		}
		return nil
	})
	if len(list) > 1 {
		sort.Sort(list)
	}
	return
}
