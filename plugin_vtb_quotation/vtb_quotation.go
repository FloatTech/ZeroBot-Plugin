package plugin_vtb_quotation

import (
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	db *gorm.DB

	regStr = ".*/(.*)"
)

//第一品类
type FirstCategory struct {
	Id                 int    `gorm:"column:id"`
	FirstCategoryIndex int    `gorm:"column:first_category_index"`
	FirstCategoryName  string `gorm:"column:first_category_name"`
	FirstCategoryUid   string `gorm:"column:first_category_uid"`
}

func (FirstCategory) TableName() string {
	return "first_category"
}

//第二品类
type SecondCategory struct {
	Id                  int    `gorm:"column:id"`
	SecondCategoryIndex int    `gorm:"column:second_category_index"`
	FirstCategoryIndex  int    `gorm:"column:first_category_index"`
	SecondCategoryName  string `gorm:"column:second_category_name"`
}

func (SecondCategory) TableName() string {
	return "second_category"
}

//第三品类
type ThirdCategory struct {
	Id                  int    `gorm:"column:id"`
	ThirdCategoryIndex  int    `gorm:"column:third_category_index"`
	SecondCategoryIndex int    `gorm:"column:second_category_index"`
	FirstCategoryIndex  int    `gorm:"column:first_category_index"`
	ThirdCategoryName   string `gorm:"column:third_category_name"`
	ThirdCategoryUrl    string `gorm:"column:third_category_url"`
}

func (ThirdCategory) TableName() string {
	return "third_category"
}

func init() {
	engine := control.Register("vtbquotation", &control.Options{
		DisableOnDefault: false,
		Help: "vtb语录\n" +
			"随机vtb\n",
	})
	engine.OnFullMatch("vtb语录").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var firstIndex int
			var secondIndex int
			var thirdIndex int
			echo, cancel := ctx.FutureEvent("message",
				ctx.CheckSession()). // 只复读开启复读模式的人的消息
				Repeat()             // 不断监听复读
			db, err := gorm.Open("sqlite3", "data/VtbQuotation/vtb.db")
			if err != nil {
				panic("failed to connect database")
			}
			defer db.Close()
			firstStepMessage := getAllFirstCategoryMessage(db)
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(firstStepMessage))
			//步骤1，2，3，依次选择3个类别
			step := 1
			//错误次数
			errorCount := 0
			for {
				select {
				case e := <-echo: // 接收到需要复读的消息
					//错误次数达到3次，结束命令
					if errorCount == 3 {
						ctx.SendChain(message.Reply(e.MessageID), message.Text("输入错误太多,请重新发指令"))
						cancel()
						return
					}
					if step == 1 {
						firstIndex, err = strconv.Atoi(e.RawMessage)
						log.Println(fmt.Sprintf("当前在第%d步", step))
						log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号,三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							SecondStepMessage := getAllSecondCategoryMessageByFirstIndex(db, firstIndex)
							log.Println(SecondStepMessage)
							if SecondStepMessage == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
								ctx.SendChain(message.Reply(e.MessageID), message.Text(getAllFirstCategoryMessage(db)))
								errorCount++
							} else {
								ctx.SendChain(message.Reply(e.MessageID), message.Text(SecondStepMessage))
								step++
							}

						}
					} else if step == 2 {
						secondIndex, err = strconv.Atoi(e.RawMessage)
						log.Println(fmt.Sprintf("当前在第%d步", step))
						log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							ThirdStepMessage := getAllThirdCategoryMessageByFirstIndexAndSecondIndex(db, firstIndex, secondIndex)
							log.Println(ThirdStepMessage)
							if ThirdStepMessage == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
								ctx.SendChain(message.Reply(e.MessageID), message.Text(getAllSecondCategoryMessageByFirstIndex(db, firstIndex)))
								errorCount++
							} else {
								ctx.SendChain(message.Reply(e.MessageID), message.Text(ThirdStepMessage))
								step++
							}

						}
					} else if step == 3 {
						thirdIndex, err = strconv.Atoi(e.RawMessage)
						log.Println(fmt.Sprintf("当前在第%d步", step))
						log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							tc := getThirdCategory(db, firstIndex, secondIndex, thirdIndex)
							reg := regexp.MustCompile(regStr)
							recordUrl := tc.ThirdCategoryUrl
							if recordUrl == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("没有内容请重新选择，三次输入错误，指令可退出重输"))
								ctx.SendChain(message.Reply(e.MessageID), message.Text(getAllFirstCategoryMessage(db)))
								errorCount++
								step = 1
							} else {
								if reg.MatchString(recordUrl) {
									log.Println(reg.FindStringSubmatch(recordUrl)[1])
									log.Println(url.QueryEscape(reg.FindStringSubmatch(recordUrl)[1]))
									recordUrl = strings.Replace(recordUrl, reg.FindStringSubmatch(recordUrl)[1], url.QueryEscape(reg.FindStringSubmatch(recordUrl)[1]), -1)
									recordUrl = strings.Replace(recordUrl, "+", "%20", -1)
									log.Println(recordUrl)
								}
								ctx.SendChain(message.Reply(e.MessageID), message.Text("请欣赏《"+tc.ThirdCategoryName+"》"))
								ctx.SendChain(message.Record(recordUrl))
								cancel()
								return
							}

						}
					}
				case <-time.After(time.Second * 60):
					cancel()
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("vtb语录指令过期"))
					return
				}
			}
		})
	engine.OnFullMatch("随机vtb").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			db, err := gorm.Open("sqlite3", "data/VtbQuotation/vtb.db")
			if err != nil {
				panic("failed to connect database")
			}
			defer db.Close()
			tc := randomVtb(db)
			fc := getFirstCategoryByFirstIndex(db, tc.FirstCategoryIndex)
			if (tc != ThirdCategory{}) && (fc != FirstCategory{}) {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请欣赏"+fc.FirstCategoryName+"的《"+tc.ThirdCategoryName+"》"))
				ctx.SendChain(message.Record(tc.ThirdCategoryUrl))
			}

		})
}

