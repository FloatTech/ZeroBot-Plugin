package kanban

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/fumiama/go-registry"
	tea "github.com/fumiama/gofastTEA"
)

func TestFetchKanbanProtocol(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ack     uint8
		text    string
		wantErr error
	}{
		{"announcement", registry.ACKDATA, "本地测试公告", nil},
		{"literalNull", registry.ACKDATA, "null", nil},
		{"missingKey", registry.ACKNULL, "null", registry.ErrNoSuchKey},
		{"serverError", registry.ACKERRO, "erro", registry.ErrInternalServer},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			serverDone := make(chan error, 1)
			go func() {
				conn, err := listener.Accept()
				if err != nil {
					serverDone <- err
					return
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
				var password [16]byte
				copy(password[:], "fumiama")
				cipher := tea.NewTeaCipherLittleEndian(password[:])
				request := registry.NewCmdPacket(registry.CMDGET, nil, &cipher)
				defer request.Put()
				if _, err = request.ReadFrom(conn); err != nil {
					serverDone <- err
					return
				}
				if err = request.Decrypt(0); err != nil {
					serverDone <- err
					return
				}
				if string(request.Data) != "ZeroBot-Plugin/kanban" {
					serverDone <- errors.New("unexpected key")
					return
				}
				response := registry.NewCmdPacket(tc.ack, []byte(tc.text), &cipher)
				defer response.Put()
				_, err = conn.Write(response.Encrypt(1))
				serverDone <- err
			}()
			text, err := fetchKanban(listener.Addr().String(), time.Second)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if err == nil && text != tc.text {
				t.Fatalf("text = %q, want %q", text, tc.text)
			}
			if err := <-serverDone; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFetchKanbanReadDeadline(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		// Accept and consume the request, but never send an announcement.
		var buffer [256]byte
		for {
			if _, err := conn.Read(buffer[:]); err != nil {
				return
			}
		}
	}()
	start := time.Now()
	_, err = fetchKanban(listener.Addr().String(), 100*time.Millisecond)
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("expected timeout, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("deadline exceeded: %v", elapsed)
	}
	select {
	case <-serverDone:
	case <-time.After(time.Second):
		t.Fatal("client did not close its connection")
	}
}
