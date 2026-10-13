//go:build ignore

// Compare the current image setup with its verified private backup.
// Prints booleans only, never keys or the original settings payload.
package main

import (
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func main() {
	id := flag.String("backup-id", "", "Verified control backup before the image edit")
	flag.Parse()
	if filepath.Base(*id) != *id || !strings.HasSuffix(*id, "-control") {
		checkVisionFail("valid backup ID required")
	}
	root := filepath.Join(".local", "admin", "backups", *id)
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		checkVisionFail("manifest unavailable")
	}
	var manifest struct{ Scope, SHA256 string }
	if json.Unmarshal(raw, &manifest) != nil || manifest.Scope != "control-db" {
		checkVisionFail("invalid manifest")
	}
	data, err := os.ReadFile(filepath.Join(root, "plugins.sqlite"))
	if err != nil {
		checkVisionFail("backup unavailable")
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != manifest.SHA256 {
		checkVisionFail("backup integrity failed")
	}
	h := md5.Sum([]byte("aichat"))
	i := byte('a') % 15
	gid := int16(binary.LittleEndian.Uint16(h[i : i+2]))
	read := func(path string) map[string]any {
		db, err := sql.Open("sqlite3", "file:"+filepath.ToSlash(path)+"?mode=ro")
		if err != nil {
			checkVisionFail("database unavailable")
		}
		defer db.Close()
		var raw string
		if db.QueryRow("SELECT ext FROM __resp WHERE gid=?", gid).Scan(&raw) != nil {
			checkVisionFail("AI row unavailable")
		}
		var fields map[string]any
		if json.Unmarshal([]byte(raw), &fields) != nil {
			checkVisionFail("invalid AI settings")
		}
		return fields
	}
	before := read(filepath.Join(root, "plugins.sqlite"))
	after := read("data/control/plugins.db")
	visionFields := map[string]bool{"ImageAPI": true, "ImageType": true, "ImageModelName": true, "ImageKey": true}
	unrelated := true
	for k, v := range before {
		if !visionFields[k] && !reflect.DeepEqual(v, after[k]) {
			unrelated = false
		}
	}
	for k := range after {
		if !visionFields[k] {
			if _, ok := before[k]; !ok {
				unrelated = false
			}
		}
	}
	key, ok := before["Key"].(string)
	configured := ok && key != "" && after["ImageKey"] == key && after["ImageAPI"] == "https://api.deepseek.com/chat/completions" && after["ImageModelName"] == "deepseek-flash" && after["ImageType"] == float64(0)
	result := map[string]any{"passed": configured && unrelated, "visionConfigured": configured, "imageKeyMatchesExistingChatKey": ok && key != "" && after["ImageKey"] == key, "unrelatedAIFieldsPreserved": unrelated, "backupIntegrityPassed": true, "backupID": *id}
	_ = json.NewEncoder(os.Stdout).Encode(result)
	if !configured || !unrelated {
		os.Exit(1)
	}
}
func checkVisionFail(s string) { fmt.Fprintln(os.Stderr, s); os.Exit(1) }
