//go:build ignore

// Windows acceptance: go run scripts/test-admin-system.go -mode restart|performance
// All process/configuration writes are confined to a new .local/admin-system directory.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	qa "github.com/FloatTech/ZeroBot-Plugin/internal/adminqa"
)

var mode = flag.String("mode", "restart", "restart or performance")
var binaryName = flag.String("binary-name", "zerobot.exe", "tested executable filename under .local/bin")
var warmup = flag.Duration("warmup", 60*time.Second, "performance warmup per profile")
var duration = flag.Duration("duration", 5*time.Minute, "performance sample duration per profile")

func main() {
	flag.Parse()
	if *mode != "restart" && *mode != "performance" {
		fmt.Fprintln(os.Stderr, "invalid mode")
		os.Exit(1)
	}
	if filepath.Base(*binaryName) != *binaryName || filepath.Ext(*binaryName) != ".exe" {
		fmt.Fprintln(os.Stderr, "invalid binary filename")
		os.Exit(1)
	}
	if *warmup < 0 || *duration < 10*time.Second {
		fmt.Fprintln(os.Stderr, "invalid timing")
		os.Exit(1)
	}
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	binary := filepath.Join(root, ".local/bin", *binaryName)
	artifactRoot := filepath.Join(root, ".local/admin-system")
	if err = os.MkdirAll(artifactRoot, 0700); err != nil {
		panic(err)
	}
	output, err := os.MkdirTemp(artifactRoot, *mode+"-")
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
	defer cancel()
	result := map[string]any{"mode": *mode, "startedAt": time.Now().Format(time.RFC3339), "isolated": true, "realQQMessagesSent": false, "modelCalls": 0, "binary": *binaryName, "logicalCPUs": runtime.NumCPU()}
	data, err := os.ReadFile(binary)
	if err == nil {
		hash := sha256.Sum256(data)
		result["binarySHA256"] = hex.EncodeToString(hash[:])
		data = nil
	}
	if err == nil {
		if *mode == "restart" {
			err = restartCheck(ctx, root, binary, output, result)
		} else {
			err = performanceCheck(ctx, root, binary, output, result)
		}
	}
	result["completedAt"] = time.Now().Format(time.RFC3339)
	result["passed"] = err == nil
	if err != nil {
		result["failure"] = err.Error()
	}
	if saveErr := qa.WriteJSON(filepath.Join(output, "result.json"), result); saveErr != nil {
		fmt.Fprintln(os.Stderr, saveErr)
		os.Exit(1)
	}
	if err == nil {
		_ = qa.WriteJSON(filepath.Join(artifactRoot, *mode+"-latest.json"), result)
	}
	// Keep full resource samples in the artifact; print only the profile summaries.
	visible := make(map[string]any, len(result))
	for key, value := range result {
		visible[key] = value
	}
	if profiles, ok := result["profiles"].([]profileResult); ok {
		compact := append([]profileResult(nil), profiles...)
		for i := range compact {
			compact[i].Samples = nil
		}
		visible["profiles"] = compact
	}
	encoded, _ := json.MarshalIndent(visible, "", "  ")
	fmt.Println(string(encoded))
	fmt.Println("ARTIFACTS=" + output)
	if err != nil {
		os.Exit(1)
	}
}

