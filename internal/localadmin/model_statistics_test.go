package localadmin

import (
	"encoding/json"
	"github.com/FloatTech/ZeroBot-Plugin/internal/modelstats"
	"github.com/FloatTech/ZeroBot-Plugin/internal/processidentity"
	"strings"
	"testing"
	"time"
)

func TestModelProcessIdentityIsAuthenticatedAndHostOwned(t *testing.T) {
	cfg := testConfig(t)
	collector := modelstats.New()
	identity := &processidentity.Identity{ProcessID: 12345, ProcessStartedAt: time.Date(2026, 10, 13, 0, 0, 0, 123456700, time.UTC), ProgramSHA256: strings.Repeat("A", 64)}
	cfg.ProcessIdentity = identity
	cfg.ModelRequests = func() modelstats.Snapshot {
		stats := collector.Read()
		stats.Process = &processidentity.Identity{ProcessID: 99999, ProcessStartedAt: time.Now(), ProgramSHA256: strings.Repeat("B", 64)}
		return stats
	}
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(s, "GET", "/api/overview", nil, nil, ""), 401)
	cookie, _ := enroll(t, s)
	w := request(s, "GET", "/api/overview", nil, cookie, "")
	expectStatus(t, w, 200)
	var got Overview
	if err = json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ModelRequests == nil || got.ModelRequests.Process == nil || *got.ModelRequests.Process != *identity {
		t.Fatalf("host identity missing: %s", w.Body)
	}
	var raw map[string]any
	json.Unmarshal(w.Body.Bytes(), &raw)
	process := raw["modelRequests"].(map[string]any)["process"].(map[string]any)
	if len(process) != 3 || process["processId"] != float64(12345) || process["processStartedAt"] != "2026-10-13T00:00:00.1234567Z" || process["programSHA256"] != strings.Repeat("A", 64) {
		t.Fatalf("unexpected process fields: %v", process)
	}
	if collector.Read().Process != nil {
		t.Fatal("overview mutated collector")
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("process identity cached")
	}
}

func TestMissingOrInvalidHostIdentityKeepsStatisticsWithoutClaimingAssociation(t *testing.T) {
	for _, identity := range []*processidentity.Identity{nil, {ProcessID: 1, ProcessStartedAt: time.Now(), ProgramSHA256: "invalid"}} {
		cfg := testConfig(t)
		cfg.ProcessIdentity = identity
		cfg.ModelRequests = modelstats.New().Read
		s, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		cookie, _ := enroll(t, s)
		w := request(s, "GET", "/api/overview", nil, cookie, "")
		expectStatus(t, w, 200)
		var got Overview
		if err = json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.ModelRequests == nil || got.ModelRequests.Process != nil || strings.Contains(w.Body.String(), `"process":`) {
			t.Fatalf("invented process identity: %s", w.Body)
		}
	}
}

func TestModelStatisticsRequireAuthAndAreReadOnly(t *testing.T) {
	cfg := testConfig(t)
	collector := modelstats.New()
	reads := 0
	cfg.ModelRequests = func() modelstats.Snapshot { reads++; return collector.Read() }
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(s, "GET", "/api/overview", nil, nil, ""), 401)
	if reads != 0 {
		t.Fatal("unauthenticated metric read")
	}
	cookie, _ := enroll(t, s)
	w := request(s, "GET", "/api/overview", nil, cookie, "")
	expectStatus(t, w, 200)
	var got Overview
	if err = json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ModelRequests == nil || got.ModelRequests.Started != 0 || len(got.ModelRequests.Categories) != 4 || reads != 1 {
		t.Fatal("missing zero statistics")
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("statistics cacheable")
	}
	expectStatus(t, request(s, "POST", "/api/overview", map[string]string{}, cookie, ""), 405)
	cfg.ModelRequests = nil
	s, err = New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	login := request(s, "POST", "/api/login", map[string]string{"password": testPassword}, nil, "")
	expectStatus(t, login, 200)
	w = request(s, "GET", "/api/overview", nil, login.Result().Cookies()[0], "")
	json.Unmarshal(w.Body.Bytes(), &got)
	if got.ModelRequests != nil {
		t.Fatal("missing source reported as zero")
	}
}
