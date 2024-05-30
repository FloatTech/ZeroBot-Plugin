package control

import (
	"runtime"
	"strings"
	"sync/atomic"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	enmap       = make(map[string]*Engine)
	prio        uint64
	custpriomap map[string]uint64
)

// LoadCustomPriority 加载自定义优先级 map，适配 1.21 及以上版本
func LoadCustomPriority(m map[string]uint64) {
	if custpriomap != nil {
		panic("double-defined custpriomap")
	}
	custpriomap = m
	prio = uint64(len(custpriomap)+1) * 10
}

// AutoRegister 根据包名自动注册插件
func AutoRegister(o *ctrl.Options[*zero.Ctx]) *Engine {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to get caller")
	}
	name := runtime.FuncForPC(pc).Name()
	a := strings.LastIndex(name, "/")
	if a < 0 {
		panic("invalid package name: " + name)
	}
	name = name[a+1:]
	b := strings.Index(name, ".")
	if b < 0 {
		panic("invalid package name: " + name)
	}
	name = name[:b]
	return Register(name, o)
}

// Register 注册插件控制器
func Register(service string, o *ctrl.Options[*zero.Ctx]) *Engine {
	if custpriomap != nil {
		logrus.Debugln("[control]插件", service, "已设置自定义优先级", prio)
		engine := newengine(service, int(custpriomap[service]), o)
		enmap[service] = engine
		return engine
	}
	logrus.Debugln("[control]插件", service, "已自动设置优先级", prio)
	engine := newengine(service, int(atomic.AddUint64(&prio, 10)), o)
	enmap[service] = engine
	return engine
}

// Delete 删除插件控制器, 不会删除数据
func Delete(service string) {
	engine, ok := enmap[service]
	if ok {
		engine.Delete()
		managers.RLock()
		_, ok = managers.M[service]
		managers.RUnlock()
		if ok {
			managers.Lock()
			delete(managers.M, service)
			managers.Unlock()
		}
	}
}
