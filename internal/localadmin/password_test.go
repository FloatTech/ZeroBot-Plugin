package localadmin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const changedTestPassword = " 新密码测试-password-🙂-only "

func passwordBody(current, next string) map[string]string {
	return map[string]string{"currentPassword": current, "newPassword": next, "confirmPassword": next}
}
func TestPasswordChangeRevokesAllSessionsAndPersists(t *testing.T) {
	s := newTestServer(t)
	cookie, csrf := enroll(t, s)
	login := request(s, "POST", "/api/login", map[string]string{"password": testPassword}, nil, "")
	expectStatus(t, login, 200)
	second := login.Result().Cookies()[0]
	old, _ := os.ReadFile(filepath.Join(s.cfg.DataDir, "credential.json"))
	oldSalt := bytes.Clone(s.credential.Salt)
	expectStatus(t, request(s, "POST", "/api/password", passwordBody("wrong", changedTestPassword), cookie, csrf), 403)
	expectStatus(t, request(s, "GET", "/api/overview", nil, cookie, ""), 200)
	w := request(s, "POST", "/api/password", passwordBody(testPassword, changedTestPassword), cookie, csrf)
	expectStatus(t, w, 200)
	if len(s.sessions) != 0 || len(w.Result().Cookies()) != 1 || w.Result().Cookies()[0].MaxAge != -1 || bytes.Contains(w.Body.Bytes(), []byte(changedTestPassword)) {
		t.Fatal("session invalidation or reply failed")
	}
	expectStatus(t, request(s, "GET", "/api/overview", nil, cookie, ""), 401)
	expectStatus(t, request(s, "GET", "/api/overview", nil, second, ""), 401)
	raw, _ := os.ReadFile(filepath.Join(s.cfg.DataDir, "credential.json"))
	if bytes.Equal(old, raw) || bytes.Equal(oldSalt, s.credential.Salt) || bytes.Contains(raw, []byte(testPassword)) || bytes.Contains(raw, []byte(changedTestPassword)) {
		t.Fatal("credential replacement/hash failed")
	}
	reloaded, err := New(s.cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(reloaded, "POST", "/api/login", map[string]string{"password": testPassword}, nil, ""), 401)
	expectStatus(t, request(reloaded, "POST", "/api/login", map[string]string{"password": changedTestPassword}, nil, ""), 200)
	expectStatus(t, request(reloaded, "POST", "/api/login", map[string]string{"password": strings.TrimSpace(changedTestPassword)}, nil, ""), 401)
	files, _ := os.ReadDir(s.cfg.DataDir)
	if len(files) != 1 || files[0].Name() != "credential.json" {
		t.Fatal("credential history/temp file retained")
	}
}
func TestPasswordChangeValidationAndSharedThrottle(t *testing.T) {
	s := newTestServer(t)
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/password", passwordBody(testPassword, changedTestPassword), nil, csrf), 401)
	expectStatus(t, request(s, "POST", "/api/password", passwordBody(testPassword, changedTestPassword), cookie, "incorrect-csrf"), 403)
	invalid := []map[string]string{passwordBody(testPassword, "short"), passwordBody(testPassword, strings.Repeat("汉", 171)), passwordBody("", changedTestPassword), passwordBody(strings.Repeat("a", 513), changedTestPassword), passwordBody(testPassword, testPassword), {"currentPassword": testPassword, "newPassword": changedTestPassword, "confirmPassword": "different"}, {"currentPassword": testPassword, "newPassword": changedTestPassword, "confirmPassword": changedTestPassword, "unknown": "x"}}
	for _, body := range invalid {
		expectStatus(t, request(s, "POST", "/api/password", body, cookie, csrf), 400)
	}
	for i := 0; i < 3; i++ {
		expectStatus(t, request(s, "POST", "/api/password", passwordBody("wrong", changedTestPassword), cookie, csrf), 403)
	}
	for i := 0; i < 2; i++ {
		expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": "wrong"}, nil, ""), 401)
	}
	expectStatus(t, request(s, "POST", "/api/password", passwordBody(testPassword, changedTestPassword), cookie, csrf), 429)
	expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": testPassword}, nil, ""), 429)
	later := s.now().Add(61 * time.Second)
	s.now = func() time.Time { return later }
	expectStatus(t, request(s, "POST", "/api/password", passwordBody(testPassword, strings.Repeat("x", 512)), cookie, csrf), 200)
	expectStatus(t, request(s, "POST", "/api/login", map[string]string{"password": strings.Repeat("x", 512)}, nil, ""), 200)
}
func TestPasswordChangeFailureAndExternalConflictKeepSessions(t *testing.T) {
	s := newTestServer(t)
	cookie, csrf := enroll(t, s)
	path := filepath.Join(s.cfg.DataDir, "credential.json")
	old, _ := os.ReadFile(path)
	s.replaceCredential = func(string, string) error { return errors.New("private path and secret should not escape") }
	w := request(s, "POST", "/api/password", passwordBody(testPassword, changedTestPassword), cookie, csrf)
	expectStatus(t, w, 500)
	current, _ := os.ReadFile(path)
	if !bytes.Equal(old, current) || !s.validPassword(testPassword) || strings.Contains(w.Body.String(), "private path") {
		t.Fatal("failed replacement changed/exposed credential")
	}
	expectStatus(t, request(s, "GET", "/api/overview", nil, cookie, ""), 200)
	files, _ := os.ReadDir(s.cfg.DataDir)
	if len(files) != 1 {
		t.Fatal("failed temporary credential not removed")
	}
	if err := os.WriteFile(path, []byte("externally changed"), 0600); err != nil {
		t.Fatal(err)
	}
	w = request(s, "POST", "/api/password", passwordBody(testPassword, changedTestPassword), cookie, csrf)
	expectStatus(t, w, 409)
	current, _ = os.ReadFile(path)
	if string(current) != "externally changed" || !s.validPassword(testPassword) {
		t.Fatal("external credential overwritten")
	}
	expectStatus(t, request(s, "GET", "/api/overview", nil, cookie, ""), 200)
	if err := os.WriteFile(path, old, 0600); err != nil {
		t.Fatal(err)
	}
	s.replaceCredential = replaceConfigFile
	later := s.now().Add(9 * time.Hour)
	s.now = func() time.Time { return later }
	expectStatus(t, request(s, "POST", "/api/password", passwordBody(testPassword, changedTestPassword), cookie, csrf), 401)
	current, _ = os.ReadFile(path)
	if !bytes.Equal(old, current) {
		t.Fatal("expired session changed credential")
	}
}

