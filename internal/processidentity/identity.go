// Package processidentity reads bounded, non-secret identity metadata once at startup.
package processidentity

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

type Identity struct {
	ProcessID        int       `json:"processId"`
	ProcessStartedAt time.Time `json:"processStartedAt"`
	ProgramSHA256    string    `json:"programSHA256"`
}

func (i Identity) Valid() bool {
	if i.ProcessID <= 0 || i.ProcessStartedAt.IsZero() || len(i.ProgramSHA256) != 64 {
		return false
	}
	_, err := hex.DecodeString(i.ProgramSHA256)
	return err == nil
}

// Read uses the OS creation time, never the time this function or the admin starts.
// The caller keeps the value; there is no hashing on overview polls or HTTP requests.
func Read() (*Identity, error) {
	started, err := creationTime()
	if err != nil {
		return nil, err
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return readFile(os.Getpid(), started, executable)
}

func readFile(pid int, started time.Time, executable string) (*Identity, error) {
	if pid <= 0 || started.IsZero() {
		return nil, errors.New("process identity unavailable")
	}
	file, err := os.Open(executable)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		return nil, err
	}
	return &Identity{ProcessID: pid, ProcessStartedAt: started.UTC(), ProgramSHA256: strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))}, nil
}
