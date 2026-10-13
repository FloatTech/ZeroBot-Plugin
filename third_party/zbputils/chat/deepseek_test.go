package chat

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/fumiama/deepinfra/model"
)

func TestDeepSeekRoundTripBuilders(t *testing.T) {
	for _, effort := range []string{"none", "off", "disabled", "", "low", "high", "max"} {
		t.Run(effort, func(t *testing.T) {
			p, err := ModelType(0).ProtocolForAPI("https://api.deepseek.com/chat/completions", "deepseek-flash", .7, .9, 512, effort)
			if err != nil {
				t.Fatal(err)
			}
			p = p.System("test").User(model.NewContentText("hello")).Assistant(model.NewContentText("ok"))
			for round := 0; round < 2; round++ {
				var body struct {
					Thinking struct {
						Type string `json:"type"`
					} `json:"thinking"`
					Effort   string                `json:"reasoning_effort"`
					Messages []model.OpenAIMessage `json:"messages"`
				}
				if err := json.Unmarshal(p.Body().Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				want := "enabled"
				if effort == "none" || effort == "off" || effort == "disabled" {
					want = "disabled"
				}
				if body.Thinking.Type != want || len(body.Messages) == 0 {
					t.Fatalf("round %d lost thinking or messages", round)
				}
				if want == "disabled" && body.Effort != "" {
					t.Fatal("invalid effort sent while thinking disabled")
				}
				var raw map[string]json.RawMessage
				_ = json.Unmarshal(p.Body().Bytes(), &raw)
				if _, ok := raw["enable_thinking"]; ok {
					t.Fatal("unrelated provider parameter leaked")
				}
				p = p.Clone().User(model.NewContentText("next"))
			}
			if err := p.Parse(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"reply","reasoning_content":"private"}}]}`)); err != nil {
				t.Fatal(err)
			}
			if p.Output().String() != "reply" {
				t.Fatal("response parsing failed")
			}
		})
	}
}

func TestDeepSeekAdapterDoesNotChangeOtherProviders(t *testing.T) {
	for _, endpoint := range []string{"https://api.deepinfra.com/v1/openai/chat/completions", "https://api.deepseek.com.example/chat/completions", "https://example.test/deepseek", "http://api.deepseek.com/chat/completions"} {
		p, err := ModelType(0).ProtocolForAPI(endpoint, "deepseek-flash", .7, .9, 512, "none")
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := p.(*model.OpenAI); !ok {
			t.Fatalf("unexpected adapter for %s", endpoint)
		}
	}
}
