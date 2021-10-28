package process

import (
	"math/rand"
	"time"
)

func SleepAbout1sTo2s() {
	time.Sleep(time.Second + time.Millisecond*time.Duration(rand.Intn(1000)))
}