func pluginState(c *qa.Client, group string) (map[string]any, error) {
	list, err := c.Checked("GET", "/api/plugins?group="+group, nil)
	if err != nil {
		return nil, err
	}
	for _, item := range list["plugins"].([]any) {
		p := item.(map[string]any)
		if p["name"] == "chrev" {
			return p["state"].(map[string]any), nil
		}
	}
	return nil, fmt.Errorf("chrev missing")
}
func setPlugin(c *qa.Client, group, mode string) error {
	s, err := pluginState(c, group)
	if err != nil {
		return err
	}
	_, err = c.Checked("POST", "/api/plugin-switch", map[string]any{"name": "chrev", "groupId": group, "mode": mode, "revision": s["revision"]})
	return err
}
func setModel(c *qa.Client, model string) error {
	v, err := c.Checked("GET", "/api/ai", nil)
	if err != nil {
		return err
	}
	_, err = c.Checked("POST", "/api/ai", map[string]any{"revision": v["global"].(map[string]any)["revision"], "fields": map[string]any{"ModelName": model}})
	return err
}
func snapshot(c *qa.Client, scope string) (string, error) {
	v, err := c.Checked("POST", "/api/backups", map[string]string{"scope": scope})
	if err != nil {
		return "", err
	}
	id, _ := v["backupId"].(string)
	if id == "" {
		return "", fmt.Errorf("missing backup ID")
	}
	return id, nil
}
func restore(c *qa.Client, id string) error {
	p, err := c.Checked("POST", "/api/backups/preview", map[string]string{"id": id})
	if err != nil {
		return err
	}
	if p["changed"] != true {
		return fmt.Errorf("restore fixture did not differ")
	}
	v, err := c.Checked("POST", "/api/backups/restore", map[string]any{"id": id, "revision": p["revision"], "expiresAt": p["expiresAt"], "token": p["token"], "confirm": true})
	if err != nil {
		return err
	}
	if v["changed"] != true || v["backupId"] == "" {
		return fmt.Errorf("restore did not commit or back up current data")
	}
	return nil
}
func command(h *qa.Harness, text, want string) error {
	reply, _, err := h.Send(qa.Owner, text)
	if err != nil || !strings.Contains(reply, want) {
		return fmt.Errorf("synthetic command %s failed", text)
	}
	return nil
}
func privateGate(h *qa.Harness, uid int64, want string) error {
	reply, _, err := h.Send(uid, "翻转 abc")
	if err != nil || reply != want {
		return fmt.Errorf("synthetic message gate %d failed", uid)
	}
	return nil
}
func fileDigest(path string) ([32]byte, error) {
	b, err := os.ReadFile(path)
	return sha256.Sum256(b), err
}

