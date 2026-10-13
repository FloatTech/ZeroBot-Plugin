// Package localadmin serves the local owner's admin UI and authenticated APIs.
package localadmin

import (
	"context"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"errors"
	"github.com/FloatTech/ZeroBot-Plugin/internal/modelstats"
	"github.com/FloatTech/ZeroBot-Plugin/internal/processidentity"
	"github.com/FloatTech/ZeroBot-Plugin/internal/replystats"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

//go:embed assets/index.html assets/app.js assets/connections-add.js assets/connections-delete.js assets/overview.js assets/reply-metrics.js assets/model-observation.js assets/ai.js assets/logs.js assets/backups.js assets/security.js assets/theme.js assets/style.css assets/appearance.css
var assets embed.FS

type Overview struct {
	StartedAt             time.Time            `json:"startedAt"`
	CheckedAt             time.Time            `json:"checkedAt"`
	UptimeSeconds         int64                `json:"uptimeSeconds"`
	Version               string               `json:"version"`
	GoVersion             string               `json:"goVersion"`
	ConnectionState       string               `json:"connectionState"`
	ConfiguredConnections int                  `json:"configuredConnections"`
	BotAccounts           []string             `json:"botAccounts"`
	PluginCount           int                  `json:"pluginCount"`
	Alerts                *OverviewAlerts      `json:"alerts"`
	ModelRequests         *modelstats.Snapshot `json:"modelRequests"`
	ReplyTasks            *replystats.Snapshot `json:"replyTasks"`
}

type Plugin struct {
	Name                    string       `json:"name"`
	Brief                   string       `json:"brief"`
	Usage                   string       `json:"usage"`
	DefaultEnabledAtStartup bool         `json:"defaultEnabledAtStartup"`
	State                   *PluginState `json:"state,omitempty"`
}

type Config struct {
	Addr            string
	DataDir         string
	Overview        func() Overview
	Plugins         []Plugin
	Connections     *ConnectionStore
	PluginControl   *PluginStore
	AI              *AIStore
	Logs            *LogStore
	Backups         *BackupStore
	ModelRequests   func() modelstats.Snapshot
	ReplyTasks      func() replystats.Snapshot
	ProcessIdentity *processidentity.Identity
}

type Server struct {
	cfg               Config
	http              *http.Server
	mu                sync.Mutex
	credential        *credential
	setupCode         string
	sessions          map[string]session
	failures          []time.Time
	now               func() time.Time
	replaceCredential func(string, string) error
}

func New(cfg Config) (*Server, error) {
	host, port, err := net.SplitHostPort(cfg.Addr)
	n, parseErr := strconv.Atoi(port)
	if err != nil || parseErr != nil || host != "127.0.0.1" || n < 1 || n > 65535 || strconv.Itoa(n) != port {
		return nil, errors.New("后台地址必须是 127.0.0.1 和有效端口，例如 127.0.0.1:3010")
	}
	if cfg.DataDir == "" || cfg.Overview == nil {
		return nil, errors.New("后台数据目录与状态来源不能为空")
	}
	cfg.DataDir, err = filepath.Abs(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	cfg.Plugins = append([]Plugin{}, cfg.Plugins...)
	s := &Server{cfg: cfg, sessions: make(map[string]session), now: time.Now, replaceCredential: replaceConfigFile}
	if err := s.loadAuth(); err != nil {
		return nil, err
	}
	if cfg.Logs != nil {
		// A failed refresh does not prevent bot startup; log reads fail closed.
		_ = s.refreshLogSecrets()
	}
	return s, nil
}

// Start binds synchronously so port conflicts are reported before bot startup.
func Start(cfg Config) (*Server, error) {
	// Validate before listening or creating local credentials.
	host, port, err := net.SplitHostPort(cfg.Addr)
	n, parseErr := strconv.Atoi(port)
	if err != nil || parseErr != nil || host != "127.0.0.1" || n < 1 || n > 65535 || strconv.Itoa(n) != port {
		return nil, errors.New("后台仅支持 127.0.0.1 和有效端口")
	}
	listener, err := net.Listen("tcp4", cfg.Addr)
	if err != nil {
		return nil, errors.New("后台端口无法监听，请检查是否被占用")
	}
	s, err := New(cfg)
	if err != nil {
		listener.Close()
		return nil, err
	}
	s.http = &http.Server{Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16384}
	go func() { _ = s.http.Serve(listener) }()
	return s, nil
}

func (s *Server) Close(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/session", s.getSession)
	mux.HandleFunc("POST /api/setup", s.setup)
	mux.HandleFunc("POST /api/login", s.login)
	mux.HandleFunc("POST /api/logout", s.authorize(s.logout, true))
	mux.HandleFunc("POST /api/password", s.authorize(s.changePassword, true))
	mux.HandleFunc("GET /api/connections", s.authorize(s.getConnections, false))
	mux.HandleFunc("POST /api/connections", s.authorize(s.saveConnections, true))
	mux.HandleFunc("GET /api/overview", s.authorize(s.getOverview, false))
	mux.HandleFunc("GET /api/plugins", s.authorize(s.getPlugins, false))
	mux.HandleFunc("POST /api/plugin-switch", s.authorize(s.setPluginSwitch, true))
	mux.HandleFunc("GET /api/ai", s.authorize(s.getAI, false))
	mux.HandleFunc("POST /api/ai", s.authorize(s.saveAI, true))
	mux.HandleFunc("POST /api/ai-group", s.authorize(s.saveAIGroup, true))
	mux.HandleFunc("GET /api/logs", s.authorize(s.getLogs, false))
	mux.HandleFunc("GET /api/logs/export", s.authorize(s.exportLogs, false))
	mux.HandleFunc("GET /api/backups", s.authorize(s.getBackups, false))
	mux.HandleFunc("POST /api/backups", s.authorize(s.createBackup, true))
	mux.HandleFunc("POST /api/backups/preview", s.authorize(s.previewBackup, true))
	mux.HandleFunc("POST /api/backups/restore", s.authorize(s.restoreBackup, true))
	mux.HandleFunc("POST /api/backups/cleanup-preview", s.authorize(s.previewBackupCleanup, true))
	mux.HandleFunc("POST /api/backups/cleanup", s.authorize(s.cleanupBackups, true))
	for route, asset := range map[string]string{"/": "index.html", "/app.js": "app.js", "/connections-add.js": "connections-add.js", "/connections-delete.js": "connections-delete.js", "/overview.js": "overview.js", "/reply-metrics.js": "reply-metrics.js", "/model-observation.js": "model-observation.js", "/ai.js": "ai.js", "/logs.js": "logs.js", "/backups.js": "backups.js", "/security.js": "security.js", "/theme.js": "theme.js", "/style.css": "style.css", "/appearance.css": "appearance.css"} {
		pattern := "GET " + route
		if route == "/" {
			pattern = "GET /{$}"
		}
		mux.HandleFunc(pattern, func(w http.ResponseWriter, _ *http.Request) {
			b, err := assets.ReadFile("assets/" + asset)
			if err != nil {
				apiError(w, 500, "页面资源无法读取")
				return
			}
			kind := "text/html; charset=utf-8"
			if strings.HasSuffix(asset, ".js") {
				kind = "text/javascript; charset=utf-8"
			}
			if strings.HasSuffix(asset, ".css") {
				kind = "text/css; charset=utf-8"
			}
			w.Header().Set("Content-Type", kind)
			_, _ = w.Write(b)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || host != "127.0.0.1" || r.Host != s.cfg.Addr {
			apiError(w, 403, "仅允许本机地址访问")
			return
		}
		origin := "http://" + s.cfg.Addr
		if value := r.Header.Get("Origin"); value != "" && value != origin {
			apiError(w, 403, "请求来源不受信任")
			return
		}
		// The root serves only the public UI shell. Browser extensions and links
		// from other sites can open it as a cross-site top-level navigation.
		// Keep APIs, subresources, embedded documents and writes same-origin.
		publicNavigation := r.Method == http.MethodGet && r.URL.Path == "/" &&
			r.Header.Get("Sec-Fetch-Mode") == "navigate" && r.Header.Get("Sec-Fetch-Dest") == "document"
		if value := r.Header.Get("Sec-Fetch-Site"); value != "" && value != "same-origin" && value != "none" &&
			!((value == "cross-site" || value == "same-site") && publicNavigation) {
			apiError(w, 403, "请求来源不受信任")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if r.Header.Get("Origin") != origin || r.Header.Get("X-Local-Request") != "1" || r.Header.Get("Content-Type") != "application/json" {
				apiError(w, 403, "请通过本地后台提交请求")
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func jsonReply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func apiError(w http.ResponseWriter, status int, text string) {
	jsonReply(w, status, map[string]string{"error": text})
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	return readJSONLimit(w, r, dst, 4096)
}

func readJSONLimit(w http.ResponseWriter, r *http.Request, dst any, limit int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		apiError(w, 400, "请求格式无效")
		return false
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		apiError(w, 400, "请求格式无效")
		return false
	}
	return true
}

func (s *Server) currentSession(r *http.Request) (string, session, bool) {
	cookie, err := r.Cookie("local_admin_session")
	if err != nil {
		return "", session{}, false
	}
	value, ok := s.sessions[cookie.Value]
	if !ok || !s.now().Before(value.Expires) {
		delete(s.sessions, cookie.Value)
		return "", session{}, false
	}
	return cookie.Value, value, true
}

func (s *Server) getSession(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, value, ok := s.currentSession(r)
	response := map[string]any{"authenticated": ok, "setupRequired": s.credential == nil}
	if ok {
		response["csrfToken"] = value.CSRF
	}
	jsonReply(w, 200, response)
}

func (s *Server) authorize(next http.HandlerFunc, write bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		_, value, ok := s.currentSession(r)
		s.mu.Unlock()
		if !ok {
			apiError(w, 401, "请先登录本地后台")
			return
		}
		if write && subtle.ConstantTimeCompare([]byte(value.CSRF), []byte(r.Header.Get("X-CSRF-Token"))) != 1 {
			apiError(w, 403, "请求校验失败，请刷新页面")
			return
		}
		next(w, r)
	}
}

func (s *Server) newSession(w http.ResponseWriter) {
	s.pruneSessions()
	if len(s.sessions) >= 32 {
		apiError(w, 429, "登录会话过多，请稍后重试")
		return
	}
	key := randomCode()
	value := session{CSRF: randomCode(), Expires: s.now().Add(8 * time.Hour)}
	s.sessions[key] = value
	http.SetCookie(w, &http.Cookie{Name: "local_admin_session", Value: key, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60})
	jsonReply(w, 200, map[string]any{"authenticated": true, "csrfToken": value.CSRF})
}

func (s *Server) setup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SetupCode string `json:"setupCode"`
		Password  string `json:"password"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.credential != nil {
		apiError(w, 409, "后台密码已经设置，请登录")
		return
	}
	if subtle.ConstantTimeCompare([]byte(s.setupCode), []byte(strings.TrimSpace(body.SetupCode))) != 1 {
		apiError(w, 401, "设置码不正确")
		return
	}
	if utf8.RuneCountInString(body.Password) < 12 || len(body.Password) > 512 {
		apiError(w, 400, "密码至少 12 个字符，且不超过 512 字节")
		return
	}
	if err := s.createCredential(body.Password); err != nil {
		apiError(w, 500, "密码保存失败，请检查后台数据目录权限")
		return
	}
	s.newSession(w)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.credential == nil {
		apiError(w, 409, "请先设置后台密码")
		return
	}
	if s.authRateLimited() {
		apiError(w, 429, "尝试次数过多，请一分钟后重试")
		return
	}
	if len(body.Password) > 512 || !s.validPassword(body.Password) {
		s.failures = append(s.failures, s.now())
		apiError(w, 401, "密码不正确")
		return
	}
	s.failures = nil
	s.newSession(w)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if cookie, err := r.Cookie("local_admin_session"); err == nil {
		delete(s.sessions, cookie.Value)
	}
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "local_admin_session", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	jsonReply(w, 200, map[string]bool{"authenticated": false})
}
