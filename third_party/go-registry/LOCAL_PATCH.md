# Local registry timeout patch

Source: `github.com/fumiama/go-registry@v0.2.7`, copied from the locked Go module cache. The original source files, module metadata and GPL-3.0 license are retained. Upstream's tests requiring a live registry server were not copied.

The original `ConnectIn` bounds dialing, but `Get`, `Cat` and `IsMd5Equal` can wait indefinitely for a response. This blocks ZeroBot resource initialization even when all local startup assets exist.

Local changes: `Connect` uses a five-second dial timeout; both connection functions wrap successful connections so each request write and its response reads share a five-second deadline. A later request refreshes the deadline. Callers continue receiving normal timeout errors and retain their existing local-resource fallback.

`deadline_test.go` checks Get, Cat and MD5 comparison against a server which consumes requests without responding. No external registry writes or live tests are needed.
