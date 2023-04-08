package moegoe

import (
	"os"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
)

type apikeystore struct {
	k string
	p string
}

func newapikeystore(p string) (s apikeystore) {
	s.p = p
	if file.IsExist(p) {
		data, err := os.ReadFile(p)
		if err == nil {
			s.k = binary.BytesToString(data)
		}
	}
	return
}
