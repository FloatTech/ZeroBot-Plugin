package registry

import (
	"errors"
	"net"
	"testing"
	"time"
)

func TestRequestsStopWhenServerDoesNotRespond(t *testing.T) {
	for _, action := range []string{"Get", "Cat", "IsMd5Equal"} {
		t.Run(action, func(t *testing.T) {
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			done := make(chan struct{})
			go func() {
				defer close(done)
				var buffer [1024]byte
				for {
					if _, err := server.Read(buffer[:]); err != nil {
						return
					}
				}
			}()
			reader := NewRegReader("unused", "", "fumiama")
			reader.conn = &deadlineConn{Conn: client, timeout: 100 * time.Millisecond}
			start := time.Now()
			var err error
			switch action {
			case "Get":
				_, err = reader.Get("test")
			case "Cat":
				_, err = reader.Cat()
			case "IsMd5Equal":
				_, err = reader.IsMd5Equal([16]byte{})
			}
			var netErr net.Error
			if !errors.As(err, &netErr) || !netErr.Timeout() {
				t.Fatalf("expected timeout, got %v", err)
			}
			if elapsed := time.Since(start); elapsed > time.Second {
				t.Fatalf("request exceeded deadline: %v", elapsed)
			}
			client.Close()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("server goroutine did not finish")
			}
		})
	}
}
