package localadmin

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const passwordIterations = 600000

type credential struct {
	Version int    `json:"version"`
	Salt    []byte `json:"salt"`
	Hash    []byte `json:"hash"`
}

type session struct {
	CSRF    string
	Expires time.Time
}

func randomCode() string {
	return base64.RawURLEncoding.EncodeToString(randomBytes(32))
}

func randomBytes(size int) []byte {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

func regularFile(path string) error {
	i, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !i.Mode().IsRegular() {
		return errors.New("后台凭据必须为普通文件")
	}
	return nil
}

func (s *Server) loadAuth() error {
	if err := os.MkdirAll(s.cfg.DataDir, 0700); err != nil {
		return err
	}
	i, err := os.Lstat(s.cfg.DataDir)
	if err != nil {
		return err
	}
	if !i.IsDir() || i.Mode()&os.ModeSymlink != 0 {
		return errors.New("后台数据目录不能为符号链接")
	}
	if err := protectPath(s.cfg.DataDir, true); err != nil {
		return err
	}
	path := filepath.Join(s.cfg.DataDir, "credential.json")
	if err = regularFile(path); err == nil {
		if err = protectPath(path, false); err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var c credential
		if err = json.Unmarshal(b, &c); err != nil || c.Version != 1 || len(c.Salt) != 16 || len(c.Hash) != 32 {
			return errors.New("后台凭据文件损坏或版本不兼容")
		}
		s.credential = &c
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	path = filepath.Join(s.cfg.DataDir, "setup-code.txt")
	if err = regularFile(path); err == nil {
		if err = protectPath(path, false); err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		code := strings.TrimSpace(string(b))
		decoded, err := base64.RawURLEncoding.DecodeString(code)
		if err != nil || len(decoded) != 32 {
			return errors.New("后台设置码文件损坏")
		}
		s.setupCode = code
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.setupCode = randomCode()
	return writeNewPrivateFile(path, []byte(s.setupCode+"\n"))
}

// Exclusive creation prevents overwriting an existing credential or a link.
func writeNewPrivateFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(path)
		}
	}()
	if err = protectPath(path, false); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func derivePassword(password string, salt []byte) ([]byte, error) {
	return pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
}

func (s *Server) createCredential(password string) error {
	c := &credential{Version: 1, Salt: randomBytes(16)}
	var err error
	c.Hash, err = derivePassword(password, c.Salt)
	if err != nil {
		return err
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err = writeNewPrivateFile(filepath.Join(s.cfg.DataDir, "credential.json"), b); err != nil {
		return err
	}
	s.credential = c
	s.setupCode = ""
	// A leftover setup file cannot be used after credential creation.
	_ = os.Remove(filepath.Join(s.cfg.DataDir, "setup-code.txt"))
	return nil
}

func (s *Server) validPassword(password string) bool {
	hash, err := derivePassword(password, s.credential.Salt)
	return err == nil && subtle.ConstantTimeCompare(hash, s.credential.Hash) == 1
}

func (s *Server) pruneSessions() {
	for key, value := range s.sessions {
		if !s.now().Before(value.Expires) {
			delete(s.sessions, key)
		}
	}
}
