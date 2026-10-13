package localadmin

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func aiFixture(t *testing.T) (*AIStore, *ctrl.Control[*zero.Ctx], string, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.ToSlash(filepath.Join(dir, "plugins.db"))
	m := ctrl.NewManager[*zero.Ctx](path)
	c := m.NewControl("aichatcfg", &ctrl.Options[*zero.Ctx]{Extra: 12345})
	llm := m.NewControl("llm", nil)
	backups := filepath.Join(dir, "private", "backups")
	NewPluginStore([]*ctrl.Control[*zero.Ctx]{c, llm}, backups)
	s := NewAIStore([]*ctrl.Control[*zero.Ctx]{c, llm})
	t.Cleanup(func() { m.Lock(); defer m.Unlock(); _ = m.D.Close() })
	return s, c, path, backups
}
func aiView(t *testing.T, s *AIStore, gid int64) AIView {
	t.Helper()
	v, err := s.Get(gid)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func aiFieldsRaw(values map[string]any) map[string]json.RawMessage {
	fields := map[string]json.RawMessage{}
	for name, value := range values {
		fields[name], _ = json.Marshal(value)
	}
	return fields
}

func TestAISharedPersistenceUnknownFieldsAndRedaction(t *testing.T) {
	s, c, path, backups := aiFixture(t)
	initial := map[string]any{"ModelName": "partial-model", "API": "https://user:legacy-url-secret@example.test/v1?key=legacy-query-secret", "Key": "fake-key-must-stay-hidden", "ImageKey": "fake-image-key", "AgentKey": "fake-agent-key", "SystemP": "Never print fake-key-must-stay-hidden", "Future": map[string]any{"n": 9007199254740993}}
	if err := c.SetExtra(initial); err != nil {
		t.Fatal(err)
	}
	v := aiView(t, s, 101)
	b, _ := json.Marshal(v)
	for _, secret := range []string{"fake-key-must-stay-hidden", "fake-image-key", "fake-agent-key", "legacy-url-secret", "legacy-query-secret"} {
		if strings.Contains(string(b), secret) {
			t.Fatalf("read exposed secret %q", secret)
		}
	}
	if !v.Global.AddressRedacted["API"] || !v.Global.Keys["Key"] {
		t.Fatal("missing secret state")
	}
	if err := s.Save(AIUpdate{Revision: v.Global.Revision, Fields: aiFieldsRaw(map[string]any{"ModelName": "new-model"})}); err != nil {
		t.Fatal(err)
	}
	raw, _, err := c.ExtraJSON()
	if err != nil || !strings.Contains(string(raw), "legacy-query-secret") || !strings.Contains(string(raw), "fake-image-key") || !strings.Contains(string(raw), "Future") {
		t.Fatal("unedited fields lost")
	}
	cfg, _, err := s.repository.Read()
	if err != nil || cfg.ModelName != "new-model" || cfg.Key != "fake-key-must-stay-hidden" {
		t.Fatal("runtime not seeing persisted config")
	}
	entries, err := os.ReadDir(backups)
	if err != nil || len(entries) != 2 {
		t.Fatal("write-before snapshots missing")
	}
	if err := s.Save(AIUpdate{Revision: aiView(t, s, 0).Global.Revision}); err != nil {
		t.Fatal(err)
	}
	entries, _ = os.ReadDir(backups)
	if len(entries) != 2 {
		t.Fatal("unchanged save created backup")
	}
	if err := c.Manager.D.Close(); err != nil {
		t.Fatal(err)
	}
	m2 := ctrl.NewManager[*zero.Ctx](path)
	defer m2.D.Close()
	c2 := m2.NewControl("aichatcfg", &ctrl.Options[*zero.Ctx]{Extra: 12345})
	r := chat.NewConfigRepository(c2)
	cfg, _, err = r.Read()
	if err != nil || cfg.ModelName != "new-model" || cfg.ImageKey != "fake-image-key" {
		t.Fatal("restart lost configuration")
	}
}

func TestAIValidationKeepClearAndWriteFailures(t *testing.T) {
	s, c, _, backups := aiFixture(t)
	v := aiView(t, s, 0)
	for _, fields := range []map[string]any{{"Type": 99}, {"TopP": 2}, {"MaxN": 131073}, {"Type": json.RawMessage(" null ")}, {"API": "file:///tmp/a"}, {"API": "http://127.0.0.1:65536"}, {"API": "http://127.0.0.1:0"}, {"API": "https://user:password@example.test"}, {"API": "https://example.test/?key=secret"}, {"ModelName": ""}, {"Unknown": 1}, {"SystemP": strings.Repeat("x", 32769)}} {
		if err := s.Save(AIUpdate{Revision: v.Global.Revision, Fields: aiFieldsRaw(fields)}); !errors.Is(err, errAIValidation) {
			t.Fatalf("invalid fields accepted: %v", err)
		}
		if aiView(t, s, 0).Global.Revision != v.Global.Revision {
			t.Fatal("invalid save changed state")
		}
	}
	if err := s.Save(AIUpdate{Revision: v.Global.Revision, Keys: map[string]AIKeyChange{"Key": {Action: "replace", Value: "synthetic-secret"}}}); err != nil {
		t.Fatal(err)
	}
	v = aiView(t, s, 0)
	if err := s.Save(AIUpdate{Revision: v.Global.Revision, Keys: map[string]AIKeyChange{"Key": {Action: "clear"}}}); !errors.Is(err, errAIValidation) {
		t.Fatal("unconfirmed clear accepted")
	}
	if err := s.Save(AIUpdate{Revision: v.Global.Revision, Keys: map[string]AIKeyChange{"Key": {Action: "keep"}}}); err != nil {
		t.Fatal(err)
	}
	c.Manager.Lock()
	oldHook := c.Manager.BeforeExtra
	c.Manager.BeforeExtra = func() error { return errors.New("synthetic-secret-backup-error") }
	c.Manager.Unlock()
	if err := s.Save(AIUpdate{Revision: v.Global.Revision, Fields: aiFieldsRaw(map[string]any{"ModelName": "not-committed"})}); err == nil {
		t.Fatal("backup failure ignored")
	}
	if aiView(t, s, 0).Global.Revision != v.Global.Revision {
		t.Fatal("failed backup changed state")
	}
	c.Manager.Lock()
	c.Manager.BeforeExtra = oldHook
	_, err := c.Manager.D.Exec("CREATE TRIGGER fail_extra BEFORE INSERT ON __resp BEGIN SELECT RAISE(ABORT,'synthetic-secret-sql-error'); END")
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(AIUpdate{Revision: v.Global.Revision, Keys: map[string]AIKeyChange{"Key": {Action: "clear", Confirm: true}}}); err == nil {
		t.Fatal("SQL failure ignored")
	}
	if aiView(t, s, 0).Global.Revision != v.Global.Revision {
		t.Fatal("failed SQL changed runtime")
	}
	c.Manager.Lock()
	_, err = c.Manager.D.Exec("DROP TRIGGER fail_extra")
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(AIUpdate{Revision: v.Global.Revision, Keys: map[string]AIKeyChange{"Key": {Action: "clear", Confirm: true}}}); err != nil {
		t.Fatal(err)
	}
	if aiView(t, s, 0).Global.Keys["Key"] {
		t.Fatal("confirmed clear failed")
	}
	entries, _ := os.ReadDir(backups)
	if len(entries) != 3 {
		t.Fatalf("snapshot count %d", len(entries))
	}
}

func TestAIGroupAtomicMergeAndSwitchPreservation(t *testing.T) {
	s, c, _, _ := aiFixture(t)
	if err := c.SetData(101, 1<<30); err != nil {
		t.Fatal(err)
	}
	if err := c.Disable(101); err != nil {
		t.Fatal(err)
	}
	before := switchState(t, c, 101)
	v := aiView(t, s, 101)
	temp, rate := 84, 25
	if err := s.SaveGroup(AIGroupUpdate{Service: "aichatcfg", GroupID: "101", Revision: v.Groups[0].Revision, Temperature: &temp, Rate: &rate}); err != nil {
		t.Fatal(err)
	}
	if c.IsEnabledIn(101) || switchState(t, c, 101).GroupMode != before.GroupMode || c.GetData(101)&(1<<30) == 0 {
		t.Fatal("group settings changed switch or unknown bits")
	}
	other := aiView(t, s, 202)
	if other.Groups[0].Temperature != 0 || other.Groups[0].Rate != 0 {
		t.Fatal("group settings leaked")
	}
	ctx := &zero.Ctx{State: zero.State{"manager": c}}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, setting := range []struct{ mask, value int64 }{{chat.BitmapRate, 31}, {chat.BitmapTemp, 77}} {
		wg.Add(1)
		go func(mask, value int64) {
			defer wg.Done()
			errs <- ctxext.UpdateStorage(ctx, 101, func(old ctxext.Storage) ctxext.Storage { return old.Set(value, mask) })
		}(setting.mask, setting.value)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	updated := aiView(t, s, 101).Groups[0]
	if updated.Rate != 31 || updated.Temperature != 77 {
		t.Fatal("concurrent field update lost")
	}
	if err := s.SaveGroup(AIGroupUpdate{Service: "aichatcfg", GroupID: "101", Revision: v.Groups[0].Revision, Temperature: &temp}); !errors.Is(err, ctrl.ErrConfigConflict) {
		t.Fatal("chat change did not invalidate web version")
	}
	c.Manager.Lock()
	_, err := c.Manager.D.Exec("CREATE TRIGGER fail_group BEFORE INSERT ON __switch_inherit BEGIN SELECT RAISE(ABORT,'group-failure'); END")
	c.Manager.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SaveGroup(AIGroupUpdate{Service: "aichatcfg", GroupID: "202", Revision: other.Groups[0].Revision, Temperature: &temp}); err == nil {
		t.Fatal("group transaction failure ignored")
	}
	if aiView(t, s, 202).Groups[0].Temperature != 0 {
		t.Fatal("failed transaction changed group")
	}
}

func TestAIConcurrentWebAndChatAndManagerIsolation(t *testing.T) {
	s, c, _, _ := aiFixture(t)
	v := aiView(t, s, 0)
	results := make(chan error, 2)
	for _, name := range []string{"winner-a", "winner-b"} {
		go func(name string) {
			results <- s.Save(AIUpdate{Revision: v.Global.Revision, Fields: aiFieldsRaw(map[string]any{"ModelName": name})})
		}(name)
	}
	win, conflict := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			win++
		} else if errors.Is(err, ctrl.ErrConfigConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if win != 1 || conflict != 1 {
		t.Fatal("old pages both committed")
	}
	v = aiView(t, s, 0)
	if _, err := s.repository.Update("", aiFieldsRaw(map[string]any{"TopP": 0.6})); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(AIUpdate{Revision: v.Global.Revision, Fields: aiFieldsRaw(map[string]any{"ModelName": "stale"})}); !errors.Is(err, ctrl.ErrConfigConflict) {
		t.Fatal("chat write not detected")
	}
	s2, c2, _, _ := aiFixture(t)
	if err := c.Manager.Response(0); err != nil {
		t.Fatal(err)
	}
	if c2.Manager.CanResponse(101) {
		t.Fatal("response cache crossed Managers")
	}
	if aiView(t, s2, 0).Global.Fields["ModelName"] == aiView(t, s, 0).Global.Fields["ModelName"] {
		t.Fatal("AI configuration crossed Managers")
	}
}

func TestAIAPIAuthConflictAndSafeErrors(t *testing.T) {
	s := newTestServer(t)
	store, c, _, _ := aiFixture(t)
	s.cfg.AI = store
	update := AIUpdate{Revision: aiView(t, store, 0).Global.Revision, Fields: aiFieldsRaw(map[string]any{"ModelName": "api-model"})}
	expectStatus(t, request(s, "GET", "/api/ai", nil, nil, ""), 401)
	expectStatus(t, request(s, "POST", "/api/ai", update, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	expectStatus(t, request(s, "POST", "/api/ai", update, cookie, "bad"), 403)
	expectStatus(t, request(s, "POST", "/api/ai", update, cookie, csrf), 200)
	expectStatus(t, request(s, "POST", "/api/ai", update, cookie, csrf), 409)
	update.Revision = aiView(t, store, 0).Global.Revision
	update.Fields = aiFieldsRaw(map[string]any{"TopP": 2})
	expectStatus(t, request(s, "POST", "/api/ai", update, cookie, csrf), 400)
	update.Fields = aiFieldsRaw(map[string]any{"ModelName": "fail"})
	c.Manager.Lock()
	c.Manager.BeforeExtra = func() error { return errors.New("synthetic-api-key-error") }
	c.Manager.Unlock()
	rsp := request(s, "POST", "/api/ai", update, cookie, csrf)
	expectStatus(t, rsp, 500)
	if strings.Contains(rsp.Body.String(), "synthetic-api-key-error") {
		t.Fatal("error leaked")
	}
	for _, group := range []string{"-1", "01", "9223372036854775808"} {
		expectStatus(t, request(s, "GET", "/api/ai?group="+group, nil, cookie, ""), 400)
	}
	g := aiView(t, store, 101).Groups[0]
	temp := 85
	body := AIGroupUpdate{Service: "aichatcfg", GroupID: "101", Revision: g.Revision, Temperature: &temp}
	expectStatus(t, request(s, "POST", "/api/ai-group", body, cookie, "bad"), 403)
	expectStatus(t, request(s, "POST", "/api/ai-group", body, cookie, csrf), 200)
	expectStatus(t, request(s, "POST", "/api/ai-group", body, cookie, csrf), 409)
	bad := AIGroupUpdate{Service: "llm", GroupID: "101", Revision: aiView(t, store, 101).Groups[1].Revision, Rate: &temp}
	expectStatus(t, request(s, "POST", "/api/ai-group", bad, cookie, csrf), 400)
}
