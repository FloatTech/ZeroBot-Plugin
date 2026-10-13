package localadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func deleteIndex(v int) *int { return &v }
func TestConnectionRemovalPreservesFieldsBackupRuntimeAndRestores(t *testing.T) {
	for _, target := range []ConnectionRemoval{{Kind: "ws", Index: deleteIndex(0), Confirm: true}, {Kind: "ws", Index: deleteIndex(1), Confirm: true}, {Kind: "wss", Index: deleteIndex(0), Confirm: true}} {
		t.Run(target.Kind+string(rune('0'+*target.Index)), func(t *testing.T) {
			s := fixtureConnectionStore(t)
			v := getSettings(t, s)
			saved, err := s.Save(ConnectionUpdate{Revision: v.Revision, RemoveConnection: &target})
			if err != nil {
				t.Fatal(err)
			}
			if len(saved.Connections) != 2 || len(saved.RunningConnections) != 3 || !saved.RestartRequired || saved.BackupID == "" {
				t.Fatal("runtime changed or missing backup")
			}
			raw, _ := os.ReadFile(s.path)
			root, _ := decodeConfig(raw)
			old, _ := decodeConfig([]byte(connectionFixture))
			for key := range old {
				if key != target.Kind && !equivalentConfig(root[key], old[key]) {
					t.Fatal("unselected field changed", key)
				}
			}
			var a, b []json.RawMessage
			_ = json.Unmarshal(root[target.Kind], &a)
			_ = json.Unmarshal(old[target.Kind], &b)
			b = append(b[:*target.Index], b[*target.Index+1:]...)
			if len(a) != len(b) {
				t.Fatal("wrong item removed")
			}
			for i := range a {
				if !equivalentConfig(a[i], b[i]) {
					t.Fatal("remaining item changed")
				}
			}
			if !bytes.Contains(raw, []byte("9007199254740993")) {
				t.Fatal("precision lost")
			}
			out, _ := json.Marshal(saved)
			if bytes.Contains(out, []byte("secret-")) {
				t.Fatal("secret exposed")
			}
			restarted, err := NewConnectionStore(s.path, s.backups, raw)
			if err != nil {
				t.Fatal(err)
			}
			if view := getSettings(t, restarted); view.RestartRequired || len(view.RunningConnections) != 2 {
				t.Fatal("new startup snapshot wrong")
			}
			store := NewBackupStore(s.backups, s, nil, NewLogStore())
			preview, err := store.Preview(context.Background(), saved.BackupID, "remove-test")
			if err != nil {
				t.Fatal(err)
			}
			_, err = store.Restore(context.Background(), RestoreRequest{ID: saved.BackupID, Revision: preview.Revision, ExpiresAt: preview.ExpiresAt, Token: preview.Token, Confirm: true}, "remove-test")
			if err != nil {
				t.Fatal(err)
			}
			actual, _ := os.ReadFile(s.path)
			if !equivalentConfig(actual, []byte(connectionFixture)) || getSettings(t, s).RestartRequired {
				t.Fatal("deletion backup not restorable")
			}
		})
	}
}
func TestConnectionRemovalValidationLastConnectionAliasesAndMixedWrites(t *testing.T) {
	for _, target := range []ConnectionRemoval{{Kind: "ws", Index: deleteIndex(0)}, {Kind: "ws", Confirm: true}, {Kind: "../ws", Index: deleteIndex(0), Confirm: true}, {Kind: "WS", Index: deleteIndex(0), Confirm: true}, {Kind: "ws", Index: deleteIndex(-1), Confirm: true}, {Kind: "ws", Index: deleteIndex(2), Confirm: true}} {
		if _, err := applyConnectionUpdate([]byte(connectionFixture), ConnectionUpdate{RemoveConnection: &target}); err == nil {
			t.Fatal("invalid deletion accepted")
		}
	}
	target := &ConnectionRemoval{Kind: "ws", Index: deleteIndex(0), Confirm: true}
	users := []string{"12345678"}
	for _, update := range []ConnectionUpdate{{RemoveConnection: target, SuperUsers: &users}, {RemoveConnection: target, Connection: &ConnectionChange{Kind: "ws", Index: 0, TokenAction: "keep"}}, {RemoveConnection: target, AddConnection: &ConnectionAddition{URL: "ws://localhost:18000", WithoutToken: true}}} {
		if _, err := applyConnectionUpdate([]byte(connectionFixture), update); err == nil {
			t.Fatal("mixed deletion accepted")
		}
	}
	for _, raw := range []string{`{"zero":{},"ws":[{"URL":"ws://localhost:17000"}]}`, `{"zero":{},"ws":[{}],"wss":null}`} {
		if _, err := applyConnectionUpdate([]byte(raw), ConnectionUpdate{RemoveConnection: target}); err == nil {
			t.Fatal("last connection removed")
		}
	}
	target.Kind = "wss"
	if _, err := applyConnectionUpdate([]byte(`{"zero":{},"wss":[{}],"ws":null}`), ConnectionUpdate{RemoveConnection: target}); err == nil {
		t.Fatal("last reverse connection removed")
	}
	raw := []byte(`{"zero":{"super_users":[12345678]},"WS":[{}],"WSS":[{}],"private":9007199254740993}`)
	next, err := applyConnectionUpdate(raw, ConnectionUpdate{RemoveConnection: target})
	if err != nil {
		t.Fatal(err)
	}
	root, _ := decodeConfig(next)
	if !bytes.Equal(root["WSS"], []byte(`[]`)) || root["wss"] != nil || !bytes.Contains(root["private"], []byte("9007199254740993")) {
		t.Fatal("aliases or empty list changed")
	}
}
func TestConnectionRemovalFailureAndConcurrentConflictKeepCurrent(t *testing.T) {
	for _, kind := range []string{"backup", "replace", "invalid"} {
		t.Run(kind, func(t *testing.T) {
			s := fixtureConnectionStore(t)
			v := getSettings(t, s)
			target := &ConnectionRemoval{Kind: "ws", Index: deleteIndex(0), Confirm: true}
			switch kind {
			case "backup":
				_ = os.MkdirAll(filepath.Dir(s.backups), 0700)
				_ = os.WriteFile(s.backups, []byte("blocked"), 0600)
			case "replace":
				s.replace = func(string, string) error { return errors.New("private-failure") }
			case "invalid":
				target.Index = nil
			}
			_, err := s.Save(ConnectionUpdate{Revision: v.Revision, RemoveConnection: target})
			if err == nil || strings.Contains(err.Error(), "private-failure") {
				t.Fatal("failure accepted or exposed")
			}
			actual, _ := os.ReadFile(s.path)
			if !bytes.Equal(actual, []byte(connectionFixture)) {
				t.Fatal("failed removal changed current file")
			}
		})
	}
	s := fixtureConnectionStore(t)
	v := getSettings(t, s)
	var wg sync.WaitGroup
	results := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Save(ConnectionUpdate{Revision: v.Revision, RemoveConnection: &ConnectionRemoval{Kind: "ws", Index: deleteIndex(0), Confirm: true}})
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
	if success != 1 || len(getSettings(t, s).Connections) != 2 {
		t.Fatal("concurrent deletion accepted")
	}
}
func TestConnectionRemovalAPIRequiresExplicitTargetAuthAndConfirmation(t *testing.T) {
	store := fixtureConnectionStore(t)
	cfg := testConfig(t)
	cfg.Connections = store
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	body := ConnectionUpdate{Revision: getSettings(t, store).Revision, RemoveConnection: &ConnectionRemoval{Kind: "ws", Index: deleteIndex(0), Confirm: true}}
	expectStatus(t, request(s, "POST", "/api/connections", body, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/connections", body, cookie, "bad"), 403)
	for _, target := range []map[string]any{{"kind": "ws", "confirm": true}, {"kind": "ws", "index": nil, "confirm": true}, {"kind": "ws", "index": 0, "confirm": false}, {"kind": "ws", "index": 0, "confirm": true, "url": "private"}} {
		expectStatus(t, request(s, "POST", "/api/connections", map[string]any{"revision": body.Revision, "removeConnection": target}, cookie, csrf), 400)
	}
	w := request(s, "POST", "/api/connections", body, cookie, csrf)
	expectStatus(t, w, 200)
	if strings.Contains(w.Body.String(), "secret-") {
		t.Fatal("secret returned")
	}
	expectStatus(t, request(s, "POST", "/api/connections", body, cookie, csrf), 409)
	asset := request(s, "GET", "/connections-delete.js", nil, nil, "")
	expectStatus(t, asset, 200)
	if asset.Header().Get("Cache-Control") != "no-store" || !strings.Contains(asset.Body.String(), "deletionCurrent") {
		t.Fatal("embedded deletion script missing")
	}
}
