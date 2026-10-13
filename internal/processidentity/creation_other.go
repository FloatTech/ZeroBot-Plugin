//go:build !windows

package processidentity

import (
	"errors"
	"time"
)

func creationTime() (time.Time, error) {
	return time.Time{}, errors.New("OS process creation time unsupported")
}
