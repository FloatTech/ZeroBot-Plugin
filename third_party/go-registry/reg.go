package registry

import (
	"bytes"
	"crypto/md5"
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	spb "github.com/fumiama/go-simple-protobuf"
	tea "github.com/fumiama/gofastTEA"
)

var (
	ErrGetKeyTooLong    = errors.New("reg: get key too long")
	ErrDecAck           = errors.New("reg: decrypt ack error")
	ErrInternalServer   = errors.New("reg: internal server error")
	ErrPermissionDenied = errors.New("reg: permission denied")
	ErrSetKeyTooLong    = errors.New("reg: set key too long")
	ErrSetValTooLong    = errors.New("reg: set val too long")
	ErrUnknownAck       = errors.New("reg: unknown ack error")
	ErrNoSuchKey        = errors.New("reg: no such key")
	ErrRawDataTooLong   = errors.New("reg: raw data too long")
	ErrMd5NotEqual      = errors.New("reg: md5 not equal")
	ErrInvalidCatData   = errors.New("reg: invalid cat data")
	ErrNilStorData      = errors.New("reg: nil stor data")
)

type Regedit struct {
	mu   sync.Mutex
	conn net.Conn
	addr string
	stor string
	tp   tea.TEA
	ts   *tea.TEA
	seq  byte
}

func NewRegedit(addr, stor, pwd, sps string) *Regedit {
	var tp, ts [16]byte
	if len(pwd) > 15 {
		pwd = pwd[:15]
	}
	if len(sps) > 15 {
		sps = sps[:15]
	}
	copy(tp[:], pwd)
	copy(ts[:], sps)
	s := tea.NewTeaCipherLittleEndian(ts[:])
	if stor != "" {
		f, err := os.Open(stor)
		if err != nil {
			f, err = os.Create(stor)
			if err != nil {
				panic(err)
			}
		}
		_ = f.Close()
	}
	return &Regedit{addr: addr, stor: stor, tp: tea.NewTeaCipherLittleEndian(tp[:]), ts: &s}
}

func NewRegReader(addr, stor, pwd string) *Regedit {
	var tp [16]byte
	if len(pwd) > 15 {
		pwd = pwd[:15]
	}
	copy(tp[:], pwd)
	if stor != "" {
		f, err := os.Open(stor)
		if err != nil {
			f, err = os.Create(stor)
			if err != nil {
				panic(err)
			}
		}
		_ = f.Close()
	}
	return &Regedit{addr: addr, stor: stor, tp: tea.NewTeaCipherLittleEndian(tp[:])}
}

func (r *Regedit) Connect() (err error) {
	r.mu.Lock()
	if r.conn == nil {
		r.conn, err = net.DialTimeout("tcp", r.addr, 5*time.Second)
		if err == nil {
			r.conn = &deadlineConn{Conn: r.conn, timeout: 5 * time.Second}
		}
	}
	r.mu.Unlock()
	return
}

func (r *Regedit) ConnectIn(timeout time.Duration) (err error) {
	r.mu.Lock()
	if r.conn == nil {
		r.conn, err = net.DialTimeout("tcp", r.addr, timeout)
		if err == nil {
			r.conn = &deadlineConn{Conn: r.conn, timeout: 5 * time.Second}
		}
	}
	r.mu.Unlock()
	return
}

func (r *Regedit) Close() (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn != nil {
		p := NewCmdPacket(CMDEND, fill(), &r.tp)
		r.conn.Write(p.Encrypt(r.seq))
		p.Put()
		r.seq = 0
		err = r.conn.Close()
		r.conn = nil
		return
	}
	return
}

func (r *Regedit) Get(key string) (string, error) {
	if len(key) > 127 {
		return "", ErrGetKeyTooLong
	}
	p := NewCmdPacket(CMDGET, StringToBytes(key), &r.tp)
	defer p.Put()
	r.mu.Lock()
	r.conn.Write(p.Encrypt(r.seq))
	r.seq++
	err := r.ack(p)
	if err != nil {
		r.mu.Unlock()
		return "", err
	}
	err = p.Decrypt(r.seq)
	r.seq++
	r.mu.Unlock()
	if err != nil {
		return "", ErrDecAck
	}
	a := string(p.Data)
	if a == "erro" && p.cmd == ACKERRO {
		return "", ErrInternalServer
	}
	if a == "null" && p.cmd == ACKNULL {
		return "", ErrNoSuchKey
	}
	return a, nil
}

