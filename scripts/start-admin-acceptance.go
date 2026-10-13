//go:build ignore

// A disposable Windows admin instance for manual page acceptance.
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
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	qa "github.com/FloatTech/ZeroBot-Plugin/internal/adminqa"
)

type acceptance struct {
	h                                        *qa.Harness
	bot                                      *qa.Bot
	client                                   *qa.Client
	password, originalAddress, editedAddress string
	backups                                  map[string]string
	credentialHash                           [32]byte
	starts                                   int
	result                                   map[string]any
}

func main() {
	binary := flag.String("binary-name", "zerobot.exe", "executable filename under .local/bin")
	lifetime := flag.Duration("lifetime", 30*time.Minute, "maximum lifetime, 1m to 2h")
	selfCheck := flag.Bool("self-check", false, "check preparation/restore/restart and exit")
	flag.Parse()
	if flag.NArg() != 0 || filepath.Base(*binary) != *binary || strings.ContainsAny(*binary, "/\\:") || filepath.Ext(*binary) != ".exe" || *lifetime < time.Minute || *lifetime > 2*time.Hour {
		fmt.Fprintln(os.Stderr, "Invalid arguments: use a local executable filename and lifetime 1m to 2h.")
		os.Exit(2)
	}
	if err := run(*binary, *lifetime, *selfCheck); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(binaryName string, lifetime time.Duration, selfCheck bool) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return err
	}
	source := filepath.Join(root, ".local/bin", binaryName)
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("cannot read the selected local binary")
	}
	defer input.Close()
	parent := filepath.Join(root, ".local/admin-acceptance")
	if err = os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	output, err := os.MkdirTemp(parent, "run-")
	if err != nil {
		return err
	}
	copied := filepath.Join(output, "zerobot-acceptance.exe")
	executable, err := os.OpenFile(copied, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(executable, hash), input)
	closeErr := executable.Close()
	inputCloseErr := input.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if inputCloseErr != nil {
		return inputCloseErr
	}
	signalContext, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignal()
	ctx, cancel := context.WithTimeout(signalContext, lifetime)
	defer cancel()
	deadline, _ := ctx.Deadline()
	a := &acceptance{result: map[string]any{"startedAt": time.Now().Format(time.RFC3339), "isolated": true, "binary": binaryName, "binarySHA256": hex.EncodeToString(hash.Sum(nil)), "selfCheck": selfCheck, "realQQMessagesSent": false, "manualAcceptancePassed": false, "environmentPrepared": false, "output": output, "expiresAt": deadline.Format(time.RFC3339)}}
	defer func() {
		if a.bot != nil {
			a.bot.Stop()
		}
		if a.h != nil {
			a.h.Close()
		}
		a.result["stoppedAt"] = time.Now().Format(time.RFC3339)
		a.result["processStarts"] = a.starts
		a.result["instanceStopped"] = true
		if err != nil {
			a.result["failure"] = err.Error()
		}
		if saveErr := qa.WriteJSON(filepath.Join(output, "result.json"), a.result); saveErr != nil && err == nil {
			err = saveErr
		}
		_ = qa.WriteJSON(filepath.Join(output, "session.json"), map[string]any{"state": "stopped", "output": output})
		fmt.Println("ARTIFACTS=" + output)
	}()
	a.h, err = qa.New(root, copied, output)
	if err != nil {
		return err
	}
	a.originalAddress = a.h.Address
	for attempt := 0; attempt < 8; attempt++ {
		a.editedAddress, err = qa.FreeAddress()
		if err != nil {
			return err
		}
		if a.editedAddress != a.originalAddress {
			break
		}
	}
	if a.editedAddress == a.originalAddress {
		return errors.New("cannot prepare a different isolated admin address")
	}
	secret := make([]byte, 24)
	if _, err = rand.Read(secret); err != nil {
		return err
	}
	a.password = "QA-" + hex.EncodeToString(secret)
	if err = os.WriteFile(filepath.Join(output, "login.txt"), []byte("Isolated admin password (not a daily credential):\n"+a.password+"\n"), 0600); err != nil {
		return err
	}
	if err = a.start(ctx, "initial", true); err != nil {
		return err
	}
	if err = a.prepare(); err != nil {
		return err
	}
	a.result["environmentPrepared"] = true
	a.result["baselineBackups"] = a.backups
	if selfCheck {
		return a.check(ctx)
	}
	if err = a.session(output); err != nil {
		return err
	}
	fmt.Println("Commands: status, restart, stop. Browser login uses the private login.txt file.")
	commands := make(chan string)
	finished := make(chan error, 1)
	monitor := time.NewTicker(3 * time.Second)
	defer monitor.Stop()
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case commands <- strings.TrimSpace(scanner.Text()):
			case <-ctx.Done():
				return
			}
		}
		finished <- scanner.Err()
	}()
	for {
		select {
		case <-monitor.C:
			if ctx.Err() != nil {
				continue
			}
			if !a.bot.Running() {
				if ctx.Err() != nil {
					continue
				}
				return errors.New("the isolated process exited; inspect this run's logs")
			}
		case <-ctx.Done():
			a.result["stopReason"] = "signal-or-time-limit"
			return nil
		case inputErr := <-finished:
			a.result["stopReason"] = "input-ended"
			return inputErr
		case command := <-commands:
			switch command {
			case "stop":
				a.result["stopReason"] = "stop-command"
				return nil
			case "status":
				if err = a.session(output); err != nil {
					return err
				}
			case "restart":
				if guardErr := a.guard(); guardErr != nil {
					fmt.Println("Restart refused: " + guardErr.Error())
					continue
				}
				a.bot.Stop()
				a.bot = nil
				if err = a.start(ctx, fmt.Sprintf("restart-%d", a.starts), false); err != nil {
					return err
				}
				if err = a.session(output); err != nil {
					return err
				}
			default:
				fmt.Println("Commands: status, restart, stop.")
			}
		}
	}
}

