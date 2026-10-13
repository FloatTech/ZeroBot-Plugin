// Package adminqa supplies isolated Windows processes for local admin acceptance checks.
package adminqa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/RomiChan/websocket"
	"golang.org/x/sys/windows"
)

const Owner = int64(12345678)
const Password = "Isolated-admin-acceptance-only"

type peer struct {
	conn *websocket.Conn
	mu   sync.Mutex
}
type Harness struct {
	Root, Binary, ConfigPath, AdminDir, Address string
	server                                      *http.Server
	listener                                    net.Listener
	connections                                 chan *peer
	replies                                     chan string
	current                                     *peer
	messageID                                   int
	mockQQ                                      int64
}

func WriteJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func FreeAddress() (string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return l.Addr().String(), nil
}

func PrepareFixture(root, destination string) error {
	for _, relative := range []string{"data/Control/kanban.png", "data/Control/zbpuwu.png", "data/Font/GlowSansSC-Normal-ExtraBold.ttf", "data/Font/Impact.ttf", "data/Font/Torus-Regular.otf", "data/Font/FOT-Yuruka Std.ttf", "data/Chat/dict.txt", "data/hyaku/小倉百人一首.csv", "data/Dish/dishes.db", "data/Pig/pig_data.json", "data/McFish/articlesInfo.json"} {
		source, err := os.Open(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return err
		}
		path := filepath.Join(destination, filepath.FromSlash(relative))
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			source.Close()
			return err
		}
		target, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			source.Close()
			return err
		}
		_, err = io.Copy(target, source)
		closeErr := target.Close()
		source.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func New(root, binary, destination string) (*Harness, error) {
	return NewWithIdentity(root, binary, destination, 87654321)
}

// NewWithIdentity prepares a separate synthetic OneBot account for multi-connection checks.
func NewWithIdentity(root, binary, destination string, mockQQ int64) (*Harness, error) {
	if mockQQ <= 0 {
		return nil, fmt.Errorf("synthetic account must be positive")
	}
	if err := PrepareFixture(root, destination); err != nil {
		return nil, err
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	address, err := FreeAddress()
	if err != nil {
		l.Close()
		return nil, err
	}
	h := &Harness{Root: destination, Binary: binary, ConfigPath: filepath.Join(destination, "config.json"), AdminDir: filepath.Join(destination, "admin"), Address: address, listener: l, connections: make(chan *peer, 8), replies: make(chan string, 64), mockQQ: mockQQ}
	if err = os.MkdirAll(h.AdminDir, 0700); err != nil {
		l.Close()
		return nil, err
	}
	config := map[string]any{"zero": map[string]any{"nickname": []string{"本机测试"}, "command_prefix": "/", "super_users": []int64{Owner}, "ring_len": 4096, "latency": 233000000, "max_process_time": 240000000000, "mark_message": true}, "ws": []map[string]any{{"Url": "ws://" + l.Addr().String(), "AccessToken": "isolated-lifecycle-token"}}, "wss": nil, "force_base64_file": false}
	if err = WriteJSON(h.ConfigPath, config); err != nil {
		l.Close()
		return nil, err
	}
	if err = h.Options(true, address); err != nil {
		l.Close()
		return nil, err
	}
	// Only synthetic configuration tables are initialized. No live database is read.
	m := ctrl.NewManager[any](filepath.ToSlash(filepath.Join(destination, "data/control/plugins.db")))
	if err = m.Response(0); err == nil {
		for _, name := range []string{"aichat", "chat", "atri", "thesaurus"} {
			c := m.NewControl(name, &ctrl.Options[any]{})
			if err = c.Disable(0); err != nil {
				break
			}
		}
	}
	closeErr := m.D.Close()
	if err != nil {
		l.Close()
		return nil, err
	}
	if closeErr != nil {
		l.Close()
		return nil, closeErr
	}
	upgrader := websocket.Upgrader{}
	h.server = &http.Server{ReadHeaderTimeout: 5 * time.Second}
	h.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer isolated-lifecycle-token" {
			http.Error(w, "unauthorized", 401)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		p := &peer{conn: conn}
		p.mu.Lock()
		err = conn.WriteJSON(map[string]any{"self_id": h.mockQQ, "post_type": "meta_event", "meta_event_type": "lifecycle", "sub_type": "connect", "time": time.Now().Unix()})
		p.mu.Unlock()
		if err != nil {
			return
		}
		select {
		case h.connections <- p:
		default:
			return
		}
		for {
			var request struct {
				Action string          `json:"action"`
				Echo   json.RawMessage `json:"echo"`
				Params struct {
					Message json.RawMessage `json:"message"`
				} `json:"params"`
			}
			if conn.ReadJSON(&request) != nil {
				return
			}
			p.mu.Lock()
			userID := Owner
			if request.Action == "get_login_info" {
				userID = h.mockQQ
			}
			err = conn.WriteJSON(map[string]any{"status": "ok", "retcode": 0, "data": map[string]any{"message_id": 999, "user_id": userID, "nickname": "Synthetic acceptance user"}, "echo": request.Echo})
			p.mu.Unlock()
			if err != nil {
				return
			}
			if request.Action == "send_private_msg" || request.Action == "send_group_msg" {
				var segments []struct {
					Type string            `json:"type"`
					Data map[string]string `json:"data"`
				}
				_ = json.Unmarshal(request.Params.Message, &segments)
				var text strings.Builder
				for _, segment := range segments {
					if segment.Type == "text" {
						text.WriteString(segment.Data["text"])
					}
				}
				select {
				case h.replies <- text.String():
				default:
				}
			}
		}
	})
	go h.server.Serve(l)
	return h, nil
}

