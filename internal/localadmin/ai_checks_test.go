package localadmin

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/FloatTech/zbputils/chat"
)

func TestAIConfigCheckTypesRequiredFieldsAndIndependence(t *testing.T) {
	for _, kind := range []chat.ModelType{-1, 0, 1, 2, 3} {
		for _, key := range []string{"", "synthetic-key"} {
			c := chat.AgentConfig{Type: kind, ModelName: "fixture", API: "http://fixture.invalid/chat", Key: chat.ModelKey(key)}
			checks := aiConfigChecks(c)
			got := checks["chat"]
			want := kind >= 0 && kind <= 2 && (kind == 1 || key != "")
			if got.Complete != want || got.KeyRequired != (kind != 1) || got.KeyConfigured != (key != "") || !got.ModelRequired {
				t.Fatal("type/key check disagrees", kind, got)
			}
			if checks["agent"].Complete || checks["image"].Complete || checks["image"].ModelRequired {
				t.Fatal("independent or optional model state incorrect")
			}
		}
	}
	c := chat.AgentConfig{Type: 1, ModelName: " \t", API: "http://fixture.invalid/chat", AgentType: 1, AgentModelName: "agent", AgentAPI: "http://fixture.invalid/agent", ImageType: 1, ImageAPI: "http://fixture.invalid/image"}
	checks := aiConfigChecks(c)
	if checks["chat"].Complete || checks["chat"].ModelConfigured || !checks["agent"].Complete || !checks["image"].Complete || checks["image"].ModelConfigured {
		t.Fatal("blank required/optional models incorrect", checks)
	}
}

func TestAIConfigCheckAddressRulesAndBooleanOnlyOutput(t *testing.T) {
	for _, address := range []string{"", " ", "file:///private-key", "ftp://fixture.invalid", "http://fixture.invalid:0", "http://fixture.invalid:65536", "https://private-user:secret@fixture.invalid/v1", "https://fixture.invalid/v1?key=secret", "https://fixture.invalid/v1#secret", "http://bad host"} {
		check := checkAIConfig(1, "private-model-name", address, "private-model-key", "API", true)
		if check.AddressValid || check.Complete {
			t.Fatal("invalid/legacy address passed")
		}
		raw, err := json.Marshal(check)
		if err != nil || strings.Contains(string(raw), "private") || strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "fixture.invalid") {
			t.Fatal("check serialized field contents")
		}
		var fields map[string]any
		_ = json.Unmarshal(raw, &fields)
		if len(fields) != 8 {
			t.Fatal("unexpected check fields")
		}
		for _, value := range fields {
			if _, ok := value.(bool); !ok {
				t.Fatal("check contains non-boolean metadata")
			}
		}
	}
	for _, address := range []string{"http://127.0.0.1:11434/api/chat", "https://fixture.invalid:65535/v1"} {
		if !checkAIConfig(1, "fixture", address, "", "API", true).Complete {
			t.Fatal("valid HTTP address rejected")
		}
	}
}

func TestAIConfigChecksUseRawSnapshotAndReadDoesNotWrite(t *testing.T) {
	s, c, _, backups := aiFixture(t)
	if err := c.SetExtra(map[string]any{"Type": 1, "ModelName": "synthetic-model-secret", "API": "http://fixture.invalid/synthetic-model-secret", "Key": "synthetic-model-secret", "AgentType": 0, "AgentModelName": "agent", "AgentAPI": "https://user:hidden-url-secret@fixture.invalid/v1", "AgentKey": "synthetic-agent-secret", "Future": map[string]any{"n": 9007199254740993}}); err != nil {
		t.Fatal(err)
	}
	before, revision, err := c.ExtraJSON()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(backups)
	if err != nil {
		t.Fatal(err)
	}
	for _, gid := range []int64{0, 101, -101} {
		v := aiView(t, s, gid)
		if !v.Global.Checks["chat"].Complete || v.Global.Checks["agent"].Complete || !v.Global.Checks["agent"].AddressConfigured || v.Global.Checks["agent"].AddressValid {
			t.Fatal("raw/redacted check state incorrect")
		}
		raw, _ := json.Marshal(v)
		for _, secret := range []string{"synthetic-model-secret", "synthetic-agent-secret", "hidden-url-secret"} {
			if strings.Contains(string(raw), secret) {
				t.Fatal("AI read exposes secret")
			}
		}
	}
	after, afterRevision, err := c.ExtraJSON()
	if err != nil || string(before) != string(after) || revision != afterRevision {
		t.Fatal("check mutated configuration")
	}
	afterEntries, _ := os.ReadDir(backups)
	if len(entries) != len(afterEntries) {
		t.Fatal("check created a backup")
	}
}

func TestAIConfigChecksAuthenticatedAPIAndSaveSnapshot(t *testing.T) {
	s := newTestServer(t)
	store, _, _, _ := aiFixture(t)
	s.cfg.AI = store
	expectStatus(t, request(s, "GET", "/api/ai", nil, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	read := func() AIView {
		t.Helper()
		rsp := request(s, "GET", "/api/ai", nil, cookie, "")
		expectStatus(t, rsp, 200)
		var v AIView
		if err := json.Unmarshal(rsp.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	v := read()
	if v.Global.Checks["chat"].Complete {
		t.Fatal("unconfigured key assumed ready")
	}
	update := AIUpdate{Revision: v.Global.Revision, Fields: aiFieldsRaw(map[string]any{"Type": 1, "ModelName": "local-model", "API": "http://fixture.invalid/chat", "AgentType": 1, "AgentModelName": "agent", "AgentAPI": "http://fixture.invalid/agent"})}
	rsp := request(s, "POST", "/api/ai", update, cookie, csrf)
	expectStatus(t, rsp, 200)
	var saved AIView
	if err := json.Unmarshal(rsp.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if !saved.Global.Checks["chat"].Complete || !saved.Global.Checks["agent"].Complete || saved.Global.Checks["image"].Complete || !reflect.DeepEqual(saved.Global.Checks, read().Global.Checks) {
		t.Fatal("saved and read check snapshots differ")
	}
	expectStatus(t, request(s, "POST", "/api/ai", update, cookie, csrf), 409)
	if !reflect.DeepEqual(saved.Global.Checks, read().Global.Checks) {
		t.Fatal("conflict changed checks")
	}
	s.cfg.AI = nil
	if v := read(); v.Available || v.Global != nil {
		t.Fatal("missing store inferred configured")
	}
}
