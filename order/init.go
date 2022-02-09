package order

import "sync"

var wg sync.WaitGroup

// DoneOnExit 在退出时执行 Done
func DoneOnExit() func() {
	wg.Add(1)
	return wg.Done
}

// Wait 等待
var Wait = wg.Wait
