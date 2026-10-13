//go:build ignore

// Authorized real provider calls with synthetic prompts. No OneBot connection
// or action executor is created; credentials remain in the existing database.
package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/chat"
	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	goba "github.com/fumiama/go-onebot-agent"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

type memory struct{}

func (memory) Load(int64) []string      { return nil }
func (memory) Save(int64, string) error { return errors.New("memory writes disabled in QA") }

type capture struct {
	calls    []map[string]any
	lastText string
}

func (c *capture) RoundTrip(req *http.Request) (*http.Response, error) {
	started := time.Now()
	restoredResponse := false
	if req.GetBody != nil {
		if body, err := req.GetBody(); err == nil {
			var payload struct {
				Messages []struct {
					Content json.RawMessage `json:"content"`
				} `json:"messages"`
			}
			if json.NewDecoder(body).Decode(&payload) == nil {
				for _, msg := range payload.Messages {
					var text string
					if json.Unmarshal(msg.Content, &text) == nil && strings.Contains(text, `"status":"ok"`) && strings.Contains(text, `"message_id":2`) {
						restoredResponse = true
					}
				}
			}
			body.Close()
		}
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	row := map[string]any{"headersElapsedMS": time.Since(started).Milliseconds(), "historyResponseInPrompt": restoredResponse}
	if err != nil {
		row["transportFailed"] = true
		c.calls = append(c.calls, row)
		return nil, err
	}
	row["status"] = resp.StatusCode
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	if readErr == nil && resp.StatusCode == 200 {
		var v struct {
			Model   string `json:"model"`
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Usage struct {
				Prompt     int `json:"prompt_tokens"`
				Completion int `json:"completion_tokens"`
				Total      int `json:"total_tokens"`
			} `json:"usage"`
		}
		if json.Unmarshal(body, &v) == nil {
			if len(v.Choices) > 0 {
				c.lastText = v.Choices[0].Message.Content
			}
			row["model"] = v.Model
			row["promptTokens"] = v.Usage.Prompt
			row["completionTokens"] = v.Usage.Completion
			row["totalTokens"] = v.Usage.Total
		}
	}
	c.calls = append(c.calls, row)
	return resp, readErr
}
func main() {
	configured := flag.Bool("configured", false, "Verify saved settings instead of the proposed setup")
	persistent := flag.Bool("persistent", false, "Exercise private history across a new Agent instance and SQLite reopen")
	flag.Parse()
	logrus.SetOutput(io.Discard)
	db, err := sql.Open("sqlite3", "file:data/control/plugins.db?mode=ro")
	if err != nil {
		fail("database unavailable")
	}
	defer db.Close()
	h := md5.Sum([]byte("aichat"))
	i := byte('a') % 15
	gid := int16(binary.LittleEndian.Uint16(h[i : i+2]))
	var raw string
	if db.QueryRow("SELECT ext FROM __resp WHERE gid = ?", gid).Scan(&raw) != nil {
		fail("AI settings unavailable")
	}
	var cfg chat.AgentConfig
	if json.Unmarshal([]byte(raw), &cfg) != nil {
		fail("AI settings invalid")
	}
	if !*configured {
		cfg.API = "https://api.deepseek.com/chat/completions"
		cfg.AgentAPI = cfg.API
		cfg.Type = 0
		cfg.AgentType = 0
		cfg.ModelName = "deepseek-flash"
		cfg.AgentModelName = cfg.ModelName
		cfg.ReasoningEffort = "none"
	}
	if cfg.API != "https://api.deepseek.com/chat/completions" || cfg.AgentAPI != cfg.API || !cfg.Ready() || cfg.AgentKey == "" || cfg.Type != 0 || cfg.AgentType != 0 || cfg.ModelName != "deepseek-flash" || cfg.AgentModelName != cfg.ModelName || cfg.ReasoningEffort != "none" {
		fail("expected official DeepSeek settings required")
	}
	transport := &capture{}
	client := &http.Client{Timeout: 30 * time.Second, Transport: transport, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	report := map[string]any{"configuredSettings": *configured, "chatModel": cfg.ModelName, "agentModel": cfg.AgentModelName, "thinking": "disabled", "realQQActionsExecuted": 0}
	p, err := cfg.Type.ProtocolForAPI(cfg.API, cfg.ModelName, .7, 1, 512, cfg.ReasoningEffort)
	if err != nil {
		fail("chat protocol invalid")
	}
	api := deepinfra.NewAPI(cfg.API, string(cfg.Key))
	api.SetHTTPClient(client)
	started := time.Now()
	answer, err := api.Request(p.System("This is a synthetic connection test. Follow the user's requested output exactly.").User(model.NewContentText("Reply with exactly DeepSeekChatOK")))
	report["chatRequestMS"] = time.Since(started).Milliseconds()
	report["chatPassed"] = err == nil && strings.TrimSpace(answer) == "DeepSeekChatOK"
	if err == nil {
		p, err = cfg.AgentType.ProtocolForAPI(cfg.AgentAPI, cfg.AgentModelName, .7, 1, 512, cfg.ReasoningEffort)
		if err != nil {
			fail("agent protocol invalid")
		}
		ag := goba.NewAgent(&goba.Config{Nickname: "QA", Sex: "unknown", Chars: "You are testing only. For private messages use send_private_msg with user_id from the event; for group messages use send_group_msg with group_id from the event. Reply with DeepSeekAgentOK, then end_action. Never save memories or perform management actions."}, 16, 8, time.Hour, "", memory{}, true, true)
		character := goba.Config{Nickname: "QA", Sex: cfg.AgentSex, Chars: cfg.AgentChar + "\n本次是隔离测试：只给当前虚拟私聊回复 DeepSeekAgentOK，响应成功后 end_action，不执行其他动作。"}
		var historyRoot string
		var closeHistory func() error
		if *persistent {
			if cfg.AgentChar == "" {
				fail("saved persona required for persistent test")
			}
			if os.MkdirAll(".local/agent-history-tests", 0700) != nil {
				fail("history fixture root unavailable")
			}
			historyRoot, err = os.MkdirTemp(".local/agent-history-tests", "run-")
			if err != nil {
				fail("history fixture unavailable")
			}
			ag = goba.NewAgent(&character, 16, 8, time.Hour, "", memory{}, true, true)
			history, err := chat.OpenAgentHistory(historyRoot, 90000, func() (chat.AgentConfig, error) { return cfg, nil })
			if err != nil {
				fail("history store unavailable")
			}
			defer history.Close()
			closeHistory = history.Close
			ag.SetHistoryStore(history)
		}
		if ag.AddEvent(-90001, &goba.Event{Time: time.Now().Unix(), PostType: "message", MessageType: "private", MessageID: 1, UserID: 90001, SelfID: 90000, Message: json.RawMessage(`[{"type":"text","data":{"text":"这是私聊，只使用 send_private_msg，user_id=90001，回复 DeepSeekAgentOK，然后 end_action。"}}]`)}) != nil {
			fail("synthetic event save failed")
		}
		api = deepinfra.NewAPI(cfg.AgentAPI, string(cfg.AgentKey))
		api.SetHTTPClient(client)
		report["agentTestRole"] = "owner"
		started = time.Now()
		reqs, callErr := ag.GetAction(api, p, -90001, goba.PermRoleOwner, 1, false)
		report["agentRound1MS"] = time.Since(started).Milliseconds()
		report["agentReturnedActions"] = len(reqs)
		report["agentPermissionError"] = errors.Is(callErr, goba.ErrPermissionDenied)
		report["agentEndedFirstRound"] = errors.Is(callErr, io.EOF)
		report["agentOtherError"] = callErr != nil && !errors.Is(callErr, io.EOF) && !errors.Is(callErr, goba.ErrPermissionDenied)
		if len(transport.lastText) < 500 {
			report["syntheticAgentReply"] = cfg.Redact(transport.lastText)
		}
		passed := callErr == nil || errors.Is(callErr, io.EOF)
		replyCount := 0
		for _, r := range reqs {
			report["agentAction"] = r.Action
			report["agentTargetMatched"] = fmt.Sprint(r.Params["user_id"]) == "90001"
			if r.Action != "send_private_msg" || fmt.Sprint(r.Params["user_id"]) != "90001" {
				passed = false
				continue
			}
			msg, _ := json.Marshal(r.Params["message"])
			if !bytes.Contains(msg, []byte("DeepSeekAgentOK")) {
				passed = false
			}
			replyCount++
			ag.AddRequest(-90001, &r)
			ag.AddResponse(-90001, &goba.APIResponse{Status: "ok", Data: json.RawMessage(`{"message_id":2}`)})
		}
		report["agentActionParsingPassed"] = passed && replyCount == 1
		if passed && replyCount == 1 {
			if *persistent {
				if closeHistory() != nil {
					fail("history close before reopen failed")
				}
				ag = goba.NewAgent(&character, 16, 8, time.Hour, "", memory{}, true, true)
				history, err := chat.OpenAgentHistory(historyRoot, 90000, func() (chat.AgentConfig, error) { return cfg, nil })
				if err != nil {
					fail("history reopen failed")
				}
				defer history.Close()
				ag.SetHistoryStore(history)
			}
			started = time.Now()
			next, nextErr := ag.GetAction(api, p.Clone(), -90001, goba.PermRoleOwner, 2, false)
			report["agentRound2MS"] = time.Since(started).Milliseconds()
			report["agentEndActionPassed"] = errors.Is(nextErr, io.EOF) && len(next) == 0
			if *persistent {
				report["historyRestoredInModelRequest"] = len(transport.calls) > 0 && transport.calls[len(transport.calls)-1]["historyResponseInPrompt"] == true
				historyDB, err := sql.Open("sqlite3", "file:"+filepath.ToSlash(filepath.Join(historyRoot, "90000.sqlite"))+"?mode=ro")
				if err != nil {
					fail("history report unavailable")
				}
				var count int
				readErr := historyDB.QueryRow("SELECT COUNT(*) FROM history WHERE conversation=-90001").Scan(&count)
				historyDB.Close()
				report["archivedRecords"] = count
				report["persistentHistoryPassed"] = readErr == nil && count == 4 && report["historyRestoredInModelRequest"] == true
				report["historyFixture"] = filepath.ToSlash(historyRoot)
			}
		}
	}
	report["calls"] = transport.calls
	report["completedAt"] = time.Now().Format(time.RFC3339)
	passed := report["chatPassed"] == true && report["agentActionParsingPassed"] == true && report["agentEndActionPassed"] == true
	if *persistent {
		passed = passed && report["persistentHistoryPassed"] == true
	}
	report["passed"] = passed
	body, _ := json.MarshalIndent(report, "", "  ")
	_ = os.MkdirAll(".local", 0700)
	reportFile := ".local/deepseek-test.json"
	if *persistent {
		reportFile = ".local/agent-history-test.json"
	}
	if os.WriteFile(reportFile, body, 0600) != nil {
		fail("QA report could not be saved")
	}
	fmt.Println(string(body))
	if !passed {
		os.Exit(1)
	}
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
