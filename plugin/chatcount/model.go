package chatcount

import (
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	chatInterval = 300
)

var (
	// ctdb 聊天时长数据库全局变量
	ctdb *chattimedb
	// 水群提醒时间提醒段，单位分钟
	levelArray = [...]int{15, 30, 60, 120, 240}
)

// chattimedb 聊天时长数据库结构体
type chattimedb gorm.DB

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
	gdb.AutoMigrate(&ChatTime{})
	return (*chattimedb)(gdb)
}

// Close 关闭
func (ctdb *chattimedb) Close() error {
	db := (*gorm.DB)(ctdb)
	return db.Close()
}

// ChatTime 聊天时长，时间的单位都是秒
type ChatTime struct {
	ID        uint      `gorm:"primary_key"`
	GroupID   int64     `gorm:"column:group_id"`
	UserID    int64     `gorm:"column:user_id"`
	LastTime  time.Time `gorm:"column:last_time"`
	TodayTime int64     `gorm:"column:today_time;default:0"`
	TotalTime int64     `gorm:"column:total_time;default:0"`
}

// TableName 表名
func (ChatTime) TableName() string {
	return "chat_time"
}

// sleep 更新发言时间,todayTime的单位是分钟
func (ctdb *chattimedb) updateChatTime(gid, uid int64) (todayTime int64, remindFlag bool) {
	db := (*gorm.DB)(ctdb)
	now := time.Now()
	st := ChatTime{
		GroupID:  gid,
		UserID:   uid,
		LastTime: now,
	}
	if err := db.Model(&ChatTime{}).Where("group_id = ? and user_id = ?", gid, uid).First(&st).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			db.Model(&ChatTime{}).Create(&st)
		}
	} else {
		// 如果不是同一天，把todayTime重置
		if st.LastTime.YearDay() != now.YearDay() {
			db.Model(&ChatTime{}).Where("group_id = ? and user_id = ?", gid, uid).Update(
				map[string]any{
					"last_time":  now,
					"today_time": 0,
				})
		} else {
			userChatTime := int64(now.Sub(st.LastTime).Seconds())
			// 当聊天时间间隔很大的话，则不计入时长
			if userChatTime < chatInterval {
				db.Model(&ChatTime{}).Where("group_id = ? and user_id = ?", gid, uid).Update(
					map[string]any{
						"last_time":  now,
						"today_time": st.TodayTime + userChatTime,
						"total_time": st.TotalTime + userChatTime,
					})
				todayTime = (st.TodayTime + userChatTime) / 60
				remindFlag = getLevel(int(st.TodayTime+userChatTime)/60) > getLevel(int(st.TodayTime/60))
			}else{
				db.Model(&ChatTime{}).Where("group_id = ? and user_id = ?", gid, uid).Update(
					map[string]any{
						"last_time":  now,
					})
			}

		}
	}
	return
}

// getChatTime 获得用户聊天时长,todayTime,totalTime的单位是分钟
func (ctdb *chattimedb) getChatTime(gid, uid int64) (todayTime int64, totalTime int64) {
	db := (*gorm.DB)(ctdb)
	st := ChatTime{}
	db.Model(&ChatTime{}).Where("group_id = ? and user_id = ?", gid, uid).First(&st)
	todayTime = st.TodayTime / 60
	totalTime = st.TotalTime / 60
	return
}

// getChatRank 获得水群排名
func (ctdb *chattimedb) getChatRank(gid int64) (chatTimeList []ChatTime) {
	db := (*gorm.DB)(ctdb)
	db.Model(&ChatTime{}).Where("group_id = ?", gid).Order("today_time DESC").Find(&chatTimeList)
	return
}

// getLevel 用时长判断等级
func getLevel(t int) int {
	for i := len(levelArray) - 1; i >= 0; i-- {
		if t >= levelArray[i] {
			return i + 1
		}
	}
	return 0
}
