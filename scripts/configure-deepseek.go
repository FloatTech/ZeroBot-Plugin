//go:build ignore

// Offline, backed-up edit of the shared AI row, preserving every unrelated field.
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
	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	backupID := flag.String("backup-id", "", "Existing consistent private control snapshot")
	restore := flag.Bool("restore-ai", false, "Restore only the AI settings from this private snapshot")
	personaFile := flag.String("persona-file", "", "Apply only Agent personality/sex from the specified YAML")
	vision := flag.Bool("vision", false, "Configure only image fields using the verified existing DeepSeek provider and key")
	flag.Parse()
	if *vision && (*restore || *personaFile != "") {
		fail("vision cannot be combined with restore or persona edits")
	}
	if filepath.Base(*backupID) != *backupID || !strings.HasSuffix(*backupID, "-control") {
		fail("valid backup ID required")
	}
	conn, err := net.DialTimeout("tcp", "127.0.0.1:3010", time.Second)
	if err == nil {
		conn.Close()
		fail("stop ZeroBot before editing settings")
	}
	backupRoot := filepath.Join(".local", "admin", "backups", *backupID)
	manifestBytes, err := os.ReadFile(filepath.Join(backupRoot, "manifest.json"))
	if err != nil {
		fail("backup manifest unavailable")
	}
	var manifest struct{ Scope, SHA256 string }
	if json.Unmarshal(manifestBytes, &manifest) != nil || manifest.Scope != "control-db" {
		fail("backup manifest invalid")
	}
	backupBytes, err := os.ReadFile(filepath.Join(backupRoot, "plugins.sqlite"))
	if err != nil {
		fail("backup unavailable")
	}
	sum := sha256.Sum256(backupBytes)
	if hex.EncodeToString(sum[:]) != manifest.SHA256 {
		fail("backup integrity failed")
	}
	db, err := sql.Open("sqlite3", "file:data/control/plugins.db?mode=rw")
	if err != nil {
		fail("database unavailable")
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		fail("transaction unavailable")
	}
	defer tx.Rollback()
	h := md5.Sum([]byte("aichat"))
	i := byte('a') % 15
	gid := int16(binary.LittleEndian.Uint16(h[i : i+2]))
	var raw string
	if tx.QueryRow("SELECT ext FROM __resp WHERE gid = ?", gid).Scan(&raw) != nil {
		fail("AI settings unavailable")
	}
	backupDB, err := sql.Open("sqlite3", "file:"+filepath.ToSlash(filepath.Join(backupRoot, "plugins.sqlite"))+"?mode=ro")
	if err != nil {
		fail("backup database unavailable")
	}
	defer backupDB.Close()
	var previous string
	if backupDB.QueryRow("SELECT ext FROM __resp WHERE gid = ?", gid).Scan(&previous) != nil {
		fail("backup AI settings unavailable")
	}
	if *restore {
		result, err := tx.Exec("UPDATE __resp SET ext = ? WHERE gid = ? AND ext = ?", previous, gid, raw)
		if err != nil {
			fail("AI settings rollback failed")
		}
		n, err := result.RowsAffected()
		if err != nil || n != 1 || tx.Commit() != nil {
			fail("AI settings rollback commit failed")
		}
		fmt.Println(`{"aiSettingsRestored":true}`)
		return
	}
	if previous != raw {
		fail("backup does not match current AI settings")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &fields) != nil {
		fail("AI settings invalid")
	}
	var key, agentKey string
	_ = json.Unmarshal(fields["Key"], &key)
	_ = json.Unmarshal(fields["AgentKey"], &agentKey)
	if key == "" || agentKey == "" || key != agentKey {
		fail("verified matching keys required")
	}
	changes := map[string]any{"API": "https://api.deepseek.com/chat/completions", "AgentAPI": "https://api.deepseek.com/chat/completions", "Type": 0, "AgentType": 0, "ModelName": "deepseek-flash", "AgentModelName": "deepseek-flash", "ReasoningEffort": "none"}
	if *vision {
		var api, agentAPI, model string
		_ = json.Unmarshal(fields["API"], &api)
		_ = json.Unmarshal(fields["AgentAPI"], &agentAPI)
		_ = json.Unmarshal(fields["ModelName"], &model)
		if api != "https://api.deepseek.com/chat/completions" || agentAPI != api || model != "deepseek-flash" {
			fail("existing verified DeepSeek provider required for vision")
		}
		changes = map[string]any{"ImageAPI": api, "ImageModelName": "deepseek-flash", "ImageType": 0, "ImageKey": key}
	}
	if *personaFile != "" {
		content, err := os.ReadFile(*personaFile)
		if err != nil {
			fail("persona file unavailable")
		}
		var persona struct {
			Sex  string `yaml:"sex"`
			Char string `yaml:"char"`
		}
		if yaml.Unmarshal(content, &persona) != nil || persona.Char == "" || len(persona.Char) > 32768 || len(persona.Sex) > 256 {
			fail("persona file invalid")
		}
		changes = map[string]any{"AgentChar": persona.Char, "AgentSex": persona.Sex}
	}
	for name, value := range changes {
		fields[name], _ = json.Marshal(value)
	}
	body, err := json.Marshal(fields)
	if err != nil {
		fail("settings encoding failed")
	}
	result, err := tx.Exec("UPDATE __resp SET ext = ? WHERE gid = ? AND ext = ?", string(body), gid, raw)
	if err != nil {
		fail("settings write failed")
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		fail("AI settings changed concurrently")
	}
	if tx.Commit() != nil {
		fail("settings commit failed")
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"configured": true, "visionConfigured": *vision, "personaUpdated": *personaFile != "", "credentialsPreserved": true, "unrelatedFieldsPreserved": true, "backupID": *backupID})
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
