package registry

import (
	"net"
	"time"
)

// Registry requests begin with a Write. Keep a single deadline for that write
// and all subsequent response reads, including Cat's multiple chunks.
type deadlineConn struct {
	net.Conn
	timeout time.Duration
}

func (c *deadlineConn) Write(data []byte) (int, error) {
	if err := c.Conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return 0, err
	}
	return c.Conn.Write(data)
}
