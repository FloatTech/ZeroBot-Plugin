package minecraftobserver

import (
	"encoding/json"
	"time"

	"github.com/RomiChan/syncx"
	"github.com/Tnze/go-mc/bot"
)

var (
	// pingServerUnreachableCounter Ping服务器不可达计数器，防止bot本体网络抖动导致误报
	pingServerUnreachableCounter = syncx.Map[string, pingServerUnreachableCounterDef]{}
	// 计数器阈值
	pingServerUnreachableCounterThreshold = int64(3)
	// 时间阈值
	pingServerUnreachableCounterTimeThreshold = time.Minute * 30
)

type pingServerUnreachableCounterDef struct {
	count                int64
	firstUnreachableTime time.Time
}

func addPingServerUnreachableCounter(addr string, ts time.Time) (int64, time.Time) {
	key := addr
	get, ok := pingServerUnreachableCounter.Load(key)
	if !ok {
		pingServerUnreachableCounter.Store(key, pingServerUnreachableCounterDef{
			count:                1,
			firstUnreachableTime: ts,
		})
		return 1, ts
	}
	// 存在则更新，时间戳不变
	pingServerUnreachableCounter.Store(key, pingServerUnreachableCounterDef{
		count:                get.count + 1,
		firstUnreachableTime: get.firstUnreachableTime,
	})
	return get.count + 1, get.firstUnreachableTime
}

func resetPingServerUnreachableCounter(addr string) {
	key := addr
	pingServerUnreachableCounter.Delete(key)
}

// getMinecraftServerStatus 获取Minecraft服务器状态
func getMinecraftServerStatus(addr string) (*serverPingAndListResp, error) {
	var s serverPingAndListResp
	resp, delay, err := bot.PingAndListTimeout(addr, time.Second*5)
	if err != nil {
		// logrus.Errorln(logPrefix+"PingAndList error: ", err)
		return nil, err
	}
	err = json.Unmarshal(resp, &s)
	if err != nil {
		// logrus.Errorln(logPrefix+"Parse json response fail: ", err)
		return nil, err
	}
	s.Delay = delay
	return &s, nil
}
