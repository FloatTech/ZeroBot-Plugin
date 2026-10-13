package localadmin

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestStartupDifferencesSameCountsDetectSecretsAndUnknownChanges(t *testing.T) {
	a := []byte(connectionFixture)
	b := []byte(strings.ReplaceAll(strings.ReplaceAll(connectionFixture, "secret-A", "new-private-token"), "9007199254740993", "9007199254740994"))
	items := startupDifferences(a, b)
	if len(items) != 5 || !items[0].Changed || !items[3].Changed {
		t.Fatal("same-count changes were omitted")
	}
	for i, expected := range []int{2, 1, 1, 2, 2} {
		if *items[i].RunningCount != expected || *items[i].SavedCount != expected {
			t.Fatal("wrong category counts", i)
		}
		if items[i].Changed != (i == 0 || i == 3) {
			t.Fatal("wrong category change", i)
		}
	}
	out, _ := json.Marshal(items)
	for _, value := range []string{"secret", "new-private", "900719925", "future", "nickname", "bot.example", "extension", "hash"} {
		if bytes.Contains(out, []byte(value)) {
			t.Fatal("summary exposed a value or field", value)
		}
	}
}

func TestStartupDifferencesRepresentationAndEmptyConfigStayConsistent(t *testing.T) {
	for _, pair := range [][2]string{
		{`{"zero":{},"ws":null}`, `{"zero":{},"ws":[]}`},
		{`{"zero":{},"ws":[]}`, `{"ZERO":{},"WS":[]}`},
		{`{"zero":{"super_users":null}}`, `{"zero":{"super_users":[]}}`},
		{`{"zero":{},"other":9007199254740993}`, `{"zero":{},"other":9007199254740994}`},
		{`{"zero":{},"other":1}`, `{"other":1,"zero":{}}`},
		{`{"zero":{}}`, `{"zero":{}}`},
	} {
		items := startupDifferences([]byte(pair[0]), []byte(pair[1]))
		changed := false
		for _, item := range items {
			changed = changed || item.Changed
		}
		if changed == equivalentConfig([]byte(pair[0]), []byte(pair[1])) {
			t.Fatal("summary disagrees with restart decision", pair)
		}
		if len(items) < 5 {
			t.Fatal("empty categories missing")
		}
	}
	items := startupDifferences([]byte(`{"zero":{},"ws":[]}`), []byte(`{"ZERO":{},"WS":[]}`))
	last := items[len(items)-1]
	if last.Label != "字段名称或表示差异" || !last.Changed || last.RunningCount != nil || last.SavedCount != nil {
		t.Fatal("representation difference was assigned invented counts")
	}
}

func TestStartupDifferencesFollowSaveRestoreAndNewStartupWithoutMutating(t *testing.T) {
	s := fixtureConnectionStore(t)
	v := getSettings(t, s)
	for _, item := range v.StartupDifferences {
		if item.Changed {
			t.Fatal("initial state changed")
		}
	}
	saved, err := s.Save(ConnectionUpdate{Revision: v.Revision, AddConnection: &ConnectionAddition{URL: "ws://127.0.0.1:17005", Token: "new-token"}})
	if err != nil {
		t.Fatal(err)
	}
	item := saved.StartupDifferences[0]
	if !saved.RestartRequired || !item.Changed || *item.RunningCount != 2 || *item.SavedCount != 3 {
		t.Fatal("save summary incorrect")
	}
	raw, _ := os.ReadFile(s.path)
	newStartup, err := NewConnectionStore(s.path, s.backups, raw)
	if err != nil {
		t.Fatal(err)
	}
	if v := getSettings(t, newStartup); v.RestartRequired || v.StartupDifferences[0].Changed || *v.StartupDifferences[0].RunningCount != 3 {
		t.Fatal("new startup kept old snapshot")
	}
	if err := os.WriteFile(s.path, []byte(connectionFixture), 0600); err != nil {
		t.Fatal(err)
	}
	if v := getSettings(t, s); v.RestartRequired || v.StartupDifferences[0].Changed {
		t.Fatal("restored saved config did not become consistent")
	}
	before, _ := os.ReadFile(s.path)
	_ = getSettings(t, s)
	after, _ := os.ReadFile(s.path)
	if !bytes.Equal(before, after) || !bytes.Equal(s.running, []byte(connectionFixture)) {
		t.Fatal("read comparison mutated saved or running config")
	}
}

func TestStartupDifferencesAPIAuthAndSameCountSecretChange(t *testing.T) {
	store := fixtureConnectionStore(t)
	cfg := testConfig(t)
	cfg.Connections = store
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(s, "GET", "/api/connections", nil, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	v := getSettings(t, store)
	update := ConnectionUpdate{Revision: v.Revision, Connection: &ConnectionChange{Kind: "ws", Index: 0, TokenAction: "replace", Token: "different-token-private"}}
	w := request(s, "POST", "/api/connections", update, cookie, csrf)
	expectStatus(t, w, 200)
	for _, w := range []string{w.Body.String(), request(s, "GET", "/api/connections", nil, cookie, "").Body.String()} {
		var value ConnectionSettings
		if err := json.Unmarshal([]byte(w), &value); err != nil {
			t.Fatal(err)
		}
		if !value.RestartRequired || !value.StartupDifferences[0].Changed || *value.StartupDifferences[0].SavedCount != 2 || *value.StartupDifferences[0].RunningCount != 2 {
			t.Fatal("secret-only change lost in API")
		}
		if strings.Contains(w, "different-token-private") || strings.Contains(w, "secret-A") {
			t.Fatal("API exposed secret")
		}
	}
}
