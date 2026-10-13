package sql

import "strings"

// QuerySet returns "q op (?,?,...,?)", []T
func QuerySet[T any](q, op string, s []T) (string, []any) {
	sz := len(s)
	if sz == 0 {
		panic("len(s) must > 0")
	}
	sb := strings.Builder{}
	qs := make([]any, sz)
	sb.WriteString(q)
	sb.WriteByte(' ')
	sb.WriteString(op)
	sb.WriteString(" (?")
	qs[0] = s[0]
	for i := 1; i < sz; i++ {
		sb.WriteString(",?")
		qs[i] = s[i]
	}
	sb.WriteByte(')')
	return sb.String(), qs
}
