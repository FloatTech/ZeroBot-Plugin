//go:build ignore

// An isolated synthetic workload for real-browser performance observation.
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
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	qa "github.com/FloatTech/ZeroBot-Plugin/internal/adminqa"
)

type resourceSample struct {
	AtSeconds       float64 `json:"atSeconds"`
	CPUSeconds      float64 `json:"cpuSeconds"`
	PrivateBytes    uint64  `json:"privateBytes"`
	WorkingSetBytes uint64  `json:"workingSetBytes"`
}

func main() {
	binary := flag.String("binary-name", "zerobot.exe", "executable under .local/bin")
	profile := flag.String("profile", "paused", "paused or updating browser label")
	warmup := flag.Duration("warmup", time.Minute, "synthetic workload warmup")
	duration := flag.Duration("duration", 6*time.Minute, "process sample window; must cover browser window")
	flag.Parse()
	if flag.NArg() != 0 || filepath.Base(*binary) != *binary || strings.ContainsAny(*binary, "/\\:") || filepath.Ext(*binary) != ".exe" || (*profile != "paused" && *profile != "updating") || *warmup < 0 || *warmup > 5*time.Minute || *duration < 10*time.Second || *duration > 15*time.Minute {
		fmt.Fprintln(os.Stderr, "Invalid local binary, profile or timing")
		os.Exit(2)
	}
	if err := run(*binary, *profile, *warmup, *duration); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(binary, profile string, warmup, duration time.Duration) (err error) {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	dir := filepath.Join(root, ".local/browser-performance")
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	out, err := os.MkdirTemp(dir, "run-")
	if err != nil {
		return err
	}
	source, err := os.Open(filepath.Join(root, ".local/bin", binary))
	if err != nil {
		return err
	}
	target, err := os.OpenFile(filepath.Join(out, "zerobot-browser-performance.exe"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
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
	signalCtx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignal()
	ctx, cancel := context.WithTimeout(signalCtx, 45*time.Minute)
	defer cancel()
	result := map[string]any{"profile": profile, "startedAt": time.Now().Format(time.RFC3339Nano), "binary": binary, "binarySHA256": hex.EncodeToString(hash.Sum(nil)), "isolated": true, "realQQMessagesSent": false, "modelCalls": 0, "browserRendererMeasured": false, "warmupSeconds": warmup.Seconds(), "sampleSeconds": duration.Seconds(), "fullTiming": warmup >= time.Minute && duration >= 5*time.Minute, "passed": false, "completed": false}
	var h *qa.Harness
	var bot *qa.Bot
	defer func() {
		if bot != nil {
			bot.Stop()
		}
		if h != nil {
			h.Close()
		}
		result["instanceStopped"] = true
		result["stoppedAt"] = time.Now().Format(time.RFC3339Nano)
		if err != nil {
			result["failure"] = err.Error()
		}
		if e := qa.WriteJSON(filepath.Join(out, "result.json"), result); e != nil && err == nil {
			err = e
		}
		_ = qa.WriteJSON(filepath.Join(out, "session.json"), map[string]any{"state": "stopped", "output": out})
		fmt.Println("ARTIFACTS=" + out)
	}()
	h, err = qa.New(root, filepath.Join(out, "zerobot-browser-performance.exe"), out)
	if err != nil {
		return err
	}
	bot, err = h.Start(ctx, "browser-performance")
	if err != nil {
		return err
	}
	secret := make([]byte, 24)
	if _, err = rand.Read(secret); err != nil {
		return err
	}
	password := "QA-" + hex.EncodeToString(secret)
	if err = os.WriteFile(filepath.Join(out, "login.txt"), []byte("Isolated browser performance password:\n"+password+"\n"), 0600); err != nil {
		return err
	}
	c := qa.NewClient(h.Address)
	if _, err = c.AuthWithPassword(h.AdminDir, password); err != nil {
		return err
	}
	session := map[string]any{"state": "ready", "url": "http://" + h.Address + "/", "output": out, "loginFile": filepath.Join(out, "login.txt"), "profile": profile, "warmupSeconds": warmup.Seconds(), "sampleSeconds": duration.Seconds(), "realQQMessagesSent": false}
	if err = qa.WriteJSON(filepath.Join(out, "session.json"), session); err != nil {
		return err
	}
	encoded, _ := json.Marshal(session)
	fmt.Println(string(encoded))
	fmt.Println("Commands: start, status, stop. Measurements begin only after start.")
	commands := make(chan string, 8)
	inputDone := make(chan error, 1)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			select {
			case commands <- strings.TrimSpace(s.Text()):
			case <-ctx.Done():
				return
			}
		}
		inputDone <- s.Err()
	}()
	var began time.Time
	for began.IsZero() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-inputDone:
			if e != nil {
				return e
			}
			result["stopReason"] = "input-ended"
			return nil
		case command := <-commands:
			switch command {
			case "start":
				began = time.Now()
			case "stop":
				result["stopReason"] = "stopped-before-start"
				return nil
			case "status":
				fmt.Println(string(encoded))
			default:
				fmt.Println("Commands: start, status, stop.")
			}
		}
	}
	result["measurementStartedAt"] = began.Format(time.RFC3339Nano)
	sampleStart := began.Add(warmup)
	sampleEnd := sampleStart.Add(duration)
	result["sampleStartsAt"] = sampleStart.Format(time.RFC3339Nano)
	result["sampleEndsAt"] = sampleEnd.Format(time.RFC3339Nano)
	nextMessage, nextSample, nextProgress := began, sampleStart, began
	messageTimes := []float64{}
	warmMessages := 0
	samples := []resourceSample{}
	defer func() {
		// Keep partial measurements when input ends, a test is stopped, or a read fails.
		result["samples"] = samples
		result["messageDurationsMs"] = messageTimes
		result["messageCount"] = len(messageTimes)
		result["warmupMessageCount"] = warmMessages
	}()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for time.Now().Before(sampleEnd) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-inputDone:
			if e != nil {
				return e
			}
			result["stopReason"] = "input-ended-before-completion"
			return nil
		case command := <-commands:
			if command == "stop" {
				result["stopReason"] = "early-stop"
				return nil
			}
			if command == "start" {
				fmt.Println("Measurement already started")
			}
			if command == "status" {
				fmt.Printf("PROFILE_STATUS=%s elapsed_seconds=%.0f\n", profile, time.Since(began).Seconds())
			}
		case <-ticker.C:
			now := time.Now()
			if !now.Before(sampleEnd) {
				break
			}
			if !bot.Running() {
				return errors.New("owned performance process exited")
			}
			if !now.Before(nextMessage) {
				reply, t, e := h.Send(qa.Owner, "翻转 abc")
				if e != nil || reply != "ɔqɐ" {
					return errors.New("synthetic workload reply failed")
				}
				if now.Before(sampleStart) {
					warmMessages++
				} else {
					messageTimes = append(messageTimes, float64(t)/float64(time.Millisecond))
				}
				nextMessage = nextMessage.Add(5 * time.Second)
			}
			if !now.Before(nextSample) {
				cpu, private, working, e := bot.Metrics()
				if e != nil {
					return e
				}
				samples = append(samples, resourceSample{time.Since(sampleStart).Seconds(), cpu, private, working})
				nextSample = nextSample.Add(2 * time.Second)
			}
			if !now.Before(nextProgress) {
				phase := "warmup"
				if !now.Before(sampleStart) {
					phase = "sampling"
				}
				progress := map[string]any{"profile": profile, "phase": phase, "elapsedSeconds": time.Since(began).Seconds(), "sampleStartsAt": sampleStart.Format(time.RFC3339Nano), "sampleEndsAt": sampleEnd.Format(time.RFC3339Nano), "sampleCount": len(samples), "messages": len(messageTimes), "warmupMessages": warmMessages}
				if e := qa.WriteJSON(filepath.Join(out, "progress.json"), progress); e != nil {
					return e
				}
				fmt.Printf("BROWSER_PROFILE_PROGRESS=%s phase=%s elapsed_seconds=%.0f messages=%d samples=%d\n", profile, phase, time.Since(began).Seconds(), len(messageTimes), len(samples))
				nextProgress = nextProgress.Add(10 * time.Second)
			}
		}
	}
	if len(samples) < 2 || len(messageTimes) == 0 {
		return errors.New("insufficient process workload samples")
	}
	first, last := samples[0], samples[len(samples)-1]
	maxPrivate, maxWorking := first.PrivateBytes, first.WorkingSetBytes
	for _, s := range samples {
		if s.PrivateBytes > maxPrivate {
			maxPrivate = s.PrivateBytes
		}
		if s.WorkingSetBytes > maxWorking {
			maxWorking = s.WorkingSetBytes
		}
	}
	sort.Float64s(messageTimes)
	p95 := messageTimes[int(math.Ceil(float64(len(messageTimes))*.95))-1]
	privateGrowth := float64(maxPrivate-first.PrivateBytes) / (1 << 20)
	cpuPercent := (last.CPUSeconds - first.CPUSeconds) / (last.AtSeconds - first.AtSeconds) / float64(runtime.NumCPU()) * 100
	result["samples"] = samples
	result["messageDurationsMs"] = messageTimes
	result["messageCount"] = len(messageTimes)
	result["warmupMessageCount"] = warmMessages
	result["messageP95Ms"] = p95
	result["privateGrowthMiB"] = privateGrowth
	result["privateLastBytes"] = last.PrivateBytes
	result["workingSetMaxBytes"] = maxWorking
	result["cpuPercentOfMachine"] = cpuPercent
	result["logicalCPUs"] = runtime.NumCPU()
	result["completed"] = true
	result["measurementEndedAt"] = time.Now().Format(time.RFC3339Nano)
	result["thresholdsPassed"] = privateGrowth <= 64 && p95 <= 1000
	result["passed"] = result["thresholdsPassed"]
	result["stopReason"] = "measurement-completed"
	if result["passed"] != true {
		return errors.New("process workload thresholds exceeded")
	}
	fmt.Printf("BROWSER_PROFILE_DONE=%s messages=%d message_p95_ms=%.2f private_growth_mib=%.2f cpu_percent=%.4f\n", profile, len(messageTimes), p95, privateGrowth, cpuPercent)
	return nil
}
