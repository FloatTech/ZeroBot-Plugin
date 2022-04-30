package recordcombo

import (
	"errors"
	"os"
	"strconv"
	"time"

	_ "github.com/fumiama/sqlite3" // use sql
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
	Number       uint   `gorm:"column:No_ID"`
	ComboName    string `gorm:"column:combo_name"`
	UserName     string `gorm:"column:user_name"`
	UserID       int64  `gorm:"column:user_id"`
	CreateData   string `gorm:"column:create_data"`
	ComboContent string `gorm:"column:combo_content"`
}

// TableName 表名
func (ComboManage) TableName() string {
	return "combo_manage"
}

//添加combo
func (sdb *combodb) addmanage(cname string, uname string, uid int64, content string) error {
	db := (*gorm.DB)(sdb)
	var errors error
	var rows int64
	if err := db.Model(&ComboManage{}).Count(&rows).Error; err != nil {
		errors = errorstofind
	}
	now := time.Now().Format("2006/01/02")
	st := ComboManage{
		Number:       uint(rows) + 1,
		ComboName:    cname,
		UserName:     uname,
		UserID:       uid,
		CreateData:   now,
		ComboContent: content,
	}
	if err := db.Model(&ComboManage{}).Where("combo_name = ?", cname).First(&st).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			if err = db.Debug().Model(&ComboManage{}).Create(&st).Error; err != nil { // newUser not user
				errors = errorsofcreate
			}
		}
	} else {
		errors = errorstoreiterated
	}
	return errors
}

//删除combo
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
			if err := db.Debug().Model(&ComboManage{}).Where("No_ID = ?", index+1).Update(
				map[string]interface{}{
					"No_ID": index,
				}).Error; err != nil {
				errors = errorstoupdate
			}
		}
	}
	return errors
}

//combolist
func (sdb *combodb) managelist() (state []string, err error) {
	db := (*gorm.DB)(sdb)
	var errors error
	var rows int64
	errors = db.Debug().Model(&ComboManage{}).Count(&rows).Error
	row := strconv.FormatInt(int64(rows), 10)
	state = []string{"[Combo List 已收录了" + row + "条]\n"}
	if errors != nil {
		return state, errors
	}
	for {
		if rows == 0 {
			break
		}
		var si ComboManage
		errors = db.Debug().Model(&ComboManage{}).Where("No_ID = ?", rows).First(&si).Error
		rows--
		if errors != nil {
			continue
		}
		state = append(state, si.ComboName)
		state = append(state, ":\n    由")
		state = append(state, si.UserName)
		state = append(state, "(")
		state = append(state, strconv.FormatInt(int64(si.UserID), 10))
		state = append(state, ")创建\n")
	}
	return state, errors
}

//查询combo
func (sdb *combodb) lookupmanage(cname string) (state []string, err error) {
	db := (*gorm.DB)(sdb)
	var errors error
	st := ComboManage{
		ComboName: cname,
	}
	if err := db.Debug().Model(&ComboManage{}).Where("combo_name LIKE  ?", cname).First(&st).Error; err != nil {
		errors = errorstofind
	} else {
		state = []string{"combo名称：" + st.ComboName + "\n"}
		state = append(state, "创建人："+st.UserName+"("+strconv.FormatInt(int64(st.UserID), 10)+")\n")
		state = append(state, "创建时间："+st.CreateData+"\n内容：\n")
		state = append(state, st.ComboContent)
	}
	return state, errors
}
