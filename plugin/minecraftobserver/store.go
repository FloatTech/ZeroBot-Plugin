package minecraftobserver

import (
	"errors"
	"os"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/jinzhu/gorm"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
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
	sdb           *gorm.DB
	statusLock    sync.RWMutex
	subscribeLock sync.RWMutex
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
		// logrus.Errorln(logPrefix+"initializeDB ERROR: ", err)
		return err
	}
	gdb.AutoMigrate(&serverStatus{}, &serverSubscribe{})
	dbInstance = &db{
		sdb:           gdb,
		statusLock:    sync.RWMutex{},
		subscribeLock: sync.RWMutex{},
	}
	return nil
}

var (
	// dbInstance 数据库实例
	dbInstance *db
	// 开启并检查数据库链接
	getDB = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		var err = initializeDB(engine.DataFolder() + dbPath)
		if err != nil {
			// logrus.Errorln(logPrefix+"initializeDB ERROR: ", err)
			ctx.SendChain(message.Text("[mc-ob] ERROR: ", err))
			return false
		}
		return true
	})
)

// 通过群组id和服务器地址获取状态
func (d *db) getServerStatus(addr string) (*serverStatus, error) {
	if d == nil {
		return nil, errDBConn
	}
	if addr == "" {
		return nil, errParam
	}
	var ss serverStatus
	if err := d.sdb.Model(&ss).Where("server_addr = ?", addr).First(&ss).Error; err != nil {
		// logrus.Errorln(logPrefix+"getServerStatus ERROR: ", err)
		return nil, err
	}
	return &ss, nil
}

// 更新服务器状态
func (d *db) updateServerStatus(ss *serverStatus) (err error) {
	if d == nil {
		return errDBConn
	}
	d.statusLock.Lock()
	defer d.statusLock.Unlock()
	if ss == nil || ss.ServerAddr == "" {
		return errParam
	}
	ss.LastUpdate = time.Now().Unix()
	ss2 := ss.deepCopy()
	if err = d.sdb.Where(&serverStatus{ServerAddr: ss.ServerAddr}).Assign(ss2).FirstOrCreate(ss).Debug().Error; err != nil {
		// logrus.Errorln(logPrefix, fmt.Sprintf("updateServerStatus %v ERROR: %v", ss, err))
		return
	}
	return
}

func (d *db) delServerStatus(addr string) (err error) {
	if d == nil {
		return errDBConn
	}
	if addr == "" {
		return errParam
	}
	d.statusLock.Lock()
	defer d.statusLock.Unlock()
	if err = d.sdb.Where("server_addr = ?", addr).Delete(&serverStatus{}).Error; err != nil {
		// logrus.Errorln(logPrefix+"deleteSubscribe ERROR: ", err)
		return
	}
	return
}

// 新增订阅
func (d *db) newSubscribe(addr string, targetID, targetType int64) (err error) {
	if d == nil {
		return errDBConn
	}
	if targetID == 0 || (targetType != 1 && targetType != 2) {
		// logrus.Errorln(logPrefix+"newSubscribe ERROR: 参数错误 ", targetID, " ", targetType)
		return errParam
	}
	d.subscribeLock.Lock()
	defer d.subscribeLock.Unlock()
	// 如果已经存在，需要报错
	existedRec := &serverSubscribe{}
	err = d.sdb.Model(&serverSubscribe{}).Where("server_addr = ? and target_id = ? and target_type = ?", addr, targetID, targetType).First(existedRec).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		// logrus.Errorln(logPrefix+"newSubscribe ERROR: ", err)
		return
	}
	if existedRec.ID != 0 {
		return errors.New("已经存在的订阅")
	}
	ss := &serverSubscribe{
		ServerAddr: addr,
		TargetID:   targetID,
		TargetType: targetType,
		LastUpdate: time.Now().Unix(),
	}
	if err = d.sdb.Model(&ss).Create(ss).Error; err != nil {
		// logrus.Errorln(logPrefix+"newSubscribe ERROR: ", err)
		return
	}
	return
}

// 删除订阅
func (d *db) deleteSubscribe(addr string, targetID int64, targetType int64) (err error) {
	if d == nil {
		return errDBConn
	}
	if addr == "" || targetID == 0 || targetType == 0 {
		return errParam
	}
	d.subscribeLock.Lock()
	defer d.subscribeLock.Unlock()
	// 检查是否存在
	if err = d.sdb.Model(&serverSubscribe{}).Where("server_addr = ? and target_id = ? and target_type = ?", addr, targetID, targetType).First(&serverSubscribe{}).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("未找到订阅")
		}
		// logrus.Errorln(logPrefix+"deleteSubscribe ERROR: ", err)
		return
	}

	if err = d.sdb.Where("server_addr = ? and target_id = ? and target_type = ?", addr, targetID, targetType).Delete(&serverSubscribe{}).Error; err != nil {
		// logrus.Errorln(logPrefix+"deleteSubscribe ERROR: ", err)
		return
	}

	// 扫描是否还有订阅，如果没有则删除服务器状态
	var cnt int
	err = d.sdb.Model(&serverSubscribe{}).Where("server_addr = ?", addr).Count(&cnt).Error
	if err != nil {
		// logrus.Errorln(logPrefix+"deleteSubscribe ERROR: ", err)
		return
	}
	if cnt == 0 {
		_ = d.delServerStatus(addr)
	}
	return
}

// 获取所有订阅
func (d *db) getAllSubscribes() (subs []serverSubscribe, err error) {
	if d == nil {
		return nil, errDBConn
	}
	subs = []serverSubscribe{}
	if err = d.sdb.Find(&subs).Error; err != nil {
		// logrus.Errorln(logPrefix+"getAllSubscribes ERROR: ", err)
		return
	}
	return
}

// 获取渠道对应的订阅列表
func (d *db) getSubscribesByTarget(targetID, targetType int64) (subs []serverSubscribe, err error) {
	if d == nil {
		return nil, errDBConn
	}
	subs = []serverSubscribe{}
	if err = d.sdb.Model(&serverSubscribe{}).Where("target_id = ? and target_type = ?", targetID, targetType).Find(&subs).Error; err != nil {
		// logrus.Errorln(logPrefix+"getSubscribesByTarget ERROR: ", err)
		return
	}
	return
}