func (a *acceptance) guard() error {
	raw, err := os.ReadFile(a.h.ConfigPath)
	if err != nil {
		return err
	}
	var configuration struct {
		WS []struct {
			URL   string `json:"Url"`
			Token string `json:"AccessToken"`
		} `json:"ws"`
		WSS []json.RawMessage `json:"wss"`
	}
	if json.Unmarshal(raw, &configuration) != nil || len(configuration.WS) != 1 || len(configuration.WSS) != 0 || configuration.WS[0].URL != a.h.MockURL() || configuration.WS[0].Token != "isolated-lifecycle-token" {
		return errors.New("restore the original synthetic OneBot connection before restarting")
	}
	raw, err = os.ReadFile(filepath.Join(a.h.AdminDir, "options.json"))
	if err != nil {
		return err
	}
	var options struct {
		Enabled bool   `json:"enabled"`
		Address string `json:"address"`
	}
	if json.Unmarshal(raw, &options) != nil || !options.Enabled || (options.Address != a.originalAddress && options.Address != a.editedAddress) {
		return errors.New("restore the fixture's enabled admin options and prepared address before restarting")
	}
	return nil
}
func (a *acceptance) start(ctx context.Context, label string, authenticate bool) error {
	if err := a.guard(); err != nil {
		return err
	}
	bot, err := a.h.Start(ctx, label)
	if err != nil {
		return err
	}
	a.bot = bot
	a.starts++
	a.client = qa.NewClient(a.h.Address)
	if authenticate {
		_, err = a.client.AuthWithPassword(a.h.AdminDir, a.password)
	}
	return err
}
func (a *acceptance) session(output string) error {
	if !a.bot.Running() {
		return errors.New("the isolated process is not running")
	}
	value := map[string]any{"state": "ready", "url": "http://" + a.h.Address + "/", "loginFile": filepath.Join(output, "login.txt"), "output": output, "baselineBackups": a.backups, "expiresAt": a.result["expiresAt"], "realQQMessagesSent": false, "manualAcceptancePassed": false}
	if err := qa.WriteJSON(filepath.Join(output, "session.json"), value); err != nil {
		return err
	}
	encoded, _ := json.Marshal(value)
	fmt.Println(string(encoded))
	return nil
}
func (a *acceptance) pluginState() (map[string]any, error) {
	value, err := a.client.Checked("GET", "/api/plugins?group=101", nil)
	if err != nil {
		return nil, err
	}
	for _, item := range value["plugins"].([]any) {
		plugin := item.(map[string]any)
		if plugin["name"] == "chrev" {
			return plugin["state"].(map[string]any), nil
		}
	}
	return nil, errors.New("fixture chrev plugin is not available")
}
func (a *acceptance) setPlugin(mode string) error {
	state, err := a.pluginState()
	if err != nil {
		return err
	}
	_, err = a.client.Checked("POST", "/api/plugin-switch", map[string]any{"name": "chrev", "groupId": "101", "mode": mode, "revision": state["revision"]})
	return err
}
func (a *acceptance) setModel(model string) error {
	value, err := a.client.Checked("GET", "/api/ai", nil)
	if err != nil {
		return err
	}
	_, err = a.client.Checked("POST", "/api/ai", map[string]any{"revision": value["global"].(map[string]any)["revision"], "fields": map[string]any{"ModelName": model}})
	return err
}
func (a *acceptance) prepare() error {
	if err := a.setPlugin("disabled"); err != nil {
		return err
	}
	if err := a.setModel("isolated-baseline-model"); err != nil {
		return err
	}
	a.backups = map[string]string{}
	for _, scope := range []string{"startup-config", "control-db", "admin-options"} {
		value, err := a.client.Checked("POST", "/api/backups", map[string]string{"scope": scope})
		if err != nil {
			return err
		}
		id, _ := value["backupId"].(string)
		if id == "" {
			return errors.New("fixture backup ID is missing")
		}
		a.backups[scope] = id
	}
	value, err := a.client.Checked("GET", "/api/connections", nil)
	if err != nil {
		return err
	}
	if _, err = a.client.Checked("POST", "/api/connections", map[string]any{"revision": value["revision"], "superUsers": []string{"12345678", "23456789"}}); err != nil {
		return err
	}
	if err = a.setPlugin("enabled"); err != nil {
		return err
	}
	if err = a.setModel("isolated-edited-model"); err != nil {
		return err
	}
	if err = a.h.Options(true, a.editedAddress); err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join(a.h.AdminDir, "credential.json"))
	if err != nil {
		return err
	}
	a.credentialHash = sha256.Sum256(raw)
	return nil
}
func (a *acceptance) check(ctx context.Context) error {
	// Deliberately invalid synthetic options must not stop or restart the current bot.
	if err := a.h.Options(true, "127.0.0.1:3010"); err != nil {
		return err
	}
	if a.guard() == nil {
		return errors.New("daily-address restart guard failed")
	}
	if err := a.h.Options(false, a.originalAddress); err != nil {
		return err
	}
	if a.guard() == nil {
		return errors.New("disabled-admin restart guard failed")
	}
	if err := a.h.Options(true, a.editedAddress); err != nil {
		return err
	}
	raw, err := os.ReadFile(a.h.ConfigPath)
	if err != nil {
		return err
	}
	var changed map[string]any
	if err = json.Unmarshal(raw, &changed); err != nil {
		return err
	}
	changed["ws"] = []any{map[string]any{"Url": "ws://127.0.0.1:1", "AccessToken": "synthetic-invalid"}}
	if err = qa.WriteJSON(a.h.ConfigPath, changed); err != nil {
		return err
	}
	if a.guard() == nil {
		return errors.New("non-fixture OneBot restart guard failed")
	}
	if err = os.WriteFile(a.h.ConfigPath, raw, 0600); err != nil {
		return err
	}
	if _, err = a.client.Checked("GET", "/api/overview", nil); err != nil {
		return errors.New("guard rejection stopped the current fixture")
	}
	a.result["restartGuardsPassed"] = true
	for _, scope := range []string{"startup-config", "control-db", "admin-options"} {
		id := a.backups[scope]
		preview, err := a.client.Checked("POST", "/api/backups/preview", map[string]string{"id": id})
		if err != nil {
			return err
		}
		if preview["changed"] != true {
			return fmt.Errorf("%s fixture did not differ", scope)
		}
		value, err := a.client.Checked("POST", "/api/backups/restore", map[string]any{"id": id, "revision": preview["revision"], "expiresAt": preview["expiresAt"], "token": preview["token"], "confirm": true})
		if err != nil {
			return err
		}
		if value["changed"] != true || value["backupId"] == "" {
			return fmt.Errorf("%s restore was not confirmed", scope)
		}
	}
	old := a.client
	a.bot.Stop()
	a.bot = nil
	if err = a.start(ctx, "restored-restart", true); err != nil {
		return err
	}
	if status, _, _, _ := old.Request("GET", "/api/overview", nil); status != http.StatusUnauthorized {
		return errors.New("old session survived fixture restart")
	}
	connections, err := a.client.Checked("GET", "/api/connections", nil)
	if err != nil {
		return err
	}
	owners, ok := connections["superUsers"].([]any)
	if !ok || len(owners) != 1 || owners[0] != "12345678" || connections["restartRequired"] != false {
		return errors.New("restored owner or restart state differs")
	}
	state, err := a.pluginState()
	if err != nil {
		return err
	}
	if state["groupMode"] != "disabled" {
		return errors.New("restored plugin mode differs")
	}
	ai, err := a.client.Checked("GET", "/api/ai", nil)
	if err != nil {
		return err
	}
	if ai["global"].(map[string]any)["fields"].(map[string]any)["ModelName"] != "isolated-baseline-model" {
		return errors.New("restored model differs")
	}
	credential, err := os.ReadFile(filepath.Join(a.h.AdminDir, "credential.json"))
	if err != nil {
		return err
	}
	if sha256.Sum256(credential) != a.credentialHash {
		return errors.New("fixture credential changed")
	}
	if a.h.Address != a.originalAddress {
		return errors.New("restored address differs")
	}
	a.result["automaticEnvironmentCheckPassed"] = true
	a.result["threeScopeRestorePassed"] = true
	a.result["oldSessionRejected"] = true
	a.result["passwordPreserved"] = true
	return nil
}
