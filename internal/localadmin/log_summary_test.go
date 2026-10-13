package localadmin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func summaryLog(t *testing.T, l *LogStore, at time.Time, level logrus.Level, message string) {
	t.Helper()
	if err := l.Fire(&logrus.Entry{Time: at, Level: level, Message: message}); err != nil {
		t.Fatal(err)
	}
}

func TestLogSummaryIndependentOfLimitCursorAndReset(t *testing.T) {
	l := NewLogStore()
	now := time.Now()
	for i := 0; i < 250; i++ {
		summaryLog(t, l, now, logrus.InfoLevel, "info fixture")
	}
	for _, level := range []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.DebugLevel, logrus.TraceLevel} {
		summaryLog(t, l, now, level, "issue fixture")
	}
	first := l.page(logQuery{limit: 200})
	if len(first.Entries) != 200 || first.Summary.Matched != 256 || first.Summary.Levels["info"] != 250 || first.Cursor != "256" {
		t.Fatal("first page summary limited to displayed rows")
	}
	for _, q := range []logQuery{{after: 1, limit: 2}, {after: 256, limit: 200}, {after: 999, limit: 2}, {after: 256, instance: "0000000000000000", limit: 2}} {
		p := l.page(q)
		if !reflect.DeepEqual(p.Summary, first.Summary) {
			t.Fatal("cursor/limit/reset changed snapshot summary")
		}
		if q.after == 1 && (len(p.Entries) != 2 || p.Cursor != "3") {
			t.Fatal("incremental cursor semantics changed")
		}
		if q.after == 256 && q.instance == "" && len(p.Entries) != 0 {
			t.Fatal("old entries returned again")
		}
	}
	issues := l.page(logQuery{after: 250, level: "issues", limit: 2})
	if len(issues.Entries) != 2 || issues.Cursor != "252" || issues.Summary.Matched != 4 || issues.Summary.Levels["panic"] != 1 || issues.Summary.Levels["fatal"] != 1 || issues.Summary.Levels["error"] != 1 || issues.Summary.Levels["warning"] != 1 || issues.Summary.Levels["info"] != 0 {
		t.Fatal("issues summary or cursor incorrect")
	}
}

func TestLogSummaryRedactedKeywordsTimeBoundsAndLateSecrets(t *testing.T) {
	l := NewLogStore()
	base := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	summaryLog(t, l, base, logrus.WarnLevel, "Needle before")
	summaryLog(t, l, base.Add(time.Second), logrus.ErrorLevel, "Needle unicode 汉🙂")
	summaryLog(t, l, base.Add(2*time.Second), logrus.InfoLevel, "Needle after")
	summaryLog(t, l, base.Add(time.Second), logrus.WarnLevel, "unknown-late-credential")
	l.AddSecrets("unknown-late-credential")
	for _, after := range []uint64{0, 2, 4} {
		p := l.page(logQuery{after: after, limit: 1, keyword: "unknown-late-credential"})
		if p.Summary.Matched != 0 || len(p.Entries) != 0 {
			t.Fatal("summary inferred a raw credential")
		}
		p = l.page(logQuery{after: after, limit: 1, keyword: "[已隐藏]"})
		if p.Summary.Matched != 1 || p.Summary.Levels["warning"] != 1 {
			t.Fatal("old row keyword did not use current redaction")
		}
	}
	bound := base.Add(time.Second)
	p := l.page(logQuery{limit: 1, from: &bound, until: &bound, keyword: "NEEDLE", level: "issues"})
	if p.Summary.Matched != 1 || p.Summary.Levels["error"] != 1 || len(p.Entries) != 1 {
		t.Fatal("inclusive time/severity/Unicode keyword combination incorrect")
	}
	before := l.page(logQuery{limit: 1}).Summary
	l.maskAll = true
	p = l.page(logQuery{after: 4, limit: 1, keyword: "needle"})
	if p.Summary.Matched != 0 {
		t.Fatal("mask-all keyword leaked old content")
	}
	if !reflect.DeepEqual(before, l.page(logQuery{after: 4, limit: 1}).Summary) {
		t.Fatal("metadata-only counts changed with current credentials")
	}
}

func TestLogSummaryEvictionByteBoundAndUnknownLevelBucket(t *testing.T) {
	l := NewLogStore()
	for i := 0; i < 1100; i++ {
		level := logrus.InfoLevel
		if i < 100 {
			level = logrus.WarnLevel
		}
		summaryLog(t, l, time.Now(), level, "eviction fixture")
	}
	p := l.page(logQuery{after: 1, limit: 2})
	if p.Summary.Matched != 1000 || p.Summary.Levels["warning"] != 0 || p.Summary.Levels["info"] != 1000 || !p.Gap {
		t.Fatal("summary includes evicted logs")
	}
	l.mu.Lock()
	l.rows[l.start].Level = "<private-unknown-level>"
	l.mu.Unlock()
	p = l.page(logQuery{after: 1100, limit: 2})
	raw, _ := json.Marshal(p.Summary)
	if p.Summary.Levels["other"] != 1 || p.Summary.Levels["info"] != 999 || strings.Contains(string(raw), "private") || len(p.Summary.Levels) != 8 {
		t.Fatal("unknown level leaked or broke fixed summary")
	}
	for i := 0; i < 1000; i++ {
		summaryLog(t, l, time.Now(), logrus.DebugLevel, strings.Repeat("汉", 600))
	}
	p = l.page(logQuery{limit: 1})
	if p.Retained >= 1000 || p.Summary.Matched != p.Retained {
		t.Fatal("summary ignores byte-capacity eviction")
	}
	var total int
	for _, count := range p.Summary.Levels {
		total += count
	}
	if total != p.Summary.Matched {
		t.Fatal("level counts do not sum to matching retained rows")
	}
}

