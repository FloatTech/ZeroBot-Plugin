package localadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func agedBackup(t *testing.T, s *BackupStore, scope string, age int) string {
	t.Helper()
	id, err := s.Create(scope)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.directory, id)
	if scope == "control-db" {
		path = filepath.Join(path, "manifest.json")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	createdAt := s.now().AddDate(0, 0, -age).UTC()
	if scope == "control-db" {
		var value controlManifest
		if json.Unmarshal(raw, &value) != nil {
			t.Fatal("bad fixture")
		}
		value.CreatedAt = createdAt
		raw, _ = json.Marshal(value)
	} else {
		// Preserve RawMessage and large integers so aging metadata cannot
		// invalidate the checksum of the actual backed-up configuration.
		var value fileBackup
		if json.Unmarshal(raw, &value) != nil {
			t.Fatal("bad fixture")
		}
		value.CreatedAt = createdAt
		raw, _ = json.Marshal(value)
	}
	if os.WriteFile(path, raw, 0600) != nil {
		t.Fatal("fixture write failed")
	}
	return id
}

func cleanupRequest(v CleanupPreview) CleanupRequest {
	return CleanupRequest{Policy: v.Policy, Revision: v.Revision, AsOf: v.AsOf, ExpiresAt: v.ExpiresAt, Token: v.Token, Confirm: true}
}

func TestCleanupRetentionScopesIntegrityAndLiveState(t *testing.T) {
	s, c := backupFixture(t)
	s.now = func() time.Time { return time.Date(2026, 10, 3, 4, 0, 0, 0, time.UTC) }
	protected := map[string]bool{}
	for _, scope := range []string{"startup-config", "control-db", "admin-options"} {
		protected[agedBackup(t, s, scope, 0)] = true
		for i := 1; i <= 6; i++ {
			id := agedBackup(t, s, scope, i*10)
			if i <= 2 {
				protected[id] = true
			}
		}
	}
	invalid := agedBackup(t, s, "startup-config", 90)
	if os.WriteFile(filepath.Join(s.directory, invalid), []byte("invalid"), 0600) != nil {
		t.Fatal("fixture")
	}
	legacy := agedBackup(t, s, "startup-config", 90)
	path := filepath.Join(s.directory, legacy)
	raw, _ := os.ReadFile(path)
	var legacyFile fileBackup
	_ = json.Unmarshal(raw, &legacyFile)
	legacyFile.SHA256 = ""
	raw, _ = json.Marshal(legacyFile)
	_ = os.WriteFile(path, raw, 0600)
	extra := agedBackup(t, s, "control-db", 90)
	if os.WriteFile(filepath.Join(s.directory, extra, "unrelated.txt"), []byte("keep"), 0600) != nil {
		t.Fatal("extra fixture")
	}
	startup, _ := os.ReadFile(s.connections.path)
	options, _ := os.ReadFile(s.optionsPath)
	dataBefore := c.GetData(101)
	v, err := s.PreviewCleanup(context.Background(), CleanupPolicy{3, 7}, "session")
	if err != nil || len(v.Candidates) != 12 || v.Skipped != 3 || v.Bytes <= 0 {
		t.Fatal(v, err)
	}
	for _, entry := range v.Candidates {
		if protected[entry.ID] {
			t.Fatal("protected backup selected")
		}
	}
	if s.cleanupTicket(cleanupRequest(v), "session") != v.Token {
		t.Fatal("ticket changed without mutation")
	}
	check, _, checkErr := s.cleanupPlan(context.Background(), v.Policy, v.AsOf)
	if checkErr != nil || check.Revision != v.Revision {
		t.Fatalf("inventory unstable: before=%s after=%s candidates=%d skipped=%d error=%v", v.Revision, check.Revision, len(check.Candidates), check.Skipped, checkErr)
	}
	result, err := s.Cleanup(context.Background(), cleanupRequest(v), "session")
	if err != nil || result.Partial || len(result.Deleted) != 12 || result.Bytes != v.Bytes {
		t.Fatal(result, err)
	}
	for id := range protected {
		if _, err := os.Stat(filepath.Join(s.directory, id)); err != nil {
			t.Fatal("protected backup lost", err)
		}
	}
	for _, id := range []string{invalid, legacy, extra} {
		if _, err := os.Stat(filepath.Join(s.directory, id)); err != nil {
			t.Fatal("unverified backup removed")
		}
	}
	nowStartup, _ := os.ReadFile(s.connections.path)
	nowOptions, _ := os.ReadFile(s.optionsPath)
	if !bytes.Equal(startup, nowStartup) || !bytes.Equal(options, nowOptions) || c.GetData(101) != dataBefore {
		t.Fatal("live state changed")
	}
	if _, err = s.Cleanup(context.Background(), cleanupRequest(v), "session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("replayed deletion accepted", err)
	}
}

