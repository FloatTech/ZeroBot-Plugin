package localadmin

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func fillIssueLogs(t *testing.T, l *LogStore) {
	t.Helper()
	// Fire captures these levels without invoking logger.Fatal/Panic.
	for _, level := range logrus.AllLevels {
		if err := l.Fire(&logrus.Entry{Level: level, Time: time.Now(), Message: "Needle " + level.String() + " synthetic-log-secret"}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLogIssuesFilterPaginationAndRedaction(t *testing.T) {
	l := NewLogStore()
	fillIssueLogs(t, l)
	l.AddSecrets("synthetic-log-secret")
	q, err := parseLogQuery(url.Values{"level": {"ISSUES"}, "q": {"needle"}, "limit": {"2"}})
	if err != nil || q.level != "issues" {
		t.Fatal("issue filter normalization failed", err)
	}
	p := l.page(q)
	if len(p.Entries) != 2 || p.Entries[0].Level != "error" || p.Entries[1].Level != "warning" || p.Cursor != "7" {
		t.Fatal("initial latest page did not select only issue levels", p)
	}
	q.after, q.instance = 1, l.instance
	p = l.page(q)
	if len(p.Entries) != 2 || p.Entries[0].Level != "fatal" || p.Entries[1].Level != "error" || p.Cursor != "3" {
		t.Fatal("incremental issue page lost order/cursor", p)
	}
	q.after = 3
	p = l.page(q)
	if len(p.Entries) != 1 || p.Entries[0].Level != "warning" || p.Cursor != "7" {
		t.Fatal("incremental issue tail incorrect", p)
	}
	q.after, q.limit = 0, logCapacity
	q.keyword = "synthetic-log-secret"
	if len(l.page(q).Entries) != 0 {
		t.Fatal("issue keyword matched unredacted secret")
	}
	q.keyword = "[已隐藏]"
	if len(l.page(q).Entries) != 4 {
		t.Fatal("issue keyword did not match redacted text")
	}
	q.level, q.keyword = "info", "needle"
	if p := l.page(q); len(p.Entries) != 1 || p.Entries[0].Level != "info" {
		t.Fatal("exact level filter changed")
	}
}

func TestLogIssuesAPIAndExportShareSelection(t *testing.T) {
	l := NewLogStore()
	fillIssueLogs(t, l)
	l.AddSecrets("synthetic-log-secret")
	cfg := testConfig(t)
	cfg.Logs = l
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/logs?level=issues", "/api/logs/export?level=issues"} {
		expectStatus(t, request(s, "GET", path, nil, nil, ""), 401)
	}
	cookie, _ := enroll(t, s)
	w := request(s, "GET", "/api/logs?level=ISSUES&q=needle", nil, cookie, "")
	expectStatus(t, w, 200)
	var page LogPage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || len(page.Entries) != 4 {
		t.Fatal("issues API returned wrong selection", err)
	}
	w = request(s, "GET", "/api/logs/export?level=issues&q=needle", nil, cookie, "")
	expectStatus(t, w, 200)
	ids := regexp.MustCompile(`\] #([0-9]+) `).FindAllStringSubmatch(w.Body.String(), -1)
	if len(ids) != len(page.Entries) || strings.Contains(w.Body.String(), "synthetic-log-secret") || strings.Contains(w.Body.String(), "[INFO]") || strings.Contains(w.Body.String(), "[DEBUG]") || strings.Contains(w.Body.String(), "[TRACE]") || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("issues export selection/redaction differs from viewer")
	}
	for i, row := range page.Entries {
		if ids[i][1] != row.ID {
			t.Fatal("issues export lost chronological order")
		}
	}
	for _, path := range []string{"/api/logs?level=unknown", "/api/logs/export?level=unknown", "/api/logs/export?level=issues&level=info"} {
		expectStatus(t, request(s, "GET", path, nil, cookie, ""), 400)
	}
}
