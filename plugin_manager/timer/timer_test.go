package timer

import (
	"testing"
	"time"

	sql "github.com/FloatTech/sqlite"
	"github.com/sirupsen/logrus"
)

func TestNextWakeTime(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ts := &Timer{}
	ts.SetMonth(-1)
	ts.SetWeek(6)
	ts.SetHour(16)
	ts.SetMinute(30)
	t1 := time.Until(ts.nextWakeTime())
	if t1 < 0 {
		t.Log(t1)
		t.Fail()
	}
	t.Log(t1)
	t.Fail()
}

func TestClock(t *testing.T) {
	db := &sql.Sqlite{DBPath: "test.db"}
	c := NewClock(db)
	c.AddTimerIntoDB(GetFilledTimer([]string{"", "12", "-1", "12", "0", "", "test"}, 0, 0, false))
	t.Log(c.ListTimers(0))
	t.Fail()
}
