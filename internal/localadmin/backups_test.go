package localadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func backupFixture(t *testing.T) (*BackupStore, *ctrl.Control[*zero.Ctx]) {
	t.Helper()
	connections := fixtureConnectionStore(t)
	c, _ := pluginFixture(t, false)
	plugins := NewPluginStore([]*ctrl.Control[*zero.Ctx]{c}, connections.backups)
	s := NewBackupStore(connections.backups, connections, plugins, NewLogStore())
	if err := os.MkdirAll(filepath.Dir(s.optionsPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.optionsPath, []byte(`{"enabled":true,"address":"127.0.0.1:3010"}`), 0600); err != nil {
		t.Fatal(err)
	}
	return s, c
}
func previewRestore(t *testing.T, s *BackupStore, id string) RestoreRequest {
	t.Helper()
	v, err := s.Preview(context.Background(), id, "fixture-session")
	if err != nil {
		t.Fatal(err)
	}
	return RestoreRequest{ID: id, Revision: v.Revision, ExpiresAt: v.ExpiresAt, Token: v.Token, Confirm: true}
}

func TestBackupFileScopesRestoreAndConflict(t *testing.T) {
	s, _ := backupFixture(t)
	for _, scope := range []string{"startup-config", "admin-options"} {
		id, err := s.Create(scope)
		if err != nil {
			t.Fatal(err)
		}
		path := s.optionsPath
		if scope == "startup-config" {
			path = s.connections.path
		}
		old, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var next []byte
		if scope == "startup-config" {
			next = []byte(strings.ReplaceAll(string(old), "secret-A", "new-synthetic-key"))
		} else {
			next = []byte(`{"enabled":false,"address":"127.0.0.1:3111"}`)
		}
		if err = os.WriteFile(path, next, 0600); err != nil {
			t.Fatal(err)
		}
		preview := previewRestore(t, s, id)
		refused := preview
		refused.Confirm = false
		if _, err = s.Restore(context.Background(), refused, "fixture-session"); !errors.Is(err, errBackupInvalid) {
			t.Fatal("missing confirmation accepted", err)
		}
		if _, err = s.Restore(context.Background(), preview, "other-session"); !errors.Is(err, errBackupConflict) {
			t.Fatal("ticket not tied to session", err)
		}
		// Failure leaves current bytes and running configuration untouched.
		failure := func(string, string) error { return errors.New("injected replacement failure") }
		if scope == "startup-config" {
			s.connections.replace = failure
		} else {
			s.replace = failure
		}
		if _, err = s.Restore(context.Background(), preview, "fixture-session"); err == nil {
			t.Fatal("write failure accepted")
		}
		actual, _ := os.ReadFile(path)
		if !bytes.Equal(actual, next) {
			t.Fatal("failed restore modified original")
		}
		if scope == "startup-config" {
			s.connections.replace = replaceConfigFile
		} else {
			s.replace = replaceConfigFile
		}
		result, err := s.Restore(context.Background(), preview, "fixture-session")
		if err != nil || !result.Changed || !result.RestartRequired || result.BackupID == "" {
			t.Fatal(result, err)
		}
		actual, _ = os.ReadFile(path)
		if !equivalentConfig(actual, old) {
			t.Fatal("restore lost unknown configuration")
		}
		_, saved, err := s.load(result.BackupID)
		if err != nil || !equivalentConfig(saved, next) {
			t.Fatal("pre-restore backup wrong", err)
		}
		if _, err = s.Restore(context.Background(), preview, "fixture-session"); !errors.Is(err, errBackupConflict) {
			t.Fatal("stale ticket accepted", err)
		}
		unchanged := previewRestore(t, s, id)
		result, err = s.Restore(context.Background(), unchanged, "fixture-session")
		if err != nil || result.Changed || result.BackupID != "" {
			t.Fatal("no-op created backup or failed", result, err)
		}
	}
	if !bytes.Equal(s.connections.running, []byte(connectionFixture)) {
		t.Fatal("running configuration changed")
	}
}

func TestBackupControlRestoreAtomicityAndCaches(t *testing.T) {
	s, c := backupFixture(t)
	if err := c.SetData(101, 123456); err != nil {
		t.Fatal(err)
	}
	if err := c.Reset(101); err != nil {
		t.Fatal(err)
	}
	if err := c.SetExtra(map[string]string{"Key": "old-synthetic-model-key"}); err == nil {
		t.Fatal("fixture must not register extension")
	}
	id, err := s.Create("control-db")
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Disable(101); err != nil {
		t.Fatal(err)
	}
	if err = c.SetData(101, 987654); err != nil {
		t.Fatal(err)
	}
	if err = c.Flip(); err != nil {
		t.Fatal(err)
	}
	if err = c.Manager.Response(101); err != nil {
		t.Fatal(err)
	}
	if err = c.Manager.DoBlock(456); err != nil {
		t.Fatal(err)
	}
	c.Ban(789, 101)
	// Prime every control cache before restore.
	_ = c.IsEnabledIn(101)
	_ = c.Manager.CanResponse(101)
	_ = c.Manager.IsBlocked(456)
	_ = c.IsBannedIn(789, 101)
	before := switchState(t, c, 101)
	preview := previewRestore(t, s, id)
	s.afterTable = func(i int) error {
		if i == 1 {
			return errors.New("injected mid-restore failure")
		}
		return nil
	}
	if _, err = s.Restore(context.Background(), preview, "fixture-session"); err == nil {
		t.Fatal("mid-transaction failure accepted")
	}
	if c.GetData(101) != 987654 || switchState(t, c, 101) != before || !c.Manager.CanResponse(101) || !c.Manager.IsBlocked(456) || !c.IsBannedIn(789, 101) {
		t.Fatal("failed restore altered data/cache")
	}
	s.afterTable = nil
	result, err := s.Restore(context.Background(), preview, "fixture-session")
	if err != nil || !result.Changed || result.BackupID == "" {
		t.Fatal(result, err)
	}
	state := switchState(t, c, 101)
	if c.GetData(101) != 123456 || state.GroupMode != "inherit" || !state.DefaultEnabled || !c.IsEnabledIn(101) || c.Manager.CanResponse(101) || c.Manager.IsBlocked(456) || c.IsBannedIn(789, 101) {
		t.Fatal("restore/cache invalidation wrong", state)
	}
	// Recover the state before restoration using its mandatory backup.
	recovery := previewRestore(t, s, result.BackupID)
	if _, err = s.Restore(context.Background(), recovery, "fixture-session"); err != nil {
		t.Fatal(err)
	}
	if c.GetData(101) != 987654 || switchState(t, c, 101) != before || !c.Manager.CanResponse(101) || !c.Manager.IsBlocked(456) || !c.IsBannedIn(789, 101) {
		t.Fatalf("recovery did not restore pre-restore state: data=%d state=%+v want=%+v response=%t block=%t ban=%t", c.GetData(101), switchState(t, c, 101), before, c.Manager.CanResponse(101), c.Manager.IsBlocked(456), c.IsBannedIn(789, 101))
	}
}

func TestBackupRejectsDamagePathsSchemaExpiredAndChangedSource(t *testing.T) {
	s, c := backupFixture(t)
	id, err := s.Create("startup-config")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"../config.json", id + "/../credential.json", filepath.Join(s.directory, id), "credential.json"} {
		if _, err = s.Preview(context.Background(), bad, "fixture-session"); !errors.Is(err, errBackupInvalid) {
			t.Fatal("path allowed", bad, err)
		}
	}
	preview := previewRestore(t, s, id)
	s.now = func() time.Time { return time.Now().Add(11 * time.Minute) }
	if _, err = s.Restore(context.Background(), preview, "fixture-session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("expired accepted", err)
	}
	s.now = time.Now
	path := filepath.Join(s.directory, id)
	raw, _ := os.ReadFile(path)
	var file fileBackup
	_ = json.Unmarshal(raw, &file)
	file.Config = json.RawMessage(strings.ReplaceAll(string(file.Config), "secret-A", "changed-key"))
	modified, _ := json.Marshal(file)
	_ = os.WriteFile(path, modified, 0600)
	if _, err = s.Preview(context.Background(), id, "fixture-session"); !errors.Is(err, errBackupInvalid) {
		t.Fatal("corrupt digest accepted", err)
	}
	file.SHA256 = digestBytes(file.Config)
	modified, _ = json.Marshal(file)
	_ = os.WriteFile(path, modified, 0600)
	if _, err = s.Restore(context.Background(), preview, "fixture-session"); !errors.Is(err, errBackupConflict) {
		t.Fatal("changed source accepted", err)
	}
	file.Version = 2
	modified, _ = json.Marshal(file)
	_ = os.WriteFile(path, modified, 0600)
	if _, err = s.Preview(context.Background(), id, "fixture-session"); !errors.Is(err, errBackupInvalid) {
		t.Fatal("unsupported version accepted", err)
	}
	dbID, err := s.Create("control-db")
	if err != nil {
		t.Fatal(err)
	}
	c.Manager.Lock()
	_, err = c.Manager.D.Exec("CREATE TABLE future_schema (id INTEGER PRIMARY KEY)")
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Preview(context.Background(), dbID, "fixture-session"); !errors.Is(err, errBackupInvalid) {
		t.Fatal("incompatible schema accepted", err)
	}
	rows, _, err := s.List()
	if err != nil || len(rows) < 2 {
		t.Fatal(rows, err)
	}
	found := false
	for _, entry := range rows {
		if entry.ID == id {
			found = true
			if entry.Valid {
				t.Fatal("damaged backup marked valid")
			}
		}
	}
	if !found {
		t.Fatal("damaged row disappeared")
	}
}

