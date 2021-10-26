package timer

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNextWakeTime(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ts := &Timer{
		Month:  10,
		Day:    -1,
		Week:   0,
		Hour:   -1,
		Minute: 6,
	}
	t1 := time.Until(ts.nextWakeTime())
	if t1 < 0 {
		t.Log(t1)
		t.Fail()
	}
	t.Log(t1)
	t.Fail()
}
