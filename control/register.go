package control

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

// Register 注册插件控制器
func Register(service string, o *Options) *zero.Engine {
	engine := zero.New()
	engine.UsePreHandler(newctrl(service, o).Handler())
	return engine
}

// Delete 删除插件控制器，不会删除数据
func Delete(engine *zero.Engine, service string) {
	engine.Delete()
	mu.RLock()
	_, ok := managers[service]
	mu.RUnlock()
	if ok {
		mu.Lock()
		delete(managers, service)
		mu.Unlock()
	}
}
