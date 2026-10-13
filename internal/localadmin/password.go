package localadmin

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"
)

// Caller holds s.mu; login and password changes share failed-attempt throttling.
func (s *Server) authRateLimited() bool {
	cutoff := s.now().Add(-time.Minute)
	kept := s.failures[:0]
	for _, at := range s.failures {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	s.failures = kept
	return len(kept) >= 5
}

var errCredentialConflict = errors.New("后台凭据已被外部修改，请检查后重新启动后台")

// A private temporary file is synced and atomically replaces the credential.
// Memory and sessions change only after the file replacement succeeds.
func (s *Server) persistCredential(ctx context.Context, next *credential) error {
	path := filepath.Join(s.cfg.DataDir, "credential.json")
	old, err := readBoundedFile(path, 4096)
	if err != nil {
		return errCredentialConflict
	}
	var disk credential
	if json.Unmarshal(old, &disk) != nil || disk.Version != s.credential.Version || !bytes.Equal(disk.Salt, s.credential.Salt) || !bytes.Equal(disk.Hash, s.credential.Hash) {
		return errCredentialConflict
	}
	b, err := json.Marshal(next)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(s.cfg.DataDir, ".admin-credential-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = protectPath(f.Name(), false); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	current, err := readBoundedFile(path, 4096)
	if err != nil || !bytes.Equal(old, current) {
		return errCredentialConflict
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return s.replaceCredential(f.Name(), path)
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	if !readJSONLimit(w, r, &body, 12<<10) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// Authorization may have preceded another password change or expiry.
	_, active, ok := s.currentSession(r)
	if !ok {
		apiError(w, 401, "请先登录本地后台")
		return
	}
	if subtle.ConstantTimeCompare([]byte(active.CSRF), []byte(r.Header.Get("X-CSRF-Token"))) != 1 {
		apiError(w, 403, "请求校验失败，请刷新页面")
		return
	}
	if len(body.CurrentPassword) == 0 || len(body.CurrentPassword) > 512 || !utf8.ValidString(body.NewPassword) || utf8.RuneCountInString(body.NewPassword) < 12 || len(body.NewPassword) > 512 {
		apiError(w, 400, "请填写当前密码；新密码至少12个字符，且不超过512字节")
		return
	}
	if body.NewPassword != body.ConfirmPassword {
		apiError(w, 400, "两次输入的新密码不一致")
		return
	}
	if body.NewPassword == body.CurrentPassword {
		apiError(w, 400, "新密码需与当前密码不同")
		return
	}
	if s.authRateLimited() {
		apiError(w, 429, "尝试次数过多，请一分钟后重试")
		return
	}
	if !s.validPassword(body.CurrentPassword) {
		s.failures = append(s.failures, s.now())
		apiError(w, 403, "当前后台密码不正确")
		return
	}
	next := &credential{Version: 1, Salt: randomBytes(16)}
	var err error
	next.Hash, err = derivePassword(body.NewPassword, next.Salt)
	if err != nil {
		apiError(w, 500, "密码保存失败，旧密码保留")
		return
	}
	if r.Context().Err() != nil {
		return
	}
	if _, _, ok = s.currentSession(r); !ok {
		apiError(w, 401, "请先登录本地后台")
		return
	}
	if err = s.persistCredential(r.Context(), next); err != nil {
		if errors.Is(err, errCredentialConflict) {
			apiError(w, 409, errCredentialConflict.Error())
		} else {
			apiError(w, 500, "密码保存失败，旧密码保留，请检查后台数据目录权限")
		}
		return
	}
	s.credential = next
	clear(s.sessions)
	s.failures = nil
	http.SetCookie(w, &http.Cookie{Name: "local_admin_session", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	jsonReply(w, 200, map[string]bool{"authenticated": false, "passwordChanged": true})
}
