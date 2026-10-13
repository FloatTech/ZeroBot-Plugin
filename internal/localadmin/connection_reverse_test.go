package localadmin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestReverseAdditionPreservesForwardOtherFieldsAndRuntime(t *testing.T) {
	s := fixtureConnectionStore(t)
	v := getSettings(t, s)
	saved, err := s.Save(ConnectionUpdate{Revision: v.Revision, AddConnection: &ConnectionAddition{Kind: "wss", URL: "ws://127.0.0.1:17002/", Token: "reverse-new-secret"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Connections) != 4 || saved.Connections[3].Kind != "wss" || saved.Connections[3].Index != 1 || !saved.Connections[3].TokenConfigured || len(saved.RunningConnections) != 3 || saved.BackupID == "" || !saved.RestartRequired {
		t.Fatal("reverse target/runtime/backup wrong")
	}
	raw, _ := os.ReadFile(s.path)
	root, _ := decodeConfig(raw)
	old, _ := decodeConfig([]byte(connectionFixture))
	for k := range old {
		if k != "wss" && !equivalentConfig(root[k], old[k]) {
			t.Fatal("other field changed", k)
		}
	}
	if !bytes.Contains(raw, []byte("9007199254740993")) {
		t.Fatal("precision lost")
	}
	out, _ := json.Marshal(saved)
	if bytes.Contains(out, []byte("secret")) {
		t.Fatal("token returned")
	}
	reopened, err := NewConnectionStore(s.path, s.backups, raw)
	if err != nil {
		t.Fatal(err)
	}
	if v := getSettings(t, reopened); v.RestartRequired || len(v.RunningConnections) != 4 {
		t.Fatal("startup snapshot wrong")
	}
}
func TestReverseAdditionValidatesDriverAddressPortsTokenAndLimits(t *testing.T) {
	for _, address := range []string{"wss://127.0.0.1:17002", "ws://localhost:17002", "ws://0.0.0.0:17002", "ws://[::1]:17002", "ws://bot.example:17002", "ws://127.0.0.1", "ws://127.0.0.1:017002", "ws://127.0.0.1:0", "ws://127.0.0.1:65536", "ws://127.0.0.1:17002/custom", "ws://127.0.0.1:17002?key=private", "ws://user:private@127.0.0.1:17002", "ws://127.0.0.1:6700", "ws://127.0.0.1:6701"} {
		_, err := applyConnectionUpdate([]byte(connectionFixture), ConnectionUpdate{AddConnection: &ConnectionAddition{Kind: "wss", URL: address, Token: "valid"}})
		if err == nil {
			t.Fatal("bad reverse address accepted", address)
		}
	}
	for _, add := range []ConnectionAddition{{Kind: "unknown", URL: "ws://127.0.0.1:17002", Token: "valid"}, {Kind: "wss", URL: "ws://127.0.0.1:17002", WithoutToken: true}, {Kind: "wss", URL: "ws://127.0.0.1:17002"}, {Kind: "wss", URL: "ws://127.0.0.1:17002", Token: "***"}, {Kind: "wss", URL: "ws://127.0.0.1:17002", Token: "line\nbreak"}, {Kind: "wss", URL: "ws://127.0.0.1:17002", Token: strings.Repeat("中", 700)}} {
		if _, err := applyConnectionUpdate([]byte(connectionFixture), ConnectionUpdate{AddConnection: &add}); err == nil {
			t.Fatal("bad direction/token accepted")
		}
	}
	for _, raw := range []string{`{"zero":{},"WSS":null}`, `{"zero":{}}`} {
		next, err := applyConnectionUpdate([]byte(raw), ConnectionUpdate{AddConnection: &ConnectionAddition{Kind: "wss", URL: "ws://127.0.0.1:17002", Token: "valid"}})
		if err != nil {
			t.Fatal(err)
		}
		root, _ := decodeConfig(next)
		k, items, _ := field(root, "wss")
		if strings.Contains(raw, "WSS") && k != "WSS" || !bytes.Contains(items, []byte("17002")) {
			t.Fatal("alias or missing list incorrect")
		}
	}
	for _, address := range []string{"ws://0.0.0.0:17002/path", "tcp://127.0.0.1:17002", "127.0.0.1:17002"} {
		root, _ := decodeConfig([]byte(connectionFixture))
		root["wss"], _ = json.Marshal([]map[string]string{{"URL": address}})
		raw, _ := json.Marshal(root)
		if _, err := applyConnectionUpdate(raw, ConnectionUpdate{AddConnection: &ConnectionAddition{Kind: "wss", URL: "ws://127.0.0.1:17002", Token: "valid"}}); err == nil {
			t.Fatal("listener port conflict ignored")
		}
	}
	root, _ := decodeConfig([]byte(connectionFixture))
	var list []map[string]string
	for i := 0; i < 64; i++ {
		list = append(list, map[string]string{"URL": fmt.Sprintf("ws://127.0.0.1:%d", 18000+i)})
	}
	root["wss"], _ = json.Marshal(list)
	raw, _ := json.Marshal(root)
	if _, err := applyConnectionUpdate(raw, ConnectionUpdate{AddConnection: &ConnectionAddition{Kind: "wss", URL: "ws://127.0.0.1:17002", Token: "valid"}}); err == nil {
		t.Fatal("reverse limit bypassed")
	}
}
func TestReverseAdditionReplacementFailureAndConcurrentVersion(t *testing.T) {
	s := fixtureConnectionStore(t)
	v := getSettings(t, s)
	s.replace = func(string, string) error { return errors.New("secret-private-failure") }
	_, err := s.Save(ConnectionUpdate{Revision: v.Revision, AddConnection: &ConnectionAddition{Kind: "wss", URL: "ws://127.0.0.1:17002", Token: "valid"}})
	if err == nil || strings.Contains(err.Error(), "secret-private") {
		t.Fatal("failure accepted or exposed")
	}
	raw, _ := os.ReadFile(s.path)
	if !bytes.Equal(raw, []byte(connectionFixture)) {
		t.Fatal("failure changed config")
	}
	s = fixtureConnectionStore(t)
	v = getSettings(t, s)
	var wg sync.WaitGroup
	results := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := s.Save(ConnectionUpdate{Revision: v.Revision, AddConnection: &ConnectionAddition{Kind: "wss", URL: fmt.Sprintf("ws://127.0.0.1:%d", 17002+i), Token: "valid"}})
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
	if success != 1 {
		t.Fatal("reverse version not serialized")
	}
}
func TestReverseAdditionAPIAuthAndAdminPortConflict(t *testing.T) {
	store := fixtureConnectionStore(t)
	cfg := testConfig(t)
	cfg.Connections = store
	cfg.Logs = NewLogStore()
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	v := getSettings(t, store)
	body := ConnectionUpdate{Revision: v.Revision, AddConnection: &ConnectionAddition{Kind: "wss", URL: "ws://127.0.0.1:3010", Token: "reverse-api-secret"}}
	expectStatus(t, request(s, "POST", "/api/connections", body, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/connections", body, cookie, "bad"), 403)
	w := request(s, "POST", "/api/connections", body, cookie, csrf)
	expectStatus(t, w, 400)
	if !strings.Contains(w.Body.String(), "后台端口冲突") || getSettings(t, store).Revision != v.Revision {
		t.Fatal("backend collision changed config")
	}
	body.AddConnection.URL = "ws://127.0.0.1:17002"
	w = request(s, "POST", "/api/connections", body, cookie, csrf)
	expectStatus(t, w, 200)
	if strings.Contains(w.Body.String(), "reverse-api-secret") {
		t.Fatal("API leaked token")
	}
	expectStatus(t, request(s, "POST", "/api/connections", body, cookie, csrf), 409)
}
