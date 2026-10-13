package goba

import (
	"errors"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	errEmptyMemory     = errors.New("empty memory")
	errMemoryHasReturn = errors.New("memory has \\r|\\n")
)

// MemoryStorage can be a db or just some files
type MemoryStorage interface {
	Save(grp int64, text string) error
	Load(grp int64) []string
}

func extractMemory(r *zero.APIRequest) (string, error) {
	txt, ok := r.Params["text"].(string)
	if !ok || txt == "" {
		return "", errEmptyMemory
	}
	for _, c := range txt {
		if c == '\r' || c == '\n' {
			return "", errMemoryHasReturn
		}
	}
	return txt, nil
}

func (ag *Agent) memoryof(grp int64) string {
	mems := ag.mem.Load(grp)
	if len(mems) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for _, m := range mems {
		sb.WriteByte('\n')
		sb.WriteString(m)
	}
	return sb.String()
}
