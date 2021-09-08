package control

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

func Register(service string, o *Options) *zero.Engine {
	engine := zero.New()
	m := *New(service, o)
	engine.UsePreHandler(m.Handler())
	return engine
}
