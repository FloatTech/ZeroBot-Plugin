package localadmin

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestOverviewAlertRetentionOrderLevelsAndMask(t *testing.T) {
	l, logger := logFixture()
	if v := l.recentAlerts(); len(v.Entries) != 0 || v.WarningCount != 0 || v.ErrorCount != 0 || v.Evicted {
		t.Fatal("empty alert state incorrect")
	}
	for i, level := range []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.InfoLevel, logrus.FatalLevel, logrus.PanicLevel, logrus.WarnLevel, logrus.ErrorLevel} {
		_ = l.Fire(&logrus.Entry{Time: time.Now(), Level: level, Message: fmt.Sprintf("alert-%d new-model-credential", i)})
	}
	l.AddSecrets("new-model-credential") // protect records that predate a saved key
	v := l.recentAlerts()
	if v.WarningCount != 2 || v.ErrorCount != 4 || v.Retained != 7 || len(v.Entries) != 5 {
		t.Fatal("level counts or bounds incorrect", v)
	}
	if !strings.HasPrefix(v.Entries[0].Message, "alert-6") || !strings.HasPrefix(v.Entries[4].Message, "alert-1") {
		t.Fatal("alerts not newest first")
	}
	for _, row := range v.Entries {
		if strings.Contains(row.Message, "new-model-credential") || row.Level == "info" {
			t.Fatal("alert redaction or level filter failed")
		}
	}
	for i := 0; i < logCapacity; i++ {
		logger.Info("new retained ordinary log")
	}
	v = l.recentAlerts()
	if !v.Evicted || v.Retained != logCapacity || v.ErrorCount != 0 || v.WarningCount != 0 || len(v.Entries) != 0 {
		t.Fatal("counts included evicted alerts")
	}
}

func TestOverviewAlertsAuthCurrentSecretsAndFailClosed(t *testing.T) {
	cfg := testConfig(t)
	l, logger := logFixture()
	cfg.Logs = l
	ai, c, _, _ := aiFixture(t)
	cfg.AI = ai
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(s, "GET", "/api/overview", nil, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	logger.Error("bare newer-overview-credential <script>alert(1)</script>")
	if err := c.SetExtra(map[string]any{"Key": "newer-overview-credential"}); err != nil {
		t.Fatal(err)
	}
	w := request(s, "GET", "/api/overview", nil, cookie, "")
	expectStatus(t, w, 200)
	var v Overview
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.ConnectionState != "connected" || v.Alerts == nil || v.Alerts.ErrorCount < 1 || strings.Contains(w.Body.String(), "newer-overview-credential") || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("authenticated overview did not safely enrich runtime state")
	}
	if err := c.Manager.D.Close(); err != nil {
		t.Fatal(err)
	}
	w = request(s, "GET", "/api/overview", nil, cookie, "")
	expectStatus(t, w, 200)
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.ConnectionState != "connected" || v.Alerts != nil || strings.Contains(w.Body.String(), "credential") || strings.Contains(w.Body.String(), "script") {
		t.Fatal("log read failure exposed text or changed connection status")
	}
	s.cfg.Logs = nil
	w = request(s, "GET", "/api/overview", nil, cookie, "")
	expectStatus(t, w, 200)
	if !strings.Contains(w.Body.String(), `"alerts":null`) {
		t.Fatal("missing log source misreported as empty logs")
	}
	expectStatus(t, request(s, "POST", "/api/logout", map[string]any{}, cookie, csrf), 200)
	expectStatus(t, request(s, "GET", "/api/overview", nil, cookie, ""), 401)
}
