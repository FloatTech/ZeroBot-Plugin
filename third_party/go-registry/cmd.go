package registry

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"io"
	"unsafe"

	base14 "github.com/fumiama/go-base16384"
	tea "github.com/fumiama/gofastTEA"
)

const (
	CMDGET uint8 = iota
	CMDCAT
	CMDMD5
	CMDACK
	CMDEND
	CMDSET
	CMDDEL
	CMDDAT
)

const (
	ACKNONE uint8 = iota<<4 + 3
	ACKSUCC
	ACKDATA
	ACKNULL
	ACKNEQU
	ACKERRO
)

var (
	ErrMd5Mismatch = errors.New("cmd: md5 mismatch")
)

type CmdPacket struct {
	io.ReaderFrom
	t    *tea.TEA
	Data []byte
	rawCmdPacket
}

type rawCmdPacket struct {
	cmd uint8
	len uint8
	md5 [16]byte
	raw [255]byte // raw will expand to len
}

//go:nosplit
func NewCmdPacket(cmd uint8, data []byte, t *tea.TEA) (c *CmdPacket) {
	c = pool.Get().(*CmdPacket)
	c.t = t
	c.Data = data
	c.cmd = cmd
	c.md5 = md5.Sum(data)
	return
}

//go:nosplit
func ParseCmdPacket(data []byte, t *tea.TEA) (c *CmdPacket) {
	if len(data) < 1+1+16 {
		return nil
	}
	if len(data)-1-1-16 < int(data[1]) {
		return nil
	}
	r := (*rawCmdPacket)(*(*unsafe.Pointer)(unsafe.Pointer(&data)))
	c = pool.Get().(*CmdPacket)
	c.t = t
	c.cmd = r.cmd
	c.len = r.len
	c.md5 = r.md5
	copy(c.raw[:], data[1+1+16:])
	return c
}

//go:nosplit
func ReadCmdPacket(f io.Reader, t *tea.TEA) (c *CmdPacket, err error) {
	c = pool.Get().(*CmdPacket)
	buf := (*[1 + 1 + 16 + 255]byte)(unsafe.Pointer(&c.rawCmdPacket))
	_, err = io.ReadFull(f, buf[:1+1+16])
	if err != nil {
		c.Put()
		return nil, err
	}
	_, err = io.ReadFull(f, c.raw[:c.len])
	if err != nil {
		c.Put()
		return nil, err
	}
	return
}

//go:nosplit
func (c *CmdPacket) Refresh(cmd uint8, data []byte, t *tea.TEA) {
	c.t = t
	c.Data = data
	c.cmd = cmd
	c.md5 = md5.Sum(data)
}

//go:nosplit
func (c *CmdPacket) ReadFrom(f io.Reader) (n int64, err error) {
	if c.cmd > 0 {
		err = io.EOF
		return
	}
	buf := (*[1 + 1 + 16 + 255]byte)(unsafe.Pointer(&c.rawCmdPacket))
	cnt, err := io.ReadFull(f, buf[:1+1+16])
	if err != nil {
		return int64(cnt), err
	}
	cnt, err = io.ReadFull(f, c.raw[:c.len])
	cnt += 1 + 1 + 16
	if err != nil {
		return int64(cnt), err
	}
	return
}

// Write should not be used due to the full-copy of buf
func (c *CmdPacket) Write(buf []byte) (n int, err error) {
	oldlen := len(c.Data)
	c.Data = append(c.Data, buf...)
	if len(c.Data) < 1+1+16 {
		return len(buf), nil
	}
	if len(c.Data) < 1+1+16+int(c.len) {
		return len(buf), nil
	}
	r := (*rawCmdPacket)(*(*unsafe.Pointer)(unsafe.Pointer(&c.Data)))
	c.cmd = r.cmd
	c.len = r.len
	c.md5 = r.md5
	copy(c.raw[:], r.raw[:c.len])
	c.Data = c.Data[1+1+16+int(c.len):]
	return 1 + 1 + 16 + int(c.len) - oldlen, nil
}

//go:nosplit
func (c *CmdPacket) Encrypt(seq uint8) (raw []byte) {
	setseq(c.t, seq)
	c.len = uint8(c.t.EncryptLittleEndianTo(c.Data, sumtable, c.raw[:]))
	(*slice)(unsafe.Pointer(&raw)).data = unsafe.Pointer(&c.rawCmdPacket)
	(*slice)(unsafe.Pointer(&raw)).len = 1 + 1 + 16 + int(c.len)
	(*slice)(unsafe.Pointer(&raw)).cap = 1 + 1 + 16 + 255
	return
}

//go:nosplit
func (c *CmdPacket) Decrypt(seq uint8) error {
	setseq(c.t, seq)
	d := c.t.DecryptLittleEndian(c.raw[:c.len], sumtable)
	if d != nil && c.md5 == md5.Sum(d) {
		c.Data = d
		return nil
	}
	return ErrMd5Mismatch
}

//go:nosplit
func (c *CmdPacket) Put() {
	c.cmd = 0
	c.Data = nil
	pool.Put(c)
}

//go:nosplit
func setseq(t *tea.TEA, seq uint8) {
	*(*uint8)(unsafe.Add(unsafe.Pointer(t), 15)) = seq
}

// randuint32 returns a lock free uint32 value.
//
//go:linkname randuint32 runtime.fastrand
func randuint32() uint32

//go:nosplit
func fill() []byte {
	var b [8]byte
	binary.LittleEndian.PutUint32(b[:4], randuint32())
	binary.LittleEndian.PutUint32(b[4:8], randuint32())
	return base14.Encode(b[:7])
}

// TEA encoding sumtable
var sumtable = [0x10]uint32{
	0x9e3579b9,
	0x3c6ef172,
	0xd2a66d2b,
	0x78dd36e4,
	0x17e5609d,
	0xb54fda56,
	0x5384560f,
	0xf1bb77c8,
	0x8ff24781,
	0x2e4ac13a,
	0xcc653af3,
	0x6a9964ac,
	0x08d12965,
	0xa708081e,
	0x451221d7,
	0xe37793d0,
}
