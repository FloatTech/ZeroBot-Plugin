package wordle

import (
	"strconv"
	"unsafe"

	goBinary "encoding/binary"

	"github.com/FloatTech/zbputils/binary"
)

type wordpack [3]byte

/*
func pack(word string) (w wordpack) {
	if len(word) != 5 {
		panic("word must be 5 letters")
	}
	r := []rune(word)
	for i := range r {
		if r[i] < 'k' { // 0-9
			r[i] -= 'a' - '0'
		} else {
			r[i] -= 10
		}
	}
	word = string(r)
	n, err := strconv.ParseUint(word, 26, 32)
	if err != nil {
		panic(err)
	}
	wt := binary.SelectWriter()
	wt.WriteUInt32LE(uint32(n))
	copy(w[:], wt.Bytes())
	binary.PutWriter(wt)
	return
}
*/

func (w wordpack) String() (word string) {
	wt := binary.SelectWriter()
	_, _ = wt.Write(w[:])
	_ = wt.WriteByte(0)
	n := goBinary.LittleEndian.Uint32(wt.Bytes())
	binary.PutWriter(wt)
	word = strconv.FormatUint(uint64(n), 26)
	for len(word) < 5 {
		word = "0" + word
	}
	r := []rune(word)
	for i := range r {
		if r[i] < 'a' { // 0-9
			r[i] += 'a' - '0'
		} else {
			r[i] += 10
		}
	}
	word = string(r)
	return
}

func loadwords(data []byte) (wordpacks []wordpack) {
	(*slice)(unsafe.Pointer(&wordpacks)).data = (*slice)(unsafe.Pointer(&data)).data
	(*slice)(unsafe.Pointer(&wordpacks)).len = len(data) / 3
	(*slice)(unsafe.Pointer(&wordpacks)).cap = (*slice)(unsafe.Pointer(&data)).cap / 3
	return
}

// slice is the runtime representation of a slice.
// It cannot be used safely or portably and its representation may
// change in a later release.
//
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}
