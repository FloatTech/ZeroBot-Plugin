package localadmin

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	db "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const maxBackupBytes = 32 << 20

var backupIDPattern = regexp.MustCompile(`^[0-9]{8}T[0-9]{6}\.[0-9]{9}Z-[0-9a-f]{16}(\.json|-control)$`)
var errBackupInvalid = errors.New("备份损坏、格式不支持或与当前程序不兼容")
var errBackupConflict = errors.New("当前配置或备份已变化，请重新预览后确认")

type BackupStore struct {
	mu                     sync.Mutex
	directory, optionsPath string
	connections            *ConnectionStore
	manager                *ctrl.Manager[*zero.Ctx]
	logs                   *LogStore
	key                    []byte
	now                    func() time.Time
	// Failure injection runs inside the database transaction, before commit.
	afterTable    func(int) error
	replace       func(string, string) error
	cleanupRemove func(*os.Root, string) error
}

func NewBackupStore(directory string, connections *ConnectionStore, plugins *PluginStore, logs *LogStore) *BackupStore {
	directory, _ = filepath.Abs(directory)
	s := &BackupStore{directory: directory, optionsPath: filepath.Join(filepath.Dir(directory), "options.json"), connections: connections, logs: logs, key: randomBytes(32), now: time.Now, replace: replaceConfigFile}
	if plugins != nil {
		for _, c := range plugins.controls {
			if s.manager != nil && s.manager != c.Manager { // Do not silently omit a second database.
				s.manager = nil
				break
			}
			s.manager = c.Manager
		}
	}
	return s
}

type BackupEntry struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Version   int       `json:"version"`
	Scope     string    `json:"scope"`
	Size      int64     `json:"size"`
	Integrity bool      `json:"integrity"`
	Valid     bool      `json:"valid"`
}

