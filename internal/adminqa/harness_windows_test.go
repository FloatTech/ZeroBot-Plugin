package adminqa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixtureAuthenticationKeepsCustomAndDefaultPasswordsSeparate(t *testing.T) {
	for _, setup := range []bool{true, false} {
		for _, custom := range []bool{true, false} {
			dir := t.TempDir()
			if setup {
				if err := os.WriteFile(filepath.Join(dir, "setup-code.txt"), []byte(" synthetic-code\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			wantPassword := Password
			if custom {
				wantPassword = "QA-synthetic-random-password-only"
			}
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				wantPath := "/api/login"
				if setup {
					wantPath = "/api/setup"
				}
				var body map[string]string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if r.Method != "POST" || r.URL.Path != wantPath || body["password"] != wantPassword {
					t.Error("incorrect fixture authentication request")
				}
				if setup && body["setupCode"] != "synthetic-code" {
					t.Error("setup code not trimmed")
				}
				if r.Header.Get("Origin") != "http://"+r.Host || r.Header.Get("X-Local-Request") != "1" {
					t.Error("local authentication headers missing")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"csrfToken":"synthetic-session"}`))
			}))
			client := NewClient(strings.TrimPrefix(server.URL, "http://"))
			var err error
			if custom {
				_, err = client.AuthWithPassword(dir, wantPassword)
			} else {
				_, err = client.Auth(dir)
			}
			server.Close()
			if err != nil || !called || client.CSRF != "synthetic-session" {
				t.Fatalf("fixture authentication did not establish a session: %v", err)
			}
		}
	}
}

func TestFixtureAuthenticationRejectsUnconfirmedSessions(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusUnauthorized} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":"synthetic rejection"}`))
		}))
		client := NewClient(strings.TrimPrefix(server.URL, "http://"))
		_, err := client.AuthWithPassword(t.TempDir(), "QA-custom-password-only")
		server.Close()
		if err == nil || client.CSRF != "" {
			t.Fatal("unconfirmed fixture session was accepted")
		}
	}
}
