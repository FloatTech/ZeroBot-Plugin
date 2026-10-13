package localadmin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func impactNamed(t *testing.T, items []RestoreImpact, label string) RestoreImpact {
	t.Helper()
	for _, item := range items {
		if item.Label == label {
			return item
		}
	}
	t.Fatalf("missing impact: %s", label)
	return RestoreImpact{}
}
func checkImpactCounts(t *testing.T, item RestoreImpact, current, backup int) {
	t.Helper()
	if item.CurrentCount == nil || item.BackupCount == nil || *item.CurrentCount != current || *item.BackupCount != backup {
		t.Fatal("impact counts wrong")
	}
}
func TestStartupRestoreImpactsUseSameSnapshotHideValuesAndDoNotWrite(t *testing.T) {
	s, _ := backupFixture(t)
	id, err := s.Create("startup-config")
	if err != nil {
		t.Fatal(err)
	}
	root, _ := decodeConfig([]byte(connectionFixture))
	var connections []jsonObject
	_ = json.Unmarshal(root["ws"], &connections)
	connections[0]["AccessToken"] = json.RawMessage(`"new-private-key"`)
	root["ws"], _ = json.Marshal(connections)
	root["wss"] = json.RawMessage(`[]`)
	zero, _ := object(root["zero"])
	zero["super_users"] = json.RawMessage(`[87654321]`)
	zero["future"] = json.RawMessage(`{"n":9007199254740992}`)
	root["zero"], _ = json.Marshal(zero)
	root["private-name-key"] = json.RawMessage(`"private-name-value"`)
	current, _ := json.Marshal(root)
	if err = os.WriteFile(s.connections.path, current, 0600); err != nil {
		t.Fatal(err)
	}
	before, _, _ := s.List()
	v, err := s.Preview(context.Background(), id, "fixture-subject")
	if err != nil {
		t.Fatal(err)
	}
	if !v.Changed || len(v.Impacts) != 5 {
		t.Fatal("startup impact incomplete")
	}
	for _, item := range v.Impacts {
		if !item.Changed {
			t.Fatal("changed category marked equal")
		}
	}
	checkImpactCounts(t, impactNamed(t, v.Impacts, "正向 OneBot 连接"), 2, 2)
	checkImpactCounts(t, impactNamed(t, v.Impacts, "反向 OneBot 连接"), 0, 1)
	checkImpactCounts(t, impactNamed(t, v.Impacts, "机器人主人"), 1, 1)
	checkImpactCounts(t, impactNamed(t, v.Impacts, "其他机器人参数"), 2, 2)
	checkImpactCounts(t, impactNamed(t, v.Impacts, "其他启动设置"), 3, 2)
	raw, _ := json.Marshal(v)
	for _, secret := range []string{"secret-A", "secret-B", "secret-C", "new-private-key", "87654321", "private-name-key", "private-name-value", "9007199254740992"} {
		if bytes.Contains(raw, []byte(secret)) {
			t.Fatal("impact preview leaked a value or unknown field name")
		}
	}
	after, _, _ := s.List()
	actual, _ := os.ReadFile(s.connections.path)
	if len(after) != len(before) || !bytes.Equal(actual, current) || !bytes.Equal(s.connections.running, []byte(connectionFixture)) {
		t.Fatal("preview wrote config, backup or running state")
	}
	request := RestoreRequest{ID: id, Revision: v.Revision, ExpiresAt: v.ExpiresAt, Token: v.Token, Confirm: true}
	if err = os.WriteFile(s.connections.path, []byte(connectionFixture), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Restore(context.Background(), request, "fixture-subject"); !errors.Is(err, errBackupConflict) {
		t.Fatal("impact preview bypassed stale snapshot protection")
	}
}
func TestFileRestoreImpactsSemanticEqualityAliasesMissingAndLargeNumbers(t *testing.T) {
	cases := []struct {
		a, b              string
		changed, fallback bool
	}{
		{`{"zero":{},"ws":[],"wss":[]}`, "{\n \"wss\": [], \"ws\": [], \"zero\": {} }", false, false},
		{`{"zero":{}}`, `{"ZERO":{}}`, true, true},
		{`{"zero":{}}`, `{"zero":{"super_users":null}}`, true, true},
		{`{"zero":{},"ws":[]}`, `{"zero":{}}`, true, false},
		{`{"zero":{},"unknown":9007199254740992}`, `{"zero":{},"unknown":9007199254740993}`, true, false},
		{`{"zero":{"super_users":[1,2]}}`, `{"zero":{"super_users":[2,1]}}`, true, false},
	}
	for _, tc := range cases {
		items, err := fileRestoreImpacts("startup-config", []byte(tc.a), []byte(tc.b))
		if err != nil {
			t.Fatal(err)
		}
		changed, fallback := false, false
		for _, item := range items {
			changed = changed || item.Changed
			fallback = fallback || item.Label == "字段名称或表示差异"
		}
		if changed != tc.changed || fallback != tc.fallback {
			t.Fatal("semantic or representation comparison wrong")
		}
	}
	if _, err := fileRestoreImpacts("startup-config", []byte(`{"zero":{},"ws":[],"WS":[]}`), []byte(connectionFixture)); err == nil {
		t.Fatal("ambiguous aliases accepted")
	}
	if _, err := fileRestoreImpacts("unknown", nil, nil); err == nil {
		t.Fatal("unsupported scope accepted")
	}
	items, err := fileRestoreImpacts("admin-options", []byte(`{"enabled":true,"address":"127.0.0.1:3010"}`), []byte(`{"address":"127.0.0.1:3111","enabled":false}`))
	if err != nil || len(items) != 2 {
		t.Fatal("options impact wrong", err)
	}
	for _, item := range items {
		if !item.Changed || item.CurrentCount != nil || item.BackupCount != nil {
			t.Fatal("options exposed value/count")
		}
	}
}
func TestControlRestoreImpactsAggregateRowsAndHideTableNames(t *testing.T) {
	s, c := backupFixture(t)
	ctx := context.Background()
	if err := c.SetData(101, 123); err != nil {
		t.Fatal(err)
	}
	exec := func(statements ...string) {
		t.Helper()
		c.Manager.Lock()
		defer c.Manager.Unlock()
		if err := c.Manager.D.TransactionContext(ctx, func(tx *sql.Tx) error {
			for _, statement := range statements {
				if _, err := tx.ExecContext(ctx, statement); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := c.SetSwitch(101, "enabled", ""); err != nil {
		t.Fatal(err)
	}
	exec(`CREATE TABLE "private-secret-table" (id INTEGER PRIMARY KEY, value TEXT)`, `INSERT INTO "private-secret-table" VALUES(1,'old-private-value')`)
	id, err := s.Create("control-db")
	if err != nil {
		t.Fatal(err)
	}
	if err = c.SetData(101, 987); err != nil {
		t.Fatal(err)
	}
	if err = c.Manager.Response(101); err != nil {
		t.Fatal(err)
	}
	if err = c.Manager.DoBlock(87654321); err != nil {
		t.Fatal(err)
	}
	state := switchState(t, c, 101)
	if _, err = c.SetSwitch(101, "inherit", state.Revision); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE "private-secret-table" SET value='new-private-value'`, `UPDATE __resp SET ext='{"Key":"private-model-key","SystemP":"private-persona"}' WHERE gid=101`)
	v, err := s.Preview(ctx, id, "fixture-subject")
	if err != nil {
		t.Fatal(err)
	}
	if !v.Changed || len(v.Impacts) != 4 {
		t.Fatal("control impacts incomplete")
	}
	sum := 0
	for _, item := range v.Impacts {
		if !item.Changed {
			t.Fatal("control category missed changes", item.Label)
		}
		sum += *item.BackupCount
	}
	if sum != v.Rows {
		t.Fatal("aggregated backup count lost records")
	}
	checkImpactCounts(t, impactNamed(t, v.Impacts, "响应规则与AI扩展"), 1, 0)
	checkImpactCounts(t, impactNamed(t, v.Impacts, "全局用户封禁"), 1, 0)
	checkImpactCounts(t, impactNamed(t, v.Impacts, "插件开关继承标记"), 1, 0)
	checkImpactCounts(t, impactNamed(t, v.Impacts, "插件与其他库内数据"), 2, 2)
	raw, _ := json.Marshal(v)
	for _, secret := range []string{"private-secret-table", "old-private-value", "new-private-value", "private-model-key", "private-persona", "87654321", "__resp", "test-plugin"} {
		if strings.Contains(string(raw), secret) {
			t.Fatal("control impact exposed private record or table name")
		}
	}
	sameID, err := s.Create("control-db")
	if err != nil {
		t.Fatal(err)
	}
	same, err := s.Preview(ctx, sameID, "fixture-subject")
	if err != nil || same.Changed {
		t.Fatal("equal database changed", err)
	}
	for _, item := range same.Impacts {
		if item.Changed || *item.CurrentCount != *item.BackupCount {
			t.Fatal("equal database impact not equal")
		}
	}
}
func TestRestoreImpactAPIRequiresAuthAndKeepsReadonlyOptions(t *testing.T) {
	store, _ := backupFixture(t)
	id, err := store.Create("admin-options")
	if err != nil {
		t.Fatal(err)
	}
	current := []byte(`{"enabled":false,"address":"127.0.0.1:3111"}`)
	if err = os.WriteFile(store.optionsPath, current, 0600); err != nil {
		t.Fatal(err)
	}
	s, err := New(Config{Addr: "127.0.0.1:3010", DataDir: t.TempDir(), Overview: func() Overview { return Overview{} }, Backups: store})
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]string{"id": id}
	expectStatus(t, request(s, "POST", "/api/backups/preview", body, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/backups/preview", body, cookie, "wrong-csrf"), 403)
	w := request(s, "POST", "/api/backups/preview", body, cookie, csrf)
	expectStatus(t, w, 200)
	var v RestorePreview
	if err = json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Impacts) != 2 || !v.Changed || !v.RestartRequired {
		t.Fatal("API omitted options impacts")
	}
	if strings.Contains(w.Body.String(), "127.0.0.1:3111") {
		t.Fatal("API exposed options value")
	}
	actual, _ := os.ReadFile(store.optionsPath)
	if !bytes.Equal(actual, current) {
		t.Fatal("preview changed current options")
	}
	expectStatus(t, request(s, "POST", "/api/backups/preview", map[string]string{"id": "../credential.json"}, cookie, csrf), 400)
}
