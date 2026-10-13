package localadmin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
)

func logFixture() (*LogStore, *logrus.Logger) {
	l := NewLogStore()
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)
	logger.SetOutput(&bytes.Buffer{})
	logger.AddHook(l)
	return l, logger
}

func TestLogExportSnapshotBoundsFiltersAndEncoding(t *testing.T) {
	l, logger := logFixture()
	cfg := testConfig(t)
	cfg.Logs = l
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, request(s, "GET", "/api/logs/export", nil, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	for i := 0; i < 1100; i++ {
		logger.Infof("record-%04d", i)
	}
	w := request(s, "GET", "/api/logs/export", nil, cookie, "")
	expectStatus(t, w, 200)
	body := w.Body.String()
	if strings.Count(body, "] [INFO] #") != 1000 || strings.Contains(body, "record-0099") || !strings.Contains(body, "record-0100") || strings.Index(body, "record-0100") > strings.Index(body, "record-1099") || !strings.Contains(body, "部分旧日志已超出") {
		t.Fatal("export did not snapshot all retained rows in order")
	}
	if w.Header().Get("Content-Type") != "text/plain; charset=utf-8" || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("X-Content-Type-Options") != "nosniff" || !regexp.MustCompile(`^attachment; filename="zerobot-logs-\d{8}T\d{6}Z\.txt"$`).MatchString(w.Header().Get("Content-Disposition")) {
		t.Fatal("unsafe download headers")
	}
	logger.Warn("测试 Needle\n[forged record] <script>alert(1)</script>")
	logger.Info("other Needle")
	w = request(s, "GET", "/api/logs/export?level=WARN&q=needle", nil, cookie, "")
	expectStatus(t, w, 200)
	if strings.Count(w.Body.String(), "] [WARNING] #") != 1 || strings.Contains(w.Body.String(), "other Needle") || !strings.Contains(w.Body.String(), `Needle\n[forged record]`) || !utf8.Valid(w.Body.Bytes()) {
		t.Fatal("filter/Unicode/multiline export failed")
	}
	for _, query := range []string{"after=0", "instance=", "limit=1000", "path=stderr.log", "q=a&q=b", "level=unknown", "q=" + strings.Repeat("x", 257), "q=%00", "q=%FF"} {
		w = request(s, "GET", "/api/logs/export?"+query, nil, cookie, "")
		expectStatus(t, w, 400)
		if w.Header().Get("Content-Disposition") != "" {
			t.Fatal("error became a download")
		}
	}
	for i := 0; i < 1100; i++ {
		logger.Info(strings.Repeat("汉🙂\\\n", 600))
	}
	w = request(s, "GET", "/api/logs/export", nil, cookie, "")
	expectStatus(t, w, 200)
	if !utf8.Valid(w.Body.Bytes()) || w.Body.Len() > 3<<20 || !strings.Contains(w.Body.String(), "已截断") {
		t.Fatal("export bytes or truncation exceeded bounds")
	}
	expectStatus(t, request(s, "POST", "/api/logout", map[string]any{}, cookie, csrf), 200)
	expectStatus(t, request(s, "GET", "/api/logs/export", nil, cookie, ""), 401)
}

func TestLogExportReadTimeSecretsAndFailClosed(t *testing.T) {
	l, logger := logFixture()
	cfg := testConfig(t)
	cfg.Logs = l
	ai, c, _, _ := aiFixture(t)
	cfg.AI = ai
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cookie, _ := enroll(t, s)
	logger.Info("new-export-credential")
	logger.Info("[api] 发送群消息(101): secret message body")
	if err := c.SetExtra(map[string]any{"Key": "new-export-credential"}); err != nil {
		t.Fatal(err)
	}
	w := request(s, "GET", "/api/logs/export?q=new-export-credential", nil, cookie, "")
	expectStatus(t, w, 200)
	if strings.Contains(w.Body.String(), "new-export-credential") {
		t.Fatal("query or old record exposed new credential")
	}
	w = request(s, "GET", "/api/logs/export", nil, cookie, "")
	expectStatus(t, w, 200)
	if strings.Contains(w.Body.String(), "secret message body") || strings.Contains(w.Body.String(), "new-export-credential") || !strings.Contains(w.Body.String(), "正文已隐藏") {
		t.Fatal("export bypassed redaction")
	}
	if err := c.Manager.D.Close(); err != nil {
		t.Fatal(err)
	}
	w = request(s, "GET", "/api/logs/export", nil, cookie, "")
	expectStatus(t, w, 503)
	if w.Header().Get("Content-Disposition") != "" || strings.Contains(w.Body.String(), "credential") {
		t.Fatal("failed export returned sensitive file")
	}
	s.cfg.Logs = nil
	expectStatus(t, request(s, "GET", "/api/logs/export", nil, cookie, ""), 409)
}

func TestLogRedactionAndExistingOutput(t *testing.T) {
	l, logger := logFixture()
	output := &bytes.Buffer{}
	logger.SetOutput(output)
	logger.SetFormatter(RedactingFormatter{Logs: l, Next: &logrus.TextFormatter{DisableColors: true}})
	l.AddSecrets("fake-token-long", "fake-token", "fake/key+quoted\"secret")
	logger.WithFields(logrus.Fields{"Authorization": "Bearer header-secret", "Cookie": "session=cookie-secret", "payload": map[string]string{"Key": "json-secret"}, "extra": "fake-token-long"}).Error("failure https://user:url-password@example.test/path?token=query-secret#fragment-secret Bearer bearer-secret")
	logger.Info("encoded fake%2Fkey%2Bquoted%22secret plain fake/key+quoted\"secret")
	logger.Info("[bot] 收到私聊消息 user : 设置AI聊天密钥chat-secret")
	logger.Info("[api] 发送群消息(101): output-secret")
	logger.Info("Authorization: Basic basic-secret\nCookie: cookie-on-next-line\npassword=assignment-secret")
	page := l.page(logQuery{limit: 200})
	raw, _ := json.Marshal(page)
	for _, secret := range []string{"fake-token", "fake%2Fkey", "header-secret", "cookie-secret", "json-secret", "url-password", "query-secret", "fragment-secret", "bearer-secret", "chat-secret", "output-secret", "basic-secret", "cookie-on-next-line", "assignment-secret"} {
		if bytes.Contains(raw, []byte(secret)) || strings.Contains(output.String(), secret) {
			t.Fatalf("credential leaked: %q", secret)
		}
	}
	if len(page.Entries) != 5 || !strings.Contains(output.String(), "level=error") || !strings.Contains(output.String(), "https://example.test/path") {
		t.Fatal("normal logger output or capture missing")
	}
	// Preserve the original Entry; the formatter works on a copy.
	entry := &logrus.Entry{Logger: logger, Message: "fake-token-long", Data: logrus.Fields{"password": "do-not-mutate"}, Time: time.Now()}
	_, err := (RedactingFormatter{Logs: l, Next: &logrus.JSONFormatter{}}).Format(entry)
	if err != nil || entry.Message != "fake-token-long" || entry.Data["password"] != "do-not-mutate" {
		t.Fatal("formatter mutated original entry")
	}
}

func TestLogBoundsCursorFiltersAndRestart(t *testing.T) {
	l, logger := logFixture()
	for i := 0; i < 1100; i++ {
		logger.Infof("record-%04d", i)
	}
	first := l.page(logQuery{limit: 200})
	if first.Retained != 1000 || len(first.Entries) != 200 || first.Entries[0].ID != "901" || first.Cursor != "1100" {
		t.Fatalf("wrong bounded initial page: retained %d entries %d cursor %s", first.Retained, len(first.Entries), first.Cursor)
	}
	gap := l.page(logQuery{after: 1, limit: 2})
	if !gap.Gap || gap.Entries[0].ID != "101" || gap.Cursor != "102" {
		t.Fatal("missing cursor expiry/backlog indication")
	}
	logger.Warn("needle <script>alert('literal')</script>")
	logger.Info("other needle")
	filtered := l.page(logQuery{after: 1100, limit: 200, level: "warning", keyword: "NEEDLE"})
	if len(filtered.Entries) != 1 || filtered.Cursor != "1102" || !strings.Contains(filtered.Entries[0].Message, "<script>") {
		t.Fatal("filter did not match text or advance cursor")
	}
	reset := l.page(logQuery{after: 1101, instance: "0000000000000000", limit: 200})
	if !reset.Reset || len(reset.Entries) == 0 {
		t.Fatal("new instance not detected")
	}
	// New secrets also protect already retained records on subsequent reads.
	logger.Info("new-model-credential")
	l.AddSecrets("new-model-credential")
	if strings.Contains(l.page(logQuery{limit: 200}).Entries[199].Message, "new-model-credential") {
		t.Fatal("new credential leaked from older record")
	}
	for i := 0; i < 1100; i++ {
		logger.Info(strings.Repeat("汉🙂", 900))
	}
	l.mu.Lock()
	if l.bytes > logByteCapacity || l.count > logCapacity {
		t.Fatal("byte or record limit exceeded")
	}
	l.mu.Unlock()
	for _, row := range l.page(logQuery{limit: 200}).Entries {
		if !row.Truncated || !utf8.ValidString(row.Message) || len(row.Message) > logMessageLimit+32 {
			t.Fatal("Unicode truncation failed")
		}
	}
}

func TestLogConcurrentReadsAndSecretHistoryBound(t *testing.T) {
	l, _ := logFixture()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			for j := 0; j < 100; j++ {
				_ = l.Fire(&logrus.Entry{Message: "concurrent ordinary log", Level: logrus.InfoLevel, Time: time.Now()})
				_ = l.page(logQuery{limit: 10})
				l.AddSecrets("concurrent-credential")
			}
		})
	}
	wg.Wait()
	if l.page(logQuery{limit: 200}).Cursor != "800" {
		t.Fatal("concurrent writes lost")
	}
	for i := 0; i < 260; i++ {
		l.AddSecrets(fmt.Sprintf("credential-%04d", i))
	}
	page := l.page(logQuery{limit: 200})
	if !l.maskAll || len(l.secrets) > 256 || !strings.Contains(page.Entries[0].Message, "密钥历史达到上限") {
		t.Fatal("secret history not bounded/fail closed")
	}
}