// The second request is authorized before the first change, then blocks reading its body.
type delayedPasswordReader struct {
	reader           io.Reader
	once             sync.Once
	entered, release chan struct{}
}

func (r *delayedPasswordReader) Read(p []byte) (int, error) {
	r.once.Do(func() { close(r.entered); <-r.release })
	return r.reader.Read(p)
}
func TestPasswordChangeRechecksAlreadyAuthorizedRequest(t *testing.T) {
	s := newTestServer(t)
	cookie, csrf := enroll(t, s)
	raw, _ := json.Marshal(passwordBody(testPassword, "other-test-password-only"))
	gated := &delayedPasswordReader{reader: bytes.NewReader(raw), entered: make(chan struct{}), release: make(chan struct{})}
	r := httptest.NewRequest(http.MethodPost, "http://"+s.cfg.Addr+"/api/password", gated)
	r.RemoteAddr = "127.0.0.1:54321"
	r.AddCookie(cookie)
	r.Header.Set("Origin", "http://"+s.cfg.Addr)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Local-Request", "1")
	r.Header.Set("X-CSRF-Token", csrf)
	w := httptest.NewRecorder()
	done := make(chan struct{})
	go func() { s.Handler().ServeHTTP(w, r); close(done) }()
	<-gated.entered
	expectStatus(t, request(s, "POST", "/api/password", passwordBody(testPassword, changedTestPassword), cookie, csrf), 200)
	close(gated.release)
	<-done
	expectStatus(t, w, 401)
	if !s.validPassword(changedTestPassword) || s.validPassword("other-test-password-only") {
		t.Fatal("preauthorized stale request replaced new password")
	}
}
