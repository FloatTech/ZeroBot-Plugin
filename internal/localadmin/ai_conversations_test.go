package localadmin

import (
	"errors"
	"os"
	"testing"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func TestAIPrivateStorageIsolationBackupAndChatRevision(t *testing.T) {
	s, c, path, backups := aiFixture(t)
	if err := c.SetData(-101, 1<<30); err != nil {
		t.Fatal(err)
	}
	if err := c.Disable(-101); err != nil {
		t.Fatal(err)
	}
	global := aiView(t, s, 0).Global.Revision
	v := aiView(t, s, -101)
	if v.Groups[0].PrivateID != "101" || v.Groups[0].GroupID != "" {
		t.Fatal("private identity not explicit")
	}
	temp, rate, enabled := 84, 25, true
	update := AIGroupUpdate{Service: "aichatcfg", PrivateID: "101", Revision: v.Groups[0].Revision, Temperature: &temp, Rate: &rate, UseAgent: &enabled}
	before, err := os.ReadDir(backups)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SaveGroup(update); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadDir(backups)
	if err != nil || len(after) != len(before)+1 {
		t.Fatal("private update lacks backup")
	}
	if c.IsEnabledIn(-101) || c.GetData(-101)&(1<<30) == 0 {
		t.Fatal("private update changed switch or unknown bits")
	}
	actual := aiView(t, s, -101).Groups[0]
	if actual.Temperature != 84 || actual.Rate != 25 || !actual.UseAgent {
		t.Fatal("private values wrong")
	}
	for _, id := range []int64{101, -102} {
		other := aiView(t, s, id).Groups[0]
		if other.Temperature != 0 || other.Rate != 0 {
			t.Fatal("private update leaked into another conversation")
		}
	}
	if aiView(t, s, 0).Global.Revision != global {
		t.Fatal("private settings changed global model")
	}
	ctx := &zero.Ctx{State: zero.State{"manager": c}}
	if err := ctxext.UpdateStorage(ctx, -101, func(old ctxext.Storage) ctxext.Storage { return old.Set(31, chat.BitmapRate) }); err != nil {
		t.Fatal(err)
	}
	if aiView(t, s, -101).Groups[0].Rate != 31 {
		t.Fatal("chat and admin private storage differ")
	}
	if err := s.SaveGroup(update); !errors.Is(err, ctrl.ErrConfigConflict) {
		t.Fatal("stale private revision accepted")
	}
	c.Manager.Lock()
	c.BeforeData = func() error { return errors.New("synthetic backup failure") }
	c.Manager.Unlock()
	update.Revision = aiView(t, s, -101).Groups[0].Revision
	if err := s.SaveGroup(update); err == nil || aiView(t, s, -101).Groups[0].Rate != 31 {
		t.Fatal("failed private backup changed settings")
	}
	if err := c.Manager.D.Close(); err != nil {
		t.Fatal(err)
	}
	m := ctrl.NewManager[*zero.Ctx](path)
	c2 := m.NewControl("aichatcfg", &ctrl.Options[*zero.Ctx]{Extra: 12345})
	l2 := m.NewControl("llm", nil)
	t.Cleanup(func() { m.Lock(); defer m.Unlock(); _ = m.D.Close() })
	reopened := aiView(t, NewAIStore([]*ctrl.Control[*zero.Ctx]{c2, l2}), -101).Groups[0]
	if reopened.Rate != 31 || reopened.Temperature != 84 {
		t.Fatal("private settings did not survive database reopening")
	}
}

func TestAIPrivateAPIValidationAuthAndTarget(t *testing.T) {
	s := newTestServer(t)
	store, _, _, _ := aiFixture(t)
	s.cfg.AI = store
	cookie, csrf := enroll(t, s)
	for _, path := range []string{"/api/ai?private=0", "/api/ai?private=-101", "/api/ai?private=01", "/api/ai?private=9223372036854775808", "/api/ai?private=101&group=101", "/api/ai?group=-101"} {
		expectStatus(t, request(s, "GET", path, nil, cookie, ""), 400)
	}
	expectStatus(t, request(s, "GET", "/api/ai?private=101", nil, nil, ""), 401)
	expectStatus(t, request(s, "GET", "/api/ai?private=101", nil, cookie, ""), 200)
	temp := 80
	update := AIGroupUpdate{Service: "llm", PrivateID: "101", Revision: aiView(t, store, -101).Groups[1].Revision, Temperature: &temp}
	expectStatus(t, request(s, "POST", "/api/ai-group", update, nil, ""), 401)
	expectStatus(t, request(s, "POST", "/api/ai-group", update, cookie, "bad"), 403)
	expectStatus(t, request(s, "POST", "/api/ai-group", update, cookie, csrf), 200)
	expectStatus(t, request(s, "POST", "/api/ai-group", update, cookie, csrf), 409)
	if aiView(t, store, -101).Groups[1].Temperature != 80 || aiView(t, store, 101).Groups[1].Temperature != 0 {
		t.Fatal("API saved wrong conversation")
	}
	update.Revision = aiView(t, store, -101).Groups[1].Revision
	update.GroupID = "101"
	expectStatus(t, request(s, "POST", "/api/ai-group", update, cookie, csrf), 400)
	update.GroupID = ""
	update.Rate = &temp
	expectStatus(t, request(s, "POST", "/api/ai-group", update, cookie, csrf), 400)
}
