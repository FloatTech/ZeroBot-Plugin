package localadmin

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func pluginFixture(t *testing.T, disabled bool) (*ctrl.Control[*zero.Ctx], string) {
	t.Helper()
	path := filepath.ToSlash(filepath.Join(t.TempDir(), "plugins.db"))
	m := ctrl.NewManager[*zero.Ctx](path)
	c := m.NewControl("test-plugin", &ctrl.Options[*zero.Ctx]{DisableOnDefault: disabled})
	t.Cleanup(func() { m.Lock(); defer m.Unlock(); _ = m.D.Close() })
	return c, path
}

func switchState(t *testing.T, c *ctrl.Control[*zero.Ctx], gid int64) ctrl.SwitchState {
	t.Helper()
	v, err := c.SwitchState(gid)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestPluginSwitchPriorityAndLocationIsolation(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		c, _ := pluginFixture(t, disabled)
		for _, global := range []string{"inherit", "enabled", "disabled"} {
			for _, group := range []string{"inherit", "enabled", "disabled"} {
				if _, err := c.SetSwitch(0, global, ""); err != nil {
					t.Fatal(err)
				}
				if _, err := c.SetSwitch(101, group, ""); err != nil {
					t.Fatal(err)
				}
				enabled, source := !disabled, "default"
				if group != "inherit" {
					enabled, source = group == "enabled", "group"
				}
				if global != "inherit" {
					enabled, source = global == "enabled", "global"
				}
				v := switchState(t, c, 101)
				if v.EffectiveEnabled != enabled || v.Source != source || c.IsEnabledIn(101) != enabled {
					t.Fatalf("priority mismatch %v %s %s: %+v", disabled, global, group, v)
				}
				other := !disabled
				if global != "inherit" {
					other = global == "enabled"
				}
				if c.IsEnabledIn(202) != other || c.IsEnabledIn(-123) != other || c.IsEnabledIn(0) != other {
					t.Fatal("group override leaked")
				}
			}
		}
	}
}

func TestPluginInheritancePreservesDataAndFlipAcrossRestart(t *testing.T) {
	c, path := pluginFixture(t, false)
	if err := c.SetData(101, 123456789); err != nil {
		t.Fatal(err)
	}
	if err := c.Disable(101); err != nil {
		t.Fatal(err)
	}
	if err := c.Reset(101); err != nil {
		t.Fatal(err)
	}
	if c.GetData(101) != 123456789 || switchState(t, c, 101).GroupMode != "inherit" {
		t.Fatal("reset lost group parameters")
	}
	if err := c.SetData(101, 987654321); err != nil {
		t.Fatal(err)
	}
	if switchState(t, c, 101).GroupMode != "inherit" {
		t.Fatal("parameter write changed inheritance")
	}
	if err := c.Flip(); err != nil {
		t.Fatal(err)
	}
	if c.IsEnabledIn(101) || switchState(t, c, 0).GlobalMode != "inherit" {
		t.Fatal("flip created a global override")
	}
	if err := c.SetData(0, 42); err != nil {
		t.Fatal(err)
	}
	if err := c.Enable(0); err != nil {
		t.Fatal(err)
	}
	if err := c.Reset(0); err != nil {
		t.Fatal(err)
	}
	if c.GetData(0) != 42 || c.IsEnabledIn(101) {
		t.Fatal("global inheritance lost data or default")
	}
	c.Manager.Lock()
	err := c.Manager.D.Close()
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	reopened := ctrl.NewManager[*zero.Ctx](path)
	defer reopened.D.Close()
	r := reopened.NewControl("test-plugin", &ctrl.Options[*zero.Ctx]{})
	if r.IsEnabledIn(101) || r.GetData(101) != 987654321 || r.GetData(0) != 42 || switchState(t, r, 101).GroupMode != "inherit" {
		t.Fatal("restart did not retain default/inheritance/data")
	}
	if err := r.Flip(); err != nil {
		t.Fatal(err)
	}
	if !r.IsEnabledIn(101) {
		t.Fatal("second flip failed")
	}
}

func TestPluginLegacyRowsAndTransactionalFailure(t *testing.T) {
	c, _ := pluginFixture(t, false)
	c.Manager.Lock()
	err := c.Manager.D.Insert(c.Service, &ctrl.GroupConfig{GroupID: 101, Disable: 19})
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if c.IsEnabledIn(101) || switchState(t, c, 101).GroupMode != "disabled" {
		t.Fatal("legacy group row changed")
	}
	before := switchState(t, c, 101)
	// Fail the second transaction statement, after the plugin row would change.
	c.Manager.Lock()
	_, err = c.Manager.D.Exec("CREATE TRIGGER fail_inheritance BEFORE INSERT ON __switch_inherit BEGIN SELECT RAISE(ABORT,'injected-secret'); END")
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Reset(101); err == nil {
		t.Fatal("expected transaction failure")
	}
	if switchState(t, c, 101) != before || c.IsEnabledIn(101) || c.GetData(101) != 9 {
		t.Fatal("failed transaction changed disk or cache")
	}
}

