package recordcombo

import (
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	errorsofcreate     = errors.New("创建combo列表失败！")
	errorstofind       = errors.New("查找combo失败！")
	errorstoreiterated = errors.New("combo重复！")
	errorsfornumber    = errors.New("操作数量返回值返回为空！")
	errorstoupdate     = errors.New("更新数据库失败！")
)

// sdb
var sdb *combodb

type combodb gorm.DB

// initialize 初始化
func initialize(dbpath string) *combodb {
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
	gdb.AutoMigrate(&ComboManage{})
	return (*combodb)(gdb)
}

// Close 关闭
func (sdb *combodb) Close() error {
	db := (*gorm.DB)(sdb)
	return db.Close()
}

// ComboManage combo信息
type ComboManage struct {
	Number       int64  `gorm:"column:No_ID"`
	ComboName    string `gorm:"column:combo_name"`
	CreateID     int64  `gorm:"column:create_id"`
	UserID       int64  `gorm:"column:user_id"`
	GroupID      int64  `gorm:"column:group_id"`
	CreateData   string `gorm:"column:create_data"`
	ComboContent string `gorm:"column:combo_content"`
}

// TableName 表名
func (ComboManage) TableName() string {
	return "combo_manage"
}

// 添加combo
func (sdb *combodb) addmanage(cname string, cid, uid, gid int64, content string) error {
	db := (*gorm.DB)(sdb)
	var errors error
	var rows int64
	if err := db.Model(&ComboManage{}).Count(&rows).Error; err != nil {
		errors = errorstofind
	}
	now := time.Now().Format("2006/01/02")
	st := ComboManage{
		Number:       rows,
		ComboName:    cname,
		CreateID:     cid,
		UserID:       uid,
		GroupID:      gid,
		CreateData:   now,
		ComboContent: content,
	}
	if err := db.Model(&ComboManage{}).Where("combo_name = ?", cname).First(&st).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			if err = db.Model(&ComboManage{}).Create(&st).Error; err != nil { // newUser not user
				errors = errorsofcreate
			}
		}
	} else {
		errors = errorstoreiterated
	}
	return errors
}

// 删除combo
func (sdb *combodb) removemanage(cname string) error {
	db := (*gorm.DB)(sdb)
	var errors error
	st := ComboManage{
		ComboName: cname,
	}
	var rows int64
	if err := db.Model(&ComboManage{}).Count(&rows).Error; err != nil {
		errors = errorstofind
	}
	if err := db.Model(&ComboManage{}).Where("combo_name = ?", cname).First(&st).Error; err != nil {
		errors = errorstofind
	} else {
		if err := db.Model(&ComboManage{}).Where("combo_name = ?", cname).Delete(&st).Error; err != nil {
			errors = err
		}
		for index := st.Number; int64(index) < rows; index++ {
			if err := db.Model(&ComboManage{}).Where("No_ID = ?", index+1).Update(
				map[string]interface{}{
					"No_ID": index,
				}).Error; err != nil {
				errors = errorstoupdate
			}
		}
	}
	return errors
}

// combolist
func (sdb *combodb) managelist() (rows int, state map[int]ComboManage, err error) {
	db := (*gorm.DB)(sdb)
	err = db.Model(&ComboManage{}).Count(&rows).Error
	if err != nil || rows == 0 {
		return
	}
	var si ComboManage
	state = make(map[int]ComboManage, rows)
	for row := 0; row < rows; row++ {
		err = db.Model(&ComboManage{}).Where("No_ID = ?", row).First(&si).Error
		if err != nil {
			continue
		}
		state[row] = si
	}
	return
}

// 查询combo
func (sdb *combodb) lookupmanage(cname string) (state ComboManage, err error) {
	db := (*gorm.DB)(sdb)
	err = db.Model(&ComboManage{}).Where("combo_name LIKE  ?", cname).First(&state).Error
	return
}

func (sdb *combodb) randinfo() (state ComboManage, err error) {
	db := (*gorm.DB)(sdb)
	var rows int64
	err = db.Model(&ComboManage{}).Count(&rows).Error
	if err != nil || rows == 0 {
		return state, errorsfornumber
	}
	err = db.Model(&ComboManage{}).Where("No_ID = ?", rand.Intn(int(rows))).First(&state).Error
	return
}
