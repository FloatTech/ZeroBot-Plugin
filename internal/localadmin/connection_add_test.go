package localadmin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestConnectionAdditionBackupPreservesOtherFieldsAndRuntime(t *testing.T) {
	s := fixtureConnectionStore(t)
	initial := getSettings(t, s)
	saved, err := s.Save(ConnectionUpdate{Revision: initial.Revision, AddConnection: &ConnectionAddition{URL: "ws://127.0.0.1:7702/ws", Token: "new-add-secret"}})
	if err != nil {
		t.Fatal(err)
	}
	if saved.BackupID == "" || !saved.RestartRequired || len(saved.Connections) != 4 || len(saved.RunningConnections) != 3 {
		t.Fatal("addition changed running state or omitted backup")
	}
	if saved.Connections[2].Kind != "ws" || saved.Connections[2].Index != 2 || !saved.Connections[2].TokenConfigured {
		t.Fatal("addition target incorrect")
	}
	raw, _ := os.ReadFile(s.path)
	root, _ := decodeConfig(raw)
	old, _ := decodeConfig([]byte(connectionFixture))
	for _, key := range []string{"zero", "wss", "force_base64_file", "unknown"} {
		if !equivalentConfig(root[key], old[key]) {
			t.Fatal("lost existing field", key)
		}
	}
	var a, b []json.RawMessage
	_ = json.Unmarshal(root["ws"], &a)
	_ = json.Unmarshal(old["ws"], &b)
	for i := range b {
		if !bytes.Equal(a[i], b[i]) && !equivalentConfig(a[i], b[i]) {
			t.Fatal("changed old connection")
		}
	}
	if !bytes.Contains(raw, []byte("9007199254740993")) {
		t.Fatal("integer precision lost")
	}
	encoded, _ := json.Marshal(saved)
	if bytes.Contains(encoded, []byte("secret")) {
		t.Fatal("token exposed")
	}
	backup, _ := os.ReadFile(filepath.Join(s.backups, saved.BackupID))
	var snapshot fileBackup
	if json.Unmarshal(backup, &snapshot) != nil || !equivalentConfig(snapshot.Config, []byte(connectionFixture)) {
		t.Fatal("wrong backup")
	}
	next, err := NewConnectionStore(s.path, s.backups, raw)
	if err != nil {
		t.Fatal(err)
	}
	if view := getSettings(t, next); view.RestartRequired || len(view.RunningConnections) != 4 {
		t.Fatal("restart snapshot did not load addition")
	}
}
func TestConnectionAdditionValidationAliasesMissingAndLimit(t *testing.T) {
	users := []string{"12345678"}
	for _, add := range []ConnectionAddition{
		{URL: "http://127.0.0.1:7777", Token: "valid"}, {URL: "ws://127.0.0.1:0", Token: "valid"}, {URL: "ws://user:secret@localhost/ws", Token: "valid"}, {URL: "ws://localhost/ws?key=secret", Token: "valid"},
		{URL: "ws://localhost/ws", Token: ""}, {URL: "ws://localhost/ws", Token: "***"}, {URL: "ws://localhost/ws", Token: "line\nbreak"}, {URL: "ws://localhost/ws", Token: strings.Repeat("中", 700)}, {URL: "ws://localhost/ws", Token: "valid", WithoutToken: true}, {URL: "ws://127.0.0.1:6700/", Token: "valid"}, {URL: "wss://BOT.EXAMPLE:0443/ws", Token: "valid"},
	} {
		if _, err := applyConnectionUpdate([]byte(connectionFixture), ConnectionUpdate{AddConnection: &add}); err == nil {
			t.Fatal("invalid addition accepted", add.URL)
		}
	}
	add := &ConnectionAddition{URL: "ws://localhost:7777", WithoutToken: true}
	for _, update := range []ConnectionUpdate{{AddConnection: add, SuperUsers: &users}, {AddConnection: add, Connection: &ConnectionChange{Kind: "ws", Index: 0, TokenAction: "keep"}}} {
		if _, err := applyConnectionUpdate([]byte(connectionFixture), update); err == nil {
			t.Fatal("mixed write accepted")
		}
	}
	for _, raw := range []string{`{"zero":{}}`, `{"zero":{},"ws":null}`, `{"zero":{},"WS":[],"unknown":9007199254740993}`} {
		next, err := applyConnectionUpdate([]byte(raw), ConnectionUpdate{AddConnection: add})
		if err != nil {
			t.Fatal(err)
		}
		root, _ := decodeConfig(next)
		key, list, _ := field(root, "ws")
		var conns []ConnectionAddition
		_ = json.Unmarshal(list, &conns)
		if len(conns) != 1 || conns[0].Token != "" || strings.Contains(raw, "WS") && key != "WS" {
			t.Fatal("missing/null/alias addition incorrect")
		}
	}
	root, _ := decodeConfig([]byte(connectionFixture))
	var list []map[string]string
	for i := 0; i < 64; i++ {
		list = append(list, map[string]string{"URL": fmt.Sprintf("ws://127.0.0.1:%d", 10000+i)})
	}
	root["ws"], _ = json.Marshal(list)
	raw, _ := json.Marshal(root)
	if _, err := applyConnectionUpdate(raw, ConnectionUpdate{AddConnection: add}); err == nil {
		t.Fatal("64 limit bypassed")
	}
	// The new limit does not block editing an existing large configuration.
	if _, err := applyConnectionUpdate(raw, ConnectionUpdate{Connection: &ConnectionChange{Kind: "ws", Index: 63, TokenAction: "keep"}}); err != nil {
		t.Fatal("old list cannot be edited", err)
	}
	for _, pair := range [][2]string{{"ws://LOCALHOST/", "ws://localhost:80"}, {"wss://localhost:0443/ws", "wss://LOCALHOST/ws"}, {"ws://[::1]/", "ws://[::1]:80"}} {
		if connectionEndpoint(pair[0]) != connectionEndpoint(pair[1]) {
			t.Fatal("equivalent address missed")
		}
	}
	if connectionEndpoint("ws://localhost/ws") == connectionEndpoint("wss://localhost/ws") || connectionEndpoint("ws://localhost/ws/") == connectionEndpoint("ws://localhost/ws") {
		t.Fatal("different endpoint collapsed")
	}
}
func TestConnectionAdditionFailuresAndConcurrentConflictKeepCurrent(t *testing.T) {
	for _, failure := range []string{"backup", "replace", "invalid"} {
		t.Run(failure, func(t *testing.T) {
			s := fixtureConnectionStore(t)
			view := getSettings(t, s)
			add := ConnectionAddition{URL: "ws://localhost:7777", Token: "new-private-secret"}
			switch failure {
			case "backup":
				_ = os.MkdirAll(filepath.Dir(s.backups), 0700)
				_ = os.WriteFile(s.backups, []byte("blocked"), 0600)
			case "replace":
				s.replace = func(string, string) error { return errors.New("injected-private-error") }
			case "invalid":
				add.URL = "http://localhost"
			}
			_, err := s.Save(ConnectionUpdate{Revision: view.Revision, AddConnection: &add})
			if err == nil || strings.Contains(err.Error(), "private") {
				t.Fatal("failure accepted or leaked")
			}
			actual, _ := os.ReadFile(s.path)
			if !bytes.Equal(actual, []byte(connectionFixture)) {
				t.Fatal("failed add overwrote config")
			}
		})
	}
	s := fixtureConnectionStore(t)
	view := getSettings(t, s)
	var wg sync.WaitGroup
	results := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := s.Save(ConnectionUpdate{Revision: view.Revision, AddConnection: &ConnectionAddition{URL: fmt.Sprintf("ws://localhost:%d", 17000+i), WithoutToken: true}})
			results <- err
		}(i)
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
	if success != 1 || len(getSettings(t, s).Connections) != 4 {
		t.Fatal("concurrent adds bypassed version")
	}
}
func TestConnectionAdditionAPIAuthCSRFRedactionAndUnknownFields(t *testing.T) {
	store := fixtureConnectionStore(t)
	cfg := testConfig(t)
	cfg.Connections = store
	cfg.Logs = NewLogStore()
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	update := ConnectionUpdate{Revision: getSettings(t, store).Revision, AddConnection: &ConnectionAddition{URL: "ws://localhost:17700", Token: "api-new-private-secret"}}
	expectStatus(t, request(s, "POST", "/api/connections", update, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/connections", update, cookie, "bad"), 403)
	unknown := map[string]any{"revision": update.Revision, "addConnection": map[string]any{"url": "ws://localhost:17700", "token": "api-new-private-secret", "future": "unknown"}}
	expectStatus(t, request(s, "POST", "/api/connections", unknown, cookie, csrf), 400)
	w := request(s, "POST", "/api/connections", update, cookie, csrf)
	expectStatus(t, w, 200)
	if strings.Contains(w.Body.String(), "private-secret") {
		t.Fatal("API exposed token")
	}
	expectStatus(t, request(s, "POST", "/api/connections", update, cookie, csrf), 409)
	w = request(s, "GET", "/api/connections", nil, cookie, "")
	expectStatus(t, w, 200)
	if strings.Contains(w.Body.String(), "private-secret") {
		t.Fatal("read exposed token")
	}
	asset := request(s, "GET", "/connections-add.js", nil, nil, "")
	expectStatus(t, asset, 200)
	if asset.Header().Get("Cache-Control") != "no-store" || !strings.Contains(asset.Body.String(), "connectionAddSnapshot") {
		t.Fatal("addition asset absent")
	}
}
