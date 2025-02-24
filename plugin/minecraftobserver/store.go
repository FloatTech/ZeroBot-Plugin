package minecraftobserver

import (
	"errors"
	"fmt"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"os"
	"sync"
	"time"
)

const (
	dbPath = "minecraft_observer"

	targetTypeGroup = 1
	targetTypeUser  = 2
)

var (
	// 数据库连接失败
	errDBConn = errors.New("数据库连接失败")
	// 参数错误
	errParam = errors.New("参数错误")
)

type db struct {
	sdb  *gorm.DB
	lock sync.RWMutex
}

// initializeDB 初始化数据库
func initializeDB(dbpath string) error {
	if _, err := os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	gdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		logrus.Errorln(logPrefix+"initializeDB ERROR: ", err)
		return err
	}
	gdb.AutoMigrate(&ServerStatus{}, &ServerSubscribe{})
	dbInstance = &db{
		sdb:  gdb,
		lock: sync.RWMutex{},
	}
	return nil
}

var (
	// dbInstance 数据库实例
	dbInstance *db
	// 开启并检查数据库链接
	getDB = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		var err error
		err = initializeDB(engine.DataFolder() + dbPath)
		if err != nil {
			logrus.Errorln(logPrefix+"initializeDB ERROR: ", err)
			ctx.SendChain(message.Text("[mc-ob] ERROR: ", err))
			return false
		}
		return true
	})
)

// 通过群组id和服务器地址获取状态
func (d *db) getServerStatus(addr string) (*ServerStatus, error) {
	if d == nil {
		return nil, errDBConn
	}
	var ss ServerStatus
	if err := d.sdb.Model(&ss).Where("server_addr = ?", addr).First(&ss).Error; err != nil {
		logrus.Errorln(logPrefix+"getServerStatus ERROR: ", err)
		return nil, err
	}
	return &ss, nil
}

// 更新服务器状态
func (d *db) updateServerStatus(ss *ServerStatus) (err error) {
	if d == nil {
		return errDBConn
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if ss == nil {
		return errors.New("参数错误")
	}
	ss.LastUpdate = time.Now().Unix()
	ss2 := ss.DeepCopy()
	if err = d.sdb.Where(&ServerStatus{ServerAddr: ss.ServerAddr}).Assign(ss2).FirstOrCreate(ss).Debug().Error; err != nil {
		logrus.Errorln(logPrefix, fmt.Sprintf("updateServerStatus %v ERROR: %v", ss, err))
		return
	}
	return
}

func (d *db) delServerStatus(addr string) (err error) {
	if err = d.sdb.Model(&ServerStatus{}).Delete(&ServerStatus{}).Where("server_addr = ?", addr).Error; err != nil {
		logrus.Errorln(logPrefix+"deleteSubscribe ERROR: ", err)
		return
	}
	return
}

// 新增订阅
func (d *db) newSubscribe(addr string, targetID, targetType int64) (err error) {
	if d == nil {
		return errDBConn
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if targetID == 0 || (targetType != 1 && targetType != 2) {
		logrus.Errorln(logPrefix+"newSubscribe ERROR: 参数错误 ", targetID, " ", targetType)
		return errParam
	}
	ss := &ServerSubscribe{
		ServerAddr: addr,
		TargetID:   targetID,
		TargetType: targetType,
		LastUpdate: time.Now().Unix(),
	}
	if err = d.sdb.Model(&ss).Create(ss).Error; err != nil {
		logrus.Errorln(logPrefix+"newSubscribe ERROR: ", err)
		return
	}
	return
}

// 删除订阅
func (d *db) deleteSubscribe(addr string, targetID int64, targetType int64) (err error) {
	if d == nil {
		return errDBConn
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if addr == "" || targetID == 0 {
		return errParam
	}
	if err = d.sdb.Model(&ServerSubscribe{}).Delete(&ServerSubscribe{}).Where("server_addr = ? and target_id = ? and target_type = ?", addr, targetID, targetType).Error; err != nil {
		logrus.Errorln(logPrefix+"deleteSubscribe ERROR: ", err)
		return
	}

	// 扫描是否还有订阅，如果没有则删除服务器状态
	var cnt int
	err = d.sdb.Model(&ServerSubscribe{}).Where("server_addr = ?", addr).Count(&cnt).Error
	if err != nil {
		logrus.Errorln(logPrefix+"deleteSubscribe ERROR: ", err)
		return
	}
	if cnt == 0 {
		dErr := d.delServerStatus(addr)
		if dErr != nil {
			logrus.Errorln(logPrefix+"deleteSubscribe-delServerStatus ERROR: ", dErr)
		}
	}
	return
}

// 获取所有订阅
func (d *db) getAllSubscribes() (subs []ServerSubscribe, err error) {
	if d == nil {
		return nil, errDBConn
	}
	subs = []ServerSubscribe{}
	if err = d.sdb.Model(&subs).Find(&subs).Error; err != nil {
		logrus.Errorln(logPrefix+"getAllSubscribes ERROR: ", err)
		return
	}
	return
}
