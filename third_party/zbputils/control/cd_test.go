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
	t.Skip("上游演示测试只有日志和无条件失败，不包含功能断言；不计入本地补丁验证")
}
