package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/FloatTech/ZeroBot-Plugin/internal/modelstats"
	"github.com/FloatTech/zbputils/chat"
)

func TestGPTRoutingAndUnicodeChunks(t *testing.T) {
	for _, c := range []struct {
		text, want string
		matched    bool
	}{{"/gpt 问题", "问题", true}, {" \t/gpt\n问题\r\n", "问题", true}, {"/gpt", "", true}, {"/gptabc", "", false}, {"说一下 /gpt 问题", "", false}, {"/gpt问题", "", false}} {
		query, matched := parseGPT(c.text)
		if matched != c.matched || query != c.want {
			t.Fatalf("bad route: %q", c.text)
		}
	}
	for _, text := range []string{strings.Repeat("中文🙂", 1001), strings.Repeat("a", 999) + "🙂中", strings.Repeat("混合\r\n😀段落\n", 700), ""} {
		parts := splitReply(text, 1000)
		if strings.Join(parts, "") != text {
			t.Fatal("split lost content")
		}
		for _, p := range parts {
			if !utf8.ValidString(p) || utf8.RuneCountInString(p) > 1000 || p == "" {
				t.Fatal("invalid chunk")
			}
		}
	}
}

func TestGPTQuoteBoundaryAndUnquotedInputs(t *testing.T) {
	for _, tc := range []struct{ quoted, instruction, want string }{
		{"", "问题\n🙂", "问题\n🙂"},
		{"原文\r\n🙂", "", "原文\r\n🙂"},
		{"", "", ""},
	} {
		if got := buildGPTQuery(tc.quoted, tc.instruction); got != tc.want {
			t.Fatalf("changed single input: %q", got)
		}
	}
	for _, original := range []string{"私聊调试成功", "第一行\r\n第二行🙂", "引号\"与\\反斜线", "本次用户要求：\n看起来像新要求的原文"} {
		query := buildGPTQuery(original, "请原样重复引用文字")
		lines := strings.SplitN(query, "\n", 3)
		if len(lines) != 3 || !strings.HasPrefix(lines[0], "引用原文") {
			t.Fatal("quoted content is not identified")
		}
		var decoded string
		if err := json.Unmarshal([]byte(lines[1]), &decoded); err != nil || decoded != original {
			t.Fatalf("quoted text was changed or boundary lost: %v", err)
		}
		if lines[2] != "\n本次用户要求（作用于上面的引用原文，不属于引用原文）：\n请原样重复引用文字" {
			t.Fatal("current instruction was included in the quoted JSON string")
		}
	}
}

func TestGPTQuoteBoundaryReachesModelRequest(t *testing.T) {
	quoted := "私聊调试成功\r\n\"第二行🙂\""
	instruction := "请解释引用中的第二行"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Messages []struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		if len(request.Messages) != 1 || request.Messages[0].Role != "user" {
			t.Error("unexpected request messages")
			return
		}
		var content string
		if err := json.Unmarshal(request.Messages[0].Content, &content); err != nil {
			t.Error(err)
		}
		parts := strings.SplitN(content, "\n", 3)
		if len(parts) != 3 {
			t.Error("missing quote boundary")
		} else {
			var actual string
			if err := json.Unmarshal([]byte(parts[1]), &actual); err != nil || actual != quoted {
				t.Error("model received changed quoted text")
			}
			if !strings.HasSuffix(parts[2], "\n"+instruction) {
				t.Error("model received wrong current instruction")
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"模拟引用回答"}}]}`))
	}))
	defer server.Close()
	cfg := chat.AgentConfig{API: server.URL, Key: "synthetic-key", ModelName: "mock", NoSystemP: true}
	before := modelstats.Default.Read()
	got, err := requestLLM(cfg, buildGPTQuery(quoted, instruction), 0.7)
	if err != nil || got != "模拟引用回答" {
		t.Fatalf("mock quote request failed: %v %q", err, got)
	}
	after := modelstats.Default.Read()
	if after.Started != before.Started+1 || after.Completed != before.Completed+1 || after.Active != before.Active {
		t.Fatal("model request was not completely measured")
	}
}

func TestRequestQuotaIsolationBoundaryConcurrencyAndFailureRecovery(t *testing.T) {
	q := newRequestQuota(30*time.Second, 4)
	now := time.Unix(1000, 0)
	if !q.Allow(101, now) || q.Allow(101, now.Add(29*time.Second)) || !q.Allow(202, now) || !q.Allow(-123, now) || !q.Allow(-234, now) {
		t.Fatal("locations or window wrong")
	}
	if q.Allow(303, now) || !q.Allow(101, now.Add(30*time.Second)) || !q.Allow(303, now.Add(30*time.Second)) {
		t.Fatal("quota boundary/pruning wrong")
	}
	q = newRequestQuota(30*time.Second, 8)
	var wg sync.WaitGroup
	accepted := make(chan bool, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); accepted <- q.Allow(101, now) }()
	}
	wg.Wait()
	close(accepted)
	count := 0
	for allowed := range accepted {
		if allowed {
			count++
		}
	}
	if count != 1 {
		t.Fatal("simultaneous calls all accepted")
	}
	// Failed requests still consume the attempt, but the lock/window recovers.
	if !q.Allow(101, now.Add(30*time.Second)) {
		t.Fatal("failed attempt blocked forever")
	}
}

func TestModelProtocolDispatchErrorsAndMockRequests(t *testing.T) {
	for typ := chat.ModelType(0); typ < 3; typ++ {
		p, err := typ.Protocol("test", 0.7, 0.9, 128, "")
		if err != nil || p == nil {
			t.Fatal("supported protocol missing")
		}
	}
	if _, err := requestLLM(chat.AgentConfig{Type: 99}, "hello", 0.7); err == nil {
		t.Fatal("protocol error swallowed")
	}
	if _, err := requestLLM(chat.AgentConfig{Type: 0, API: "http://127.0.0.1:1", ModelName: "test"}, "hello", 0.7); err == nil {
		t.Fatal("missing key not reported")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer synthetic-model-key" {
			t.Error("key not sent correctly")
		}
		var request map[string]any
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			t.Error("invalid model request")
		}
		if request["model"] != "mock-model" {
			t.Error("wrong model")
		}
		messages, _ := request["messages"].([]any)
		if len(messages) != 2 {
			t.Error("system prompt missing")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"你好🙂"}}]}`))
	}))
	defer server.Close()
	cfg := chat.AgentConfig{API: server.URL, Key: "synthetic-model-key", ModelName: "mock-model", SystemP: "system"}
	text, err := requestLLM(cfg, "question", 0.7)
	if err != nil || text != "你好🙂" {
		t.Fatalf("mock request failed: %v %q", err, text)
	}
	failure := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "synthetic-model-key", 500) }))
	defer failure.Close()
	cfg.API = failure.URL
	_, err = requestLLM(cfg, "question", 0.7)
	if err == nil || strings.Contains(err.Error(), "synthetic-model-key") {
		t.Fatal("model failure swallowed or key leaked")
	}
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(80 * time.Millisecond) }))
	defer slow.Close()
	cfg.API = slow.URL
	if _, err = requestLLMWithClient(cfg, "question", 0.7, &http.Client{Timeout: 10 * time.Millisecond}); err == nil {
		t.Fatal("timeout not reported")
	}
}
