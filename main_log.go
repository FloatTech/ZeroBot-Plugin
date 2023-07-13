package main

import (
    _ "net/http/pprof"
    "net/http"
)

func init() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()
}