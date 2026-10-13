//go:build ignore

// Disposable model/OneBot fixture for real browser observation report checks.
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	qa "github.com/FloatTech/ZeroBot-Plugin/internal/adminqa"
)

func main() {
	binary := flag.String("binary-name", "zerobot.exe", "filename under .local/bin")
	lifetime := flag.Duration("lifetime", 20*time.Minute, "maximum lifetime, 1m to 2h")
	modelDelay := flag.Duration("model-delay", 0, "synthetic model delay, 0 to 2s")
	flag.Parse()
	if flag.NArg() != 0 || filepath.Base(*binary) != *binary || strings.ContainsAny(*binary, "/\\:") || filepath.Ext(*binary) != ".exe" || *lifetime < time.Minute || *lifetime > 2*time.Hour || *modelDelay < 0 || *modelDelay > 2*time.Second {
		fmt.Fprintln(os.Stderr, "Invalid arguments")
		os.Exit(2)
	}
	if err := run(*binary, *lifetime, *modelDelay); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(binary string, lifetime, modelDelay time.Duration) (err error) {
	root, err := os.Getwd()
	if err != nil {
		return
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return
	}
	parent := filepath.Join(root, ".local/model-observation-acceptance")
	if err = os.MkdirAll(parent, 0700); err != nil {
		return
	}
	out, err := os.MkdirTemp(parent, "run-")
	if err != nil {
		return
	}
	input, err := os.Open(filepath.Join(root, ".local/bin", binary))
	if err != nil {
		return
	}
	copied := filepath.Join(out, "zerobot-model-observation-acceptance.exe")
	target, err := os.OpenFile(copied, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		input.Close()
		return
	}
	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(target, hash), input)
	targetClose := target.Close()
	input.Close()
	if err != nil {
		return
	}
	if targetClose != nil {
		return targetClose
	}
	signalCtx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignal()
	ctx, cancel := context.WithTimeout(signalCtx, lifetime)
	defer cancel()
	var h *qa.Harness
	var bot *qa.Bot
	var modelCalls atomic.Int32
	starts, batches := 0, 0
	result := map[string]any{"startedAt": time.Now().Format(time.RFC3339Nano), "binarySHA256": hex.EncodeToString(hash.Sum(nil)), "isolated": true, "realQQMessagesSent": false, "realModelCalls": 0, "browserAcceptancePassed": false, "passed": false}
	defer func() {
		if bot != nil {
			bot.Stop()
		}
		if h != nil {
			h.Close()
		}
		result["processStarts"], result["requestBatches"], result["syntheticModelCalls"] = starts, batches, modelCalls.Load()
		result["instanceStopped"], result["stoppedAt"] = true, time.Now().Format(time.RFC3339Nano)
		if err != nil {
			result["failure"] = err.Error()
		}
		if saveErr := qa.WriteJSON(filepath.Join(out, "result.json"), result); saveErr != nil && err == nil {
			err = saveErr
		}
		_ = qa.WriteJSON(filepath.Join(out, "session.json"), map[string]any{"state": "stopped"})
		fmt.Println("ARTIFACTS=" + out)
	}()
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer synthetic-observation-key" {
			http.Error(w, "denied", 403)
			return
		}
		var body struct {
			Model string `json:"model"`
		}
		if json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body) != nil || body.Model != "synthetic-observation-model" {
			http.Error(w, "invalid", 400)
			return
		}
		modelCalls.Add(1)
		if modelDelay > 0 {
			timer := time.NewTimer(modelDelay)
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-r.Context().Done():
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"模型观察模拟回复"}}]}`)
	}))
	defer model.Close()
	h, err = qa.New(root, copied, out)
	if err != nil {
		return
	}
	secret := make([]byte, 24)
	if _, err = rand.Read(secret); err != nil {
		return
	}
	password := "QA-" + hex.EncodeToString(secret)
	if err = os.WriteFile(filepath.Join(out, "login.txt"), []byte("Isolated observation password:\n"+password+"\n"), 0600); err != nil {
		return
	}
	client := qa.NewClient(h.Address)
	guard := func() error {
		raw, e := os.ReadFile(h.ConfigPath)
		if e != nil {
			return e
		}
		var cfg struct {
			WS []struct {
				URL   string `json:"Url"`
				Token string `json:"AccessToken"`
			} `json:"ws"`
			WSS []json.RawMessage `json:"wss"`
		}
		if json.Unmarshal(raw, &cfg) != nil || len(cfg.WS) != 1 || len(cfg.WSS) != 0 || cfg.WS[0].URL != h.MockURL() || cfg.WS[0].Token != "isolated-lifecycle-token" {
			return errors.New("fixture requires its original synthetic OneBot endpoint")
		}
		return nil
	}
	start := func() error {
		if e := guard(); e != nil {
			return e
		}
		var e error
		bot, e = h.Start(ctx, fmt.Sprintf("start-%d", starts+1))
		if e != nil {
			return e
		}
		starts++
		_, e = client.AuthWithPassword(h.AdminDir, password)
		return e
	}
	if err = start(); err != nil {
		return
	}
	value, err := client.Checked("GET", "/api/ai", nil)
	if err != nil {
		return
	}
	_, err = client.Checked("POST", "/api/ai", map[string]any{"revision": value["global"].(map[string]any)["revision"], "fields": map[string]any{"API": model.URL, "ModelName": "synthetic-observation-model", "Type": 0, "NoSystemP": true}, "keys": map[string]any{"Key": map[string]any{"action": "replace", "value": "synthetic-observation-key"}}})
	if err != nil {
		return
	}
	credential, err := os.ReadFile(filepath.Join(h.AdminDir, "credential.json"))
	if err != nil {
		return
	}
	credentialHash := sha256.Sum256(credential)
	status := func() error {
		snapshot, e := client.Checked("GET", "/api/overview", nil)
		if e != nil {
			return e
		}
		s := map[string]any{"state": "ready", "url": "http://" + h.Address + "/", "output": out, "loginFile": filepath.Join(out, "login.txt"), "processStarts": starts, "requestBatches": batches, "syntheticModelCalls": modelCalls.Load(), "modelRequests": snapshot["modelRequests"], "replyTasks": snapshot["replyTasks"], "syntheticModelDelayMilliseconds": modelDelay.Milliseconds(), "realQQMessagesSent": false, "realModelCalls": 0}
		if e = qa.WriteJSON(filepath.Join(out, "session.json"), s); e != nil {
			return e
		}
		encoded, _ := json.Marshal(s)
		fmt.Println(string(encoded))
		return nil
	}
	if err = status(); err != nil {
		return
	}
	fmt.Println("Commands: status, requests (3 synthetic GPT replies), restart, stop")
	commands := make(chan string)
	done := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case commands <- strings.TrimSpace(scanner.Text()):
			case <-ctx.Done():
				return
			}
		}
		done <- scanner.Err()
	}()
	for {
		select {
		case <-ctx.Done():
			result["stopReason"] = "signal-or-time-limit"
			return nil
		case err = <-done:
			result["stopReason"] = "input-ended"
			return
		case command := <-commands:
			switch command {
			case "status":
				if err = status(); err != nil {
					return
				}
			case "requests":
				if err = guard(); err != nil {
					return
				}
				value, err = client.Checked("GET", "/api/ai", nil)
				if err != nil {
					return
				}
				fields := value["global"].(map[string]any)["fields"].(map[string]any)
				if fields["API"] != model.URL || fields["ModelName"] != "synthetic-observation-model" || fields["Type"] != float64(0) {
					return errors.New("model calls require this run's original synthetic endpoint")
				}
				if batches >= 20 {
					return errors.New("synthetic request batch limit reached")
				}
				before := modelCalls.Load()
				for i := 0; i < 3; i++ {
					reply, _, e := h.Send(23456789+int64(batches*3+i), "/gpt 模型观察测试")
					if e != nil || reply != "模型观察模拟回复" {
						return errors.New("synthetic GPT roundtrip failed")
					}
				}
				if modelCalls.Load() != before+3 {
					return errors.New("synthetic model request count differs")
				}
				batches++
				result["syntheticRoundtripPassed"] = true
				if err = status(); err != nil {
					return
				}
			case "restart":
				if err = guard(); err != nil {
					return
				}
				bot.Stop()
				bot = nil
				if err = start(); err != nil {
					return
				}
				nowCredential, e := os.ReadFile(filepath.Join(h.AdminDir, "credential.json"))
				if e != nil || sha256.Sum256(nowCredential) != credentialHash {
					return errors.New("fixture credential changed")
				}
				result["credentialPreserved"] = true
				if err = status(); err != nil {
					return
				}
			case "stop":
				result["stopReason"], result["passed"] = "stop-command", true
				return nil
			default:
				fmt.Println("Commands: status, requests, restart, stop")
			}
		}
	}
}
