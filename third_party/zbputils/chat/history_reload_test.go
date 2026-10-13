package chat

import (
	"strings"
	"testing"
	"time"

	goba "github.com/fumiama/go-onebot-agent"
)

func TestAgentPersonaReloadKeepsArchivedContext(t *testing.T) {
	h, err := openAgentHistory(t.TempDir(), 789, 900000, historyConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	if err = h.Append(10, "event", historyEvent(10, "pre-persona-change")); err != nil {
		t.Fatal(err)
	}
	ag := goba.NewAgent(&goba.Config{}, 16, 8, time.Hour, "", testMemory{}, true, true)
	ag.SetHistoryStore(h)
	agentsMu.Lock()
	previous := histories[789]
	histories[789] = h
	ags.Store(789, &ag)
	agentsMu.Unlock()
	t.Cleanup(func() {
		agentsMu.Lock()
		if previous == nil {
			delete(histories, 789)
		} else {
			histories[789] = previous
		}
		ags.Delete(789)
		agentsMu.Unlock()
	})
	generation := AgentGeneration()
	ReloadAgents()
	if AgentGeneration() == generation {
		t.Fatal("old task generation remains active")
	}
	if _, ok := ags.Load(789); ok {
		t.Fatal("old persona still cached")
	}
	if !strings.Contains(historyBody(t, h, 10), "pre-persona-change") {
		t.Fatal("persona reload removed history")
	}
}
