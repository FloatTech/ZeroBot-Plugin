package localadmin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicNavigationKeepsAPIsIsolated(t *testing.T) {
	s := newTestServer(t)
	setupCode := s.setupCode
	cookie, _ := enroll(t, s)
	baseline := request(s, http.MethodGet, "/", nil, nil, "")
	expectStatus(t, baseline, http.StatusOK)
	for _, tc := range []struct {
		name, method, path, site, mode, dest, origin string
		status                                       int
	}{
		{"extension navigation", "GET", "/", "cross-site", "navigate", "document", "", 200},
		{"same-site navigation", "GET", "/", "same-site", "navigate", "document", "", 200},
		{"normal navigation", "GET", "/", "none", "navigate", "document", "", 200},
		{"root fetch", "GET", "/", "cross-site", "cors", "empty", "", 403},
		{"root image", "GET", "/", "cross-site", "no-cors", "image", "", 403},
		{"root iframe", "GET", "/", "cross-site", "navigate", "iframe", "", 403},
		{"missing destination", "GET", "/", "cross-site", "navigate", "", "", 403},
		{"missing mode", "GET", "/", "cross-site", "", "document", "", 403},
		{"unknown site", "GET", "/", "unknown", "navigate", "document", "", 403},
		{"foreign origin", "GET", "/", "cross-site", "navigate", "document", "https://evil.example", 403},
		{"null origin", "GET", "/", "cross-site", "navigate", "document", "null", 403},
		{"session navigation", "GET", "/api/session", "cross-site", "navigate", "document", "", 403},
		{"authenticated navigation", "GET", "/api/overview", "cross-site", "navigate", "document", "", 403},
		{"same-site API", "GET", "/api/overview", "same-site", "navigate", "document", "", 403},
		{"script navigation", "GET", "/app.js", "cross-site", "navigate", "document", "", 403},
		{"login submission", "POST", "/api/login", "cross-site", "navigate", "document", "http://" + s.cfg.Addr, 403},
		{"root submission", "POST", "/", "cross-site", "navigate", "document", "http://" + s.cfg.Addr, 403},
		{"authenticated same-origin API", "GET", "/api/overview", "same-origin", "cors", "empty", "", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://"+s.cfg.Addr+tc.path, nil)
			r.RemoteAddr = "127.0.0.1:1234"
			r.AddCookie(cookie)
			r.Header.Set("Sec-Fetch-Site", tc.site)
			r.Header.Set("Sec-Fetch-Mode", tc.mode)
			r.Header.Set("Sec-Fetch-Dest", tc.dest)
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("X-Local-Request", "1")
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			expectStatus(t, w, tc.status)
			if tc.status == 200 && tc.path == "/" {
				if w.Body.String() != baseline.Body.String() || strings.Contains(w.Body.String(), setupCode) {
					t.Fatal("navigation returned private or session-dependent content")
				}
				if len(w.Result().Cookies()) != 0 || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("X-Frame-Options") != "DENY" {
					t.Fatal("public navigation changed session or browser protections")
				}
			}
		})
	}
	for _, tc := range []struct {
		name, remote, host string
	}{
		{"remote browser", "192.168.1.2:1234", s.cfg.Addr},
		{"rebinding host", "127.0.0.1:1234", "evil.example:3010"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "http://"+s.cfg.Addr+"/", nil)
			r.RemoteAddr, r.Host = tc.remote, tc.host
			r.Header.Set("Sec-Fetch-Site", "cross-site")
			r.Header.Set("Sec-Fetch-Mode", "navigate")
			r.Header.Set("Sec-Fetch-Dest", "document")
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			expectStatus(t, w, 403)
		})
	}
}