//取出所有vtb
func getAllFirstCategoryMessage(db *gorm.DB) string {
	firstStepMessage := "请选择一个vtb并发送序号:\n"
	var fc FirstCategory
	rows, err := db.Model(&FirstCategory{}).Rows()
	if err != nil {
		log.Println("数据库读取错误", err)
	}
	if rows == nil {
		return ""
	}
	for rows.Next() {
		db.ScanRows(rows, &fc)
		log.Println(fc)
		firstStepMessage = firstStepMessage + strconv.Itoa(fc.FirstCategoryIndex) + ". " + fc.FirstCategoryName + "\n"
	}
	return firstStepMessage
}

//取得同一个vtb所有语录类别
func getAllSecondCategoryMessageByFirstIndex(db *gorm.DB, firstIndex int) string {
	SecondStepMessage := "请选择一个语录类别并发送序号:\n"
	var sc SecondCategory
	var count int
	db.Model(&SecondCategory{}).Where("first_category_index = ?", firstIndex).Count(&count)
	if count == 0 {
		return ""
	}
	rows, err := db.Model(&SecondCategory{}).Where("first_category_index = ?", firstIndex).Rows()
	if err != nil {
		log.Println("数据库读取错误", err)
	}

	for rows.Next() {
		db.ScanRows(rows, &sc)
		log.Println(sc)
		SecondStepMessage = SecondStepMessage + strconv.Itoa(sc.SecondCategoryIndex) + ". " + sc.SecondCategoryName + "\n"
	}
	return SecondStepMessage
}

//取得同一个vtb同个类别的所有语录
func getAllThirdCategoryMessageByFirstIndexAndSecondIndex(db *gorm.DB, firstIndex, secondIndex int) string {
	ThirdStepMessage := "请选择一个语录并发送序号:\n"
	var count int
	db.Model(&ThirdCategory{}).Where("first_category_index = ? and second_category_index = ?", firstIndex, secondIndex).Count(&count)
	if count == 0 {
		return ""
	}
	var tc ThirdCategory
	rows, err := db.Model(&ThirdCategory{}).Where("first_category_index = ? and second_category_index = ?", firstIndex, secondIndex).Rows()
	if err != nil {
		log.Println("数据库读取错误", err)
	}
	for rows.Next() {
		db.ScanRows(rows, &tc)
		log.Println(tc)
		ThirdStepMessage = ThirdStepMessage + strconv.Itoa(tc.ThirdCategoryIndex) + ". " + tc.ThirdCategoryName + "\n"
	}
	return ThirdStepMessage
}
func getThirdCategory(db *gorm.DB, firstIndex, secondIndex, thirdIndex int) ThirdCategory {
	var tc ThirdCategory
	db.Model(&ThirdCategory{}).Where("first_category_index = ? and second_category_index = ? and third_category_index = ?", firstIndex, secondIndex, thirdIndex).Take(&tc)
	return tc
}

func randomVtb(db *gorm.DB) ThirdCategory {
	var tc ThirdCategory
	db.Model(&ThirdCategory{}).Take(&tc)
	return tc
}

func getFirstCategoryByFirstIndex(db *gorm.DB, firstIndex int) FirstCategory {
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).Take(&fc)
	return fc
}
