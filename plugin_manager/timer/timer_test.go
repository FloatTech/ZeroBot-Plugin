package timer

import (
	"testing"
	"time"

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