func restartCheck(ctx context.Context, root, binary, output string, result map[string]any) error {
	h, err := qa.New(root, binary, output)
	if err != nil {
		return err
	}
	defer h.Close()
	bot, err := h.Start(ctx, "initial")
	if err != nil {
		return err
	}
	defer func() { bot.Stop() }()
	c := qa.NewClient(h.Address)
	if _, err = c.Auth(h.AdminDir); err != nil {
		return err
	}
	credentialPath := filepath.Join(h.AdminDir, "credential.json")
	credentialBefore, err := fileDigest(credentialPath)
	if err != nil {
		return err
	}
	if err = setPlugin(c, "101", "disabled"); err != nil {
		return err
	}
	if err = setModel(c, "qa-restart-original"); err != nil {
		return err
	}
	if err = command(h, "/全局禁止 chrev 23456789", "已禁止"); err != nil {
		return err
	}
	if err = command(h, "/封禁 34567890", "已封禁"); err != nil {
		return err
	}
	startupID, err := snapshot(c, "startup-config")
	if err != nil {
		return err
	}
	controlID, err := snapshot(c, "control-db")
	if err != nil {
		return err
	}
	// Save a future admin address while the current process still uses address A.
	addressA := h.Address
	addressB, err := qa.FreeAddress()
	if err != nil {
		return err
	}
	if err = h.Options(true, addressB); err != nil {
		return err
	}
	optionsID, err := snapshot(c, "admin-options")
	if err != nil {
		return err
	}
	if err = h.Options(true, addressA); err != nil {
		return err
	}
	connections, err := c.Checked("GET", "/api/connections", nil)
	if err != nil {
		return err
	}
	if _, err = c.Checked("POST", "/api/connections", map[string]any{"revision": connections["revision"], "superUsers": []string{"12345678", "23456789"}}); err != nil {
		return err
	}
	connections, err = c.Checked("GET", "/api/connections", nil)
	if err != nil {
		return err
	}
	if connections["restartRequired"] != true || len(connections["runningSuperUsers"].([]any)) != 1 {
		return fmt.Errorf("saved/running configuration boundary failed")
	}
	if err = setPlugin(c, "101", "enabled"); err != nil {
		return err
	}
	if err = setPlugin(c, "0", "disabled"); err != nil {
		return err
	}
	if err = setModel(c, "qa-restart-temporary"); err != nil {
		return err
	}
	if err = command(h, "/全局允许 chrev 23456789", "已允许"); err != nil {
		return err
	}
	if err = command(h, "/解封 34567890", "已解封"); err != nil {
		return err
	}
	if err = command(h, "/全局沉默", "休息"); err != nil {
		return err
	}
	bot.Stop()
	bot, err = h.Start(ctx, "changed-restart")
	if err != nil {
		return err
	}
	if status, _, _, _ := c.Request("GET", "/api/overview", nil); status != 401 {
		return fmt.Errorf("old session survived process restart")
	}
	c = qa.NewClient(h.Address)
	if _, err = c.Auth(h.AdminDir); err != nil {
		return err
	}
	connections, err = c.Checked("GET", "/api/connections", nil)
	if err != nil {
		return err
	}
	if connections["restartRequired"] != false || len(connections["runningSuperUsers"].([]any)) != 2 {
		return fmt.Errorf("restart did not apply saved owners")
	}
	if err = restore(c, startupID); err != nil {
		return err
	}
	if err = restore(c, controlID); err != nil {
		return err
	}
	if err = restore(c, optionsID); err != nil {
		return err
	}
	connections, err = c.Checked("GET", "/api/connections", nil)
	if err != nil {
		return err
	}
	if connections["restartRequired"] != true || len(connections["runningSuperUsers"].([]any)) != 2 {
		return fmt.Errorf("restored startup did not mark restart")
	}
	if err = privateGate(h, qa.Owner, "ɔqɐ"); err != nil {
		return err
	}
	bot.Stop()
	bot, err = h.Start(ctx, "restored-restart")
	if err != nil {
		return err
	}
	if h.Address != addressB {
		return fmt.Errorf("restored admin address not selected")
	}
	if status, _, _, _ := c.Request("GET", "/api/overview", nil); status != 0 {
		return fmt.Errorf("old admin address remained available")
	}
	oldAtNew := *c
	oldAtNew.Base = "http://" + addressB
	if status, _, _, _ := oldAtNew.Request("GET", "/api/overview", nil); status != 401 {
		return fmt.Errorf("old session accepted at restored address")
	}
	c = qa.NewClient(h.Address)
	if _, err = c.Auth(h.AdminDir); err != nil {
		return err
	}
	connections, err = c.Checked("GET", "/api/connections", nil)
	if err != nil {
		return err
	}
	if connections["restartRequired"] != false || len(connections["runningSuperUsers"].([]any)) != 1 || len(connections["superUsers"].([]any)) != 1 {
		return fmt.Errorf("restored owners not loaded after restart")
	}
	s, err := pluginState(c, "101")
	if err != nil {
		return err
	}
	if s["groupMode"] != "disabled" || s["globalMode"] != "inherit" || s["effectiveEnabled"] != false {
		return fmt.Errorf("restored switch inheritance lost on restart")
	}
	ai, err := c.Checked("GET", "/api/ai", nil)
	if err != nil {
		return err
	}
	if ai["global"].(map[string]any)["fields"].(map[string]any)["ModelName"] != "qa-restart-original" {
		return fmt.Errorf("restored AI configuration lost on restart")
	}
	if err = privateGate(h, qa.Owner, "ɔqɐ"); err != nil {
		return err
	}
	if err = privateGate(h, 23456789, ""); err != nil {
		return err
	}
	if err = privateGate(h, 34567890, ""); err != nil {
		return err
	}
	credentialAfter, err := fileDigest(credentialPath)
	if err != nil || credentialAfter != credentialBefore {
		return fmt.Errorf("restore changed admin credential")
	}
	// Restore a disabled admin setting; its launcher-selected flag must be absent.
	if err = h.Options(false, addressB); err != nil {
		return err
	}
	disabledID, err := snapshot(c, "admin-options")
	if err != nil {
		return err
	}
	if err = h.Options(true, addressB); err != nil {
		return err
	}
	if err = restore(c, disabledID); err != nil {
		return err
	}
	bot.Stop()
	bot, err = h.Start(ctx, "disabled-restart")
	if err != nil {
		return err
	}
	probe := &http.Client{Timeout: time.Second}
	if rsp, e := probe.Get("http://" + addressB + "/api/session"); e == nil {
		rsp.Body.Close()
		return fmt.Errorf("disabled admin still listening")
	}
	if err = privateGate(h, qa.Owner, "ɔqɐ"); err != nil {
		return err
	}
	for _, key := range []string{"startupSaveRestartPassed", "startupRestoreRestartPassed", "controlRestoreRestartPassed", "adminAddressRestoreRestartPassed", "adminDisableRestoreRestartPassed", "oldSessionRejectedPassed", "credentialPreservedPassed", "restoredResponseBanBlockPassed"} {
		result[key] = true
	}
	result["processLaunches"] = 4
	result["launcher"] = "isolated process; reads options.json into the same flags as service-host.ps1; scheduler not exercised"
	return nil
}

