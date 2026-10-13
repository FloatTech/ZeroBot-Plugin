package localadmin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	sql "github.com/FloatTech/sqlite"
)

// BackupControlDatabase uses SQLite's consistent snapshot, including WAL data.
// Call under the control Manager lock to keep the snapshot before its mutation.
func BackupControlDatabase(database *sql.Sqlite, directory string) (string, error) {
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", errors.New("控制数据库备份目录不可写")
	}
	info, err := os.Lstat(directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("控制数据库备份目录无效")
	}
	if err = protectPath(directory, true); err != nil {
		return "", errors.New("控制数据库备份权限设置失败")
	}
	id := time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + hex.EncodeToString(randomBytes(8)) + "-control"
	root := filepath.Join(directory, id)
	if err = os.Mkdir(root, 0700); err != nil {
		return "", err
	}
	if err = protectPath(root, true); err != nil {
		return "", err
	}
	path := filepath.Join(root, "plugins.sqlite")
	if err = writeNewPrivateFile(path, nil); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err = database.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return "", errors.New("控制数据库快照失败，开关未修改")
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return "", err
	}
	digest := sha256.New()
	_, err = io.Copy(digest, f)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	manifest, _ := json.MarshalIndent(struct {
		Version   int       `json:"version"`
		CreatedAt time.Time `json:"createdAt"`
		Scope     string    `json:"scope"`
		File      string    `json:"file"`
		SHA256    string    `json:"sha256"`
	}{1, time.Now().UTC(), "control-db", "plugins.sqlite", hex.EncodeToString(digest.Sum(nil))}, "", "  ")
	if err = writeNewPrivateFile(filepath.Join(root, "manifest.json"), manifest); err != nil {
		return "", err
	}
	return id, nil
}
