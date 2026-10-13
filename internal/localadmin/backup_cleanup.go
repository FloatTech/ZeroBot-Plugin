package localadmin

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxCleanupScanBytes = 256 << 20

type CleanupPolicy struct {
	KeepPerScope int `json:"keepPerScope"`
	KeepDays     int `json:"keepDays"`
}

func (p CleanupPolicy) valid() bool {
	return p.KeepPerScope >= 3 && p.KeepPerScope <= 200 && p.KeepDays >= 1 && p.KeepDays <= 3650
}

type CleanupPreview struct {
	Policy     CleanupPolicy `json:"policy"`
	Candidates []BackupEntry `json:"candidates"`
	Bytes      int64         `json:"bytes"`
	Skipped    int           `json:"skipped"`
	More       bool          `json:"more"`
	Revision   string        `json:"revision"`
	AsOf       int64         `json:"asOf"`
	ExpiresAt  int64         `json:"expiresAt"`
	Token      string        `json:"token"`
}

type CleanupRequest struct {
	Policy    CleanupPolicy `json:"policy"`
	Revision  string        `json:"revision"`
	AsOf      int64         `json:"asOf"`
	ExpiresAt int64         `json:"expiresAt"`
	Token     string        `json:"token"`
	Confirm   bool          `json:"confirm"`
}

type CleanupResult struct {
	Deleted []string `json:"deleted"`
	Bytes   int64    `json:"bytes"`
	Partial bool     `json:"partial"`
	Message string   `json:"message"`
}

type cleanupItem struct {
	Entry BackupEntry
	Hash  string
	Bytes int64
	Files []string
}

// Only the expected regular files are eligible. Unknown children are retained.
// All deletion is relative to os.Root, so links cannot lead outside this tree.
func (s *BackupStore) cleanupFiles(id string) ([]string, int64, error) {
	if !backupIDPattern.MatchString(id) || filepath.Base(id) != id {
		return nil, 0, errBackupInvalid
	}
	files := []string{id}
	if strings.HasSuffix(id, "-control") {
		dir := filepath.Join(s.directory, id)
		if safeDirectory(dir) != nil {
			return nil, 0, errBackupInvalid
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 2 {
			return nil, 0, errBackupInvalid
		}
		files = []string{filepath.Join(id, "plugins.sqlite"), filepath.Join(id, "manifest.json")}
	}
	var bytes int64
	for _, file := range files {
		path := filepath.Join(s.directory, file)
		if regularFile(path) != nil {
			return nil, 0, errBackupInvalid
		}
		info, err := os.Lstat(path)
		if err != nil || info.Size() > maxBackupBytes+4096 {
			return nil, 0, errBackupInvalid
		}
		bytes += info.Size()
	}
	return files, bytes, nil
}

func (s *BackupStore) cleanupPlan(ctx context.Context, policy CleanupPolicy, asOf int64) (CleanupPreview, []cleanupItem, error) {
	v := CleanupPreview{Policy: policy, AsOf: asOf, Candidates: []BackupEntry{}}
	if !policy.valid() || asOf <= 0 {
		return v, nil, errBackupInvalid
	}
	if _, err := os.Lstat(s.directory); os.IsNotExist(err) {
		v.Revision = s.revision([]any{policy, asOf, "empty"})
		return v, nil, nil
	}
	if safeDirectory(s.directory) != nil {
		return v, nil, errBackupInvalid
	}
	dir, err := os.Open(s.directory)
	if err != nil {
		return v, nil, errBackupInvalid
	}
	defer dir.Close()
	entries, err := dir.ReadDir(4097)
	if err != nil && err != io.EOF {
		return v, nil, errBackupInvalid
	}
	if len(entries) > 4096 {
		return v, nil, errors.New("备份目录超过 4096 项，请先在本机整理后再预览")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	all := []cleanupItem{}
	inventory := []any{}
	var scanned int64
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return v, nil, err
		}
		id := entry.Name()
		if !backupIDPattern.MatchString(id) {
			continue
		}
		files, size, err := s.cleanupFiles(id)
		// Ineligible items are never deleted; sign their presence/eligibility,
		// not filesystem timestamps that may change when directories are read.
		if err != nil {
			v.Skipped++
			inventory = append(inventory, []any{id, "skipped"})
			continue
		}
		scanned += size
		if scanned > maxCleanupScanBytes {
			return v, nil, errors.New("备份扫描超过 256 MiB，请先在本机整理后再预览")
		}
		backup, raw, err := s.load(id)
		if err != nil || !backup.Integrity {
			v.Skipped++
			inventory = append(inventory, []any{id, "skipped"})
			continue
		}
		item := cleanupItem{Entry: backup, Hash: digestBytes(raw), Bytes: size, Files: files}
		all = append(all, item)
		inventory = append(inventory, item)
	}
	v.Revision = s.revision([]any{policy, asOf, inventory})
	sort.Slice(all, func(i, j int) bool {
		if all[i].Entry.CreatedAt.Equal(all[j].Entry.CreatedAt) {
			return all[i].Entry.ID > all[j].Entry.ID
		}
		return all[i].Entry.CreatedAt.After(all[j].Entry.CreatedAt)
	})
	counts := map[string]int{}
	selected := []cleanupItem{}
	cutoff := time.Unix(asOf, 0).AddDate(0, 0, -policy.KeepDays)
	for _, item := range all {
		counts[item.Entry.Scope]++
		if counts[item.Entry.Scope] <= policy.KeepPerScope || !item.Entry.CreatedAt.Before(cutoff) {
			continue
		}
		if len(selected) == 100 {
			v.More = true
			continue
		}
		selected = append(selected, item)
		v.Candidates = append(v.Candidates, item.Entry)
		v.Bytes += item.Bytes
	}
	return v, selected, nil
}

