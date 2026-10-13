package registry

import "sync"

var pool sync.Pool

func init() {
	pool.New = func() interface{} {
		return new(CmdPacket)
	}
}
