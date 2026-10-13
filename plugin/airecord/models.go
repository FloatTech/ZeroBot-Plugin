package airecord

import "github.com/tidwall/gjson"

type voiceModel struct{ name, id, preview string }

// QQ repeats some voices across categories. Use IDs, not names, to keep the
// displayed index tied to the exact model the user selects.
func availableVoiceModels(data gjson.Result) []voiceModel {
	models := []voiceModel{}
	seen := map[string]bool{}
	data.ForEach(func(_, bucket gjson.Result) bool {
		bucket.Get("characters").ForEach(func(_, character gjson.Result) bool {
			name, id := character.Get("character_name").String(), character.Get("character_id").String()
			if name != "" && id != "" && !seen[id] {
				seen[id] = true
				models = append(models, voiceModel{name: name, id: id, preview: character.Get("preview_url").String()})
			}
			return true
		})
		return true
	})
	return models
}
