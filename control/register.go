package control

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

func Register(service string, o *Options) *zero.Engine {
	engine := zero.New()
	engine.UsePreHandler(new(service, o).Handler())
	return engine
}
