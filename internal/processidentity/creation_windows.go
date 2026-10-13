//go:build windows

package processidentity

import (
	"time"

	"golang.org/x/sys/windows"
)

func creationTime() (time.Time, error) {
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(windows.CurrentProcess(), &created, &exited, &kernel, &user); err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, created.Nanoseconds()).UTC(), nil
}