type sample struct {
	AtSeconds, CPUSeconds         float64
	PrivateBytes, WorkingSetBytes uint64
}
type profileResult struct {
	Name                                                 string `json:"name"`
	WarmupSeconds, SampleSeconds                         float64
	ReadP95Ms                                            map[string]float64
	ReadCounts                                           map[string]int
	LoginP95Ms, MessageP95Ms, CPUPercent                 float64
	MessageCount, LoginCount                             int
	PrivateFirst, PrivateLast, PrivateMax, WorkingSetMax uint64
	PrivateGrowthMiB                                     float64
	Samples                                              []sample `json:"samples,omitempty"`
}

func p95(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	return v[int(math.Ceil(float64(len(v))*.95))-1]
}

func performanceCheck(ctx context.Context, root, binary, output string, result map[string]any) error {
	profiles := []string{"admin-off", "admin-idle", "dashboard-poll-10s", "logs-poll-2s"}
	results := []profileResult{}
	for _, name := range profiles {
		dir := filepath.Join(output, name)
		if err := os.Mkdir(dir, 0700); err != nil {
			return err
		}
		h, err := qa.New(root, binary, dir)
		if err != nil {
			return err
		}
		pr, err := measureProfile(ctx, h, name)
		h.Close()
		if err != nil {
			result["profiles"] = results
			return err
		}
		results = append(results, pr)
		result["profiles"] = results
		if err = qa.WriteJSON(filepath.Join(output, "progress.json"), result); err != nil {
			return err
		}
		fmt.Printf("PROFILE_DONE=%s messages=%d message_p95_ms=%.2f cpu_percent=%.4f private_growth_mib=%.2f\n", name, pr.MessageCount, pr.MessageP95Ms, pr.CPUPercent, pr.PrivateGrowthMiB)
	}
	fullTiming := *warmup >= 60*time.Second && *duration >= 5*time.Minute
	thresholds := true
	baseline := results[0]
	for _, p := range results {
		if p.PrivateGrowthMiB > 64 || p.MessageP95Ms-baseline.MessageP95Ms > 100 || p.CPUPercent-baseline.CPUPercent > 1 || p.LoginP95Ms > 1000 {
			thresholds = false
		}
		for _, value := range p.ReadP95Ms {
			if value > 500 {
				thresholds = false
			}
		}
	}
	result["fullTiming"] = fullTiming
	result["thresholdsPassed"] = thresholds
	result["workload"] = "one synthetic chrev private message every five seconds; Windows process metrics every two seconds; 10 successful logins and 20 reads per endpoint outside resource interval; server-side polling only"
	result["browserRendererMeasured"] = false
	result["schedulerMeasured"] = false
	if !thresholds {
		return fmt.Errorf("performance thresholds exceeded; inspect profile measurements")
	}
	return nil
}

