package localadmin

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLogTimeRangeBoundariesCombinationAndCursor(t *testing.T) {
	l, logger := logFixture()
	start := time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC)
	logger.WithTime(start.Add(-time.Nanosecond)).Warn("before needle")
	logger.WithTime(start).Warn("start needle")
	logger.WithTime(start.Add(time.Second)).Info("normal needle")
	logger.WithTime(start.Add(2 * time.Second)).Error("end needle")
	logger.WithTime(start.Add(2*time.Second + time.Nanosecond)).Warn("after needle")
	cases := []struct {
		values url.Values
		ids    string
	}{
		{url.Values{"from": {start.Format(time.RFC3339Nano)}, "until": {start.Add(2 * time.Second).Format(time.RFC3339Nano)}, "level": {"issues"}, "q": {"NEEDLE"}}, "2,4"},
		{url.Values{"from": {"2026-10-09T16:00:00+08:00"}, "until": {"2026-10-09T08:00:00Z"}}, "2"},
		{url.Values{"from": {start.Add(2 * time.Second).Format(time.RFC3339Nano)}}, "4,5"},
		{url.Values{"until": {start.Format(time.RFC3339Nano)}}, "1,2"},
		{url.Values{"from": {"2100-01-01T00:00:00Z"}}, ""},
		{url.Values{"from": {"0001-01-01T00:00:00Z"}}, "1,2,3,4,5"},
	}
	for _, c := range cases {
		q, err := parseLogQuery(c.values)
		if err != nil {
			t.Fatal(err)
		}
		p := l.page(q)
		ids := []string{}
		for _, row := range p.Entries {
			ids = append(ids, row.ID)
		}
		if strings.Join(ids, ",") != c.ids || p.Cursor != "5" {
			t.Fatalf("%v: got %v cursor %s", c.values, ids, p.Cursor)
		}
	}
	q, _ := parseLogQuery(cases[0].values)
	q.after = 1
	q.limit = 1
	first := l.page(q)
	if len(first.Entries) != 1 || first.Cursor != "2" {
		t.Fatal("filtered cursor did not stop at first match")
	}
	q.after = 2
	next := l.page(q)
	if len(next.Entries) != 1 || next.Cursor != "4" {
		t.Fatal("filtered cursor skipped boundary")
	}
	q.after = 4
	last := l.page(q)
	if len(last.Entries) != 0 || last.Cursor != "5" {
		t.Fatal("filtered cursor did not advance past nonmatching rows")
	}
	q.instance = "0000000000000000"
	if !l.page(q).Reset {
		t.Fatal("time filter bypassed instance reset")
	}
}

func TestLogTimeValidationAndAuthentication(t *testing.T) {
	l, _ := logFixture()
	cfg := testConfig(t)
	cfg.Logs = l
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cookie, csrf := enroll(t, s)
	invalid := []url.Values{
		{"from": {"not-a-date"}}, {"until": {"2026-10-09T08:00:00"}}, {"from": {"2026-02-30T08:00:00Z"}},
		{"from": {"2026-10-09T08:00:00Z", "2026-10-09T08:00:00Z"}}, {"until": {"", ""}},
		{"from": {"2026-10-09T08:00:01Z"}, "until": {"2026-10-09T16:00:00+08:00"}}, {"from": {strings.Repeat("x", 65)}},
	}
	for _, endpoint := range []string{"/api/logs", "/api/logs/export"} {
		for _, values := range invalid {
			w := request(s, "GET", endpoint+"?"+values.Encode(), nil, cookie, "")
			expectStatus(t, w, 400)
			if w.Header().Get("Content-Disposition") != "" {
				t.Fatal("error became download")
			}
		}
		expectStatus(t, request(s, "GET", endpoint+"?from=2026-10-09T08:00:00Z", nil, nil, ""), 401)
	}
	expectStatus(t, request(s, "POST", "/api/logout", map[string]any{}, cookie, csrf), 200)
	expectStatus(t, request(s, "GET", "/api/logs/export?until=2026-10-09T08:00:00Z", nil, cookie, ""), 401)
}

func TestLogTimeExportMatchesViewerAndKeepsRedaction(t *testing.T) {
	l, logger := logFixture()
	cfg := testConfig(t)
	cfg.Logs = l
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cookie, _ := enroll(t, s)
	base := time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC)
	l.AddSecrets("new-time-test-key")
	logger.WithTime(base.Add(-time.Second)).Warn("outside")
	logger.WithTime(base).Warn("inside new-time-test-key")
	logger.WithTime(base.Add(time.Second)).Error("inside second")
	logger.WithTime(base.Add(2 * time.Second)).Warn("outside after")
	query := url.Values{"from": {"2026-10-09T16:00:00+08:00"}, "until": {"2026-10-09T08:00:01Z"}, "level": {"issues"}, "q": {"inside"}}
	w := request(s, "GET", "/api/logs?"+query.Encode(), nil, cookie, "")
	expectStatus(t, w, 200)
	var page LogPage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	w = request(s, "GET", "/api/logs/export?"+query.Encode(), nil, cookie, "")
	expectStatus(t, w, 200)
	body := w.Body.String()
	if len(page.Entries) != 2 || strings.Count(body, "] #") != 2 || strings.Contains(body, "outside") || strings.Contains(body, "new-time-test-key") || !strings.Contains(body, "2026-10-09T08:00:00Z 至 2026-10-09T08:00:01Z") {
		t.Fatal("range/header/redaction mismatch")
	}
	for _, row := range page.Entries {
		if !strings.Contains(body, "#"+row.ID+" ") {
			t.Fatal("export/view mismatch")
		}
	}
	w = request(s, "GET", "/api/logs/export?from=2100-01-01T00:00:00Z", nil, cookie, "")
	expectStatus(t, w, 200)
	if !strings.Contains(w.Body.String(), "符合已应用筛选：0 条") || strings.Contains(w.Body.String(), "] #") {
		t.Fatal("empty range wrong")
	}
}