func (h *Harness) Close() {
	if h.current != nil {
		h.current.conn.Close()
	}
	h.server.Close()
}

// MockURL identifies this fixture's synthetic OneBot endpoint.
func (h *Harness) MockURL() string { return "ws://" + h.listener.Addr().String() }

// WaitConnection attaches an extra fixture endpoint to the already started test process.
func (h *Harness) WaitConnection(ctx context.Context) error {
	select {
	case h.current = <-h.connections:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (h *Harness) Options(enabled bool, address string) error {
	return WriteJSON(filepath.Join(h.AdminDir, "options.json"), map[string]any{"enabled": enabled, "address": address})
}

type Bot struct {
	cmd    *exec.Cmd
	done   chan error
	handle windows.Handle
	logs   []*os.File
	stop   sync.Once
}

// Running checks only this fixture's owned process handle.
func (b *Bot) Running() bool {
	if b == nil || b.handle == 0 {
		return false
	}
	var code uint32
	return windows.GetExitCodeProcess(b.handle, &code) == nil && code == 259
}

func (b *Bot) Stop() {
	if b == nil {
		return
	}
	b.stop.Do(func() {
		_ = b.cmd.Process.Kill()
		<-b.done
		windows.CloseHandle(b.handle)
		for _, f := range b.logs {
			f.Close()
		}
	})
}

func (h *Harness) Start(ctx context.Context, label string) (*Bot, error) {
	var options struct {
		Enabled bool   `json:"enabled"`
		Address string `json:"address"`
	}
	raw, err := os.ReadFile(filepath.Join(h.AdminDir, "options.json"))
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(raw, &options); err != nil {
		return nil, err
	}
	args := []string{"-c", h.ConfigPath}
	if options.Enabled {
		args = append(args, "-admin", "-admin-addr", options.Address, "-admin-data", h.AdminDir)
	}
	cmd := exec.CommandContext(ctx, h.Binary, args...)
	cmd.Dir = h.Root
	cmd.Env = append(os.Environ(), "DEBUG_MODE=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	b := &Bot{cmd: cmd, done: make(chan error, 1)}
	for _, stream := range []string{"stdout", "stderr"} {
		f, err := os.Create(filepath.Join(h.Root, label+"-"+stream+".log"))
		if err != nil {
			for _, old := range b.logs {
				old.Close()
			}
			return nil, err
		}
		b.logs = append(b.logs, f)
	}
	cmd.Stdout = b.logs[0]
	cmd.Stderr = b.logs[1]
	if err = cmd.Start(); err != nil {
		for _, f := range b.logs {
			f.Close()
		}
		return nil, err
	}
	go func() { b.done <- cmd.Wait() }()
	b.handle, err = windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, uint32(cmd.Process.Pid))
	if err != nil {
		_ = cmd.Process.Kill()
		<-b.done
		for _, f := range b.logs {
			f.Close()
		}
		return nil, err
	}
	select {
	case h.current = <-h.connections:
	case <-ctx.Done():
		b.Stop()
		return nil, ctx.Err()
	case <-time.After(45 * time.Second):
		b.Stop()
		return nil, fmt.Errorf("isolated bot handshake timed out")
	}
	if options.Enabled {
		h.Address = options.Address
		client := &http.Client{Timeout: time.Second}
		ready := false
		for i := 0; i < 50; i++ {
			rsp, e := client.Get("http://" + h.Address + "/api/session")
			if e == nil {
				rsp.Body.Close()
				if rsp.StatusCode == 200 {
					ready = true
					break
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		if !ready {
			b.Stop()
			return nil, fmt.Errorf("isolated admin did not become ready")
		}
	}
	time.Sleep(300 * time.Millisecond)
	return b, nil
}

func (h *Harness) Send(uid int64, text string) (string, time.Duration, error) {
	h.messageID++
	start := time.Now()
	p := h.current
	p.mu.Lock()
	err := p.conn.WriteJSON(map[string]any{"time": time.Now().Unix(), "self_id": h.mockQQ, "post_type": "message", "message_type": "private", "sub_type": "friend", "message_id": h.messageID, "user_id": uid, "raw_message": text, "message": []map[string]any{{"type": "text", "data": map[string]string{"text": text}}}, "sender": map[string]any{"user_id": uid, "nickname": "Synthetic acceptance user", "role": "member"}})
	p.mu.Unlock()
	if err != nil {
		return "", time.Since(start), err
	}
	select {
	case reply := <-h.replies:
		return reply, time.Since(start), nil
	case <-time.After(3 * time.Second):
		return "", time.Since(start), nil
	}
}

type Client struct {
	Base string
	HTTP *http.Client
	CSRF string
}

func NewClient(address string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{Base: "http://" + address, HTTP: &http.Client{Jar: jar, Timeout: 10 * time.Second}}
}
func (c *Client) Request(method, path string, body any) (int, map[string]any, time.Duration, error) {
	var raw []byte
	var err error
	if body != nil {
		raw, err = json.Marshal(body)
		if err != nil {
			return 0, nil, 0, err
		}
	}
	req, err := http.NewRequest(method, c.Base+path, bytes.NewReader(raw))
	if err != nil {
		return 0, nil, 0, err
	}
	if method == "POST" {
		req.Header.Set("Origin", c.Base)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Local-Request", "1")
		req.Header.Set("X-CSRF-Token", c.CSRF)
	}
	start := time.Now()
	rsp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, nil, time.Since(start), err
	}
	defer rsp.Body.Close()
	var value map[string]any
	err = json.NewDecoder(io.LimitReader(rsp.Body, 2<<20)).Decode(&value)
	return rsp.StatusCode, value, time.Since(start), err
}
func (c *Client) Auth(adminDir string) (time.Duration, error) {
	return c.AuthWithPassword(adminDir, Password)
}

// AuthWithPassword keeps interactive fixture passwords separate from test defaults.
func (c *Client) AuthWithPassword(adminDir, password string) (time.Duration, error) {
	code, err := os.ReadFile(filepath.Join(adminDir, "setup-code.txt"))
	path := "/api/login"
	body := map[string]string{"password": password}
	if err == nil {
		path = "/api/setup"
		body["setupCode"] = strings.TrimSpace(string(code))
	} else if !os.IsNotExist(err) {
		return 0, err
	}
	status, value, duration, err := c.Request("POST", path, body)
	if err != nil || status != 200 {
		return duration, fmt.Errorf("isolated admin authentication failed: status %d", status)
	}
	c.CSRF, _ = value["csrfToken"].(string)
	if c.CSRF == "" {
		return duration, fmt.Errorf("missing CSRF")
	}
	return duration, nil
}
func (c *Client) Checked(method, path string, body any) (map[string]any, error) {
	status, value, _, err := c.Request(method, path, body)
	if err != nil || status != 200 {
		return nil, fmt.Errorf("isolated API %s failed: status %d", path, status)
	}
	return value, nil
}

type Memory struct {
	Size                                                                                                                                                                                 uint32
	PageFaultCount                                                                                                                                                                       uint32
	PeakWorkingSetSize, WorkingSetSize, QuotaPeakPagedPoolUsage, QuotaPagedPoolUsage, QuotaPeakNonPagedPoolUsage, QuotaNonPagedPoolUsage, PagefileUsage, PeakPagefileUsage, PrivateUsage uintptr
}

var memoryProc = windows.NewLazySystemDLL("psapi.dll").NewProc("GetProcessMemoryInfo")

func (b *Bot) Metrics() (float64, uint64, uint64, error) {
	var create, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(b.handle, &create, &exit, &kernel, &user); err != nil {
		return 0, 0, 0, err
	}
	m := Memory{Size: uint32(unsafe.Sizeof(Memory{}))}
	ok, _, err := memoryProc.Call(uintptr(b.handle), uintptr(unsafe.Pointer(&m)), uintptr(m.Size))
	if ok == 0 {
		return 0, 0, 0, err
	}
	seconds := float64((uint64(kernel.HighDateTime)<<32|uint64(kernel.LowDateTime))+(uint64(user.HighDateTime)<<32|uint64(user.LowDateTime))) / 1e7
	return seconds, uint64(m.PrivateUsage), uint64(m.WorkingSetSize), nil
}
