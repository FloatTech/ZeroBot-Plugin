package kanban

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/fumiama/go-registry"
	tea "github.com/fumiama/gofastTEA"
)

// fetchKanban bounds both dialing and protocol I/O by the same deadline.
// go-registry's ConnectIn only limits dialing, while Get has no read deadline.
func fetchKanban(address string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	dialer := net.Dialer{Deadline: deadline}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err = conn.SetDeadline(deadline); err != nil {
		return "", err
	}

	var password [16]byte
	copy(password[:], "fumiama")
	cipher := tea.NewTeaCipherLittleEndian(password[:])
	request := registry.NewCmdPacket(registry.CMDGET, []byte("ZeroBot-Plugin/kanban"), &cipher)
	defer request.Put()
	raw := request.Encrypt(0)
	if written, writeErr := conn.Write(raw); writeErr != nil {
		return "", writeErr
	} else if written != len(raw) {
		return "", io.ErrShortWrite
	}

	reader := bufio.NewReader(conn)
	header, err := reader.Peek(1)
	if err != nil {
		return "", err
	}
	ack := header[0]
	// ReadCmdPacket in the pinned library does not initialize the cipher on a
	// fresh pooled packet. NewCmdPacket + ReadFrom keeps it initialized.
	response := registry.NewCmdPacket(registry.CMDGET, nil, &cipher)
	defer response.Put()
	if _, err = response.ReadFrom(reader); err != nil {
		return "", err
	}
	if err = response.Decrypt(1); err != nil {
		return "", registry.ErrDecAck
	}
	text := string(response.Data)
	if ack == registry.ACKERRO && text == "erro" {
		return "", registry.ErrInternalServer
	}
	if ack == registry.ACKNULL && text == "null" {
		return "", registry.ErrNoSuchKey
	}
	return text, nil
}
