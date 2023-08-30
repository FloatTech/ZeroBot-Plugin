// Package service 依赖服务 数据库
package service

import (
	"os"

	"github.com/jinzhu/gorm"
)

var chessDB *gorm.DB

// ELO user elo info
type ELO struct {
	gorm.Model
	Uin  int64 `gorm:"unique_index"`
	Name string
	Rate int
}

// PGN chess pgn info
type PGN struct {
	gorm.Model
	Data      string
	WhiteUin  int64
	BlackUin  int64
	WhiteName string
	BlackName string
}

// DBService 数据库服务
type DBService struct {
	db *gorm.DB
}

// NewDBService 创建数据库服务
func NewDBService() *DBService {
	return &DBService{
		db: chessDB,
	}
}

// InitDatabase init database
func InitDatabase(dbPath string) {
	var err error
	if _, err = os.Stat(dbPath); err != nil || os.IsNotExist(err) {
		f, err := os.Create(dbPath)
		if err != nil {
			panic(err)
		}
		defer f.Close()
	}
	chessDB, err = gorm.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	chessDB.AutoMigrate(&ELO{}, &PGN{})
}

// CreateELO 创建 ELO
func (s *DBService) CreateELO(uin int64, name string, rate int) error {
	return s.db.Create(&ELO{
		Uin:  uin,
		Name: name,
		Rate: rate,
	}).Error
}

// GetELOByUin 获取 ELO
func (s *DBService) GetELOByUin(uin int64) (ELO, error) {
	var elo ELO
	err := s.db.Where("uin = ?", uin).First(&elo).Error
	return elo, err
}

// GetELORateByUin 获取 ELO 等级分
func (s *DBService) GetELORateByUin(uin int64) (int, error) {
	var elo ELO
	err := s.db.Select("rate").Where("uin = ?", uin).First(&elo).Error
	return elo.Rate, err
}

// GetHighestRateList 获取最高的等级分列表
func (s *DBService) GetHighestRateList() ([]ELO, error) {
	var eloList []ELO
	err := s.db.Order("rate desc").Limit(10).Find(&eloList).Error
	return eloList, err
}

// UpdateELOByUin 更新 ELO 等级分
func (s *DBService) UpdateELOByUin(uin int64, name string, rate int) error {
	return s.db.Model(&ELO{}).Where("uin = ?", uin).Update("name", name).Update("rate", rate).Error
}

// CleanELOByUin 清空用户 ELO 等级分
func (s *DBService) CleanELOByUin(uin int64) error {
	return s.db.Model(&ELO{}).Where("uin = ?", uin).Update("rate", 100).Error
}

// CreatePGN 创建 PGN
func (s *DBService) CreatePGN(data string, whiteUin int64, blackUin int64, whiteName string, blackName string) error {
	return s.db.Create(&PGN{
		Data:      data,
		WhiteUin:  whiteUin,
		BlackUin:  blackUin,
		WhiteName: whiteName,
		BlackName: blackName,
	}).Error
}
