package processidentity

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestExactFileHashAndFixedIdentityFields(t *testing.T) {
	file := filepath.Join(t.TempDir(), "program.bin")
	body := []byte("not a path, account or config")
	if err := os.WriteFile(file, body, 0600); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 10, 13, 8, 1, 2, 123456700, time.FixedZone("local", 8*3600))
	got, err := readFile(12345, start, file)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Valid() || got.ProcessID != 12345 || got.ProcessStartedAt.Location() != time.UTC || !got.ProcessStartedAt.Equal(start) || got.ProgramSHA256 != fmt.Sprintf("%X", sha256.Sum256(body)) {
		t.Fatalf("unexpected identity: %+v", got)
	}
	if err = os.WriteFile(file, []byte("replaced after capture"), 0600); err != nil {
		t.Fatal(err)
	}
	if got.ProgramSHA256 != fmt.Sprintf("%X", sha256.Sum256(body)) {
		t.Fatal("captured identity changed")
	}
}

func TestMissingIdentityIsNotInvented(t *testing.T) {
	for _, tc := range []struct {
		pid   int
		start time.Time
		file  string
	}{
		{0, time.Now(), "missing"}, {1, time.Time{}, "missing"}, {1, time.Now(), filepath.Join(t.TempDir(), "missing")}, {1, time.Now(), t.TempDir()},
	} {
		if got, err := readFile(tc.pid, tc.start, tc.file); err == nil || got != nil {
			t.Fatal("unreadable identity accepted")
		}
	}
	for _, got := range []Identity{{}, {ProcessID: 1, ProcessStartedAt: time.Now(), ProgramSHA256: "invalid"}} {
		if got.Valid() {
			t.Fatal("invalid identity accepted")
		}
	}
}

func TestCurrentOSIdentity(t *testing.T) {
	before := time.Now()
	got, err := Read()
	if runtime.GOOS != "windows" {
		if err == nil || got != nil {
			t.Fatal("unsupported OS invented a creation time")
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if !got.Valid() || got.ProcessID != os.Getpid() || got.ProcessStartedAt.After(before) {
		t.Fatalf("wrong OS identity: %+v", got)
	}
}
