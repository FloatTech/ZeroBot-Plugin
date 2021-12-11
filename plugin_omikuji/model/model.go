package model

import (
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
)

type OmikujiDB gorm.DB

func Initialize(dbpath string) *OmikujiDB {
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
	gdb.AutoMigrate(&Signature{})
	return (*OmikujiDB)(gdb)
}

func Open(dbpath string) (*OmikujiDB, error) {
	db, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	} else {
		return (*OmikujiDB)(db), nil
	}
}

func (odb *OmikujiDB) Close() error {
	db := (*gorm.DB)(odb)
	return db.Close()
}

type Signature struct {
	gorm.Model
	UserId      int64 `gorm:"column:user_id"`
	SignatureId int64 `gorm:"column:signature_id"`
}

func (Signature) TableName() string {
	return "signature"
}

func (odb *OmikujiDB) GetSignature(userId, signatureId int64) (newSignatureId int64, err error) {
	db := (*gorm.DB)(odb)
	s := Signature{
		UserId:      userId,
		SignatureId: signatureId,
	}
	now := time.Now()
	if err = db.Debug().Model(&Signature{}).Where("user_id = ?", userId).First(&s).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&Signature{}).Create(&s)
			newSignatureId = signatureId
		}
	} else if s.UpdatedAt.Day() != now.Day() {
		err = db.Debug().Model(&Signature{}).Where("user_id = ?", userId).Updates(
			map[string]interface{}{
				"signature_id": signatureId,
			}).Error
		newSignatureId = signatureId
	} else {
		newSignatureId = s.SignatureId
	}
	return newSignatureId, err
}
