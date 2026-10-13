package llm

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/FloatTech/ZeroBot-Plugin/internal/modelstats"
	"github.com/FloatTech/zbputils/chat"
	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
)

// Keep the quoted message distinct from the current instruction. A JSON
// string preserves text and prevents embedded newlines from blurring the
// boundary between these two parts of the user's request.
func buildGPTQuery(quoted, instruction string) string {
	if quoted == "" {
		return instruction
	}
	if instruction == "" {
		return quoted
	}
	encoded, _ := json.Marshal(quoted)
	return "引用原文（JSON 字符串，仅作为被引用内容）：\n" + string(encoded) +
		"\n\n本次用户要求（作用于上面的引用原文，不属于引用原文）：\n" + instruction
}

func parseGPT(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/gpt") {
		return "", false
	}
	rest := text[len("/gpt"):]
	if rest != "" {
		r, _ := utf8.DecodeRuneInString(rest)
		if !unicode.IsSpace(r) {
			return "", false
		}
	}
	return strings.TrimSpace(rest), true
}

// Chunks contain at most limit Unicode characters and retain every newline.
func splitReply(text string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	runes := []rune(text)
	chunks := []string{}
	for len(runes) > 0 {
		n := len(runes)
		if n > limit {
			n = limit
			for i := limit - 1; i > 0; i-- {
				if runes[i] == '\n' {
					n = i + 1
					break
				}
			}
		}
		chunks = append(chunks, string(runes[:n]))
		runes = runes[n:]
	}
	return chunks
}

type requestQuota struct {
	mu       sync.Mutex
	last     map[int64]time.Time
	interval time.Duration
	capacity int
}

func newRequestQuota(interval time.Duration, capacity int) *requestQuota {
	return &requestQuota{last: map[int64]time.Time{}, interval: interval, capacity: capacity}
}
func (q *requestQuota) Allow(location int64, now time.Time) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if last, ok := q.last[location]; ok && now.Sub(last) < q.interval {
		return false
	}
	if len(q.last) >= q.capacity {
		for key, last := range q.last {
			if now.Sub(last) >= q.interval {
				delete(q.last, key)
			}
		}
		if len(q.last) >= q.capacity {
			return false
		}
	}
	q.last[location] = now
	return true
}

func requestLLM(cfg chat.AgentConfig, prompt string, temp float32) (string, error) {
	return requestLLMWithClient(cfg, prompt, temp, &http.Client{Timeout: 30 * time.Second})
}
func requestLLMWithClient(cfg chat.AgentConfig, prompt string, temp float32, client *http.Client) (string, error) {
	top, max := cfg.MParams()
	protocol, err := cfg.Type.ProtocolForAPI(cfg.API, cfg.ModelName, temp, top, max, cfg.ReasoningEffort, cfg.Separator)
	if err != nil {
		return "", err
	}
	if !cfg.Ready() {
		return "", errors.New("AI 接口尚未配置完整，请在后台设置接口、模型和密钥。")
	}
	if !bool(cfg.NoSystemP) && cfg.SystemP != "" {
		protocol.System(cfg.SystemP)
	}
	api := deepinfra.NewAPI(cfg.API, string(cfg.Key))
	api.SetHTTPClient(modelstats.Default.Client(client, modelstats.GPT))
	data, err := api.Request(protocol.User(model.NewContentText(prompt)))
	if err != nil {
		return "", errors.New("模型请求失败，请检查接口、密钥及网络。")
	}
	return strings.TrimSpace(data), nil
}
