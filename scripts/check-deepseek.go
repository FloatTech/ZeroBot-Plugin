//go:build ignore

// Read only the existing AI settings; never print credentials or provider errors.
package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	_ "modernc.org/sqlite"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	db, err := sql.Open("sqlite3", "file:data/control/plugins.db?mode=ro")
	if err != nil {
		fail("database unavailable")
	}
	defer db.Close()
	digest := md5.Sum([]byte("aichat"))
	offset := byte('a') % 15
	gid := int16(binary.LittleEndian.Uint16(digest[offset : offset+2]))
	var raw string
	if db.QueryRow("SELECT ext FROM __resp WHERE gid = ?", gid).Scan(&raw) != nil {
		fail("AI settings unavailable")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &fields) != nil {
		fail("AI settings invalid")
	}
	str := func(k string) string { var v string; _ = json.Unmarshal(fields[k], &v); return v }
	safeURL := func(k string) string {
		u, e := url.Parse(str(k))
		if e != nil {
			return "invalid"
		}
		u.User = nil
		u.RawQuery = ""
		u.Fragment = ""
		return u.String()
	}
	report := map[string]any{"chatAPI": safeURL("API"), "chatModel": str("ModelName"), "chatKeyConfigured": str("Key") != "", "agentAPI": safeURL("AgentAPI"), "agentModel": str("AgentModelName"), "agentKeyConfigured": str("AgentKey") != "", "keysMatch": str("Key") != "" && str("Key") == str("AgentKey"), "reasoningEffort": str("ReasoningEffort")}
	if str("Key") != "" {
		req, _ := http.NewRequest("GET", "https://api.deepseek.com/models", nil)
		req.Header.Set("Authorization", "Bearer "+str("Key"))
		client := &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
		resp, e := client.Do(req)
		if e != nil {
			report["modelsTransportFailed"] = true
		} else {
			report["modelsStatus"] = resp.StatusCode
			if resp.StatusCode == 200 {
				var result struct {
					Data []struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				if json.NewDecoder(io.LimitReader(resp.Body, 65536)).Decode(&result) == nil {
					ids := []string{}
					for _, m := range result.Data {
						ids = append(ids, m.ID)
					}
					report["availableModels"] = ids
				}
			}
			resp.Body.Close()
		}
	}
	_ = json.NewEncoder(os.Stdout).Encode(report)
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
