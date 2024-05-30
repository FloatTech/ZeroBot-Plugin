package pool

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"

	"github.com/FloatTech/floatbox/binary"
)

const (
	ntcacheurlprefix = "https://multimedia.nt.qq.com.cn/download?appid=1407&fileid="
	ntcacheurl       = ntcacheurlprefix + "%s&spec=0&rkey=%s"
	ntappidlen       = 60
	ntrkeylen        = 46
	ntrawlen         = ntappidlen + ntrkeylen
)

var ntcachere = regexp.MustCompile(`^https://multimedia.nt.qq.com.cn/download\?appid=1407&fileid=([0-9a-zA-Z_-]+)&spec=0&rkey=([0-9a-zA-Z_-]+)$`)

var (
	ErrInvalidNTURL = errors.New("invalid nt url")
	ErrInvalidNTRaw = errors.New("invalid nt raw")
)

type nturl string

func unpack(raw, rkey string) (nturl, error) {
	if len(raw) != ntrawlen {
		return "", ErrInvalidNTRaw
	}
	rb := binary.StringToBytes(raw)
	b := rb[ntappidlen-1]
	fileid := base64.RawURLEncoding.EncodeToString(rb[:59])
	if len(fileid) < int(b) {
		return "", ErrInvalidNTRaw
	}
	fileid = fileid[:b]
	if rkey == "" {
		rkey = base64.RawURLEncoding.EncodeToString(rb[60:])
		b = rb[ntrawlen-1]
		if len(rkey) < int(b) {
			return "", ErrInvalidNTRaw
		}
		rkey = rkey[:b]
	}
	return nturl(fmt.Sprintf(ntcacheurl, fileid, rkey)), nil
}

// pack url into pool
func (nu nturl) pack() (string, error) {
	subs := ntcachere.FindStringSubmatch(string(nu))
	if len(subs) != 3 {
		return "", ErrInvalidNTURL
	}
	var buf [ntrawlen]byte
	fileid := subs[1]
	rkey := subs[2]
	_, err := base64.RawURLEncoding.Decode(buf[:ntappidlen], binary.StringToBytes(fileid))
	if err != nil {
		return "", err
	}
	buf[ntappidlen-1] = byte(len(fileid))
	_, err = base64.RawURLEncoding.Decode(buf[ntappidlen:], binary.StringToBytes(rkey))
	if err != nil {
		return "", err
	}
	buf[ntrawlen-1] = byte(len(rkey))
	return binary.BytesToString(buf[:]), nil
}

// rkey get the embeded rkey
func (nu nturl) rkey() (string, error) {
	subs := ntcachere.FindStringSubmatch(string(nu))
	if len(subs) != 3 {
		return "", ErrInvalidNTURL
	}
	return subs[2], nil
}
