package chatcount

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RomiChan/syncx"

	"github.com/jinzhu/gorm"
)

const (
	chatInterval = 300
)

var (
	// ctdb 聊天时长数据库全局变量
	ctdb *chattimedb
	// l 水群提醒时间提醒段，单位分钟
	l = newLeveler(60, 120, 180, 240, 300)
)

// chattimedb 聊天时长数据库结构体
type chattimedb struct {
	// ctdb.userTimestampMap 每个人发言的时间戳 key=groupID_userID
	userTimestampMap syncx.Map[string, int64]
	// ctdb.userTodayTimeMap 每个人今日水群时间 key=groupID_userID
	userTodayTimeMap syncx.Map[string, int64]
	// ctdb.userTodayMessageMap 每个人今日水群次数 key=groupID_userID
	userTodayMessageMap syncx.Map[string, int64]
	// db 数据库
	db *gorm.DB
	// chatmu 读写添加锁
	chatmu sync.Mutex
}

// initialize 初始化
func initialize(dbpath string) *chattimedb {
	var err error
	if _, err = os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil
		}
		defer f.Close()
	}
	gdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		panic(err)
	}
	gdb.AutoMigrate(&chatTime{})
	return &chattimedb{
		db: gdb,
	}
}

// Close 关闭
func (ctdb *chattimedb) Close() error {
	db := ctdb.db
	return db.Close()
}

// chatTime 聊天时长，时间的单位都是秒
type chatTime struct {
	ID           uint  `gorm:"primary_key"`
	GroupID      int64 `gorm:"column:group_id"`
	UserID       int64 `gorm:"column:user_id"`
	TodayTime    int64 `gorm:"-"`
	TodayMessage int64 `gorm:"-"`
	TotalTime    int64 `gorm:"column:total_time;default:0"`
	TotalMessage int64 `gorm:"column:total_message;default:0"`
}

// TableName 表名
func (chatTime) TableName() string {
	return "chat_time"
}

// updateChatTime 更新发言时间,todayTime的单位是分钟
func (ctdb *chattimedb) updateChatTime(gid, uid int64) (remindTime int64, remindFlag bool) {
	ctdb.chatmu.Lock()
	defer ctdb.chatmu.Unlock()
	db := ctdb.db
	now := time.Now()
	keyword := fmt.Sprintf("%v_%v", gid, uid)
	ts, ok := ctdb.userTimestampMap.Load(keyword)
	if !ok {
		ctdb.userTimestampMap.Store(keyword, now.Unix())
		ctdb.userTodayMessageMap.Store(keyword, 1)
		return
	}
	lastTime := time.Unix(ts, 0)
	todayTime, _ := ctdb.userTodayTimeMap.Load(keyword)
	totayMessage, _ := ctdb.userTodayMessageMap.Load(keyword)
	// 这个消息数是必须统计的
	ctdb.userTodayMessageMap.Store(keyword, totayMessage+1)
	st := chatTime{
		GroupID:      gid,
		UserID:       uid,
		TotalTime:    todayTime,
		TotalMessage: totayMessage,
	}

	// 如果不是同一天，把TotalTime,TotalMessage重置
	if lastTime.YearDay() != now.YearDay() {
		if err := db.Model(&st).Where("group_id = ? and user_id = ?", gid, uid).First(&st).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				db.Model(&st).Create(&st)
			}
		} else {
			db.Model(&st).Where("group_id = ? and user_id = ?", gid, uid).Update(
				map[string]any{
					"total_time":    st.TotalTime + todayTime,
					"total_message": st.TotalMessage + totayMessage,
				})
		}
		ctdb.userTimestampMap.Store(keyword, now.Unix())
		ctdb.userTodayTimeMap.Delete(keyword)
		ctdb.userTodayMessageMap.Delete(keyword)
		return
	}

	userChatTime := int64(now.Sub(lastTime).Seconds())
	// 当聊天时间在一定范围内的话，则计入时长
	if userChatTime < chatInterval {
		ctdb.userTodayTimeMap.Store(keyword, todayTime+userChatTime)
		remindTime = (todayTime + userChatTime) / 60
		remindFlag = l.level(int((todayTime+userChatTime)/60)) > l.level(int(todayTime/60))
	}
	ctdb.userTimestampMap.Store(keyword, now.Unix())
	return
}

// getChatTime 获得用户聊天时长和消息次数,todayTime,totalTime的单位是秒,todayMessage,totalMessage单位是条数
func (ctdb *chattimedb) getChatTime(gid, uid int64) (todayTime, todayMessage, totalTime, totalMessage int64) {
	ctdb.chatmu.Lock()
	defer ctdb.chatmu.Unlock()
	db := ctdb.db
	st := chatTime{}
	db.Model(&st).Where("group_id = ? and user_id = ?", gid, uid).First(&st)
	keyword := fmt.Sprintf("%v_%v", gid, uid)
	todayTime, _ = ctdb.userTodayTimeMap.Load(keyword)
	todayMessage, _ = ctdb.userTodayMessageMap.Load(keyword)
	totalTime = st.TotalTime
	totalMessage = st.TotalMessage
	return
}

// getChatRank 获得水群排名，时间单位为秒
func (ctdb *chattimedb) getChatRank(gid int64) (chatTimeList []chatTime) {
	ctdb.chatmu.Lock()
	defer ctdb.chatmu.Unlock()
	chatTimeList = make([]chatTime, 0, 100)
	keyList := make([]string, 0, 100)
	ctdb.userTimestampMap.Range(func(key string, value int64) bool {
		t := time.Unix(value, 0)
		if strings.Contains(key, strconv.FormatInt(gid, 10)) && t.YearDay() == time.Now().YearDay() {
			keyList = append(keyList, key)
		}
		return true
	})
	for _, v := range keyList {
		_, a, _ := strings.Cut(v, "_")
		uid, _ := strconv.ParseInt(a, 10, 64)
		todayTime, _ := ctdb.userTodayTimeMap.Load(v)
		todayMessage, _ := ctdb.userTodayMessageMap.Load(v)
		chatTimeList = append(chatTimeList, chatTime{
			GroupID:      gid,
			UserID:       uid,
			TodayTime:    todayTime,
			TodayMessage: todayMessage,
		})
	}
	sort.Sort(sortChatTime(chatTimeList))
	return
}

// leveler 结构体，包含一个 levelArray 字段
type leveler struct {
	levelArray []int
}

// newLeveler 构造函数，用于创建 Leveler 实例
func newLeveler(levels ...int) *leveler {
	return &leveler{
		levelArray: levels,
	}
}

// level 方法，封装了 getLevel 函数的逻辑
func (l *leveler) level(t int) int {
	for i := len(l.levelArray) - 1; i >= 0; i-- {
		if t >= l.levelArray[i] {
			return i + 1
		}
	}
	return 0
}

// sortChatTime chatTime排序数组
type sortChatTime []chatTime

// Len 实现 sort.Interface
func (a sortChatTime) Len() int {
	return len(a)
}

// Less 实现 sort.Interface，按 TodayTime 降序，TodayMessage 降序
func (a sortChatTime) Less(i, j int) bool {
	if a[i].TodayTime == a[j].TodayTime {
		return a[i].TodayMessage > a[j].TodayMessage
	}
	return a[i].TodayTime > a[j].TodayTime
}

// Swap 实现 sort.Interface
func (a sortChatTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