func TestPluginConcurrentChatAndWebWrites(t *testing.T) {
	c, _ := pluginFixture(t, false)
	v := switchState(t, c, 101)
	results := make(chan error, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := c.SetSwitch(101, "disabled", v.Revision); results <- err }()
	}
	wg.Wait()
	close(results)
	count := 0
	for err := range results {
		if err == nil {
			count++
		} else if !errors.Is(err, ctrl.ErrSwitchConflict) {
			t.Fatal(err)
		}
	}
	if count != 1 {
		t.Fatalf("%d concurrent writes succeeded", count)
	}
	v = switchState(t, c, 101)
	if err := c.Enable(101); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SetSwitch(101, "disabled", v.Revision); !errors.Is(err, ctrl.ErrSwitchConflict) {
		t.Fatal("chat update did not invalidate web revision")
	}
	if !c.IsEnabledIn(101) {
		t.Fatal("chat write not visible")
	}
}

func TestPluginSQLiteSnapshotAndBackupFailure(t *testing.T) {
	c, _ := pluginFixture(t, false)
	backups := filepath.Join(t.TempDir(), "backups")
	_ = NewPluginStore([]*ctrl.Control[*zero.Ctx]{c}, backups)
	c.Manager.Lock()
	_, err := c.Manager.D.Exec("PRAGMA journal_mode=WAL")
	if err == nil {
		err = c.Manager.D.Insert(c.Service, &ctrl.GroupConfig{GroupID: 101, Disable: 34})
	}
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Disable(101); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(backups)
	if err != nil || len(entries) != 1 {
		t.Fatalf("snapshot missing: %v", err)
	}
	root := filepath.Join(backups, entries[0].Name())
	var manifest struct {
		Version             int
		Scope, File, SHA256 string
		CreatedAt           time.Time
	}
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil || json.Unmarshal(raw, &manifest) != nil || manifest.Scope != "control-db" || manifest.Version != 1 || len(manifest.SHA256) != 64 || manifest.CreatedAt.IsZero() {
		t.Fatal("invalid manifest")
	}
	copyDB := sql.New(filepath.Join(root, manifest.File))
	if err = copyDB.Open(time.Hour); err != nil {
		t.Fatal(err)
	}
	defer copyDB.Close()
	var old ctrl.GroupConfig
	if err = copyDB.Find(c.Service, &old, "WHERE gid=101"); err != nil || old.Disable != 34 {
		t.Fatal("snapshot omitted pre-write WAL data")
	}
	if err = c.Disable(101); err != nil {
		t.Fatal(err)
	}
	entries, _ = os.ReadDir(backups)
	if len(entries) != 1 {
		t.Fatal("no-op made another snapshot")
	}
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err = os.WriteFile(blocked, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	_ = NewPluginStore([]*ctrl.Control[*zero.Ctx]{c}, blocked)
	before := switchState(t, c, 101)
	if err = c.Enable(101); err == nil {
		t.Fatal("backup failure allowed write")
	}
	if switchState(t, c, 101) != before || c.IsEnabledIn(101) {
		t.Fatal("backup failure changed state")
	}
}

func TestPluginAPIAuthenticationValidationAndConflict(t *testing.T) {
	s := newTestServer(t)
	c, _ := pluginFixture(t, false)
	s.cfg.PluginControl = NewPluginStore([]*ctrl.Control[*zero.Ctx]{c}, filepath.Join(t.TempDir(), "backups"))
	update := map[string]string{"name": c.Service, "groupId": "101", "mode": "disabled", "revision": switchState(t, c, 101).Revision}
	expectStatus(t, request(s, "POST", "/api/plugin-switch", update, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/plugin-switch", update, cookie, "wrong"), 403)
	for _, group := range []string{"-1", "01", "abc", "9223372036854775808"} {
		expectStatus(t, request(s, "GET", "/api/plugins?group="+group, nil, cookie, ""), 400)
	}
	expectStatus(t, request(s, "POST", "/api/plugin-switch", update, cookie, csrf), 200)
	expectStatus(t, request(s, "POST", "/api/plugin-switch", update, cookie, csrf), 409)
	if c.IsEnabledIn(101) {
		t.Fatal("API not visible to engine")
	}
	update["revision"] = switchState(t, c, 101).Revision
	update["name"] = "../../unknown"
	expectStatus(t, request(s, "POST", "/api/plugin-switch", update, cookie, csrf), 404)
	update["name"] = c.Service
	update["revision"] = ""
	expectStatus(t, request(s, "POST", "/api/plugin-switch", update, cookie, csrf), 400)
	c.Manager.Lock()
	c.Manager.BeforeSwitch = func() error { return errors.New("secret-must-not-leak") }
	c.Manager.Unlock()
	update["revision"] = switchState(t, c, 101).Revision
	update["mode"] = "enabled"
	rsp := request(s, "POST", "/api/plugin-switch", update, cookie, csrf)
	expectStatus(t, rsp, 500)
	if strings.Contains(rsp.Body.String(), "secret-must-not-leak") {
		t.Fatal("error leaked")
	}
	rsp = request(s, "GET", "/api/plugins?group=101", nil, cookie, "")
	expectStatus(t, rsp, 200)
	if !strings.Contains(rsp.Body.String(), `"effectiveEnabled":false`) || !strings.Contains(rsp.Body.String(), `"source":"group"`) {
		t.Fatal("actual state missing")
	}
}
