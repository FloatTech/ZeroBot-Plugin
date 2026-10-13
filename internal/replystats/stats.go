// Package replystats keeps bounded AI task counts and matcher-to-send receipt timings.
package replystats

import (
	"math"
	"sync"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/internal/processidentity"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type Kind uint8

const (
	GPT Kind = iota
	Chat
	Agent
	kindCount
)

const ingressKey = zero.StateKeyPrefixKeep + "__ai_reply_ingress__"

var ingressOnce sync.Once

// This early non-blocking rule adds only an ephemeral monotonic timestamp. It
// performs no I/O and returns false so no handler, reply or blocking is added.
func InstallIngress() {
	ingressOnce.Do(func() {
		zero.OnMessage(func(ctx *zero.Ctx) bool {
			if ctx.State != nil {
				ctx.State[ingressKey] = time.Now()
			}
			return false
		}).SetPriority(math.MinInt).SetBlock(false)
	})
}

func init() { InstallIngress() }

type Counts struct {
	Started                       uint64  `json:"started"`
	Active                        uint64  `json:"active"`
	Completed                     uint64  `json:"completed"`
	Failed                        uint64  `json:"failed"`
	NoReply                       uint64  `json:"noReply"`
	WithReply                     uint64  `json:"withReply"`
	MissingIngress                uint64  `json:"missingIngress"`
	ReplyAttempts                 uint64  `json:"replyAttempts"`
	ReplyAccepted                 uint64  `json:"replyAccepted"`
	ReplyUnconfirmed              uint64  `json:"replyUnconfirmed"`
	AverageFirstReplyMilliseconds float64 `json:"averageFirstReplyMilliseconds"`
	MaxFirstReplyMilliseconds     float64 `json:"maxFirstReplyMilliseconds"`
	AverageLastReplyMilliseconds  float64 `json:"averageLastReplyMilliseconds"`
	MaxLastReplyMilliseconds      float64 `json:"maxLastReplyMilliseconds"`
}
type Category struct {
	Kind string `json:"kind"`
	Counts
}
type Snapshot struct {
	Process   *processidentity.Identity `json:"process,omitempty"`
	StartedAt time.Time                 `json:"startedAt"`
	CheckedAt time.Time                 `json:"checkedAt"`
	Scope     string                    `json:"scope"`
	Counts
	Categories []Category `json:"categories"`
}
type aggregate struct {
	Counts
	firstSum, lastSum float64
}
type Collector struct {
	mu      sync.Mutex
	started time.Time
	total   aggregate
	rows    [kindCount]aggregate
}

func New() *Collector { return &Collector{started: time.Now().UTC()} }

var Default = New()

func (c *Collector) Read() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := Snapshot{StartedAt: c.started, CheckedAt: time.Now().UTC(), Scope: "bot-matcher-entry-to-onebot-send-receipt", Counts: c.total.Counts, Categories: make([]Category, kindCount)}
	for i, kind := range [...]string{"gpt", "chat", "agent"} {
		s.Categories[i] = Category{Kind: kind, Counts: c.rows[i].Counts}
	}
	return s
}

// Task retains only an ephemeral target needed to distinguish replies to the
// current conversation. Neither targets nor request/response contents enter the collector.
type Task struct {
	mu                         sync.Mutex
	c                          *Collector
	kind                       Kind
	start                      time.Time
	groupID, userID            int64
	hasReply, failed, finished bool
	first, last                float64
}

func (c *Collector) Begin(ctx *zero.Ctx, kind Kind) *Task {
	if c == nil || kind >= kindCount || ctx == nil || ctx.Event == nil || ctx.Event.PostType != "message" || (ctx.Event.MessageType != "group" && ctx.Event.MessageType != "private") {
		return nil
	}
	now := time.Now()
	start, ok := ctx.State[ingressKey].(time.Time)
	if !ok || start.IsZero() || start.After(now) {
		start = now
		ok = false
	}
	c.mu.Lock()
	for _, a := range []*aggregate{&c.total, &c.rows[kind]} {
		a.Started++
		a.Active++
		if !ok {
			a.MissingIngress++
		}
	}
	c.mu.Unlock()
	return &Task{c: c, kind: kind, start: start, groupID: ctx.Event.GroupID, userID: ctx.Event.UserID}
}
func (t *Task) Fail() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.finished {
		t.failed = true
	}
}
func (t *Task) record(confirmed bool) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.finished {
		return
	}
	ms := math.Max(0, float64(time.Since(t.start))/float64(time.Millisecond))
	if confirmed {
		if !t.hasReply {
			t.hasReply = true
			t.first = ms
		}
		t.last = math.Max(t.last, ms)
	} else {
		t.failed = true
	}
	t.c.mu.Lock()
	defer t.c.mu.Unlock()
	for _, a := range []*aggregate{&t.c.total, &t.c.rows[t.kind]} {
		a.ReplyAttempts++
		if confirmed {
			a.ReplyAccepted++
		} else {
			a.ReplyUnconfirmed++
		}
	}
}
func (t *Task) finish() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.finished {
		return
	}
	t.finished = true
	t.c.mu.Lock()
	defer t.c.mu.Unlock()
	for _, a := range []*aggregate{&t.c.total, &t.c.rows[t.kind]} {
		a.Active--
		if t.failed {
			a.Failed++
		} else if t.hasReply {
			a.Completed++
		} else {
			a.NoReply++
		}
		if t.hasReply {
			a.WithReply++
			a.firstSum += t.first
			a.lastSum += t.last
			a.AverageFirstReplyMilliseconds = a.firstSum / float64(a.WithReply)
			a.AverageLastReplyMilliseconds = a.lastSum / float64(a.WithReply)
			a.MaxFirstReplyMilliseconds = math.Max(a.MaxFirstReplyMilliseconds, t.first)
			a.MaxLastReplyMilliseconds = math.Max(a.MaxLastReplyMilliseconds, t.last)
		}
	}
}

// End must be deferred directly by the handler. Preserve any original panic.
func (t *Task) End() {
	if p := recover(); p != nil {
		t.Fail()
		t.finish()
		panic(p)
	}
	t.finish()
}