func (r *Regedit) Cat() (*Storage, error) {
	p := NewCmdPacket(CMDCAT, fill(), &r.tp)
	defer p.Put()
	r.mu.Lock()
	r.conn.Write(p.Encrypt(r.seq))
	r.seq++
	seq := r.seq
	r.seq++
	r.mu.Unlock()
	var buf [64]byte
	i := 0
	for {
		_, err := r.conn.Read(buf[i : i+1])
		if err != nil {
			return nil, err
		}
		if buf[i] == '$' {
			break
		}
		i++
		if i >= 64 {
			return nil, ErrRawDataTooLong
		}
	}
	n, err := strconv.ParseUint(BytesToString(buf[:i]), 10, 64)
	if err != nil {
		return nil, err
	}
	data := make([]byte, n)
	_, err = io.ReadFull(r.conn, data)
	if err != nil {
		return nil, err
	}
	setseq(&r.tp, seq)
	data = r.tp.DecryptLittleEndian(data, sumtable)
	s := new(Storage)
	s.m = make(map[string]string, 256)
	s.Md5 = md5.Sum(data)
	rd := bytes.NewReader(data)
	for i = 0; i < len(data); {
		sp, err := spb.NewSimplePB(rd)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(sp.Target) <= 1 {
			return nil, ErrInvalidCatData
		}
		s.m[BytesToString(sp.Target[0])] = BytesToString(sp.Target[1])
		i += int(sp.RealLen)
	}
	f, err := os.Create(r.stor)
	if err != nil {
		return s, err
	}
	defer f.Close()
	_, err = f.Write(data)
	return s, err
}

func (r *Regedit) Load() (*Storage, error) {
	data, err := os.ReadFile(r.stor)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrNilStorData
	}
	s := new(Storage)
	s.m = make(map[string]string, 256)
	s.Md5 = md5.Sum(data)
	rd := bytes.NewReader(data)
	for i := 0; i < len(data); {
		sp, err := spb.NewSimplePB(rd)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(sp.Target) <= 1 {
			return nil, ErrInvalidCatData
		}
		s.m[BytesToString(sp.Target[0])] = BytesToString(sp.Target[1])
		i += int(sp.RealLen)
	}
	f, err := os.Create(r.stor)
	if err != nil {
		return s, err
	}
	defer f.Close()
	_, err = f.Write(data)
	return s, err
}

func (r *Regedit) IsMd5Equal(m [md5.Size]byte) (bool, error) {
	p := NewCmdPacket(CMDMD5, m[:], &r.tp)
	defer p.Put()
	r.mu.Lock()
	r.conn.Write(p.Encrypt(r.seq))
	r.seq++
	err := r.ack(p)
	if err != nil {
		r.mu.Unlock()
		return false, err
	}
	err = p.Decrypt(r.seq)
	r.seq++
	r.mu.Unlock()
	if err != nil {
		return false, ErrDecAck
	}
	a := string(p.Data)
	if a == "erro" && p.cmd == ACKERRO {
		return false, ErrInternalServer
	}
	if a == "nequ" && p.cmd == ACKNEQU {
		return false, ErrNoSuchKey
	}
	if a == "null" && p.cmd == ACKNULL {
		return true, nil
	}
	return false, ErrUnknownAck
}

func (r *Regedit) Set(key, value string) error {
	if r.ts == nil {
		return ErrPermissionDenied
	}
	if len(key) > 127 {
		return ErrSetKeyTooLong
	}
	if len(value) > 127 {
		return ErrSetValTooLong
	}
	p := NewCmdPacket(CMDSET, StringToBytes(key), r.ts)
	defer p.Put()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conn.Write(p.Encrypt(r.seq))
	r.seq++
	ack := NewCmdPacket(CMDACK, nil, &r.tp)
	defer ack.Put()
	err := r.ack(ack)
	if err != nil {
		return err
	}
	err = ack.Decrypt(r.seq)
	r.seq++
	if err != nil {
		return ErrDecAck
	}
	a := BytesToString(ack.Data)
	if a == "erro" || ack.cmd == ACKERRO {
		return ErrInternalServer
	}
	if a != "data" && ack.cmd != ACKDATA {
		return ErrUnknownAck
	}
	p.Refresh(CMDDAT, StringToBytes(value), r.ts)
	r.conn.Write(p.Encrypt(r.seq))
	r.seq++
	err = r.ack(ack)
	if err != nil {
		return err
	}
	err = ack.Decrypt(r.seq)
	r.seq++
	if err != nil {
		return ErrDecAck
	}
	a = BytesToString(ack.Data)
	if a == "erro" || ack.cmd == ACKERRO {
		return ErrInternalServer
	}
	if a != "succ" && ack.cmd != ACKSUCC {
		return ErrUnknownAck
	}
	return nil
}

func (r *Regedit) Del(key string) error {
	if r.ts == nil {
		return ErrPermissionDenied
	}
	if len(key) > 127 {
		return ErrGetKeyTooLong
	}
	p := NewCmdPacket(CMDDEL, StringToBytes(key), r.ts)
	defer p.Put()
	r.mu.Lock()
	r.conn.Write(p.Encrypt(r.seq))
	r.seq++
	ack := NewCmdPacket(CMDACK, nil, &r.tp)
	defer ack.Put()
	err := r.ack(ack)
	if err != nil {
		r.mu.Unlock()
		return err
	}
	err = ack.Decrypt(r.seq)
	r.seq++
	r.mu.Unlock()
	if err != nil {
		return ErrDecAck
	}
	a := BytesToString(ack.Data)
	if a == "erro" || ack.cmd == ACKERRO {
		return ErrInternalServer
	}
	if a == "null" || ack.cmd == ACKNULL {
		return ErrNoSuchKey
	}
	if a != "succ" && ack.cmd != ACKSUCC {
		return ErrUnknownAck
	}
	return nil
}

func (r *Regedit) ack(c *CmdPacket) error {
	c.cmd = 0
	_, err := io.Copy(c, r.conn)
	return err
}
