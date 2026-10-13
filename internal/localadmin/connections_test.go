package localadmin

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const connectionFixture = `{"zero":{"super_users":[12345678],"nickname":["fixture"],"future":{"n":9007199254740993}},"ws":[{"Url":"ws://127.0.0.1:6700","AccessToken":"secret-A","extension":{"keep":true}},{"URL":"wss://bot.example/ws","AccessToken":"secret-B"}],"wss":[{"URL":"ws://127.0.0.1:6701","AccessToken":"secret-C"}],"force_base64_file":true,"unknown":{"nested":[1,2,3]}}`

func fixtureConnectionStore(t *testing.T) *ConnectionStore {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(connectionFixture), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := NewConnectionStore(path, filepath.Join(dir, "private", "backups"), []byte(connectionFixture))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func getSettings(t *testing.T, s *ConnectionStore) ConnectionSettings {
	t.Helper()
	v, err := s.Get()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func stringPointer(s string) *string { return &s }

func TestConnectionSavePreservesFieldsBackupAndRuntime(t *testing.T) {
	s := fixtureConnectionStore(t)
	v := getSettings(t, s)
	encoded, _ := json.Marshal(v)
	if strings.Contains(string(encoded), "secret-") || v.RestartRequired || len(v.Connections) != 3 || v.SuperUsers[0] != "12345678" {
		t.Fatal("bad initial redacted view")
	}
	users := []string{"87654321", "9007199254740993"}
	saved, err := s.Save(ConnectionUpdate{Revision: v.Revision, Connection: &ConnectionChange{Kind: "ws", Index: 0, URL: stringPointer("ws://127.0.0.1:7700/ws"), TokenAction: "keep"}, SuperUsers: &users})
	if err != nil {
		t.Fatal(err)
	}
	if !saved.RestartRequired || saved.BackupID == "" || saved.RunningConnections[0].URL != v.Connections[0].URL || saved.RunningSuperUsers[0] != "12345678" {
		t.Fatal("running state changed on save")
	}
	raw, _ := os.ReadFile(s.path)
	root, _ := decodeConfig(raw)
	original, _ := decodeConfig([]byte(connectionFixture))
	for _, key := range []string{"unknown", "force_base64_file", "wss"} {
		if !equivalentConfig(root[key], original[key]) {
			t.Fatalf("lost %s", key)
		}
	}
	_, zeroRaw, _ := field(root, "zero")
	zero, _ := object(zeroRaw)
	if !bytes.Contains(zero["future"], []byte("9007199254740993")) || !strings.Contains(string(raw), "secret-A") || !strings.Contains(string(raw), "secret-B") {
		t.Fatal("lost original fields, precision, or token")
	}
	var list []json.RawMessage
	_ = json.Unmarshal(root["ws"], &list)
	var originalList []json.RawMessage
	_ = json.Unmarshal(original["ws"], &originalList)
	if !equivalentConfig(list[1], originalList[1]) || !bytes.Contains(list[0], []byte(`"extension"`)) {
		t.Fatal("changed unedited connection")
	}
	backupRaw, err := os.ReadFile(filepath.Join(s.backups, saved.BackupID))
	if err != nil {
		t.Fatal(err)
	}
	var backup struct {
		Version int
		Scope   string
		Config  json.RawMessage
	}
	if json.Unmarshal(backupRaw, &backup) != nil || backup.Version != 1 || backup.Scope != "startup-config" || !equivalentConfig(backup.Config, []byte(connectionFixture)) {
		t.Fatal("backup did not contain old config")
	}
	restarted, err := NewConnectionStore(s.path, s.backups, raw)
	if err != nil {
		t.Fatal(err)
	}
	if getSettings(t, restarted).RestartRequired {
		t.Fatal("restart did not consume saved config")
	}
	if _, err = s.Save(ConnectionUpdate{Revision: v.Revision, SuperUsers: &users}); !errors.Is(err, errConfigConflict) {
		t.Fatal("stale write was accepted")
	}
}

func TestTokenKeepReplaceClearAndNoop(t *testing.T) {
	s := fixtureConnectionStore(t)
	v := getSettings(t, s)
	noop, err := s.Save(ConnectionUpdate{Revision: v.Revision, Connection: &ConnectionChange{Kind: "ws", Index: 0, TokenAction: "keep"}})
	if err != nil || noop.BackupID != "" || noop.Revision != v.Revision {
		t.Fatal("no-op created a backup or changed revision")
	}
	changed, err := s.Save(ConnectionUpdate{Revision: v.Revision, Connection: &ConnectionChange{Kind: "ws", Index: 0, TokenAction: "replace", Token: "new-fake-token"}})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(s.path)
	if !strings.Contains(string(raw), "new-fake-token") || strings.Contains(string(raw), "secret-A") {
		t.Fatal("replacement did not persist")
	}
	cleared, err := s.Save(ConnectionUpdate{Revision: changed.Revision, Connection: &ConnectionChange{Kind: "ws", Index: 0, TokenAction: "clear"}})
	if err != nil || cleared.Connections[0].TokenConfigured || !cleared.RunningConnections[0].TokenConfigured {
		t.Fatal("clear or runtime separation failed")
	}
}

func TestConnectionValidationDoesNotWrite(t *testing.T) {
	badChanges := []ConnectionChange{
		{Kind: "../ws", Index: 0, TokenAction: "keep"}, {Kind: "ws", Index: -1, TokenAction: "keep"}, {Kind: "ws", Index: 9, TokenAction: "keep"},
		{Kind: "ws", Index: 0, URL: stringPointer("http://127.0.0.1:6700"), TokenAction: "keep"},
		{Kind: "ws", Index: 0, URL: stringPointer("ws://127.0.0.1:0"), TokenAction: "keep"},
		{Kind: "ws", Index: 0, URL: stringPointer("ws://127.0.0.1:65536"), TokenAction: "keep"},
		{Kind: "ws", Index: 0, URL: stringPointer("ws://user:password@host/ws"), TokenAction: "keep"},
		{Kind: "ws", Index: 0, URL: stringPointer("ws://host/ws?token=secret"), TokenAction: "keep"},
		{Kind: "ws", Index: 0, TokenAction: "replace", Token: "***"}, {Kind: "ws", Index: 0, TokenAction: "replace"},
		{Kind: "ws", Index: 0, TokenAction: "replace", Token: "token\nheader"}, {Kind: "ws", Index: 0, TokenAction: "keep", Token: "unexpected"},
		{Kind: "ws", Index: 0, TokenAction: "clear", Token: "unexpected"}, {Kind: "ws", Index: 0, TokenAction: "unknown"},
	}
	s := fixtureConnectionStore(t)
	initial := getSettings(t, s)
	for i, change := range badChanges {
		if _, err := s.Save(ConnectionUpdate{Revision: initial.Revision, Connection: &change}); err == nil {
			t.Fatalf("accepted invalid change %d", i)
		}
	}
	for _, users := range [][]string{{}, {"abc"}, {"-1"}, {"01"}, {"123", "123"}, {"9223372036854775808"}} {
		if _, err := s.Save(ConnectionUpdate{Revision: initial.Revision, SuperUsers: &users}); err == nil {
			t.Fatal("accepted invalid owner")
		}
	}
	raw, _ := os.ReadFile(s.path)
	if string(raw) != connectionFixture {
		t.Fatal("invalid input changed disk")
	}
	if _, err := os.Stat(s.backups); !os.IsNotExist(err) {
		t.Fatal("invalid input created backups")
	}
}

func TestFormattedConfigSizeLimitDoesNotWrite(t *testing.T) {
	s := fixtureConnectionStore(t)
	raw := []byte(`{"zero":{"super_users":[12345678]},"ws":[],"extra":[` + strings.Repeat("0,", 200000) + `0]}`)
	if err := os.WriteFile(s.path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	view := getSettings(t, s)
	users := []string{"87654321"}
	if _, err := s.Save(ConnectionUpdate{Revision: view.Revision, SuperUsers: &users}); !errors.Is(err, errConfigValidation) {
		t.Fatal("accepted formatted output over startup size limit")
	}
	actual, err := os.ReadFile(s.path)
	if err != nil || !bytes.Equal(raw, actual) {
		t.Fatal("oversized output changed original")
	}
	if _, err := os.Stat(s.backups); !os.IsNotExist(err) {
		t.Fatal("oversized output created backups")
	}
}

func TestBackupAndReplaceFailureLeaveOriginal(t *testing.T) {
	for _, failure := range []string{"backup", "replace"} {
		t.Run(failure, func(t *testing.T) {
			s := fixtureConnectionStore(t)
			v := getSettings(t, s)
			users := []string{"87654321"}
			if failure == "backup" {
				if err := os.MkdirAll(filepath.Dir(s.backups), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(s.backups, []byte("blocked"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				s.replace = func(string, string) error { return errors.New("fake-secret-must-not-leak") }
			}
			_, err := s.Save(ConnectionUpdate{Revision: v.Revision, SuperUsers: &users})
			if err == nil || strings.Contains(err.Error(), "fake-secret") {
				t.Fatal("failure was not safely reported")
			}
			raw, _ := os.ReadFile(s.path)
			if string(raw) != connectionFixture {
				t.Fatal("failed write changed original")
			}
			files, _ := filepath.Glob(filepath.Join(filepath.Dir(s.path), ".admin-config-*.tmp"))
			if len(files) != 0 {
				t.Fatal("temporary config leaked")
			}
			if failure == "replace" {
				files, _ := os.ReadDir(s.backups)
				if len(files) != 1 {
					t.Fatal("write failure lost backup")
				}
			}
		})
	}
}

func TestConcurrentAndExternalEdits(t *testing.T) {
	s := fixtureConnectionStore(t)
	initial := getSettings(t, s)
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			users := []string{"87654321"}
			_, err := s.Save(ConnectionUpdate{Revision: initial.Revision, SuperUsers: &users})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, errConfigConflict) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("concurrent writes succeeded %d times", success)
	}
	view := getSettings(t, s)
	if err := os.WriteFile(s.path, []byte(connectionFixture), 0600); err != nil {
		t.Fatal(err)
	}
	users := []string{"76543210"}
	if _, err := s.Save(ConnectionUpdate{Revision: view.Revision, SuperUsers: &users}); !errors.Is(err, errConfigConflict) {
		t.Fatal("external edit overwritten")
	}
}

func TestConnectionsAPIAuthCSRFAndRedaction(t *testing.T) {
	s := newTestServer(t)
	s.cfg.Connections = fixtureConnectionStore(t)
	expectStatus(t, request(s, "GET", "/api/connections", nil, nil, ""), 401)
	view := getSettings(t, s.cfg.Connections)
	users := []string{"87654321"}
	update := ConnectionUpdate{Revision: view.Revision, SuperUsers: &users}
	expectStatus(t, request(s, "POST", "/api/connections", update, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/connections", update, cookie, "wrong"), 403)
	rsp := request(s, "GET", "/api/connections", nil, cookie, "")
	expectStatus(t, rsp, 200)
	if strings.Contains(rsp.Body.String(), "secret-") || rsp.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("secret leaked or response cached")
	}
	expectStatus(t, request(s, "POST", "/api/connections", update, cookie, csrf), 200)
	expectStatus(t, request(s, "POST", "/api/connections", update, cookie, csrf), 409)
	invalid := ConnectionUpdate{Revision: getSettings(t, s.cfg.Connections).Revision, Connection: &ConnectionChange{Kind: "ws", Index: 0, TokenAction: "replace", Token: "***"}}
	expectStatus(t, request(s, "POST", "/api/connections", invalid, cookie, csrf), 400)
	s.cfg.Connections.replace = func(string, string) error { return errors.New("injected disk failure") }
	update.Revision = getSettings(t, s.cfg.Connections).Revision
	users = []string{"34567890"}
	update.SuperUsers = &users
	expectStatus(t, request(s, "POST", "/api/connections", update, cookie, csrf), 500)
	expectStatus(t, request(s, "GET", "/backups/anything.json", nil, cookie, ""), 404)
}

func TestSensitiveURLAndBrokenJSON(t *testing.T) {
	s := fixtureConnectionStore(t)
	raw := strings.ReplaceAll(connectionFixture, "ws://127.0.0.1:6700", "ws://user:password@127.0.0.1:6700/ws?token=url-secret#private")
	if err := os.WriteFile(s.path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	v := getSettings(t, s)
	encoded, _ := json.Marshal(v)
	for _, secret := range []string{"password", "url-secret", "private"} {
		if bytes.Contains(encoded, []byte(secret)) {
			t.Fatal("URL credential leaked")
		}
	}
	if !v.Connections[0].AddressRedacted {
		t.Fatal("redaction not marked")
	}
	if err := os.WriteFile(s.path, []byte(`{"zero":`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(); err == nil {
		t.Fatal("broken JSON accepted")
	}
	if _, err := decodeConfig([]byte(`{"zero":{},"ws":[null]}`)); err == nil {
		t.Fatal("null connection accepted")
	}
	if _, err := decodeConfig([]byte(`{"zero":{},"ws":[{"URL":"ws://localhost","url":"ws://other"}]}`)); err == nil {
		t.Fatal("ambiguous fields accepted")
	}
	if _, err := decodeConfig([]byte(`{"zero":{},"ws":[],"ws":[]}`)); err == nil {
		t.Fatal("duplicate root field accepted")
	}
}
