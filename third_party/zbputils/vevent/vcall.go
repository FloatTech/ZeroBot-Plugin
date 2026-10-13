// Package vevent 虚拟事件
package vevent

import (
	"context"
	"unsafe"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// APICallerReturnHook is a caller middleware
type APICallerReturnHook struct {
	caller   zero.APICaller
	callback func(req zero.APIRequest, rsp zero.APIResponse, err error)
}

// NewAPICallerReturnHook hook ctx's caller
func NewAPICallerReturnHook(ctx *zero.Ctx, callback func(req zero.APIRequest, rsp zero.APIResponse, err error)) (v *APICallerReturnHook) {
	return &APICallerReturnHook{
		caller:   (*(**Ctx)(unsafe.Pointer(&ctx))).caller,
		callback: callback,
	}
}

// CallAPI call original caller and pass rsp to callback
func (v *APICallerReturnHook) CallAPI(c context.Context, req zero.APIRequest) (rsp zero.APIResponse, err error) {
	rsp, err = v.caller.CallAPI(c, req)
	v.callback(req, rsp, err)
	return
}