func TestLogSummaryAPIAuthReadOnlyAndExportAgreement(t *testing.T) {
	l := NewLogStore()
	cfg := testConfig(t)
	cfg.Logs = l
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(s, "GET", "/api/logs?q=needle", nil, nil, ""), 401)
	cookie, _ := enroll(t, s)
	for i := 0; i < 250; i++ {
		summaryLog(t, l, time.Now(), logrus.WarnLevel, "Needle synthetic-warning")
	}
	before := l.next
	rsp := request(s, "GET", "/api/logs?after=250&limit=1&level=issues&q=needle", nil, cookie, "")
	expectStatus(t, rsp, 200)
	var p LogPage
	if err := json.Unmarshal(rsp.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Entries) != 0 || p.Summary.Matched != 250 || p.Summary.Levels["warning"] != 250 || before != l.next {
		t.Fatal("read-only API summary incorrect")
	}
	exported := request(s, "GET", "/api/logs/export?level=issues&q=needle", nil, cookie, "")
	expectStatus(t, exported, 200)
	if strings.Count(exported.Body.String(), "] [WARNING] #") != p.Summary.Matched {
		t.Fatal("summary disagrees with unpaginated export")
	}
	for _, query := range []string{"level=invalid", "after=-1", "limit=201", "from=invalid"} {
		expectStatus(t, request(s, "GET", "/api/logs?"+query, nil, cookie, ""), 400)
	}
}

func TestLogSummaryReadCacheSecretChangesAndEviction(t *testing.T) {
	l := NewLogStore()
	summaryLog(t, l, time.Now(), logrus.InfoLevel, "late-model-secret needle")
	query := logQuery{limit: 200, keyword: "late-model-secret"}
	if l.page(query).Summary.Matched != 1 {
		t.Fatal("initial cache fixture mismatch")
	}
	version := l.maskVersion
	l.AddSecrets("late-model-secret")
	if l.maskVersion == version || l.page(query).Summary.Matched != 0 {
		t.Fatal("new key did not invalidate old read result")
	}
	version = l.maskVersion
	l.AddSecrets("late-model-secret")
	if l.maskVersion != version || l.page(query).Summary.Matched != 0 {
		t.Fatal("duplicate keys invalidated or changed redaction")
	}
	if l.page(logQuery{limit: 200, keyword: "[已隐藏]"}).Summary.Matched != 1 {
		t.Fatal("refreshed cache did not contain current mask")
	}
	for i := 0; i < 1000; i++ {
		summaryLog(t, l, time.Now(), logrus.DebugLevel, "fresh-cache-fixture")
	}
	if l.page(logQuery{limit: 200, keyword: "needle"}).Summary.Matched != 0 {
		t.Fatal("evicted row survived in cached search")
	}
	l.mu.Lock()
	for l.count > 0 {
		l.evictLocked()
	}
	for _, cache := range l.readMasks {
		if cache != (logReadMask{}) {
			l.mu.Unlock()
			t.Fatal("eviction retained a cached message reference")
		}
	}
	l.mu.Unlock()
	summaryLog(t, l, time.Now(), logrus.InfoLevel, "another-late-secret")
	if l.page(logQuery{limit: 200, keyword: "another-late-secret"}).Summary.Matched != 1 {
		t.Fatal("new row was replaced by old cache")
	}
	l.AddSecrets(strings.Repeat("x", 16385))
	if l.page(logQuery{limit: 200, keyword: "another-late-secret"}).Summary.Matched != 0 {
		t.Fatal("fail-closed state did not invalidate cached content")
	}
}

func BenchmarkLogSummaryIncremental(b *testing.B) {
	l := NewLogStore()
	for i := 0; i < 1000; i++ {
		_ = l.Fire(&logrus.Entry{Time: time.Now(), Level: logrus.InfoLevel, Message: fmt.Sprintf("needle synthetic log %04d %s", i, strings.Repeat("x", 256))})
	}
	for _, keyword := range []string{"", "needle"} {
		name := "metadata-only"
		if keyword != "" {
			name = "redacted-keyword"
		}
		b.Run(name, func(b *testing.B) {
			query, err := parseLogQuery(url.Values{"after": {"1000"}, "q": {keyword}})
			if err != nil {
				b.Fatal(err)
			}
			// Model an active viewer after its initial retained-page read.
			_ = l.page(logQuery{limit: 200, keyword: keyword})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if l.page(query).Summary.Matched != 1000 {
					b.Fatal("incomplete benchmark summary")
				}
			}
		})
	}
}
