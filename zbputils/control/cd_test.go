package control

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGenToken(t *testing.T) {
	tok := genToken()
	t.Log(tok)
	assert.Equal(t, true, isValidToken(tok, 10))
}

func TestMaru(t *testing.T) {
	t.Log(len("\xff"))
	t.Fail()
}
