package heisi

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"
)

const (
	template2021    = "http://hs.heisiwu.com/wp-content/uploads/%4d/%02d/%4d%02d16%06d-611a3%8s.jpg"
	templategeneral = "http://hs.heisiwu.com/wp-content/uploads/%4d/%02d/%015x"
)

type item [10]byte

// String item to url
func (it item) String() string {
	year, month := int((it[0]>>4)&0x0f), int(it[0]&0x0f)
	year += 2021
	if year == 2021 {
		num := binary.BigEndian.Uint32(it[1:5])
		dstr := hex.EncodeToString(it[5:9])
		return fmt.Sprintf(template2021, year, month, year, month, num, dstr)
	}
	d := binary.BigEndian.Uint64(it[1:9])
	isscaled := it[9]&0x80 > 0
	num := int(it[9] & 0x7f)
	trestore := fmt.Sprintf(templategeneral, year, month, d&0x0fffffff_ffffffff)
	if num > 0 {
		trestore += fmt.Sprintf("-%d", num)
	}
	if isscaled {
		trestore += "-scaled"
	}
	d = bits.RotateLeft64(d, 4) & 0x0f
	switch d {
	case 0:
		trestore += ".jpg"
	case 1:
		trestore += ".png"
	case 2:
		trestore += ".webp"
	default:
		return "invalid ext"
	}
	return trestore
}
