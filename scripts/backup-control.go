//go:build ignore

// Consistent private snapshot for a local program upgrade, without schema writes.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/internal/localadmin"
	sql "github.com/FloatTech/sqlite"
)

func main() {
	info, err := os.Lstat(filepath.FromSlash("data/control/plugins.db"))
	if err != nil || !info.Mode().IsRegular() {
		fmt.Fprintln(os.Stderr, "Existing regular control database required")
		os.Exit(1)
	}
	db := sql.New(filepath.FromSlash("data/control/plugins.db"))
	if err := db.Open(time.Hour); err != nil {
		fmt.Fprintln(os.Stderr, "Control database could not be opened")
		os.Exit(1)
	}
	defer db.Close()
	id, err := localadmin.BackupControlDatabase(&db, filepath.FromSlash(".local/admin/backups"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Control database snapshot failed")
		os.Exit(1)
	}
	fmt.Println("CONTROL_BACKUP_ID=" + id)
}
