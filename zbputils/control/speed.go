package control

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
)

// ApplySingle 应用反并发
func (e *Engine) ApplySingle(s *single.Single[int64]) *Engine {
	s.Apply(e.en)
	return e
}

// Limit 限速器
//
//	postfn 当请求被拒绝时的操作
func (m *Matcher) Limit(limiterfn func(*zero.Ctx) *rate.Limiter, postfn ...func(*zero.Ctx)) *Matcher {
	m.Rules = append(m.Rules, func(ctx *zero.Ctx) bool {
		if limiterfn(ctx).Acquire() {
			return true
		}
		if len(postfn) > 0 {
			for _, fn := range postfn {
				fn(ctx)
			}
		}
		return false
	})
	return m
}
