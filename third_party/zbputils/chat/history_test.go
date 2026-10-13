package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	goba "github.com/fumiama/go-onebot-agent"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func historyConfig() (AgentConfig, error) {
	return AgentConfig{Key: "synthetic-old-key", AgentKey: "synthetic-agent-key"}, nil
}
func historyEvent(group int64, text string) *goba.Event {
	msg, _ := json.Marshal([]map[string]any{{"type": "text", "data": map[string]string{"text": text}}})
	return &goba.Event{PostType: "message", MessageType: "group", GroupID: group, UserID: 123, SelfID: 456, MessageID: 1, Message: msg}
}
func historyBody(t *testing.T, s *agentHistory, group int64) string {
	t.Helper()
	p := model.NewOpenAI("test", "", .7, 1, 512, "")
	m, err := s.Modelize(p, group, "persona", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	return m.Body().String()
}

func TestAgentHistoryRetentionRestartAndIsolation(t *testing.T) {
	root := t.TempDir()
	s, err := openAgentHistory(root, 456, 900000, historyConfig)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		if err = s.Append(10, "event", historyEvent(10, fmt.Sprintf("record-%03d-中文🙂", i))); err != nil {
			t.Fatal(err)
		}
	}
	long := strings.Repeat("完整文本🙂", 600)
	if err = s.Append(10, "event", historyEvent(10, long)); err != nil {
		t.Fatal(err)
	}
	if err = s.Append(20, "event", historyEvent(20, "other-group-only")); err != nil {
		t.Fatal(err)
	}
	if err = s.Append(-123, "event", historyEvent(0, "private-only")); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = openAgentHistory(root, 456, 900000, historyConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	body := historyBody(t, s, 10)
	for _, want := range []string{"record-000", "record-099", long} {
		if !strings.Contains(body, want) {
			t.Fatalf("reopened history missing %s", want[:10])
		}
	}
	for _, wrong := range []string{"other-group-only", "private-only"} {
		if strings.Contains(body, wrong) {
			t.Fatal("conversation leaked")
		}
	}
	other, err := openAgentHistory(root, 999, 900000, historyConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if strings.Contains(historyBody(t, other, 10), "record-") {
		t.Fatal("bot account leaked")
	}
	if err = s.reset(); err != nil {
		t.Fatal(err)
	}
	var retained int
	if err = s.db.QueryRow("SELECT COUNT(*) FROM history WHERE kind='event'").Scan(&retained); err != nil || retained != 103 {
		t.Fatal("reset deleted archived records")
	}
	if err = s.Append(10, "event", historyEvent(10, "fresh-context")); err != nil {
		t.Fatal(err)
	}
	body = historyBody(t, s, 10)
	if strings.Contains(body, "record-000") || !strings.Contains(body, "fresh-context") {
		t.Fatal("reset boundary ineffective")
	}
}

func TestAgentHistoryRedactionAndContextBudget(t *testing.T) {
	s, err := openAgentHistory(t.TempDir(), 456, 1024, historyConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, text := range []string{"synthetic-old-key", "设置AI聊天Agent密钥synthetic-new-key"} {
		if err = s.Append(10, "event", historyEvent(10, text)); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := s.db.Query("SELECT payload FROM history")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var raw string
		if rows.Scan(&raw) != nil {
			t.Fatal("read failed")
		}
		if strings.Contains(raw, "synthetic-") {
			t.Fatal("key persisted")
		}
	}
	rows.Close()
	text := "first-marker" + strings.Repeat("超长🙂", 5000) + "last-marker"
	if err = s.Append(10, "event", historyEvent(10, text)); err != nil {
		t.Fatal(err)
	}
	items, clipped, err := s.recent(10)
	if err != nil || !clipped || len(items) != 1 || !utf8.ValidString(items[0].payload) || !json.Valid([]byte(items[0].payload)) {
		t.Fatal("invalid bounded excerpt")
	}
	if len(items[0].payload) > 1024 || !strings.Contains(items[0].payload, "first-marker") || !strings.Contains(items[0].payload, "last-marker") {
		t.Fatal("latest event excerpt missing")
	}
	var raw string
	if s.db.QueryRow("SELECT payload FROM history ORDER BY seq DESC LIMIT 1").Scan(&raw) != nil || !strings.Contains(raw, text) {
		t.Fatal("model clipping changed full stored record")
	}
}

func TestAgentHistoryWriteFailureStopsRequest(t *testing.T) {
	s, err := openAgentHistory(t.TempDir(), 456, 900000, historyConfig)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	ag := goba.NewAgent(&goba.Config{}, 16, 8, time.Hour, "", testMemory{}, true, true)
	ag.SetHistoryStore(s)
	if ag.AddEvent(10, historyEvent(10, "not-saved")) == nil {
		t.Fatal("write failure hidden")
	}
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer server.Close()
	api := deepinfra.NewAPI(server.URL, "fake")
	_, err = ag.GetAction(api, model.NewOpenAI("test", "", .7, 1, 512, ""), 10, goba.PermRoleUser, 1, false)
	if err == nil || called {
		t.Fatal("failed record reached model")
	}
}

func TestAgentHistoryPrivatePathRejectsNonFile(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "456.sqlite"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := openAgentHistory(root, 456, 900000, historyConfig); err == nil {
		t.Fatal("directory accepted as private history file")
	}
	if _, err := openAgentHistory(root, -1, 900000, historyConfig); err == nil {
		t.Fatal("invalid bot accepted")
	}
}

type testMemory struct{}

func (testMemory) Load(int64) []string      { return nil }
func (testMemory) Save(int64, string) error { return nil }

func TestAgentEndPreservesActionsAndReconstructedHistory(t *testing.T) {
	s, err := openAgentHistory(t.TempDir(), 456, 900000, historyConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ag := goba.NewAgent(&goba.Config{}, 16, 8, time.Hour, "", testMemory{}, true, true)
	ag.SetHistoryStore(s)
	if ag.AddEvent(-123, historyEvent(0, "remember-marker")) != nil {
		t.Fatal("append failed")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(raw), "remember-marker") {
			t.Error("persistent event missing from request")
		}
		w.Header().Set("Content-Type", "application/json")
		content := `{"action":"send_private_msg","params":{"user_id":123,"message":"reply"}}` + "\n" + `{"action":"end_action","params":{}}`
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": content}}}})
	}))
	defer server.Close()
	api := deepinfra.NewAPI(server.URL, "fake")
	p := model.NewOpenAI("test", "", .7, 1, 512, "")
	reqs, ended, err := NextAgentActions(&ag, true, 1, api, p, -123, goba.PermRoleOwner)
	if err != nil || !ended || len(reqs) != 1 || reqs[0].Action != "send_private_msg" {
		t.Fatal("valid action before end was lost")
	}
	if ag.AddRequest(-123, &reqs[0]) != nil || ag.AddResponse(-123, &goba.APIResponse{Status: "ok", Data: json.RawMessage(`{"message_id":9}`)}) != nil {
		t.Fatal("response append failed")
	}
	body := historyBody(t, s, -123)
	if !strings.Contains(body, "send_private_msg") || !strings.Contains(body, "message_id") {
		t.Fatal("action result not restored")
	}
}

func TestAgentRunPastEightRoundsAndStopCycles(t *testing.T) {
	nextCalls, executed := 0, 0
	err := RunAgent(func(iter int) ([]zero.APIRequest, bool, error) {
		nextCalls++
		if iter == 13 {
			return nil, true, nil
		}
		return []zero.APIRequest{{Action: "send_group_msg", Params: zero.Params{"group_id": 10, "message": fmt.Sprint(iter)}}}, false, nil
	}, func(zero.APIRequest) error { executed++; return nil }, func() bool { return true })
	if err != nil || nextCalls != 13 || executed != 12 {
		t.Fatal("task was capped at eight rounds")
	}
	nextCalls, executed = 0, 0
	err = RunAgent(func(iter int) ([]zero.APIRequest, bool, error) {
		nextCalls++
		return []zero.APIRequest{{Action: "send_group_msg", Params: zero.Params{"message": iter % 2}, Echo: uint64(iter)}}, false, nil
	}, func(zero.APIRequest) error { executed++; return nil }, func() bool { return true })
	if !errors.Is(err, ErrAgentRepeatedAction) || nextCalls != 3 || executed != 2 {
		t.Fatal("action cycle did not stop")
	}
}

func TestAgentRunReloadAndExecutionFailure(t *testing.T) {
	current := true
	executed := 0
	err := RunAgent(func(int) ([]zero.APIRequest, bool, error) {
		current = false
		return []zero.APIRequest{{Action: "send_group_msg"}}, false, nil
	}, func(zero.APIRequest) error { executed++; return nil }, func() bool { return current })
	if !errors.Is(err, ErrAgentReloaded) || executed != 0 {
		t.Fatal("reloaded task executed")
	}
	failure := errors.New("synthetic write failure")
	err = RunAgent(func(int) ([]zero.APIRequest, bool, error) {
		return []zero.APIRequest{{Action: "send_group_msg"}}, false, nil
	}, func(zero.APIRequest) error { return failure }, func() bool { return true })
	if !errors.Is(err, failure) {
		t.Fatal("execution failure ignored")
	}
}
