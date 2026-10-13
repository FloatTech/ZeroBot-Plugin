package localadmin

import (
	"encoding/json"
	"strings"

	"github.com/FloatTech/zbputils/chat"
)

// AIConfigCheck contains only local field checks, never a connectivity verdict.
type AIConfigCheck struct {
	TypeValid         bool `json:"typeValid"`
	ModelConfigured   bool `json:"modelConfigured"`
	ModelRequired     bool `json:"modelRequired"`
	AddressConfigured bool `json:"addressConfigured"`
	AddressValid      bool `json:"addressValid"`
	KeyConfigured     bool `json:"keyConfigured"`
	KeyRequired       bool `json:"keyRequired"`
	Complete          bool `json:"complete"`
}

func checkAIConfig(kind chat.ModelType, model, address, key, addressField string, modelRequired bool) AIConfigCheck {
	check := AIConfigCheck{
		TypeValid: kind >= 0 && kind <= 2, ModelConfigured: strings.TrimSpace(model) != "", ModelRequired: modelRequired,
		AddressConfigured: strings.TrimSpace(address) != "", KeyConfigured: key != "", KeyRequired: kind != 1,
	}
	if check.AddressConfigured {
		raw, _ := json.Marshal(address)
		check.AddressValid = chat.ValidateConfigFields(map[string]json.RawMessage{addressField: raw}) == nil
	}
	check.Complete = check.TypeValid && check.AddressConfigured && check.AddressValid &&
		(check.ModelConfigured || !check.ModelRequired) && (check.KeyConfigured || !check.KeyRequired)
	return check
}

func aiConfigChecks(c chat.AgentConfig) map[string]AIConfigCheck {
	return map[string]AIConfigCheck{
		"chat":  checkAIConfig(c.Type, c.ModelName, c.API, string(c.Key), "API", true),
		"agent": checkAIConfig(c.AgentType, c.AgentModelName, c.AgentAPI, string(c.AgentKey), "AgentAPI", true),
		"image": checkAIConfig(c.ImageType, c.ImageModelName, c.ImageAPI, string(c.ImageKey), "ImageAPI", false),
	}
}
