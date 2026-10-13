//go:build ignore

// A disposable two-account fixture for real-browser connection acceptance.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	qa "github.com/FloatTech/ZeroBot-Plugin/internal/adminqa"
)

func main() {
	binary := flag.String("binary-name", "zerobot.exe", "filename under .local/bin")
	lifetime := flag.Duration("lifetime", time.Hour, "maximum lifetime, 1m to 2h")
	flag.Parse()
	if flag.NArg() != 0 || filepath.Base(*binary) != *binary || strings.ContainsAny(*binary, "/\\:") || filepath.Ext(*binary) != ".exe" || *lifetime < time.Minute || *lifetime > 2*time.Hour {
		fmt.Fprintln(os.Stderr, "Invalid binary or lifetime")
		os.Exit(2)
	}
	if err := run(*binary, *lifetime); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(binary string, lifetime time.Duration) (err error) {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	parent := filepath.Join(root, ".local/connection-browser-acceptance")
	if err = os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	out, err := os.MkdirTemp(parent, "run-")
	if err != nil {
		return err
	}
	source, err := os.Open(filepath.Join(root, ".local/bin", binary))
	if err != nil {
		return err
	}
	copied := filepath.Join(out, "zerobot-connection-acceptance.exe")
	target, err := os.OpenFile(copied, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		source.Close()
		return err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(target, hash), source)
	closeErr := target.Close()
	source.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	signalCtx, cancelSignal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelSignal()
	ctx, cancel := context.WithTimeout(signalCtx, lifetime)
	defer cancel()
	result := map[string]any{"startedAt": time.Now().Format(time.RFC3339Nano), "binarySHA256": hex.EncodeToString(hash.Sum(nil)), "isolated": true, "realQQMessagesSent": false, "modelCalls": 0, "browserAcceptancePassed": false, "passed": false}
	var primary, secondary *qa.Harness
	var bot *qa.Bot
	starts, checks := 0, 0
	defer func() {
		if bot != nil {
			bot.Stop()
		}
		if secondary != nil {
			secondary.Close()
		}
		if primary != nil {
			primary.Close()
		}
		result["processStarts"], result["transportChecks"] = starts, checks
		result["instanceStopped"], result["stoppedAt"] = true, time.Now().Format(time.RFC3339Nano)
		if err != nil {
			result["failure"] = err.Error()
		}
		if e := qa.WriteJSON(filepath.Join(out, "result.json"), result); e != nil && err == nil {
			err = e
		}
		_ = qa.WriteJSON(filepath.Join(out, "session.json"), map[string]any{"state": "stopped", "output": out})
		fmt.Println("ARTIFACTS=" + out)
	}()
	primary, err = qa.NewWithIdentity(root, copied, out, 87654321)
	if err != nil {
		return err
	}
	secondaryRoot := filepath.Join(out, "secondary-endpoint")
	if err = os.MkdirAll(secondaryRoot, 0700); err != nil {
		return err
	}
	secondary, err = qa.NewWithIdentity(root, copied, secondaryRoot, 87654322)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(primary.ConfigPath)
	if err != nil {
		return err
	}
	var cfg map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err = decoder.Decode(&cfg); err != nil {
		return err
	}
	cfg["ws"] = []map[string]any{{"Url": primary.MockURL(), "AccessToken": "isolated-lifecycle-token"}, {"Url": secondary.MockURL(), "AccessToken": "isolated-lifecycle-token"}}
	cfg["qa_unknown"] = map[string]any{"large": json.Number("9007199254740993"), "keep": "connection-fixture"}
	if err = qa.WriteJSON(primary.ConfigPath, cfg); err != nil {
		return err
	}
	secret := make([]byte, 24)
	if _, err = rand.Read(secret); err != nil {
		return err
	}
	password := "QA-" + hex.EncodeToString(secret)
	if err = os.WriteFile(filepath.Join(out, "login.txt"), []byte("Isolated connection password:\n"+password+"\n"), 0600); err != nil {
		return err
	}
	guard := func() error {
		raw, e := os.ReadFile(primary.ConfigPath)
		if e != nil {
			return e
		}
		var c struct {
			WS []struct {
				URL   string `json:"Url"`
				Token string `json:"AccessToken"`
			} `json:"ws"`
			WSS  []json.RawMessage `json:"wss"`
			Zero struct {
				Owners []int64 `json:"super_users"`
			} `json:"zero"`
			Unknown struct {
				Large json.Number `json:"large"`
				Keep  string      `json:"keep"`
			} `json:"qa_unknown"`
		}
		if json.Unmarshal(raw, &c) != nil || len(c.WS) != 2 || len(c.WSS) != 0 || c.Unknown.Large.String() != "9007199254740993" || c.Unknown.Keep != "connection-fixture" {
			return errors.New("restore the two synthetic connections and fixture marker before restart")
		}
		seen := map[string]bool{}
		for _, conn := range c.WS {
			base := strings.TrimSuffix(conn.URL, "/edited")
			if (base != primary.MockURL() && base != secondary.MockURL()) || seen[base] || conn.Token != "isolated-lifecycle-token" {
				return errors.New("restart permits only this run's distinct synthetic endpoints and original test tokens")
			}
			seen[base] = true
		}
		if len(c.Zero.Owners) < 1 || len(c.Zero.Owners) > 2 || c.Zero.Owners[0] != qa.Owner || (len(c.Zero.Owners) == 2 && c.Zero.Owners[1] != 23456789) {
			return errors.New("restore the allowed synthetic owners before restart")
		}
		raw, e = os.ReadFile(filepath.Join(primary.AdminDir, "options.json"))
		if e != nil {
			return e
		}
		var options struct {
			Enabled bool   `json:"enabled"`
			Address string `json:"address"`
		}
		if json.Unmarshal(raw, &options) != nil || !options.Enabled || options.Address != primary.Address {
			return errors.New("restore this run's enabled admin address before restart")
		}
		return nil
	}
	start := func() error {
		if e := guard(); e != nil {
			return e
		}
		var e error
		bot, e = primary.Start(ctx, fmt.Sprintf("start-%d", starts+1))
		if e != nil {
			return e
		}
		attachCtx, stop := context.WithTimeout(ctx, 15*time.Second)
		defer stop()
		if e = secondary.WaitConnection(attachCtx); e != nil {
			return e
		}
		starts++
		return nil
	}
	check := func() error {
		for _, endpoint := range []*qa.Harness{primary, secondary} {
			reply, _, e := endpoint.Send(qa.Owner, "翻转 abc")
			if e != nil || reply != "ɔqɐ" {
				return errors.New("two-account synthetic transport check failed")
			}
		}
		checks++
		return nil
	}
	if err = start(); err != nil {
		return err
	}
	client := qa.NewClient(primary.Address)
	if _, err = client.AuthWithPassword(primary.AdminDir, password); err != nil {
		return err
	}
	credential, err := os.ReadFile(filepath.Join(primary.AdminDir, "credential.json"))
	if err != nil {
		return err
	}
	credentialHash := sha256.Sum256(credential)
	if err = check(); err != nil {
		return err
	}
	session := func() error {
		value := map[string]any{"state": "ready", "url": "http://" + primary.Address + "/", "output": out, "loginFile": filepath.Join(out, "login.txt"), "primaryURL": primary.MockURL(), "secondaryURL": secondary.MockURL(), "mockAccounts": []int64{87654321, 87654322}, "processStarts": starts, "transportChecks": checks, "realQQMessagesSent": false}
		if e := qa.WriteJSON(filepath.Join(out, "session.json"), value); e != nil {
			return e
		}
		encoded, _ := json.Marshal(value)
		fmt.Println(string(encoded))
		return nil
	}
	if err = session(); err != nil {
		return err
	}
	commands := make(chan string)
	done := make(chan error, 1)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			select {
			case commands <- strings.TrimSpace(s.Text()):
			case <-ctx.Done():
				return
			}
		}
		done <- s.Err()
	}()
	fmt.Println("Commands: status, check, restart, stop")
	for {
		select {
		case <-ctx.Done():
			result["stopReason"] = "signal-or-time-limit"
			return nil
		case e := <-done:
			result["stopReason"] = "input-ended"
			return e
		case command := <-commands:
			switch command {
			case "status":
				if err = session(); err != nil {
					return err
				}
			case "check":
				if err = check(); err != nil {
					return err
				}
				fmt.Println("TWO_ACCOUNT_CHECK_PASSED")
			case "restart":
				if e := guard(); e != nil {
					fmt.Println("Restart refused: " + e.Error())
					result["restartRefusedBeforeStop"] = true
					continue
				}
				bot.Stop()
				bot = nil
				if err = start(); err != nil {
					return err
				}
				status, _, _, e := client.Request("GET", "/api/connections", nil)
				if e != nil || status != 401 {
					return errors.New("old session was not rejected after restart")
				}
				client = qa.NewClient(primary.Address)
				if _, err = client.AuthWithPassword(primary.AdminDir, password); err != nil {
					return err
				}
				credential, err = os.ReadFile(filepath.Join(primary.AdminDir, "credential.json"))
				if err != nil {
					return err
				}
				if sha256.Sum256(credential) != credentialHash {
					return errors.New("test credential changed")
				}
				value, e := client.Checked("GET", "/api/connections", nil)
				if e != nil {
					return e
				}
				if value["restartRequired"] != false {
					return errors.New("saved configuration differs after restart")
				}
				if err = check(); err != nil {
					return err
				}
				result["oldSessionRejected"] = true
				result["credentialPreserved"] = true
				result["restartLoadedSavedConfig"] = true
				if err = session(); err != nil {
					return err
				}
			case "stop":
				result["stopReason"] = "stop-command"
				result["passed"] = starts >= 2 && checks >= 2
				return nil
			default:
				fmt.Println("Commands: status, check, restart, stop")
			}
		}
	}
}
