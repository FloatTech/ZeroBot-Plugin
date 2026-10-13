//go:build ignore

// Two authorized provider calls with a generated fixture. No OneBot executor,
// real conversation history, or credential output is created.
package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/chat"
	"github.com/fumiama/deepinfra"
	goba "github.com/fumiama/go-onebot-agent"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

type visionMemory struct{}

func (visionMemory) Load(int64) []string      { return nil }
func (visionMemory) Save(int64, string) error { return errors.New("synthetic memory writes disabled") }

type visionTransport struct {
	calls       []map[string]any
	description string
}

func (c *visionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	row := map[string]any{}
	imageInput := false
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err == nil {
			data, _ := io.ReadAll(io.LimitReader(body, 2<<20))
			body.Close()
			imageInput = bytes.Contains(data, []byte(`"type":"image_url"`))
			row["imageInput"] = imageInput
			row["descriptionInAgentPrompt"] = bytes.Contains(data, []byte("__agent_desc__"))
		}
	}
	started := time.Now()
	resp, err := http.DefaultTransport.RoundTrip(req)
	row["elapsedMS"] = time.Since(started).Milliseconds()
	if err != nil {
		row["transportFailed"] = true
		c.calls = append(c.calls, row)
		return nil, err
	}
	row["status"] = resp.StatusCode
	data, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(data))
	if resp.StatusCode == 200 && readErr == nil {
		var reply struct {
			Model   string `json:"model"`
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if json.Unmarshal(data, &reply) == nil {
			row["model"] = reply.Model
			if imageInput && len(reply.Choices) > 0 {
				c.description = reply.Choices[0].Message.Content
			}
		}
	}
	c.calls = append(c.calls, row)
	return resp, readErr
}

