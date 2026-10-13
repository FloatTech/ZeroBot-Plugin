package registry

import "crypto/md5"

type Storage struct {
	Md5 [md5.Size]byte
	m   map[string]string
}

func (s *Storage) Get(key string) (string, error) {
	v, ok := s.m[key]
	if ok {
		return v, nil
	}
	return "", ErrNoSuchKey
}