func TestBackupAPIAuthPreviewAndPrivateResponses(t *testing.T) {
	store, c := backupFixture(t)
	s, err := New(Config{Addr: "127.0.0.1:3010", DataDir: filepath.Dir(store.directory), Overview: func() Overview { return Overview{} }, Backups: store})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/backups", "/api/backups/preview", "/api/backups/restore"} {
		method := "POST"
		if path == "/api/backups" {
			method = "GET"
		}
		expectStatus(t, request(s, method, path, map[string]string{}, nil, ""), 401)
	}
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/backups", map[string]string{"scope": "startup-config"}, cookie, "invalid"), 403)
	w := request(s, "POST", "/api/backups", map[string]string{"scope": "control-db"}, cookie, csrf)
	expectStatus(t, w, 200)
	var created struct {
		ID string `json:"backupId"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if err = c.Disable(101); err != nil {
		t.Fatal(err)
	}
	w = request(s, "POST", "/api/backups/preview", map[string]string{"id": created.ID}, cookie, csrf)
	expectStatus(t, w, 200)
	var preview RestorePreview
	_ = json.Unmarshal(w.Body.Bytes(), &preview)
	if !preview.Changed || preview.Revision == "" || preview.Token == "" {
		t.Fatal("preview incomplete")
	}
	for _, path := range []string{"/api/backups", "/backups/" + created.ID, "/api/backups/" + created.ID + "/download"} {
		w = request(s, "GET", path, nil, cookie, "")
		if path == "/api/backups" {
			expectStatus(t, w, 200)
		} else {
			expectStatus(t, w, 404)
		}
		if strings.Contains(w.Body.String(), "secret-A") || strings.Contains(w.Body.String(), "plugins.sqlite") {
			t.Fatal("backup contents leaked")
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("sensitive data cached")
		}
	}
	req := RestoreRequest{created.ID, preview.Revision, preview.ExpiresAt, preview.Token, true}
	w = request(s, "POST", "/api/backups/restore", req, cookie, csrf)
	expectStatus(t, w, 200)
	if !c.IsEnabledIn(101) {
		t.Fatal("API restore failed")
	}
	expectStatus(t, request(s, "POST", "/api/logout", map[string]string{}, cookie, csrf), 200)
	expectStatus(t, request(s, "POST", "/api/backups/restore", req, cookie, csrf), 401)
}

func TestBackupLegacyIntegrityDatabaseDamageAndCancellation(t *testing.T) {
	s, c := backupFixture(t)
	jsonWithMarkup := []byte(strings.ReplaceAll(connectionFixture, "fixture", "<script>\u2028fixture</script>"))
	if err := s.ensureDirectory(); err != nil {
		t.Fatal(err)
	}
	markupID, err := writeFileBackup(s.directory, "startup-config", jsonWithMarkup)
	if err != nil {
		t.Fatal(err)
	}
	_, markup, err := s.load(markupID)
	if err != nil || !equivalentConfig(markup, jsonWithMarkup) {
		t.Fatal("escaped JSON hash/payload mismatch", err)
	}
	if validateOptions([]byte(`{"enabled":null,"address":"127.0.0.1:3010"}`)) == nil {
		t.Fatal("null enabled accepted")
	}
	id, err := s.Create("startup-config")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.directory, id)
	raw, _ := os.ReadFile(path)
	var file fileBackup
	_ = json.Unmarshal(raw, &file)
	file.SHA256 = ""
	raw, _ = json.Marshal(file)
	_ = os.WriteFile(path, raw, 0600)
	v, err := s.Preview(context.Background(), id, "fixture-session")
	if err != nil || v.Backup.Integrity || len(v.Warnings) < 2 {
		t.Fatal("legacy warning missing", v, err)
	}
	dbID, err := s.Create("control-db")
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Disable(101); err != nil {
		t.Fatal(err)
	}
	preview := previewRestore(t, s, dbID)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = s.Restore(ctx, preview, "fixture-session"); err == nil || c.IsEnabledIn(101) {
		t.Fatal("cancelled restore changed state", err)
	}
	root := filepath.Join(s.directory, dbID)
	manifestRaw, _ := os.ReadFile(filepath.Join(root, "manifest.json"))
	var manifest controlManifest
	_ = json.Unmarshal(manifestRaw, &manifest)
	broken := []byte("synthetic corrupted SQLite")
	manifest.SHA256 = digestBytes(broken)
	manifestRaw, _ = json.Marshal(manifest)
	_ = os.WriteFile(filepath.Join(root, "manifest.json"), manifestRaw, 0600)
	_ = os.WriteFile(filepath.Join(root, "plugins.sqlite"), broken, 0600)
	if _, err = s.Preview(context.Background(), dbID, "fixture-session"); !errors.Is(err, errBackupInvalid) {
		t.Fatal("damaged SQLite accepted", err)
	}
}

func TestBackupRestoreSerializesWithNormalControlWrites(t *testing.T) {
	s, c := backupFixture(t)
	id, err := s.Create("control-db")
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Disable(101); err != nil {
		t.Fatal(err)
	}
	preview := previewRestore(t, s, id)
	entered, release := make(chan struct{}), make(chan struct{})
	s.afterTable = func(i int) error {
		if i == 0 {
			close(entered)
			<-release
		}
		return nil
	}
	restored := make(chan error, 1)
	go func() { _, err := s.Restore(context.Background(), preview, "fixture-session"); restored <- err }()
	select {
	case <-entered:
	case err := <-restored:
		t.Fatal("restore never entered", err)
	case <-time.After(5 * time.Second):
		t.Fatal("restore stalled")
	}
	if c.Manager.TryLock() {
		c.Manager.Unlock()
		close(release)
		t.Fatal("restore did not retain Manager lock")
	}
	written := make(chan error, 1)
	go func() { written <- c.Disable(101) }()
	close(release)
	if err := <-restored; err != nil {
		t.Fatal(err)
	}
	if err := <-written; err != nil {
		t.Fatal(err)
	}
	if c.IsEnabledIn(101) {
		t.Fatal("normal writer was lost after restore")
	}
}

func TestControlBanReloadSignedIDsAndGlobalFallback(t *testing.T) {
	_, c := backupFixture(t)
	c.Ban(789, 101) // Fixture digest has its high bit set.
	c.Manager.Lock()
	c.Manager.RestoredLocked(nil, nil)
	c.Manager.Unlock()
	if !c.IsBannedIn(789, 101) {
		t.Fatal("signed persisted ban unreadable after reload")
	}
	c.Permit(789, 101)
	c.Manager.Lock()
	c.Manager.RestoredLocked(nil, nil)
	c.Manager.Unlock()
	if c.IsBannedIn(789, 101) {
		t.Fatal("signed persisted ban not removed")
	}
	c.Ban(456, 0)
	if !c.IsBannedIn(456, 101) || !c.IsBannedIn(456, 101) {
		t.Fatal("cached group miss bypassed global ban")
	}
}