func main() {
	configured := flag.Bool("configured", false, "Require the saved image configuration instead of testing the proposed setup")
	flag.Parse()
	logrus.SetOutput(io.Discard)
	db, err := sql.Open("sqlite3", "file:data/control/plugins.db?mode=ro")
	if err != nil {
		visionFail("database unavailable")
	}
	defer db.Close()
	h := md5.Sum([]byte("aichat"))
	i := byte('a') % 15
	gid := int16(binary.LittleEndian.Uint16(h[i : i+2]))
	var raw string
	if db.QueryRow("SELECT ext FROM __resp WHERE gid=?", gid).Scan(&raw) != nil {
		visionFail("AI settings unavailable")
	}
	var cfg chat.AgentConfig
	if json.Unmarshal([]byte(raw), &cfg) != nil {
		visionFail("AI settings invalid")
	}
	const endpoint = "https://api.deepseek.com/chat/completions"
	if cfg.API != endpoint || cfg.AgentAPI != endpoint || cfg.ModelName != "deepseek-flash" || cfg.AgentModelName != "deepseek-flash" || cfg.Key == "" || cfg.Key != cfg.AgentKey {
		visionFail("verified existing DeepSeek configuration required")
	}
	if !*configured {
		cfg.ImageAPI = endpoint
		cfg.ImageModelName = "deepseek-flash"
		cfg.ImageType = 0
		cfg.ImageKey = cfg.Key
	}
	if cfg.ImageAPI != endpoint || cfg.ImageModelName != "deepseek-flash" || cfg.ImageType != 0 || cfg.ImageKey != cfg.Key {
		visionFail("image settings differ from the authorized shared DeepSeek setup")
	}
	fixture := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.Draw(fixture, fixture.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(fixture, image.Rect(45, 45, 240, 195), image.NewUniform(color.RGBA{R: 235, A: 255}), image.Point{}, draw.Src)
	for y := 275; y < 465; y++ {
		for x := 275; x < 465; x++ {
			if (x-370)*(x-370)+(y-370)*(y-370) < 90*90 {
				fixture.Set(x, y, color.RGBA{B: 235, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	if png.Encode(&buf, fixture) != nil {
		visionFail("synthetic image encoding failed")
	}
	if os.MkdirAll(".local", 0700) != nil || os.WriteFile(".local/deepseek-vision-test.png", buf.Bytes(), 0600) != nil {
		visionFail("synthetic image save failed")
	}
	transport := new(visionTransport)
	client := &http.Client{Timeout: 30 * time.Second, Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	imageAPI := deepinfra.NewAPI(cfg.ImageAPI, string(cfg.ImageKey))
	imageAPI.SetHTTPClient(client)
	imageProtocol, err := cfg.ImageType.ProtocolForAPI(cfg.ImageAPI, cfg.ImageModelName, .2, 1, 512, cfg.ReasoningEffort, cfg.Separator)
	if err != nil {
		visionFail("image protocol unavailable")
	}
	ag := goba.NewAgent(&goba.Config{Nickname: "合成识图测试"}, 16, 8, time.Hour, "", visionMemory{}, true, true)
	ag.SetViewImageAPI(imageAPI, imageProtocol)
	message, _ := json.Marshal([]any{map[string]any{"type": "image", "data": map[string]string{"file": "base64://" + base64.StdEncoding.EncodeToString(buf.Bytes())}}, map[string]any{"type": "text", "data": map[string]string{"text": "这是隔离识图测试。请只使用send_private_msg，user_id=90001，用中文说出这张图片的两种形状和颜色，之后end_action。不要调用其他动作。"}}})
	if ag.AddEvent(-90001, &goba.Event{Time: time.Now().Unix(), PostType: "message", MessageType: "private", MessageID: 1, UserID: 90001, SelfID: 90000, Message: message}) != nil {
		visionFail("synthetic event failed")
	}
	api := deepinfra.NewAPI(cfg.AgentAPI, string(cfg.AgentKey))
	api.SetHTTPClient(client)
	p, err := cfg.AgentType.ProtocolForAPI(cfg.AgentAPI, cfg.AgentModelName, .2, 1, 512, cfg.ReasoningEffort, cfg.Separator)
	if err != nil {
		visionFail("Agent protocol unavailable")
	}
	reqs, callErr := ag.GetAction(api, p, -90001, goba.PermRoleOwner, 1, false)
	actionMatched := len(reqs) == 1 && reqs[0].Action == "send_private_msg" && fmt.Sprint(reqs[0].Params["user_id"]) == "90001"
	replyText := ""
	if actionMatched {
		data, _ := json.Marshal(reqs[0].Params["message"])
		replyText = string(data)
	}
	colors := func(s string) bool {
		return strings.Contains(s, "红") && strings.Contains(s, "蓝") && (strings.Contains(s, "矩形") || strings.Contains(s, "长方形")) && strings.Contains(s, "圆")
	}
	passed := len(transport.calls) == 2 && transport.calls[0]["status"] == 200 && transport.calls[0]["imageInput"] == true && transport.calls[1]["status"] == 200 && transport.calls[1]["descriptionInAgentPrompt"] == true && colors(transport.description) && actionMatched && colors(replyText) && (callErr == nil || errors.Is(callErr, io.EOF))
	report := map[string]any{"completedAt": time.Now().Format(time.RFC3339), "passed": passed, "configured": *configured, "calls": transport.calls, "visionDescription": cfg.Redact(transport.description), "syntheticReply": cfg.Redact(replyText), "imageAndTextRecognized": colors(transport.description), "descriptionInAgentContext": len(transport.calls) == 2 && transport.calls[1]["descriptionInAgentPrompt"] == true, "agentTargetMatched": actionMatched, "agentEnded": errors.Is(callErr, io.EOF), "realQQMessagesSent": 0, "realHistoryRead": false, "imageFixture": ".local/deepseek-vision-test.png"}
	data, _ := json.MarshalIndent(report, "", "  ")
	name := ".local/deepseek-vision-test.json"
	if *configured {
		name = ".local/deepseek-vision-configured-test.json"
	}
	if os.WriteFile(name, data, 0600) != nil {
		visionFail("test result save failed")
	}
	fmt.Println(string(data))
	if !passed {
		os.Exit(1)
	}
}
func visionFail(s string) { fmt.Fprintln(os.Stderr, s); os.Exit(1) }
