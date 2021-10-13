// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux,cgo darwin,cgo freebsd,cgo
// +build linux,cgo darwin,cgo freebsd,cgo

package plugin

/*
#cgo linux LDFLAGS: -ldl
#include <dlfcn.h>
#include <limits.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
static int pluginClose(void* handle, char** err) {
	int res = dlclose(handle);
	if (res != 0) {
		*err = (char*)dlerror();
	}
	return res;
}
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"
)

func unload(pg *Plugin) error {
	filepath := pg.pluginpath

	pluginsMu.Lock()
	p := plugins[filepath]
	if p != nil {
		delete(plugins, filepath)
	}
	pluginsMu.Unlock()
	if p != nil {
		var cErr *C.char
		res := C.pluginClose(unsafe.Pointer(p), &cErr)
		if res != 0 {
			return errors.New(`plugin.Close("` + filepath + `"): ` + C.GoString(cErr))
		}
	}
	return nil
}

//go:linkname pluginsMu plugin.pluginsMu
//go:linkname plugins plugin.plugins
var (
	pluginsMu sync.Mutex
	plugins   map[string]*Plugin
)
