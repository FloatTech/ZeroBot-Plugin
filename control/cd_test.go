package control

import "testing"

func TestGenToken(t *testing.T) {
	tok, err := genToken()
	if err == nil {
		t.Log(tok)
		t.Log(isValidToken(tok))
		t.Fail()
	} else {
		t.Fatal(err)
	}
}

func TestMaru(t *testing.T) {
	t.Log(len("\xff"))
	t.Fail()
}
