package localadmin

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
)

const logCapacity = 1000
const logByteCapacity = 1 << 20
const logMessageLimit = 2048

var logURL = regexp.MustCompile(`(?i)(?:https?|wss?)://[^\s<>"']+`)
var logHeader = regexp.MustCompile(`(?im)(?:authorization|proxy-authorization|set-cookie|cookie)["']?\s*[:=]\s*[^\r\n]*`)
var logSecret = regexp.MustCompile(`(?i)(?:access[_-]?token|api[_-]?key|imagekey|agentkey|token|key|password|passwd|secret|credential|setup[_-]?code|hash)["']?\s*[:=]\s*(?:"(?:\\.|[^"\\])*"|'[^']*'|[^\s,;\]}]+)`)
var logKeyCommand = regexp.MustCompile(`设置AI聊天(?:识图|Agent)?密钥.*`)
var logBearer = regexp.MustCompile(`(?i)\b(?:bearer|basic)\s+[a-z0-9_+/=.-]+`)

type LogEntry struct {
	ID        string    `json:"id"`
	Time      time.Time `json:"time"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Truncated bool      `json:"truncated"`
}

// LogStore is a bounded in-memory logrus hook. It neither changes the logger's
// output nor reads databases/files from Fire (control writers can log under lock).
type LogStore struct {
	mu          sync.Mutex
	instance    string
	next        uint64
	rows        [logCapacity]LogEntry
	start       int
	count       int
	bytes       int
	secrets     map[string]struct{}
	maskAll     bool
	replacer    *strings.Replacer
	dirty       bool
	maskVersion uint64
	readMasks   [logCapacity]logReadMask
}

type logReadMask struct {
	id, raw, message, lower string
	version                 uint64
	maskAll, truncated      bool
}

func NewLogStore() *LogStore {
	return &LogStore{instance: hex.EncodeToString(randomBytes(8)), secrets: map[string]struct{}{}}
}
func (l *LogStore) Levels() []logrus.Level { return logrus.AllLevels }

func (l *LogStore) ProtectConfig(raw []byte) { l.AddSecrets(secretsFromConfig(raw)...) }

// RedactingFormatter keeps the existing output destination and formatter style,
// while protecting subsequent standard logrus output as well as the web view.
type RedactingFormatter struct {
	Logs *LogStore
	Next logrus.Formatter
}

func (f RedactingFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	copy := *entry
	f.Logs.mu.Lock()
	copy.Message = f.Logs.maskLocked(entry.Message)
	copy.Message, _ = clipLogMessage(copy.Message)
	copy.Data = logrus.Fields{}
	for key, value := range entry.Data {
		if len(copy.Data) >= 16 {
			break
		}
		text := fmt.Sprint(value)
		if len(text) > 65536 {
			text = "[超长字段已省略]"
		}
		safeKey := f.Logs.maskLocked(key)
		if logSecret.MatchString(key+"=x") || logHeader.MatchString(key+"=x") {
			copy.Data[safeKey] = "[已隐藏]"
		} else {
			copy.Data[safeKey], _ = clipLogMessage(f.Logs.maskLocked(text))
		}
	}
	f.Logs.mu.Unlock()
	return f.Next.Format(&copy)
}

func (l *LogStore) AddSecrets(values ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, value := range values {
		if value == "" {
			continue
		}
		encoded, _ := json.Marshal(value)
		for _, variant := range []string{value, url.QueryEscape(value), url.PathEscape(value), string(encoded[1 : len(encoded)-1])} {
			if _, exists := l.secrets[variant]; exists {
				continue
			}
			// Fail closed instead of growing a secret history without a bound.
			if len(variant) > 16384 || len(l.secrets) >= 256 {
				l.maskAll = true
				return
			}
			l.secrets[variant] = struct{}{}
			l.dirty = true
			l.maskVersion++
		}
	}
}

func genericLogMask(text string) string {
	// Message bodies can contain credentials before a chat setter commits them.
	if strings.Contains(text, "[bot]") && strings.Contains(text, "收到") && strings.Contains(text, "消息") {
		return "[bot] 收到消息（正文已隐藏）"
	}
	if strings.Contains(text, "[api]") && strings.Contains(text, "发送") && strings.Contains(text, "消息") {
		return "[api] 发送消息（正文已隐藏）"
	}
	text = logKeyCommand.ReplaceAllString(text, "设置AI聊天密钥[已隐藏]")
	text = logURL.ReplaceAllStringFunc(text, func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" {
			return "[地址已隐藏]"
		}
		if u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
			u.User, u.RawQuery, u.Fragment, u.ForceQuery = nil, "", "", false
			return u.String() + "[认证参数已隐藏]"
		}
		return raw
	})
	text = logHeader.ReplaceAllString(text, "[认证头已隐藏]")
	text = logSecret.ReplaceAllString(text, "[凭据已隐藏]")
	text = logBearer.ReplaceAllString(text, "[认证值已隐藏]")
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, strings.ToValidUTF8(text, "�"))
}

func (l *LogStore) maskLocked(text string) string {
	if l.maskAll {
		return "[日志已隐藏：密钥历史达到上限，请重启后台]"
	}
	if len(text) > 65536 {
		return "[超长日志已省略]"
	}
	// A simultaneous replacer prevents overlapping keys or replacements from
	// exposing suffixes or repeatedly rewriting the redaction marker.
	if l.dirty {
		keys := make([]string, 0, len(l.secrets))
		for key := range l.secrets {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
		pairs := make([]string, 0, len(keys)*2)
		for _, key := range keys {
			pairs = append(pairs, key, "[已隐藏]")
		}
		l.replacer = strings.NewReplacer(pairs...)
		l.dirty = false
	}
	if l.replacer != nil {
		text = l.replacer.Replace(text)
	}
	return genericLogMask(text)
}

func (l *LogStore) evictLocked() {
	l.bytes -= len(l.rows[l.start].Message)
	l.rows[l.start] = LogEntry{}
	l.readMasks[l.start] = logReadMask{}
	l.start = (l.start + 1) % logCapacity
	l.count--
}

func (l *LogStore) Fire(entry *logrus.Entry) error {
	text := entry.Message
	if len(text) > 65536 {
		text = "[超长日志已省略]"
	} else {
		text = genericLogMask(text)
	}
	keys := make([]string, 0, len(entry.Data))
	for key := range entry.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		if i >= 16 {
			text += "\n[其余字段已省略]"
			break
		}
		if logSecret.MatchString(key+"=x") || strings.EqualFold(key, "Authorization") || strings.EqualFold(key, "Cookie") {
			text += "\n[凭据字段已隐藏]"
			continue
		}
		value := fmt.Sprint(entry.Data[key])
		if len(key)+len(value) > 4096 {
			text += "\n[超长字段已省略]"
			continue
		}
		text += "\n" + genericLogMask(key+"="+value)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	text = l.maskLocked(text)
	text, truncated := clipLogMessage(text)
	for l.count > 0 && (l.count == logCapacity || l.bytes+len(text) > logByteCapacity) {
		l.evictLocked()
	}
	l.next++
	row := LogEntry{ID: strconv.FormatUint(l.next, 10), Time: entry.Time, Level: entry.Level.String(), Message: text, Truncated: truncated}
	l.rows[(l.start+l.count)%logCapacity] = row
	l.readMasks[(l.start+l.count)%logCapacity] = logReadMask{}
	l.count++
	l.bytes += len(text)
	return nil
}

func clipLogMessage(text string) (string, bool) {
	if len(text) > logMessageLimit {
		end := logMessageLimit
		for !utf8.ValidString(text[:end]) {
			end--
		}
		return text[:end] + "…[已截断]", true
	}
	return text, false
}

type logQuery struct {
	after    uint64
	instance string
	level    string
	keyword  string
	from     *time.Time
	until    *time.Time
	limit    int
}
type LogPage struct {
	Entries  []LogEntry `json:"entries"`
	Cursor   string     `json:"cursor"`
	Instance string     `json:"instance"`
	Retained int        `json:"retained"`
	Capacity int        `json:"capacity"`
	Gap      bool       `json:"gap"`
	Reset    bool       `json:"reset"`
	Evicted  bool       `json:"evicted"`
	Summary  LogSummary `json:"summary"`
}

type LogSummary struct {
	Matched int            `json:"matched"`
	Levels  map[string]int `json:"levels"`
}

func parseLogQuery(v url.Values) (logQuery, error) {
	q := logQuery{instance: v.Get("instance"), level: v.Get("level"), keyword: v.Get("q"), limit: 200}
	if text := v.Get("after"); text != "" {
		n, err := strconv.ParseUint(text, 10, 64)
		if err != nil || strconv.FormatUint(n, 10) != text {
			return q, errors.New("日志游标无效")
		}
		q.after = n
	}
	if text := v.Get("limit"); text != "" {
		n, err := strconv.Atoi(text)
		if err != nil || n < 1 || n > 200 {
			return q, errors.New("每次最多读取 200 条日志")
		}
		q.limit = n
	}
	if q.instance != "" {
		b, err := hex.DecodeString(q.instance)
		if err != nil || len(b) != 8 {
			return q, errors.New("日志实例无效")
		}
	}
	q.level = strings.ToLower(q.level)
	if q.level == "warn" {
		q.level = "warning"
	}
	if q.level != "" && q.level != "issues" {
		if _, err := logrus.ParseLevel(q.level); err != nil {
			return q, errors.New("日志级别无效")
		}
	}
	if len(q.keyword) > 256 || !utf8.ValidString(q.keyword) || strings.ContainsRune(q.keyword, 0) {
		return q, errors.New("日志关键词过长或包含无效字符")
	}
	for name, target := range map[string]**time.Time{"from": &q.from, "until": &q.until} {
		items := v[name]
		if len(items) > 1 {
			return q, errors.New("日志时间条件不可重复")
		}
		if len(items) == 0 || items[0] == "" {
			continue
		}
		if len(items[0]) > 64 {
			return q, errors.New("日志时间无效，需包含时区")
		}
		value, err := time.Parse(time.RFC3339Nano, items[0])
		if err != nil || value.Year() < 1 || value.Year() > 9999 {
			return q, errors.New("日志时间无效，需包含时区")
		}
		*target = &value
	}
	if q.from != nil && q.until != nil && q.from.After(*q.until) {
		return q, errors.New("开始时间不能晚于结束时间")
	}
	return q, nil
}

func matchesLogLevel(level, filter string) bool {
	if filter == "issues" {
		return level == "warning" || level == "error" || level == "fatal" || level == "panic"
	}
	return filter == "" || level == filter
}

func (l *LogStore) page(q logQuery) LogPage {
	l.mu.Lock()
	defer l.mu.Unlock()
	p := LogPage{Entries: []LogEntry{}, Cursor: strconv.FormatUint(l.next, 10), Instance: l.instance, Retained: l.count, Capacity: logCapacity, Evicted: l.next > uint64(l.count)}
	p.Summary.Levels = map[string]int{"panic": 0, "fatal": 0, "error": 0, "warning": 0, "info": 0, "debug": 0, "trace": 0, "other": 0}
	if (q.instance != "" && q.instance != l.instance) || q.after > l.next {
		p.Reset, q.after = true, 0
	}
	if l.count > 0 && q.after > 0 {
		oldest, _ := strconv.ParseUint(l.rows[l.start].ID, 10, 64)
		p.Gap = q.after < oldest-1
	}
	needle := strings.ToLower(q.keyword)
	for i := 0; i < l.count; i++ {
		index := (l.start + i) % logCapacity
		row := l.rows[index]
		id, _ := strconv.ParseUint(row.ID, 10, 64)
		if (q.from != nil && row.Time.Before(*q.from)) || (q.until != nil && row.Time.After(*q.until)) || !matchesLogLevel(row.Level, q.level) {
			continue
		}
		returned := id > q.after && (q.after == 0 || len(p.Entries) < q.limit)
		// Keyword matches and returned messages always use current redaction.
		// With no keyword, older rows only need level/time metadata for counts.
		if returned || needle != "" {
			cached := &l.readMasks[index]
			if cached.id != row.ID || cached.raw != row.Message || cached.version != l.maskVersion || cached.maskAll != l.maskAll {
				message, clipped := clipLogMessage(l.maskLocked(row.Message))
				*cached = logReadMask{id: row.ID, raw: row.Message, message: message, version: l.maskVersion, maskAll: l.maskAll, truncated: row.Truncated || clipped}
			}
			row.Message, row.Truncated = cached.message, cached.truncated
			if needle != "" {
				if cached.lower == "" {
					cached.lower = strings.ToLower(cached.message)
				}
				if !strings.Contains(cached.lower, needle) {
					continue
				}
			}
		}
		p.Summary.Matched++
		level := row.Level
		if _, ok := p.Summary.Levels[level]; !ok {
			level = "other"
		}
		p.Summary.Levels[level]++
		if !returned {
			continue
		}
		p.Entries = append(p.Entries, row)
		if q.after > 0 && len(p.Entries) == q.limit {
			p.Cursor = row.ID
		}
	}
	if q.after == 0 && len(p.Entries) > q.limit {
		p.Entries = p.Entries[len(p.Entries)-q.limit:]
	}
	return p
}

func secretsFromConfig(raw []byte) []string {
	var root any
	if json.Unmarshal(raw, &root) != nil {
		return nil
	}
	values := []string{}
	var visit func(any)
	visit = func(node any) {
		switch n := node.(type) {
		case map[string]any:
			for name, value := range n {
				if text, ok := value.(string); ok {
					if logSecret.MatchString(name + "=x") {
						values = append(values, text)
					}
					if u, err := url.Parse(text); err == nil && u.Hostname() != "" {
						if u.User != nil {
							if password, ok := u.User.Password(); ok {
								values = append(values, password)
							}
						}
						for key, items := range u.Query() {
							if logSecret.MatchString(key + "=x") {
								values = append(values, items...)
							}
						}
					}
				}
				visit(value)
			}
		case []any:
			for _, value := range n {
				visit(value)
			}
		}
	}
	visit(root)
	return values
}

// refreshLogSecrets is called outside the hook/buffer lock. It fails closed if
// live configuration cannot be read, rather than returning unsanitized logs.
func (s *Server) refreshLogSecrets() error {
	values := []string{}
	if c := s.cfg.Connections; c != nil {
		c.mu.Lock()
		raw, err := c.read()
		if err == nil {
			values = append(values, secretsFromConfig(raw)...)
			values = append(values, secretsFromConfig(c.running)...)
		}
		c.mu.Unlock()
		if err != nil {
			return err
		}
	}
	if ai := s.cfg.AI; ai != nil && ai.repository != nil {
		c, _, err := ai.repository.Read()
		if err != nil {
			return err
		}
		raw, _ := json.Marshal(c)
		values = append(values, secretsFromConfig(raw)...)
	}
	s.cfg.Logs.AddSecrets(values...)
	return nil
}

func (s *Server) getLogs(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Logs == nil {
		apiError(w, 409, "当前程序未接入日志查看")
		return
	}
	q, err := parseLogQuery(r.URL.Query())
	if err != nil {
		apiError(w, 400, err.Error())
		return
	}
	if r.Context().Err() != nil {
		return
	}
	if err := s.refreshLogSecrets(); err != nil {
		apiError(w, 503, "日志暂不可读取，请检查配置数据库与启动配置")
		return
	}
	jsonReply(w, 200, s.cfg.Logs.page(q))
}

// Export uses the same read-time secret refresh and retained buffer as the viewer.
// It never reads historical stderr files or accepts paths/cursors from the caller.
func (s *Server) exportLogs(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Logs == nil {
		apiError(w, 409, "当前程序未接入日志查看")
		return
	}
	values := r.URL.Query()
	for name, items := range values {
		if (name != "level" && name != "q" && name != "from" && name != "until") || len(items) != 1 {
			apiError(w, 400, "导出仅支持级别、关键词与起止时间筛选")
			return
		}
	}
	q, err := parseLogQuery(values)
	if err != nil {
		apiError(w, 400, err.Error())
		return
	}
	if r.Context().Err() != nil {
		return
	}
	if err := s.refreshLogSecrets(); err != nil {
		apiError(w, 503, "日志暂不可读取，请检查配置数据库与启动配置")
		return
	}
	q.limit = logCapacity
	page := s.cfg.Logs.page(q)
	now := time.Now().UTC()
	var body strings.Builder
	fmt.Fprintf(&body, "ZeroBot 本次启动的脱敏日志\n导出时间：%s\n保留：%d / %d 条；符合已应用筛选：%d 条\n", now.Format(time.RFC3339), page.Retained, page.Capacity, len(page.Entries))
	from, until := "不限", "不限"
	if q.from != nil {
		from = q.from.UTC().Format(time.RFC3339Nano)
	}
	if q.until != nil {
		until = q.until.UTC().Format(time.RFC3339Nano)
	}
	fmt.Fprintf(&body, "时间范围（包含边界，UTC）：%s 至 %s\n", from, until)
	if page.Evicted {
		body.WriteString("部分旧日志已超出条数或容量限制，本文件不包含这些记录。\n")
	}
	body.WriteString("仅包含内存中仍保留的日志；重启前的日志不包含在内。长记录可能截断，密钥与消息正文已隐藏。\n正文使用引号包围，换行以 \\n 表示，每条记录占一行。\n\n")
	for _, row := range page.Entries {
		fmt.Fprintf(&body, "[%s] [%s] #%s %s\n", row.Time.UTC().Format(time.RFC3339Nano), strings.ToUpper(row.Level), row.ID, strconv.Quote(row.Message))
	}
	if r.Context().Err() != nil {
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="zerobot-logs-`+now.Format("20060102T150405Z")+`.txt"`)
	w.Header().Set("Content-Length", strconv.Itoa(body.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body.String()))
}
