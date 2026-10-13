package chat

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/fumiama/deepinfra/model"
)

// ProtocolForAPI keeps provider-specific parameters scoped to the official
// endpoint. Other OpenAI-compatible providers keep their existing payloads.
func (mt ModelType) ProtocolForAPI(api, name string, temp, topp float32, maxn uint, reasoning string, separators ...string) (model.Protocol, error) {
	p, err := mt.Protocol(name, temp, topp, maxn, reasoning, separators...)
	if err != nil || mt != 0 {
		return p, err
	}
	u, err := url.Parse(api)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, "api.deepseek.com") {
		return p, nil
	}
	thinking := "enabled"
	switch strings.ToLower(strings.TrimSpace(reasoning)) {
	case "none", "off", "disabled":
		thinking = "disabled"
		p.(*model.OpenAI).ReasoningEffort = ""
	}
	return &deepSeekProtocol{OpenAI: p.(*model.OpenAI), thinking: thinking}, nil
}

type deepSeekProtocol struct {
	*model.OpenAI
	thinking string
}

func (p *deepSeekProtocol) Body() *bytes.Buffer {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(p.OpenAI.Body().Bytes(), &fields); err != nil {
		panic(err)
	}
	delete(fields, "enable_thinking") // This boolean belongs to a different provider.
	fields["thinking"], _ = json.Marshal(map[string]string{"type": p.thinking})
	body, err := json.Marshal(fields)
	if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(body)
}

// Fluent builders and Clone must retain this adapter: the Agent reconstructs
// its prompt through these methods on every action round.
func (p *deepSeekProtocol) System(prompt string) model.Protocol {
	p.OpenAI.System(prompt)
	return p
}
func (p *deepSeekProtocol) User(prompt ...model.Content) model.Protocol {
	p.OpenAI.User(prompt...)
	return p
}
func (p *deepSeekProtocol) Assistant(prompt ...model.Content) model.Protocol {
	p.OpenAI.Assistant(prompt...)
	return p
}
func (p *deepSeekProtocol) Clone() model.Protocol {
	return &deepSeekProtocol{OpenAI: p.OpenAI.Clone().(*model.OpenAI), thinking: p.thinking}
}
