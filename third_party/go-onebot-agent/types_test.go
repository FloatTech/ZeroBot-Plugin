package goba

import "testing"

func TestMarshalAPIResponse(t *testing.T) {
	rsp := &APIResponse{
		Status:  "error",
		Data:    []byte(`null`),
		Message: "12345",
	}
	t.Log(rsp.String())
	rsp = &APIResponse{
		Status:  "error",
		Data:    []byte(``),
		Message: "12345",
	}
	t.Log(rsp.String())
}
