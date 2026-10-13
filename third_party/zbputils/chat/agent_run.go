package chat

import (
	"crypto/sha256"
	"encoding/json"
	"errors"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var ErrAgentRepeatedAction = errors.New("Agent 重复生成相同动作，已停止本次操作")
var ErrAgentReloaded = errors.New("Agent 设置已更新，已停止旧任务")

// RunAgent has no fixed iteration count. Explicit end, request/write failure,
// reload or a repeated action cycle terminates the current task. New messages
// can start new tasks with the same retained conversation history.
func RunAgent(next func(int) ([]zero.APIRequest, bool, error), execute func(zero.APIRequest) error, current func() bool) error {
	seen := map[[32]byte]bool{}
	for iteration := 1; ; iteration++ {
		if !current() {
			return ErrAgentReloaded
		}
		reqs, ended, err := next(iteration)
		if err != nil {
			return err
		}
		if len(reqs) > 0 {
			// Ignore Echo/metadata: changing an echo ID is not task progress.
			actions := make([]struct {
				Action string
				Params zero.Params
			}, len(reqs))
			for i, r := range reqs {
				actions[i].Action = r.Action
				actions[i].Params = r.Params
			}
			raw, err := json.Marshal(actions)
			if err != nil {
				return errors.New("Agent 动作格式无效")
			}
			fingerprint := sha256.Sum256(raw)
			if seen[fingerprint] {
				return ErrAgentRepeatedAction
			}
			seen[fingerprint] = true
			for _, req := range reqs {
				if !current() {
					return ErrAgentReloaded
				}
				if err := execute(req); err != nil {
					return err
				}
			}
		}
		if ended || len(reqs) == 0 {
			return nil
		}
	}
}
