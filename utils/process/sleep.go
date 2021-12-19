// Package process 流程控制相关
package process

import (
	"math/rand"
	"time"
)

// SleepAbout1sTo2s 随机阻塞等待 1 ~ 2s
func SleepAbout1sTo2s() {
	time.Sleep(time.Second + time.Millisecond*time.Duration(rand.Intn(1000)))
}