type fileBackup struct {
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"createdAt"`
	Scope     string          `json:"scope"`
	Config    json.RawMessage `json:"config"`
	SHA256    string          `json:"sha256,omitempty"`
}

type controlManifest struct {
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	Scope     string    `json:"scope"`
	File      string    `json:"file"`
	SHA256    string    `json:"sha256"`
}

func digestBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func safeDirectory(path string) error {
	// Reject linked ancestors too, including Windows junctions exposed as links.
	for p := filepath.Clean(path); ; p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errBackupInvalid
		}
		if filepath.Dir(p) == p {
			return nil
		}
	}
}

func readBoundedFile(path string, limit int64) ([]byte, error) {
	if err := safeDirectory(filepath.Dir(path)); err != nil {
		return nil, err
	}
	if err := regularFile(path); err != nil {
		return nil, errBackupInvalid
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, errBackupInvalid
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(raw)) > limit {
		return nil, errBackupInvalid
	}
	return raw, nil
}

func validateOptions(raw []byte) error {
	obj, err := object(raw)
	if err != nil || len(obj) != 2 {
		return errBackupInvalid
	}
	var enabled bool
	var address string
	if bytes.Equal(bytes.TrimSpace(obj["enabled"]), []byte("null")) || json.Unmarshal(obj["enabled"], &enabled) != nil || json.Unmarshal(obj["address"], &address) != nil {
		return errBackupInvalid
	}
	host, port, err := net.SplitHostPort(address)
	n, parseErr := strconv.Atoi(port)
	if err != nil || parseErr != nil || host != "127.0.0.1" || n < 1 || n > 65535 || strconv.Itoa(n) != port {
		return errBackupInvalid
	}
	return nil
}

func validateFileConfig(scope string, raw []byte) error {
	if len(raw) > 1<<20 {
		return errBackupInvalid
	}
	switch scope {
	case "startup-config":
		if _, err := decodeConfig(raw); err != nil {
			return errBackupInvalid
		}
	case "admin-options":
		return validateOptions(raw)
	default:
		return errBackupInvalid
	}
	return nil
}

func (s *BackupStore) ensureDirectory() error {
	if err := os.MkdirAll(s.directory, 0700); err != nil {
		return errors.New("备份目录不可写")
	}
	if err := safeDirectory(s.directory); err != nil {
		return err
	}
	if err := protectPath(s.directory, true); err != nil {
		return errors.New("备份目录权限设置失败")
	}
	return nil
}

func writeFileBackup(directory, scope string, raw []byte) (string, error) {
	id := time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + hex.EncodeToString(randomBytes(8)) + ".json"
	// Use the exact compact/HTML-escaped form the outer JSON encoder will store.
	canonical, err := json.Marshal(json.RawMessage(raw))
	if err != nil || validateFileConfig(scope, canonical) != nil {
		return "", errBackupInvalid
	}
	entry := fileBackup{1, time.Now().UTC(), scope, canonical, digestBytes(canonical)}
	b, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	if err = writeNewPrivateFile(filepath.Join(directory, id), b); err != nil {
		return "", errors.New("配置备份写入失败，原配置保留")
	}
	return id, nil
}

func (s *BackupStore) load(id string) (BackupEntry, []byte, error) {
	entry := BackupEntry{ID: id}
	if !backupIDPattern.MatchString(id) {
		return entry, nil, errBackupInvalid
	}
	if strings.HasSuffix(id, ".json") {
		b, err := readBoundedFile(filepath.Join(s.directory, id), (1<<20)+4096)
		if err != nil {
			return entry, nil, err
		}
		var file fileBackup
		if json.Unmarshal(b, &file) != nil || file.Version != 1 || file.CreatedAt.IsZero() || validateFileConfig(file.Scope, file.Config) != nil {
			return entry, nil, errBackupInvalid
		}
		if file.SHA256 != "" && file.SHA256 != digestBytes(file.Config) {
			return entry, nil, errBackupInvalid
		}
		entry.CreatedAt, entry.Version, entry.Scope, entry.Size, entry.Integrity, entry.Valid = file.CreatedAt, file.Version, file.Scope, int64(len(file.Config)), file.SHA256 != "", true
		return entry, file.Config, nil
	}
	root := filepath.Join(s.directory, id)
	b, err := readBoundedFile(filepath.Join(root, "manifest.json"), 4096)
	if err != nil {
		return entry, nil, err
	}
	var manifest controlManifest
	if json.Unmarshal(b, &manifest) != nil || manifest.Version != 1 || manifest.Scope != "control-db" || manifest.File != "plugins.sqlite" || manifest.CreatedAt.IsZero() || !validRevision(manifest.SHA256) {
		return entry, nil, errBackupInvalid
	}
	raw, err := readBoundedFile(filepath.Join(root, "plugins.sqlite"), maxBackupBytes)
	if err != nil || manifest.SHA256 != digestBytes(raw) {
		return entry, nil, errBackupInvalid
	}
	entry.CreatedAt, entry.Version, entry.Scope, entry.Size, entry.Integrity, entry.Valid = manifest.CreatedAt, 1, manifest.Scope, int64(len(raw)), true, true
	return entry, raw, nil
}

func (s *BackupStore) List() ([]BackupEntry, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries := []BackupEntry{}
	if _, err := os.Lstat(s.directory); os.IsNotExist(err) {
		return entries, false, nil
	}
	if err := safeDirectory(s.directory); err != nil {
		return nil, false, err
	}
	f, err := os.Open(s.directory)
	if err != nil {
		return nil, false, errBackupInvalid
	}
	defer f.Close()
	ids := []string{}
	for count := 0; ; {
		part, err := f.ReadDir(128)
		count += len(part)
		if count > 4096 {
			return nil, false, errors.New("备份目录超过 4096 项，请先在本机整理旧备份")
		}
		for _, item := range part {
			if backupIDPattern.MatchString(item.Name()) {
				ids = append(ids, item.Name())
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, false, errBackupInvalid
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	more := len(ids) > 200
	if more {
		ids = ids[:200]
	}
	for _, id := range ids {
		entry, _, err := s.load(id)
		if err != nil {
			entry = BackupEntry{ID: id}
		}
		entries = append(entries, entry)
	}
	return entries, more, nil
}

func (s *BackupStore) Create(scope string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureDirectory(); err != nil {
		return "", err
	}
	switch scope {
	case "control-db":
		if s.manager == nil {
			return "", errBackupInvalid
		}
		s.manager.Lock()
		defer s.manager.Unlock()
		return BackupControlDatabase(&s.manager.D, s.directory)
	case "startup-config":
		if s.connections == nil {
			return "", errBackupInvalid
		}
		s.connections.mu.Lock()
		defer s.connections.mu.Unlock()
		raw, err := s.connections.read()
		if err != nil {
			return "", errBackupInvalid
		}
		return writeFileBackup(s.directory, scope, raw)
	case "admin-options":
		raw, err := readBoundedFile(s.optionsPath, 4096)
		if err != nil || validateOptions(raw) != nil {
			return "", errBackupInvalid
		}
		return writeFileBackup(s.directory, scope, raw)
	default:
		return "", errBackupInvalid
	}
}

// All source SQL is inspected, never executed. Existing live schema must match.
type schemaItem struct{ Type, Name, Table, SQL string }
type tableCopy struct {
	Name    string
	Columns []string
	Rows    [][]any
}
type databaseCopy struct {
	Schema []schemaItem
	Tables []tableCopy
}

func quoteName(name string) string { return `"` + strings.ReplaceAll(name, `"`, `""`) + `"` }

func readDatabase(ctx context.Context, tx *sql.Tx) (*databaseCopy, error) {
	var check string
	if tx.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&check) != nil || check != "ok" {
		return nil, errBackupInvalid
	}
	rows, err := tx.QueryContext(ctx, "SELECT type,name,tbl_name,coalesce(sql,'') FROM sqlite_master WHERE name NOT GLOB 'sqlite_*' ORDER BY type,name")
	if err != nil {
		return nil, err
	}
	copy := &databaseCopy{}
	for rows.Next() {
		var item schemaItem
		if err = rows.Scan(&item.Type, &item.Name, &item.Table, &item.SQL); err != nil {
			break
		}
		if item.Type != "table" && item.Type != "index" {
			err = errBackupInvalid
			break
		}
		copy.Schema = append(copy.Schema, item)
		if len(copy.Schema) > 512 {
			err = errBackupInvalid
			break
		}
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return nil, err
	}
	count, memory := 0, 0
	for _, item := range copy.Schema {
		if item.Type != "table" {
			continue
		}
		rows, err := tx.QueryContext(ctx, "SELECT * FROM "+quoteName(item.Name)+" ORDER BY 1")
		if err != nil {
			return nil, err
		}
		columns, err := rows.Columns()
		if err != nil || len(columns) == 0 || len(columns) > 32 {
			rows.Close()
			return nil, errBackupInvalid
		}
		table := tableCopy{Name: item.Name, Columns: columns, Rows: [][]any{}}
		for rows.Next() {
			values := make([]any, len(columns))
			destinations := make([]any, len(columns))
			for i := range values {
				destinations[i] = &values[i]
			}
			if err = rows.Scan(destinations...); err != nil {
				break
			}
			for i, value := range values {
				switch v := value.(type) {
				case []byte:
					values[i] = bytes.Clone(v)
					memory += len(v)
				case string:
					memory += len(v)
				default:
					memory += 16
				}
			}
			count++
			memory += len(columns) * 24
			if count > 100000 || memory > 64<<20 {
				err = errBackupInvalid
				break
			}
			table.Rows = append(table.Rows, values)
		}
		if err == nil {
			err = rows.Err()
		}
		rows.Close()
		if err != nil {
			return nil, err
		}
		copy.Tables = append(copy.Tables, table)
	}
	return copy, nil
}

