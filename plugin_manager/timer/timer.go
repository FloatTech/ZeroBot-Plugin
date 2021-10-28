// Package timer 群管定时器
package timer

import (
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fumiama/cron"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
)

type Clock struct {
	// 记录每个定时器以便取消
	timersmap TimersMap
	// 定时器map
	timers   *(map[string]*Timer)
	timersmu sync.RWMutex
	// 定时器存储位置
	pbfile *string
	// cron 定时器
	cron *cron.Cron
	// entries key <-> cron
	entries map[string]cron.EntryID
	entmu   sync.Mutex
}

var (
	// @全体成员
	atall = message.MessageSegment{
		Type: "at",
		Data: map[string]string{
			"qq": "all",
		},
	}
)

func NewClock(pbfile string) (c Clock) {
	c.loadTimers(pbfile)
	c.timers = &c.timersmap.Timers
	c.pbfile = &pbfile
	c.cron = cron.New()
	c.entries = make(map[string]cron.EntryID)
	c.cron.Start()
	return
}

// RegisterTimer 注册计时器
func (c *Clock) RegisterTimer(ts *Timer, grp int64, save bool) bool {
	key := ts.GetTimerInfo(grp)
	t, ok := c.GetTimer(key)
	if t != ts && ok { // 避免重复注册定时器
		t.SetEn(false)
	}
	c.timersmu.Lock()
	(*c.timers)[key] = ts
	c.timersmu.Unlock()
	logrus.Println("[群管]注册计时器", key)
	if ts.Cron != "" {
		var ctx *zero.Ctx
		if ts.Selfid != 0 {
			ctx = zero.GetBot(ts.Selfid)
		} else {
			zero.RangeBot(func(id int64, c *zero.Ctx) bool {
				ctx = c
				ts.Selfid = id
				return false
			})
		}
		eid, err := c.cron.AddFunc(ts.Cron, func() { ts.sendmsg(grp, ctx) })
		if err == nil {
			c.entmu.Lock()
			c.entries[key] = eid
			c.entmu.Unlock()
			if save {
				c.SaveTimers()
			}
			return true
		}
		ts.Alert = err.Error()
	} else {
		if save {
			c.SaveTimers()
		}
		for ts.En() {
			nextdate := ts.nextWakeTime()
			sleepsec := time.Until(nextdate)
			logrus.Printf("[群管]计时器%s将睡眠%ds", key, sleepsec/time.Second)
			time.Sleep(sleepsec)
			if ts.En() {
				if ts.Month() < 0 || ts.Month() == time.Now().Month() {
					if ts.Day() < 0 || ts.Day() == time.Now().Day() {
						ts.judgeHM(grp)
					} else if ts.Day() == 0 {
						if ts.Week() < 0 || ts.Week() == time.Now().Weekday() {
							ts.judgeHM(grp)
						}
					}
				}
			}
		}
	}
	return false
}

// CancelTimer 取消计时器
func (c *Clock) CancelTimer(key string) bool {
	t, ok := (*c.timers)[key]
	if ok {
		if t.Cron != "" {
			c.entmu.Lock()
			e := c.entries[key]
			c.cron.Remove(e)
			delete(c.entries, key)
			c.entmu.Unlock()
		} else {
			t.SetEn(false)
		}
		c.timersmu.Lock()
		delete(*c.timers, key) // 避免重复取消
		c.timersmu.Unlock()
		_ = c.SaveTimers()
	}
	return ok
}

// SaveTimers 保存当前计时器
func (c *Clock) SaveTimers() error {
	c.timersmu.RLock()
	data, err := c.timersmap.Marshal()
	c.timersmu.RUnlock()
	if err == nil {
		c.timersmu.Lock()
		defer c.timersmu.Unlock()
		f, err1 := os.OpenFile(*c.pbfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err1 != nil {
			return err1
		} else {
			_, err2 := f.Write(data)
			f.Close()
			return err2
		}
	}
	return err
}

// ListTimers 列出本群所有计时器
func (c *Clock) ListTimers(grpID uint64) []string {
	// 数组默认长度为map长度,后面append时,不需要重新申请内存和拷贝,效率很高
	if c.timers != nil {
		g := strconv.FormatUint(grpID, 10)
		c.timersmu.RLock()
		keys := make([]string, 0, len(*c.timers))
		for k := range *c.timers {
			if strings.Contains(k, g) {
				start := strings.Index(k, "]")
				msg := strings.ReplaceAll(k[start+1:]+"\n", "-1", "每")
				msg = strings.ReplaceAll(msg, "月0日0周", "月周天")
				msg = strings.ReplaceAll(msg, "月0日", "月")
				msg = strings.ReplaceAll(msg, "日0周", "日")
				keys = append(keys, msg)
			}
		}
		c.timersmu.RUnlock()
		return keys
	} else {
		return nil
	}
}

func (c *Clock) GetTimer(key string) (t *Timer, ok bool) {
	c.timersmu.RLock()
	t, ok = (*c.timers)[key]
	c.timersmu.RUnlock()
	return
}

func (c *Clock) loadTimers(pbfile string) {
	if file.IsExist(pbfile) {
		f, err := os.Open(pbfile)
		if err == nil {
			data, err := io.ReadAll(f)
			if err == nil {
				if len(data) > 0 {
					err = c.timersmap.Unmarshal(data)
					if err == nil {
						for str, t := range c.timersmap.Timers {
							grp, err := strconv.ParseInt(str[1:strings.Index(str, "]")], 10, 64)
							if err == nil {
								go c.RegisterTimer(t, grp, false)
							}
						}
						return
					}
					logrus.Errorln("[群管]读取定时器文件失败，将在下一次保存时覆盖原文件。err:", err)
					logrus.Errorln("[群管]如不希望被覆盖，请运行源码plugin_manager/timers/migrate下的程序将timers.pb刷新为新版")
				}
			}
		}
	}
	c.timersmap.Timers = make(map[string]*Timer)
}
