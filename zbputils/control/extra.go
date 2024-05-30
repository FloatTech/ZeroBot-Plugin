package control

import (
	"crypto/md5"
	"encoding/binary"

	bin "github.com/FloatTech/floatbox/binary"
)

// ExtraFromString generate int16 extra key from string's md5
func ExtraFromString(s string) int16 {
	if s == "" {
		return 0
	}
	m := md5.Sum(bin.StringToBytes(s))
	i := s[0] % (md5.Size - 1)
	return int16(binary.LittleEndian.Uint16(m[i : i+2]))
}