func (s *BackupStore) sourceDatabase(ctx context.Context, raw []byte) (*databaseCopy, error) {
	if err := s.ensureDirectory(); err != nil {
		return nil, err
	}
	// Work from private verified bytes so file edits cannot race SQLite reads.
	f, err := os.CreateTemp(s.directory, ".restore-check-*.sqlite")
	if err != nil {
		return nil, errBackupInvalid
	}
	path := f.Name()
	defer os.Remove(path)
	if err = protectPath(path, false); err == nil {
		_, err = f.Write(raw)
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return nil, errBackupInvalid
	}
	database, err := sql.Open(db.DriverName, filepath.ToSlash(path))
	if err != nil {
		return nil, errBackupInvalid
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err = database.ExecContext(ctx, "PRAGMA query_only=ON"); err != nil {
		return nil, errBackupInvalid
	}
	tx, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, errBackupInvalid
	}
	defer tx.Rollback()
	copy, err := readDatabase(ctx, tx)
	if err != nil {
		return nil, errBackupInvalid
	}
	return copy, nil
}

func (s *BackupStore) revision(value any) string {
	b, _ := json.Marshal(value)
	h := hmac.New(sha256.New, s.key)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
func defaultBits(copy *databaseCopy) map[string]int64 {
	result := map[string]int64{}
	for _, table := range copy.Tables {
		if len(table.Columns) != 2 || table.Columns[0] != "gid" || table.Columns[1] != "disable" {
			continue
		}
		for _, row := range table.Rows {
			if row[0] == int64(0) {
				if value, ok := row[1].(int64); ok {
					result[table.Name] = value
				}
			}
		}
	}
	return result
}
func (s *BackupStore) protectRestored(raw []byte, copy *databaseCopy) {
	if s.logs == nil {
		return
	}
	s.logs.ProtectConfig(raw)
	if copy != nil {
		for _, table := range copy.Tables {
			if table.Name == "__resp" {
				for _, row := range table.Rows {
					for _, value := range row {
						if text, ok := value.(string); ok && json.Valid([]byte(text)) {
							s.logs.ProtectConfig([]byte(text))
						}
					}
				}
			}
		}
	}
}

type RestorePreview struct {
	Backup          BackupEntry     `json:"backup"`
	Revision        string          `json:"revision"`
	ExpiresAt       int64           `json:"expiresAt"`
	Token           string          `json:"token"`
	Changed         bool            `json:"changed"`
	RestartRequired bool            `json:"restartRequired"`
	Tables          int             `json:"tables"`
	Rows            int             `json:"rows"`
	ChangedTables   int             `json:"changedTables"`
	Warnings        []string        `json:"warnings"`
	Impacts         []RestoreImpact `json:"impacts"`
}
type RestoreRequest struct {
	ID        string `json:"id"`
	Revision  string `json:"revision"`
	ExpiresAt int64  `json:"expiresAt"`
	Token     string `json:"token"`
	Confirm   bool   `json:"confirm"`
}

func (s *BackupStore) ticket(id, digest, revision, subject string, expiry int64) string {
	return s.revision([]any{id, digest, revision, subject, expiry})
}

func (s *BackupStore) Preview(ctx context.Context, id, subject string) (RestorePreview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, raw, err := s.load(id)
	if err != nil {
		return RestorePreview{}, err
	}
	v := RestorePreview{Backup: entry, ExpiresAt: s.now().Add(10 * time.Minute).Unix(), Warnings: []string{"一次仅恢复所选备份的范围；不恢复 QQ 登录、聊天媒体、后台密码或会话。"}}
	if !entry.Integrity {
		v.Warnings = append(v.Warnings, "此旧版 JSON 备份没有校验摘要，格式有效但无法证明内容未被修改。")
	}
	if entry.Scope == "control-db" {
		if s.manager == nil {
			return v, errBackupInvalid
		}
		source, err := s.sourceDatabase(ctx, raw)
		if err != nil {
			return v, err
		}
		s.manager.Lock()
		defer s.manager.Unlock()
		err = s.manager.D.TransactionContext(ctx, func(tx *sql.Tx) error {
			current, err := readDatabase(ctx, tx)
			if err != nil {
				return err
			}
			if s.revision(source.Schema) != s.revision(current.Schema) {
				return errBackupInvalid
			}
			v.Revision = s.revision(current)
			v.Changed = s.revision(source) != v.Revision
			v.Tables = len(source.Tables)
			v.Impacts = newDatabaseImpacts()
			for i, table := range source.Tables {
				v.Rows += len(table.Rows)
				impact := &v.Impacts[databaseImpactBucket(table.Name)]
				*impact.CurrentCount += len(current.Tables[i].Rows)
				*impact.BackupCount += len(table.Rows)
				if s.revision(table) != s.revision(current.Tables[i]) {
					v.ChangedTables++
					impact.Changed = true
				}
			}
			return nil
		})
		if err != nil {
			return v, errBackupInvalid
		}
		v.RestartRequired = v.Changed
		v.Warnings = append(v.Warnings, "覆盖整个控制库：插件开关、群参数、AI 扩展、响应/封禁及库内插件数据。控制层立即更新；其他插件缓存可能需要重启。")
	} else {
		current, unlock, err := s.currentFile(entry.Scope)
		if err != nil {
			return v, err
		}
		defer unlock()
		v.Revision = s.revision(current)
		v.Changed = !equivalentConfig(current, raw)
		v.Impacts, err = fileRestoreImpacts(entry.Scope, current, raw)
		if err != nil {
			return v, err
		}
		v.RestartRequired = v.Changed
		if entry.Scope == "startup-config" {
			v.Warnings = append(v.Warnings, "覆盖完整启动 JSON，包括所有连接、Token、主人和未知字段；重启后生效。")
		} else {
			v.Warnings = append(v.Warnings, "覆盖后台启用与回环地址设置；重启后可能关闭后台或改变访问端口。")
		}
	}
	v.Token = s.ticket(id, digestBytes(raw), v.Revision, subject, v.ExpiresAt)
	return v, nil
}

// Caller must invoke unlock even when it elects not to write.
func (s *BackupStore) currentFile(scope string) ([]byte, func(), error) {
	if scope == "startup-config" && s.connections != nil {
		s.connections.mu.Lock()
		raw, err := s.connections.read()
		if err != nil {
			s.connections.mu.Unlock()
			return nil, func() {}, errBackupInvalid
		}
		return raw, s.connections.mu.Unlock, nil
	}
	if scope == "admin-options" {
		raw, err := readBoundedFile(s.optionsPath, 4096)
		if err != nil || validateOptions(raw) != nil {
			return nil, func() {}, errBackupInvalid
		}
		return raw, func() {}, nil
	}
	return nil, func() {}, errBackupInvalid
}

type RestoreResult struct {
	BackupID        string `json:"backupId"`
	Scope           string `json:"scope"`
	Changed         bool   `json:"changed"`
	RestartRequired bool   `json:"restartRequired"`
}

func (s *BackupStore) Restore(ctx context.Context, request RestoreRequest, subject string) (RestoreResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := RestoreResult{}
	if !request.Confirm || !validRevision(request.Revision) || !validRevision(request.Token) {
		return result, errBackupInvalid
	}
	if request.ExpiresAt <= s.now().Unix() || request.ExpiresAt > s.now().Add(10*time.Minute).Unix() {
		return result, errBackupConflict
	}
	entry, raw, err := s.load(request.ID)
	if err != nil {
		return result, err
	}
	want := s.ticket(request.ID, digestBytes(raw), request.Revision, subject, request.ExpiresAt)
	if !hmac.Equal([]byte(want), []byte(request.Token)) {
		return result, errBackupConflict
	}
	result.Scope = entry.Scope
	if err = s.ensureDirectory(); err != nil {
		return result, err
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if entry.Scope == "control-db" {
		if s.manager == nil {
			return result, errBackupInvalid
		}
		source, err := s.sourceDatabase(ctx, raw)
		if err != nil {
			return result, err
		}
		s.manager.Lock()
		defer s.manager.Unlock()
		var current *databaseCopy
		err = s.manager.D.TransactionContext(ctx, func(tx *sql.Tx) error { var err error; current, err = readDatabase(ctx, tx); return err })
		if err != nil {
			return result, errBackupInvalid
		}
		if s.revision(current) != request.Revision {
			return result, errBackupConflict
		}
		if s.revision(source.Schema) != s.revision(current.Schema) {
			return result, errBackupInvalid
		}
		if s.revision(source) == request.Revision {
			return result, nil
		}
		result.BackupID, err = BackupControlDatabase(&s.manager.D, s.directory)
		if err != nil {
			return result, err
		}
		s.protectRestored(nil, source)
		err = s.manager.D.TransactionContext(ctx, func(tx *sql.Tx) error {
			// Recheck inside the write transaction, including outside DB writers.
			latest, err := readDatabase(ctx, tx)
			if err != nil {
				return err
			}
			if s.revision(latest) != request.Revision {
				return errBackupConflict
			}
			for i, table := range source.Tables {
				if _, err = tx.ExecContext(ctx, "DELETE FROM "+quoteName(table.Name)); err != nil {
					return err
				}
				columns := []string{}
				for _, name := range table.Columns {
					columns = append(columns, quoteName(name))
				}
				query := "INSERT INTO " + quoteName(table.Name) + " (" + strings.Join(columns, ",") + ") VALUES (" + strings.TrimSuffix(strings.Repeat("?,", len(columns)), ",") + ")"
				stmt, err := tx.PrepareContext(ctx, query)
				if err != nil {
					return err
				}
				for _, row := range table.Rows {
					if _, err = stmt.ExecContext(ctx, row...); err != nil {
						break
					}
				}
				stmt.Close()
				if err != nil {
					return err
				}
				if s.afterTable != nil {
					if err = s.afterTable(i); err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			if errors.Is(err, errBackupConflict) {
				return result, err
			}
			return result, errors.New("控制库恢复失败，事务已回滚，原配置与缓存保留；恢复前备份已创建")
		}
		s.manager.RestoredLocked(defaultBits(current), defaultBits(source))
	} else {
		current, unlock, err := s.currentFile(entry.Scope)
		if err != nil {
			return result, err
		}
		defer unlock()
		if s.revision(current) != request.Revision {
			return result, errBackupConflict
		}
		if equivalentConfig(current, raw) {
			return result, nil
		}
		result.BackupID, err = writeFileBackup(s.directory, entry.Scope, current)
		if err != nil {
			return result, err
		}
		path, replace := s.optionsPath, s.replace
		if entry.Scope == "startup-config" {
			path, replace = s.connections.path, s.connections.replace
		}
		f, err := os.CreateTemp(filepath.Dir(path), ".admin-restore-*.tmp")
		if err != nil {
			return result, errors.New("恢复文件创建失败，原配置保留")
		}
		tmp := f.Name()
		defer os.Remove(tmp)
		if err = protectPath(tmp, false); err == nil {
			_, err = f.Write(raw)
		}
		if err == nil {
			err = f.Sync()
		}
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
		if err == nil {
			latest, readErr := readBoundedFile(path, 1<<20)
			if readErr != nil || !bytes.Equal(latest, current) {
				return result, errBackupConflict
			}
			if err = ctx.Err(); err == nil {
				s.protectRestored(raw, nil)
				err = replace(tmp, path)
			}
		}
		if err != nil {
			return result, errors.New("配置恢复失败，原文件保留；恢复前备份已创建")
		}
	}
	result.Changed, result.RestartRequired = true, true
	return result, nil
}

func (s *Server) backupError(w http.ResponseWriter, err error) {
	status := 500
	if errors.Is(err, errBackupInvalid) {
		status = 400
	}
	if errors.Is(err, errBackupConflict) {
		status = 409
	}
	// Database and filesystem errors may contain paths or configuration text.
	message := "备份操作失败，原配置保留；请检查本机备份目录或数据库"
	if status != 500 {
		message = err.Error()
	}
	apiError(w, status, message)
}
func (s *Server) getBackups(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Backups == nil {
		apiError(w, 409, "当前程序未接入备份管理")
		return
	}
	entries, more, err := s.cfg.Backups.List()
	if err != nil {
		s.backupError(w, err)
		return
	}
	jsonReply(w, 200, map[string]any{"backups": entries, "more": more, "scopes": map[string]bool{"startup-config": s.cfg.Backups.connections != nil, "control-db": s.cfg.Backups.manager != nil, "admin-options": validateOptionsFile(s.cfg.Backups.optionsPath)}})
}
func validateOptionsFile(path string) bool {
	raw, err := readBoundedFile(path, 4096)
	return err == nil && validateOptions(raw) == nil
}
func (s *Server) createBackup(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Backups == nil {
		apiError(w, 409, "当前程序未接入备份管理")
		return
	}
	var req struct {
		Scope string `json:"scope"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	id, err := s.cfg.Backups.Create(req.Scope)
	if err != nil {
		s.backupError(w, err)
		return
	}
	jsonReply(w, 200, map[string]string{"backupId": id})
}
func (s *Server) previewBackup(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Backups == nil {
		apiError(w, 409, "当前程序未接入备份管理")
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	cookie, _ := r.Cookie("local_admin_session")
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	v, err := s.cfg.Backups.Preview(ctx, req.ID, cookie.Value)
	if err != nil {
		s.backupError(w, err)
		return
	}
	jsonReply(w, 200, v)
}
func (s *Server) restoreBackup(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Backups == nil {
		apiError(w, 409, "当前程序未接入备份管理")
		return
	}
	var req RestoreRequest
	if !readJSON(w, r, &req) {
		return
	}
	cookie, _ := r.Cookie("local_admin_session")
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	v, err := s.cfg.Backups.Restore(ctx, req, cookie.Value)
	if err != nil {
		s.backupError(w, err)
		return
	}
	jsonReply(w, 200, v)
}