func (s *BackupStore) cleanupTicket(req CleanupRequest, subject string) string {
	return s.revision([]any{"cleanup", req.Policy, req.Revision, req.AsOf, req.ExpiresAt, subject})
}

func (s *BackupStore) PreviewCleanup(ctx context.Context, policy CleanupPolicy, subject string) (CleanupPreview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, _, err := s.cleanupPlan(ctx, policy, s.now().Unix())
	if err != nil {
		return v, err
	}
	v.ExpiresAt = v.AsOf + 600
	v.Token = s.cleanupTicket(CleanupRequest{Policy: v.Policy, Revision: v.Revision, AsOf: v.AsOf, ExpiresAt: v.ExpiresAt}, subject)
	return v, nil
}

func (s *BackupStore) Cleanup(ctx context.Context, req CleanupRequest, subject string) (CleanupResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := CleanupResult{Deleted: []string{}}
	if !req.Confirm || !req.Policy.valid() {
		return result, errBackupInvalid
	}
	if req.AsOf > s.now().Unix() || req.ExpiresAt <= s.now().Unix() || req.ExpiresAt-req.AsOf > 600 || req.ExpiresAt <= req.AsOf {
		return result, errBackupConflict
	}
	token, err := hex.DecodeString(req.Token)
	expected, _ := hex.DecodeString(s.cleanupTicket(req, subject))
	if err != nil || !hmac.Equal(token, expected) {
		return result, errBackupConflict
	}
	v, items, err := s.cleanupPlan(ctx, req.Policy, req.AsOf)
	if err != nil {
		return result, err
	}
	if v.Revision != req.Revision || req.ExpiresAt <= s.now().Unix() {
		return result, errBackupConflict
	}
	if len(items) == 0 {
		result.Message = "没有符合条件的旧备份。"
		return result, nil
	}
	root, err := os.OpenRoot(s.directory)
	if err != nil {
		return result, errBackupInvalid
	}
	defer root.Close()
	for _, item := range items {
		// Recheck each target before deleting; partial failures are reported
		// explicitly, never presented as an atomic all-or-nothing cleanup.
		current, raw, readErr := s.load(item.Entry.ID)
		files, _, filesErr := s.cleanupFiles(item.Entry.ID)
		if ctx.Err() != nil || readErr != nil || filesErr != nil || !current.Integrity || digestBytes(raw) != item.Hash {
			result.Partial = true
			break
		}
		for _, file := range files {
			if s.cleanupRemove != nil {
				err = s.cleanupRemove(root, file)
			} else {
				err = root.Remove(file)
			}
			if err != nil {
				result.Partial = true
				break
			}
		}
		if result.Partial {
			break
		}
		if strings.HasSuffix(item.Entry.ID, "-control") && root.Remove(item.Entry.ID) != nil {
			result.Partial = true
		}
		result.Deleted = append(result.Deleted, item.Entry.ID)
		result.Bytes += item.Bytes
		if result.Partial {
			break
		}
	}
	result.Message = "清理完成，当前配置未改变。"
	if result.Partial {
		result.Message = "清理已停止，部分文件可能已删除；请刷新列表检查实际状态后重新预览。"
	}
	return result, nil
}

func (s *Server) previewBackupCleanup(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Backups == nil {
		apiError(w, 409, "当前程序未接入备份管理")
		return
	}
	var policy CleanupPolicy
	if !readJSON(w, r, &policy) {
		return
	}
	cookie, _ := r.Cookie("local_admin_session")
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	value, err := s.cfg.Backups.PreviewCleanup(ctx, policy, cookie.Value)
	if err != nil {
		s.backupError(w, err)
		return
	}
	jsonReply(w, 200, value)
}

func (s *Server) cleanupBackups(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Backups == nil {
		apiError(w, 409, "当前程序未接入备份管理")
		return
	}
	var req CleanupRequest
	if !readJSON(w, r, &req) {
		return
	}
	cookie, _ := r.Cookie("local_admin_session")
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	value, err := s.cfg.Backups.Cleanup(ctx, req, cookie.Value)
	if err != nil {
		s.backupError(w, err)
		return
	}
	jsonReply(w, 200, value)
}
