package localadmin

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const testPassword = "local-test-password-only"

func testConfig(t *testing.T) Config {
	t.Helper()
	return Config{Addr: "127.0.0.1:3010", DataDir: t.TempDir(), Overview: func() Overview { return Overview{BotAccounts: []string{"12345678"}, ConnectionState: "connected"} }, Plugins: []Plugin{{Name: "test-plugin", Brief: "local fixture", Usage: "help"}}}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	s, err := New(testConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func request(s *Server, method, path string, body any, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = strings.NewReader(string(b))
	}
	r := httptest.NewRequest(method, "http://"+s.cfg.Addr+path, reader)
	r.RemoteAddr = "127.0.0.1:54321"
	if method != "GET" {
		r.Header.Set("Origin", "http://"+s.cfg.Addr)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Local-Request", "1")
		r.Header.Set("X-CSRF-Token", csrf)
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}

func expectStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Fatalf("HTTP %d, want %d: %s", w.Code, expected, w.Body.String())
	}
}

func enroll(t *testing.T, s *Server) (*http.Cookie, string) {
	t.Helper()
	w := request(s, "POST", "/api/setup", map[string]string{"setupCode": s.setupCode, "password": testPassword}, nil, "")
	expectStatus(t, w, 200)
	var value struct {
		CSRF string `json:"csrfToken"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &value); err != nil {
		t.Fatal(err)
	}
	return w.Result().Cookies()[0], value.CSRF
}

func TestAuthLifecycleAndPersistence(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/api/overview", "/api/plugins"} {
		expectStatus(t, request(s, "GET", path, nil, nil, ""), 401)
	}
	state := request(s, "GET", "/api/session", nil, nil, "")
	if !strings.Contains(state.Body.String(), `"setupRequired":true`) || strings.Contains(state.Body.String(), s.setupCode) {
		t.Fatal("setup status exposed code or was incorrect")
	}
	expectStatus(t, request(s, "POST", "/api/setup", map[string]string{"setupCode": "incorrect", "password": testPassword}, nil, ""), 401)
	expectStatus(t, request(s, "POST", "/api/setup", map[string]string{"setupCode": s.setupCode, "password": "short"}, nil, ""), 400)
	cookie, csrf := enroll(t, s)
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || csrf == "" {
		t.Fatal("invalid session properties")
	}
	expectStatus(t, request(s, "POST", "/api/setup", map[string]string{"setupCode": "anything", "password": testPassword}, nil, ""), 409)
	for _, path := range []string{"/api/overview", "/api/plugins", "/api/session"} {
		w := request(s, "GET", path, nil, cookie, "")
		expectStatus(t, w, 200)
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("sensitive response was cacheable")
		}
	}
	b, err := os.ReadFile(filepath.Join(s.cfg.DataDir, "credential.json"))
	if err != nil || strings.Contains(string(b), testPassword) {
		t.Fatal("password persisted as plaintext or was not saved")
	}
	if _, err = os.Stat(filepath.Join(s.cfg.DataDir, "setup-code.txt")); !os.IsNotExist(err) {
		t.Fatal("setup code was not removed")
	}
	expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": "incorrect"}, nil, ""), 401)
	expectStatus(t, request(s, "POST", "/api/logout", map[string]string{}, cookie, "incorrect"), 403)
	expectStatus(t, request(s, "POST", "/api/logout", map[string]string{}, cookie, csrf), 200)
	expectStatus(t, request(s, "GET", "/api/overview", nil, cookie, ""), 401)
	reloaded, err := New(s.cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(reloaded, "POST", "/api/login", map[string]string{"password": testPassword}, nil, ""), 200)
	expectStatus(t, request(reloaded, "GET", "/api/overview", nil, cookie, ""), 401)
}

func TestBoundaryChecks(t *testing.T) {
	s := newTestServer(t)
	cookie, _ := enroll(t, s)
	for _, tc := range []struct {
		name   string
		mutate func(*http.Request)
	}{
		{"remote peer", func(r *http.Request) { r.RemoteAddr = "192.168.1.2:1234" }},
		{"rebinding host", func(r *http.Request) { r.Host = "evil.example:3010" }},
		{"null origin", func(r *http.Request) { r.Header.Set("Origin", "null") }},
		{"cross origin", func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") }},
		{"other localhost port", func(r *http.Request) { r.Header.Set("Origin", "http://127.0.0.1:3011") }},
		{"fetch cross-site", func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "http://"+s.cfg.Addr+"/api/overview", nil)
			r.RemoteAddr = "127.0.0.1:1234"
			r.AddCookie(cookie)
			tc.mutate(r)
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			expectStatus(t, w, 403)
		})
	}
	for _, missing := range []string{"Origin", "Content-Type", "X-Local-Request"} {
		r := httptest.NewRequest("POST", "http://"+s.cfg.Addr+"/api/login", strings.NewReader(`{"password":"test"}`))
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("Origin", "http://"+s.cfg.Addr)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Local-Request", "1")
		r.Header.Del(missing)
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		expectStatus(t, w, 403)
	}
	expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": testPassword, "unknown": "x"}, nil, ""), 400)
	expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": strings.Repeat("x", 5000)}, nil, ""), 400)
	expectStatus(t, request(s, "GET", "/api/logout", nil, cookie, ""), 405)
	for _, path := range []string{"/credential.json", "/setup-code.txt", "/.local/admin/credential.json", "/data/control/plugins.db"} {
		expectStatus(t, request(s, "GET", path, nil, cookie, ""), 404)
	}
}

func TestExpiryAndLoginThrottle(t *testing.T) {
	s := newTestServer(t)
	cookie, _ := enroll(t, s)
	now := time.Now().Add(9 * time.Hour)
	s.now = func() time.Time { return now }
	expectStatus(t, request(s, "GET", "/api/overview", nil, cookie, ""), 401)
	for range 5 {
		expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": "incorrect"}, nil, ""), 401)
	}
	expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": testPassword}, nil, ""), 429)
	now = now.Add(61 * time.Second)
	expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": testPassword}, nil, ""), 200)
}

func TestCorruptCredentialFailsClosed(t *testing.T) {
	s := newTestServer(t)
	enroll(t, s)
	path := filepath.Join(s.cfg.DataDir, "credential.json")
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := New(s.cfg); err == nil {
		t.Fatal("corrupt credentials accepted")
	}
	b, _ := os.ReadFile(path)
	if string(b) != "broken" {
		t.Fatal("corrupt credential overwritten")
	}
}

func TestBindingAndOfflineAssets(t *testing.T) {
	for _, addr := range []string{"0.0.0.0:3010", "localhost:3010", "[::1]:3010", "127.0.0.1:0", "127.0.0.1:65536", "127.0.0.1:03010"} {
		cfg := testConfig(t)
		cfg.Addr = addr
		if _, err := Start(cfg); err == nil {
			t.Fatalf("accepted address %s", addr)
		}
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	cfg := testConfig(t)
	cfg.Addr = listener.Addr().String()
	if _, err := Start(cfg); err == nil {
		t.Fatal("accepted occupied port")
	}
	listener.Close()
	s, err := Start(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close(context.Background())
	client := &http.Client{Timeout: 2 * time.Second}
	for _, path := range []string{"/", "/app.js", "/theme.js", "/style.css", "/appearance.css"} {
		response, err := client.Get("http://" + cfg.Addr + path)
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || response.StatusCode != 200 || len(b) == 0 {
			t.Fatalf("offline asset failed: %s", path)
		}
		if strings.Contains(string(b), "https://") || strings.Contains(string(b), "http://cdn") {
			t.Fatalf("external asset in %s", path)
		}
		if response.Header.Get("Content-Security-Policy") == "" {
			t.Fatal("CSP missing")
		}
	}
}

type unusedCaller struct{}

func (unusedCaller) CallAPI(context.Context, zero.APIRequest) (zero.APIResponse, error) {
	panic("overview must not call OneBot APIs")
}

func TestOverviewReflectsDisconnect(t *testing.T) {
	const id int64 = 987654321
	read := RuntimeOverview(time.Now(), 1, 3)
	if read().ConnectionState != "waiting" {
		t.Fatal("unexpected initial state")
	}
	zero.APICallers.Store(id, unusedCaller{})
	defer zero.APICallers.Delete(id)
	if value := read(); value.ConnectionState != "connected" || len(value.BotAccounts) != 1 {
		t.Fatal("connection absent")
	}
	zero.APICallers.Delete(id)
	if value := read(); value.ConnectionState != "waiting" || len(value.BotAccounts) != 0 {
		t.Fatal("stale connected state")
	}
	if RuntimeOverview(time.Now(), 0, 0)().ConnectionState != "unconfigured" {
		t.Fatal("unconfigured state missing")
	}
}
