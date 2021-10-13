//go:build windows,cgo
// +build windows,cgo

package plugin

/*
#include <limits.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include <windows.h>
#define ERRBUF 20
static char *dlerror() {
	char *err=(char *) malloc(ERRBUF);
	sprintf(err, "error %i", GetLastError());
	return err;
}
static uintptr_t pluginOpen(const char* path, char** err) {
	void* h = LoadLibrary(path);
	if (h == NULL) {
		*err = (char*)dlerror();
	}
	return (uintptr_t)h;
}
static void* pluginLookup(uintptr_t h, const char* name, char** err) {
	void* r = GetProcAddress((void*)h, name);
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return r;
}
static int pluginClose(void* handle, char** err) {
	int res = FreeLibrary(handle);
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

func open(name string) (*Plugin, error) {
	cPath := make([]byte, C.PATH_MAX+1)
	cRelName := make([]byte, len(name)+1)
	copy(cRelName, name)
	if C._fullpath(
		(*C.char)(unsafe.Pointer(&cPath[0])),
		(*C.char)(unsafe.Pointer(&cRelName[0])), C.PATH_MAX) == nil {
		return nil, errors.New(`plugin.Open("` + name + `"): realpath failed`)
	}

	filepath := C.GoString((*C.char)(unsafe.Pointer(&cPath[0])))

	pluginsMu.Lock()
	if p := plugins[filepath]; p != nil {
		pluginsMu.Unlock()
		if p.err != "" {
			return nil, errors.New(`plugin.Open("` + name + `"): ` + p.err + ` (previous failure)`)
		}
		<-p.loaded
		return p, nil
	}
	var cErr *C.char
	h := C.pluginOpen((*C.char)(unsafe.Pointer(&cPath[0])), &cErr)
	if h == 0 {
		pluginsMu.Unlock()
		return nil, errors.New(`plugin.Open("` + name + `"): ` + C.GoString(cErr))
	}
	// TODO(crawshaw): look for plugin note, confirm it is a Go plugin
	// and it was built with the correct toolchain.
	if len(name) > 4 && name[len(name)-4:] == ".dll" {
		name = name[:len(name)-4]
	}
	if plugins == nil {
		plugins = make(map[string]*Plugin)
	}
	pluginpath, syms, errstr := lastmoduleinit()
	if errstr != "" {
		plugins[filepath] = &Plugin{
			pluginpath: pluginpath,
			err:        errstr,
		}
		pluginsMu.Unlock()
		return nil, errors.New(`plugin.Open("` + name + `"): ` + errstr)
	}
	// This function can be called from the init function of a plugin.
	// Drop a placeholder in the map so subsequent opens can wait on it.
	p := &Plugin{
		pluginpath: pluginpath,
		loaded:     make(chan struct{}),
	}
	plugins[filepath] = p
	pluginsMu.Unlock()

	initStr := make([]byte, len(pluginpath)+len("..inittask")+1) // +1 for terminating NUL
	copy(initStr, pluginpath)
	copy(initStr[len(pluginpath):], "..inittask")

	initTask := C.pluginLookup(h, (*C.char)(unsafe.Pointer(&initStr[0])), &cErr)
	if initTask != nil {
		doInit(initTask)
	}

	// Fill out the value of each plugin symbol.
	updatedSyms := map[string]interface{}{}
	for symName, sym := range syms {
		isFunc := symName[0] == '.'
		if isFunc {
			delete(syms, symName)
			symName = symName[1:]
		}

		fullName := pluginpath + "." + symName
		cname := make([]byte, len(fullName)+1)
		copy(cname, fullName)

		p := C.pluginLookup(h, (*C.char)(unsafe.Pointer(&cname[0])), &cErr)
		if p == nil {
			return nil, errors.New(`plugin.Open("` + name + `"): could not find symbol ` + symName + `: ` + C.GoString(cErr))
		}
		valp := (*[2]unsafe.Pointer)(unsafe.Pointer(&sym))
		if isFunc {
			(*valp)[1] = unsafe.Pointer(&p)
		} else {
			(*valp)[1] = p
		}
		// we can't add to syms during iteration as we'll end up processing
		// some symbols twice with the inability to tell if the symbol is a function
		updatedSyms[symName] = sym
	}
	p.syms = updatedSyms

	close(p.loaded)
	return p, nil
}

func unload(name string) error {
	cPath := make([]byte, C.PATH_MAX+1)
	cRelName := make([]byte, len(name)+1)
	copy(cRelName, name)
	if C._fullpath(
		(*C.char)(unsafe.Pointer(&cPath[0])),
		(*C.char)(unsafe.Pointer(&cRelName[0])), C.PATH_MAX) == nil {
		return errors.New(`plugin.Close("` + name + `"): realpath failed`)
	}

	filepath := C.GoString((*C.char)(unsafe.Pointer(&cPath[0])))

	var cErr *C.char

	pluginsMu.Lock()
	p := plugins[filepath]
	if p != nil {
		delete(plugins, filepath)
	}
	pluginsMu.Unlock()
	if p != nil {
		res := C.pluginClose(unsafe.Pointer(p), &cErr)
		if res != 0 {
			return errors.New(`plugin.Close("` + name + `"): ` + C.GoString(cErr))
		}
	}
	return nil
}

func lookup(p *Plugin, symName string) (Symbol, error) {
	if s := p.syms[symName]; s != nil {
		return s, nil
	}
	return nil, errors.New("plugin: symbol " + symName + " not found in plugin " + p.pluginpath)
}

var (
	pluginsMu sync.Mutex
	plugins   map[string]*Plugin
)

// lastmoduleinit is defined in package runtime
//go:linkname lastmoduleinit plugin.lastmoduleinit
func lastmoduleinit() (pluginpath string, syms map[string]interface{}, errstr string)

// doInit is defined in package runtime
//go:linkname doInit runtime.doInit
func doInit(t unsafe.Pointer) // t should be a *runtime.initTask
