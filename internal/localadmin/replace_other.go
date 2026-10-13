//go:build !windows

package localadmin

import "os"

func replaceConfigFile(source, target string) error { return os.Rename(source, target) }
