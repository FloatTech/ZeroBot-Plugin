package pool

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/FloatTech/floatbox/binary"
)

func TestNTPacker(t *testing.T) {
	rkeyb64 := "CAESKBkcro_MGujokpPszCPSJtCwBMcJAkNIEqxA0gVXuTCaxQLbnGx4yk4"
	idb64 := "CgoyNzE1NDM3MTQwEhS2wXeeCxuBUgynUmsZ7oYX2r80ARiO468GIP8KKIGuxu_A7IUDUIC9owE"
	nuurl := nturl(fmt.Sprintf(ntcacheurl, idb64, rkeyb64))
	t.Log(nuurl)
	raw, err := nuurl.pack()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(binary.StringToBytes(raw)))
	upknt, err := unpack(raw, "")
	if err != nil {
		t.Fatal(err)
	}
	if upknt != nuurl {
		t.Fatal("expected", nuurl, "but got", upknt)
	}
}