func TestCleanupTicketsInventoryAndCancellation(t *testing.T) {
	s, _ := backupFixture(t)
	for i := 0; i < 5; i++ {
		agedBackup(t, s, "startup-config", 30+i)
	}
	v, err := s.PreviewCleanup(context.Background(), CleanupPolicy{3, 7}, "session")
	if err != nil {
		t.Fatal(err)
	}
	req := cleanupRequest(v)
	req.Confirm = false
	if _, err = s.Cleanup(context.Background(), req, "session"); !errors.Is(err, errBackupInvalid) {
		t.Fatal("unconfirmed request accepted")
	}
	req.Confirm = true
	if _, err = s.Cleanup(context.Background(), req, "other-session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("other session accepted")
	}
	req.Policy.KeepDays = 8
	if _, err = s.Cleanup(context.Background(), req, "session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("modified policy accepted")
	}
	req = cleanupRequest(v)
	req.Token = strings.Repeat("0", 64)
	if _, err = s.Cleanup(context.Background(), req, "session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("invalid token accepted")
	}
	req = cleanupRequest(v)
	oldNow := s.now
	s.now = func() time.Time { return oldNow().Add(11 * time.Minute) }
	if _, err = s.Cleanup(context.Background(), req, "session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("expired preview accepted")
	}
	s.now = oldNow
	if _, err = s.Create("admin-options"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Cleanup(context.Background(), req, "session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("changed inventory accepted")
	}
	v, err = s.PreviewCleanup(context.Background(), CleanupPolicy{3, 7}, "session")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.directory, v.Candidates[0].ID)
	raw, _ := os.ReadFile(path)
	var file fileBackup
	_ = json.Unmarshal(raw, &file)
	file.Config = []byte(strings.ReplaceAll(string(file.Config), "secret-A", "changed-synthetic-key"))
	file.SHA256 = digestBytes(file.Config)
	raw, _ = json.Marshal(file)
	_ = os.WriteFile(path, raw, 0600)
	if _, err = s.Cleanup(context.Background(), cleanupRequest(v), "session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("changed backup accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = s.PreviewCleanup(ctx, CleanupPolicy{3, 7}, "session"); err == nil {
		t.Fatal("canceled scan accepted")
	}
	for _, policy := range []CleanupPolicy{{2, 7}, {201, 7}, {3, 0}, {3, 3651}} {
		if _, err = s.PreviewCleanup(context.Background(), policy, "session"); err == nil {
			t.Fatal("invalid policy")
		}
	}
	if _, _, err = s.cleanupFiles("../outside.json"); err == nil {
		t.Fatal("traversal accepted")
	}
}

func TestCleanupPartialDeletionIsVisible(t *testing.T) {
	s, _ := backupFixture(t)
	for i := 0; i < 5; i++ {
		agedBackup(t, s, "startup-config", 30+i)
	}
	v, err := s.PreviewCleanup(context.Background(), CleanupPolicy{3, 7}, "session")
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	s.cleanupRemove = func(root *os.Root, name string) error {
		calls++
		if calls == 2 {
			return errors.New("sensitive filesystem detail")
		}
		return root.Remove(name)
	}
	result, err := s.Cleanup(context.Background(), cleanupRequest(v), "session")
	if err != nil || !result.Partial || len(result.Deleted) != 1 || strings.Contains(result.Message, "sensitive") {
		t.Fatal(result, err)
	}
	if _, err = os.Stat(filepath.Join(s.directory, v.Candidates[0].ID)); !os.IsNotExist(err) {
		t.Fatal("first deletion not reflected")
	}
	if _, err = os.Stat(filepath.Join(s.directory, v.Candidates[1].ID)); err != nil {
		t.Fatal("failed target lost")
	}
}

func TestCleanupBatchAndControlResidualProtection(t *testing.T) {
	s, _ := backupFixture(t)
	for i := 0; i < 105; i++ {
		agedBackup(t, s, "admin-options", 30+i)
	}
	v, err := s.PreviewCleanup(context.Background(), CleanupPolicy{3, 7}, "session")
	if err != nil || len(v.Candidates) != 100 || !v.More {
		t.Fatal(v, err)
	}
	result, err := s.Cleanup(context.Background(), cleanupRequest(v), "session")
	if err != nil || result.Partial || len(result.Deleted) != 100 {
		t.Fatal(result, err)
	}
	v, err = s.PreviewCleanup(context.Background(), CleanupPolicy{3, 7}, "session")
	if err != nil || len(v.Candidates) != 2 || v.More {
		t.Fatal(v, err)
	}
	for i := 0; i < 4; i++ {
		agedBackup(t, s, "control-db", 30+i)
	}
	// A failed second file removal must not be reported as a complete deletion.
	s.cleanupRemove = func(root *os.Root, name string) error {
		if strings.HasSuffix(name, "manifest.json") {
			return errors.New("injected")
		}
		return root.Remove(name)
	}
	// The control fixture is newer than remaining eligible file backups, so
	// the partial control failure occurs before any complete deletion.
	controlID := agedBackup(t, s, "control-db", 120)
	v, err = s.PreviewCleanup(context.Background(), CleanupPolicy{3, 90}, "session")
	if err != nil {
		t.Fatal(err)
	}
	result, err = s.Cleanup(context.Background(), cleanupRequest(v), "session")
	if err != nil || !result.Partial || len(result.Deleted) != 0 {
		t.Fatal(result, err)
	}
	if _, _, err = s.load(controlID); err == nil {
		t.Fatal("incomplete control backup remains valid")
	}
	v, err = s.PreviewCleanup(context.Background(), CleanupPolicy{3, 90}, "session")
	if err != nil || v.Skipped != 1 {
		t.Fatal(v, err)
	}
}

func TestCleanupAPIAuthAndSessionBinding(t *testing.T) {
	store, _ := backupFixture(t)
	s, err := New(Config{Addr: "127.0.0.1:3010", DataDir: t.TempDir(), Overview: func() Overview { return Overview{} }, Backups: store})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/backups/cleanup-preview", "/api/backups/cleanup"} {
		expectStatus(t, request(s, "POST", path, map[string]any{}, nil, ""), http.StatusUnauthorized)
	}
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/backups/cleanup-preview", CleanupPolicy{3, 7}, cookie, "bad"), http.StatusForbidden)
	for i := 0; i < 4; i++ {
		agedBackup(t, store, "admin-options", 30+i)
	}
	w := request(s, "POST", "/api/backups/cleanup-preview", CleanupPolicy{3, 7}, cookie, csrf)
	expectStatus(t, w, 200)
	var preview CleanupPreview
	if json.Unmarshal(w.Body.Bytes(), &preview) != nil || len(preview.Candidates) != 1 {
		t.Fatal(w.Body.String())
	}
	if strings.Contains(w.Body.String(), "secret-A") || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("preview privacy")
	}
	req := cleanupRequest(preview)
	req.Confirm = false
	expectStatus(t, request(s, "POST", "/api/backups/cleanup", req, cookie, csrf), 400)
	req.Confirm = true
	expectStatus(t, request(s, "POST", "/api/backups/cleanup", req, cookie, csrf), 200)
	expectStatus(t, request(s, "POST", "/api/backups/cleanup", req, cookie, csrf), 409)
}
