package chess

import (
	"os"

	"github.com/jinzhu/gorm"
)

var chessDB *gorm.DB

// elo user elo info
type elo struct {
	gorm.Model
	Uin  int64 `gorm:"unique_index"`
	Name string
	Rate int
}

// pgn chess pgn info
type pgn struct {
	gorm.Model
	Data      string
	WhiteUin  int64
	BlackUin  int64
	WhiteName string
	BlackName string
}

// chessDBService 数据库服务
type chessDBService struct {
	db *gorm.DB
}

// newDBService 创建数据库服务
func newDBService() *chessDBService {
	return &chessDBService{
		db: chessDB,
	}
}

// initDatabase init database
func initDatabase(dbPath string) {
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
	chessDB.AutoMigrate(&elo{}, &pgn{})
}

// createELO 创建 ELO
func (s *chessDBService) createELO(uin int64, name string, rate int) error {
	return s.db.Create(&elo{
		Uin:  uin,
		Name: name,
		Rate: rate,
	}).Error
}

// getELORateByUin 获取 ELO 等级分
func (s *chessDBService) getELORateByUin(uin int64) (int, error) {
	var elo elo
	err := s.db.Select("rate").Where("uin = ?", uin).First(&elo).Error
	return elo.Rate, err
}

// getHighestRateList 获取最高的等级分列表
func (s *chessDBService) getHighestRateList() ([]elo, error) {
	var eloList []elo
	err := s.db.Order("rate desc").Limit(10).Find(&eloList).Error
	return eloList, err
}

// updateELOByUin 更新 ELO 等级分
func (s *chessDBService) updateELOByUin(uin int64, name string, rate int) error {
	return s.db.Model(&elo{}).Where("uin = ?", uin).Update("name", name).Update("rate", rate).Error
}

// cleanELOByUin 清空用户 ELO 等级分
func (s *chessDBService) cleanELOByUin(uin int64) error {
	return s.db.Model(&elo{}).Where("uin = ?", uin).Update("rate", 100).Error
}

// createPGN 创建 PGN
func (s *chessDBService) createPGN(data string, whiteUin int64, blackUin int64, whiteName string, blackName string) error {
	return s.db.Create(&pgn{
		Data:      data,
		WhiteUin:  whiteUin,
		BlackUin:  blackUin,
		WhiteName: whiteName,
		BlackName: blackName,
	}).Error
}
