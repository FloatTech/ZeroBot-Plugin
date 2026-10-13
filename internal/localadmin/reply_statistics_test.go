package localadmin

import (
	"encoding/json"
	"testing"

	"github.com/FloatTech/ZeroBot-Plugin/internal/replystats"
)

func TestReplyStatisticsRequireAuthAndMissingSourceIsNotZero(t *testing.T) {
	cfg := testConfig(t)
	c := replystats.New()
	reads := 0
	cfg.ReplyTasks = func() replystats.Snapshot { reads++; return c.Read() }
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(s, "GET", "/api/overview", nil, nil, ""), 401)
	if reads != 0 {
		t.Fatal("unauthenticated read")
	}
	cookie, _ := enroll(t, s)
	w := request(s, "GET", "/api/overview", nil, cookie, "")
	expectStatus(t, w, 200)
	var got Overview
	if err = json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ReplyTasks == nil || got.ReplyTasks.Started != 0 || len(got.ReplyTasks.Categories) != 3 || got.ReplyTasks.Scope != "bot-matcher-entry-to-onebot-send-receipt" || reads != 1 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("reply statistics missing or cacheable")
	}
	s.cfg.ReplyTasks = nil
	w = request(s, "GET", "/api/overview", nil, cookie, "")
	json.Unmarshal(w.Body.Bytes(), &got)
	if got.ReplyTasks != nil {
		t.Fatal("missing source reported as zero")
	}
	expectStatus(t, request(s, "POST", "/api/overview", map[string]string{}, cookie, ""), 405)
}