func measureProfile(ctx context.Context, h *qa.Harness, name string) (p profileResult, resultErr error) {
	p = profileResult{Name: name, WarmupSeconds: warmup.Seconds(), SampleSeconds: duration.Seconds(), ReadP95Ms: map[string]float64{}, ReadCounts: map[string]int{}}
	if err := h.Options(name != "admin-off", h.Address); err != nil {
		return p, err
	}
	bot, err := h.Start(ctx, "performance")
	if err != nil {
		return p, err
	}
	defer bot.Stop()
	var client *qa.Client
	if name != "admin-off" {
		client = qa.NewClient(h.Address)
		if _, err = client.Auth(h.AdminDir); err != nil {
			return p, err
		}
	}
	readTimes := map[string][]float64{}
	loginTimes := []float64{}
	messageTimes := []float64{}
	cursor, instance := "", ""
	read := func(path string, record bool) error {
		status, value, t, err := client.Request("GET", path, nil)
		if err != nil || status != 200 {
			return fmt.Errorf("performance read %s failed: status %d", path, status)
		}
		key := strings.SplitN(path, "?", 2)[0]
		if record {
			readTimes[key] = append(readTimes[key], float64(t)/float64(time.Millisecond))
		}
		if key == "/api/logs" {
			cursor, _ = value["cursor"].(string)
			instance, _ = value["instance"].(string)
			if value["capacity"] != float64(1000) {
				return fmt.Errorf("log capacity changed")
			}
		}
		return nil
	}
	// Probe login and endpoints before warmup, so CPU resource comparison excludes them.
	if client != nil {
		for i := 0; i < 10; i++ {
			t, err := client.Auth(h.AdminDir)
			if err != nil {
				return p, err
			}
			loginTimes = append(loginTimes, float64(t)/float64(time.Millisecond))
		}
		for i := 0; i < 20; i++ {
			for _, path := range []string{"/api/overview", "/api/plugins", "/api/connections", "/api/ai", "/api/backups", "/api/logs?limit=200"} {
				if err := read(path, true); err != nil {
					return p, err
				}
			}
		}
	}
	runInterval := func(length time.Duration, record bool) error {
		start := time.Now()
		nextMessage, nextRead := start, start
		nextProgress := start.Add(time.Minute)
		for time.Since(start) < length {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			now := time.Now()
			if !now.Before(nextMessage) {
				reply, t, err := h.Send(qa.Owner, "翻转 abc")
				if err != nil || reply != "ɔqɐ" {
					return fmt.Errorf("performance message reply failed")
				}
				if record {
					messageTimes = append(messageTimes, float64(t)/float64(time.Millisecond))
				}
				nextMessage = nextMessage.Add(5 * time.Second)
			}
			if client != nil && !now.Before(nextRead) {
				if name == "dashboard-poll-10s" {
					for _, path := range []string{"/api/overview", "/api/plugins", "/api/connections"} {
						if err := read(path, record); err != nil {
							return err
						}
					}
				} else if name == "logs-poll-2s" {
					path := "/api/logs?limit=200"
					if cursor != "" && instance != "" {
						path += "&after=" + cursor + "&instance=" + instance
					}
					if err := read(path, record); err != nil {
						return err
					}
				}
				period := 10 * time.Second
				if name == "logs-poll-2s" {
					period = 2 * time.Second
				}
				nextRead = nextRead.Add(period)
			}
			if record {
				cpu, private, working, err := bot.Metrics()
				if err != nil {
					return err
				}
				p.Samples = append(p.Samples, sample{time.Since(start).Seconds(), cpu, private, working})
				if private > p.PrivateMax {
					p.PrivateMax = private
				}
				if working > p.WorkingSetMax {
					p.WorkingSetMax = working
				}
			}
			if !time.Now().Before(nextProgress) {
				phase := "warmup"
				if record {
					phase = "sample"
				}
				status := map[string]any{"profile": name, "phase": phase, "elapsedSeconds": time.Since(start).Seconds(), "messageCount": len(messageTimes), "messageP95Ms": p95(messageTimes), "privateMaxMiB": float64(p.PrivateMax) / (1 << 20)}
				if err := qa.WriteJSON(filepath.Join(h.Root, "sampling-progress.json"), status); err != nil {
					return err
				}
				fmt.Printf("PROFILE_PROGRESS=%s phase=%s elapsed_seconds=%.0f messages=%d message_p95_ms=%.2f private_max_mib=%.2f\n", name, phase, time.Since(start).Seconds(), len(messageTimes), p95(messageTimes), float64(p.PrivateMax)/(1<<20))
				nextProgress = nextProgress.Add(time.Minute)
			}
			time.Sleep(2 * time.Second)
		}
		return nil
	}
	fmt.Printf("PROFILE_START=%s warmup_seconds=%.0f sample_seconds=%.0f\n", name, warmup.Seconds(), duration.Seconds())
	if err = runInterval(*warmup, false); err != nil {
		return p, err
	}
	if err = runInterval(*duration, true); err != nil {
		return p, err
	}
	if len(p.Samples) < 2 {
		return p, fmt.Errorf("too few resource samples")
	}
	first, last := p.Samples[0], p.Samples[len(p.Samples)-1]
	p.PrivateFirst = first.PrivateBytes
	p.PrivateLast = last.PrivateBytes
	p.PrivateGrowthMiB = float64(p.PrivateMax-p.PrivateFirst) / (1 << 20)
	p.CPUPercent = (last.CPUSeconds - first.CPUSeconds) / (last.AtSeconds - first.AtSeconds) / float64(runtime.NumCPU()) * 100
	p.MessageP95Ms = p95(messageTimes)
	p.MessageCount = len(messageTimes)
	p.LoginP95Ms = p95(loginTimes)
	p.LoginCount = len(loginTimes)
	for key, times := range readTimes {
		p.ReadP95Ms[key] = p95(times)
		p.ReadCounts[key] = len(times)
	}
	return p, nil
}
