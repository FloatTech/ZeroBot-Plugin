package timer

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNextWakeTime(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ts := &Timer{
		En1Month4Day5Week3Hour5Min6: 0xffffff,
	}
	t1 := time.Until(ts.nextWakeTime())
	if t1 < 0 {
		t.Log(t1)
		t.Fail()
	}
	t.Log(t1)
	t.Fail()
}
