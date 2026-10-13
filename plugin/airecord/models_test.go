package airecord

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestVoiceModelsKeepSelectionIDs(t *testing.T) {
	models := availableVoiceModels(gjson.Parse(`[{"characters":[{"character_name":"同名","character_id":"first"},{"character_name":"同名","character_id":"second"}]},{"characters":[{"character_name":"重复","character_id":"first"},{"character_name":"无编号"},{"character_id":"missing-name"}]}]`))
	if len(models) != 2 || models[0].id != "first" || models[1].id != "second" || models[0].name != "同名" {
		t.Fatal("displayed index lost its model ID or duplicated categories")
	}
	if len(availableVoiceModels(gjson.Parse(`[]`))) != 0 {
		t.Fatal("empty API invented voices")
	}
}
