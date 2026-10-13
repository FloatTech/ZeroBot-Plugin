package localadmin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func TestResponseStateGlobalOrGroupMatchesExistingRules(t *testing.T) {
	c, _ := pluginFixture(t, false)
	m := c.Manager
	check := func(gid int64, global, group bool, source string) {
		t.Helper()
		v, err := m.ReadResponseState(gid)
		if err != nil || v.GlobalEnabled != global || v.GroupEnabled != group || v.EffectiveEnabled != (global || group) || v.Source != source || m.CanResponse(gid) != v.EffectiveEnabled {
			t.Fatal("response gate mismatch", v, err)
		}
	}
	check(0, false, false, "none")
	check(101, false, false, "none")
	if err := m.Response(101); err != nil {
		t.Fatal(err)
	}
	check(101, false, true, "group")
	check(202, false, false, "none")
	if err := m.Response(0); err != nil {
		t.Fatal(err)
	}
	check(0, true, false, "global")
	check(101, true, true, "global")
	check(202, true, false, "global")
	if err := m.Silence(202); err != nil {
		t.Fatal(err)
	}
	check(202, true, false, "global")
	if err := m.Silence(0); err != nil {
		t.Fatal(err)
	}
	check(101, false, true, "group")
	check(202, false, false, "none")
}

func TestResponseStateReadHidesAndPreservesExtensionsAndNeverBacksUp(t *testing.T) {
	c, _ := pluginFixture(t, false)
	m := c.Manager
	secret := `{"Key":"private-model-secret","future":9007199254740993}`
	m.Lock()
	err := m.D.Insert("__resp", &ctrl.ResponseGroup{GroupID: 9223372036854775807, Extra: secret})
	m.BeforePolicy = func() error { t.Fatal("read invoked policy backup"); return nil }
	m.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	v, err := m.ReadResponseState(9223372036854775807)
	if err != nil || !v.GroupEnabled || v.Source != "group" {
		t.Fatal(v, err)
	}
	out, _ := json.Marshal(v)
	if strings.Contains(string(out), "private") || strings.Contains(string(out), "future") || strings.Contains(string(out), "9007199") {
		t.Fatal("extension leaked")
	}
	m.Lock()
	var row ctrl.ResponseGroup
	err = m.D.Find("__resp", &row, "WHERE gid = ?", int64(9223372036854775807))
	if err == nil {
		err = m.D.Insert("__resp", &ctrl.ResponseGroup{GroupID: 101, Extra: "-"})
	}
	m.Unlock()
	if err != nil || row.Extra != secret {
		t.Fatal("read changed extension", err)
	}
	if v, err := m.ReadResponseState(101); err != nil || v.EffectiveEnabled {
		t.Fatal("sentinel interpreted as responding", err)
	}
}

func TestResponseStateReadErrorsDoNotReturnCachedActiveState(t *testing.T) {
	c, _ := pluginFixture(t, false)
	if err := c.Manager.Response(0); err != nil || !c.Manager.CanResponse(101) {
		t.Fatal("fixture not active", err)
	}
	if _, err := c.Manager.ReadResponseState(-1); err == nil {
		t.Fatal("private/reserved row exposed as a group")
	}
	c.Manager.Lock()
	_, err := c.Manager.D.Exec("DROP TABLE __resp")
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Manager.ReadResponseState(101); err == nil {
		t.Fatal("database error silently converted to cached state")
	}
}

func TestPluginResponseStateAPIAuthUnavailableAndReadOnly(t *testing.T) {
	c, _ := pluginFixture(t, false)
	backups := filepath.Join(t.TempDir(), "backups")
	cfg := testConfig(t)
	cfg.Plugins = []Plugin{{Name: c.Service}}
	cfg.PluginControl = NewPluginStore([]*ctrl.Control[*zero.Ctx]{c}, backups)
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(s, "GET", "/api/plugins?group=101", nil, nil, ""), 401)
	cookie, _ := enroll(t, s)
	read := func() map[string]any {
		t.Helper()
		w := request(s, "GET", "/api/plugins?group=101", nil, cookie, "")
		expectStatus(t, w, 200)
		var value map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	if read()["response"].(map[string]any)["effectiveEnabled"] != false {
		t.Fatal("empty response assumed active")
	}
	if _, err := os.Stat(backups); !os.IsNotExist(err) {
		t.Fatal("read created backup directory", err)
	}
	if err := c.Manager.Response(101); err != nil {
		t.Fatal(err)
	}
	if read()["response"].(map[string]any)["source"] != "group" {
		t.Fatal("chat policy change not visible")
	}
	other, _ := pluginFixture(t, false)
	second := other.Manager.NewControl("second", &ctrl.Options[*zero.Ctx]{})
	s.cfg.PluginControl = NewPluginStore([]*ctrl.Control[*zero.Ctx]{c, second}, backups)
	if read()["response"] != nil {
		t.Fatal("multiple managers guessed unified response")
	}
	s.cfg.PluginControl = nil
	if read()["response"] != nil {
		t.Fatal("unavailable control guessed response")
	}
	expectStatus(t, request(s, "GET", "/api/plugins?group=-1", nil, cookie, ""), 400)
}