func TestLogAPIAuthValidationAndFailClosed(t *testing.T) {
	l, logger := logFixture()
	cfg := testConfig(t)
	cfg.Logs = l
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("authenticated record")
	expectStatus(t, request(s, "GET", "/api/logs", nil, nil, ""), 401)
	cookie, csrf := enroll(t, s)
	for _, query := range []string{"after=-1", "after=01", "after=18446744073709551616", "limit=201", "limit=0", "level=unknown", "instance=bad", "q=" + strings.Repeat("x", 257)} {
		expectStatus(t, request(s, "GET", "/api/logs?"+query, nil, cookie, ""), 400)
	}
	w := request(s, "GET", "/api/logs?level=info&q=authenticated", nil, cookie, "")
	expectStatus(t, w, 200)
	if w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Body.String(), "authenticated record") {
		t.Fatal("authenticated log API wrong content/cache policy")
	}
	// Repeated short requests need no per-client subscription or reader goroutine.
	for i := 0; i < 30; i++ {
		expectStatus(t, request(s, "GET", "/api/logs?after=1", nil, cookie, ""), 200)
	}
	ai, c, _, _ := aiFixture(t)
	s.cfg.AI = ai
	if err := c.SetExtra(map[string]any{"Key": "read-time-model-secret"}); err != nil {
		t.Fatal(err)
	}
	logger.Info("read-time-model-secret")
	w = request(s, "GET", "/api/logs", nil, cookie, "")
	expectStatus(t, w, 200)
	if strings.Contains(w.Body.String(), "read-time-model-secret") {
		t.Fatal("live model credential not protected")
	}
	if err := c.Manager.D.Close(); err != nil {
		t.Fatal(err)
	}
	w = request(s, "GET", "/api/logs", nil, cookie, "")
	expectStatus(t, w, 503)
	if strings.Contains(w.Body.String(), "secret") {
		t.Fatal("read failure exposed internal details")
	}
	expectStatus(t, request(s, "POST", "/api/logout", map[string]any{}, cookie, csrf), 200)
	expectStatus(t, request(s, "GET", "/api/logs", nil, cookie, ""), 401)
	if q, err := parseLogQuery(url.Values{"level": {"WARN"}}); err != nil || q.level != "warning" {
		t.Fatal("warn level alias failed")
	}
}
