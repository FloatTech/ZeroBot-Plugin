package localadmin

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var errConfigConflict = errors.New("配置已变化，请重新加载后再保存")
var errConfigValidation = errors.New("配置校验失败")

type jsonObject map[string]json.RawMessage

// ConnectionStore changes only explicitly selected fields in the startup JSON.
// Running drivers are never changed by saving; a restart loads the saved values.
type ConnectionStore struct {
	mu            sync.Mutex
	path, backups string
	running       []byte
	revisionKey   []byte
	// replace is injectable for write failure tests.
	replace func(string, string) error
}

type ConnectionView struct {
	Kind            string `json:"kind"`
	Index           int    `json:"index"`
	URL             string `json:"url"`
	TokenConfigured bool   `json:"tokenConfigured"`
	AddressRedacted bool   `json:"addressRedacted"`
}

type ConnectionSettings struct {
	Available          bool                `json:"available"`
	Revision           string              `json:"revision"`
	Connections        []ConnectionView    `json:"connections"`
	SuperUsers         []string            `json:"superUsers"`
	RunningConnections []ConnectionView    `json:"runningConnections"`
	RunningSuperUsers  []string            `json:"runningSuperUsers"`
	RestartRequired    bool                `json:"restartRequired"`
	BackupID           string              `json:"backupId,omitempty"`
	StartupDifferences []StartupDifference `json:"startupDifferences"`
}

type ConnectionChange struct {
	Kind        string  `json:"kind"`
	Index       int     `json:"index"`
	URL         *string `json:"url,omitempty"`
	TokenAction string  `json:"tokenAction"`
	Token       string  `json:"token,omitempty"`
}

type ConnectionUpdate struct {
	Revision         string              `json:"revision"`
	Connection       *ConnectionChange   `json:"connection,omitempty"`
	SuperUsers       *[]string           `json:"superUsers,omitempty"`
	AddConnection    *ConnectionAddition `json:"addConnection,omitempty"`
	RemoveConnection *ConnectionRemoval  `json:"removeConnection,omitempty"`
}

func NewConnectionStore(path, backups string, running []byte) (*ConnectionStore, error) {
	if path == "" {
		return nil, errors.New("未使用启动配置文件")
	}
	p, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	b, err := filepath.Abs(backups)
	if err != nil {
		return nil, err
	}
	if _, err = decodeConfig(running); err != nil {
		return nil, err
	}
	s := &ConnectionStore{path: p, backups: b, running: bytes.Clone(running), revisionKey: randomBytes(32), replace: replaceConfigFile}
	if _, err = s.read(); err != nil {
		return nil, err
	}
	return s, nil
}

func object(raw []byte) (jsonObject, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	token, err := dec.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("配置格式无效，须为 JSON 对象")
	}
	result := jsonObject{}
	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return nil, errors.New("配置格式无效")
		}
		key, ok := token.(string)
		if !ok {
			return nil, errors.New("配置格式无效")
		}
		if _, exists := result[key]; exists {
			return nil, errors.New("配置字段重复，请先修正本地文件")
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, errors.New("配置格式无效")
		}
		result[key] = value
	}
	if _, err := dec.Token(); err != nil {
		return nil, errors.New("配置格式无效")
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, errors.New("配置格式无效")
	}
	return result, nil
}

// Match Go's case-insensitive field handling but reject ambiguous aliases.
func field(obj jsonObject, name string) (string, json.RawMessage, error) {
	key := name
	var value json.RawMessage
	for k, v := range obj {
		if strings.EqualFold(k, name) {
			if value != nil {
				return "", nil, errors.New("配置字段存在重复别名")
			}
			key, value = k, v
		}
	}
	return key, value, nil
}

func decodeConfig(raw []byte) (jsonObject, error) {
	root, err := object(raw)
	if err != nil {
		return nil, err
	}
	_, z, err := field(root, "zero")
	if err != nil {
		return nil, err
	}
	zero, err := object(z)
	if err != nil {
		return nil, errors.New("启动配置缺少有效的 zero 对象")
	}
	_, users, err := field(zero, "super_users")
	if err != nil {
		return nil, err
	}
	var ids []int64
	if users != nil && json.Unmarshal(users, &ids) != nil {
		return nil, errors.New("主人账号配置格式无效")
	}
	for _, id := range ids {
		if id <= 0 {
			return nil, errors.New("主人账号须为正整数")
		}
	}
	for _, kind := range []string{"ws", "wss"} {
		_, rawList, err := field(root, kind)
		if err != nil {
			return nil, err
		}
		var list []json.RawMessage
		if rawList != nil && json.Unmarshal(rawList, &list) != nil {
			return nil, errors.New("连接列表格式无效")
		}
		for _, rawConn := range list {
			conn, err := object(rawConn)
			if err != nil {
				return nil, errors.New("连接须为 JSON 对象")
			}
			for _, key := range []string{"URL", "AccessToken"} {
				_, value, err := field(conn, key)
				if err != nil {
					return nil, err
				}
				var text string
				if value != nil && json.Unmarshal(value, &text) != nil {
					return nil, errors.New("连接字段格式无效")
				}
			}
		}
	}
	return root, nil
}

func (s *ConnectionStore) read() ([]byte, error) {
	if err := regularFile(s.path); err != nil {
		return nil, errors.New("配置文件不可读取或不是普通文件")
	}
	f, err := os.Open(s.path)
	if err != nil {
		return nil, errors.New("配置文件不可读取")
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if err != nil || len(b) > 1<<20 {
		return nil, errors.New("配置文件读取失败或超过 1 MiB")
	}
	if _, err = decodeConfig(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *ConnectionStore) revision(b []byte) string {
	h := hmac.New(sha256.New, s.revisionKey)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func configView(raw []byte) ([]ConnectionView, []string) {
	root, _ := decodeConfig(raw)
	connections := []ConnectionView{}
	for _, kind := range []string{"ws", "wss"} {
		_, rawList, _ := field(root, kind)
		var list []json.RawMessage
		_ = json.Unmarshal(rawList, &list)
		for i, item := range list {
			conn, _ := object(item)
			_, address, _ := field(conn, "URL")
			_, secret, _ := field(conn, "AccessToken")
			var endpoint, token string
			_ = json.Unmarshal(address, &endpoint)
			_ = json.Unmarshal(secret, &token)
			redacted := false
			u, err := url.Parse(endpoint)
			if err != nil {
				endpoint, redacted = "（地址格式无法展示）", true
			} else if u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
				u.User, u.RawQuery, u.Fragment, u.ForceQuery = nil, "", "", false
				endpoint, redacted = u.String(), true
			}
			connections = append(connections, ConnectionView{kind, i, endpoint, token != "", redacted})
		}
	}
	_, z, _ := field(root, "zero")
	zero, _ := object(z)
	_, users, _ := field(zero, "super_users")
	var ids []int64
	_ = json.Unmarshal(users, &ids)
	result := []string{}
	for _, id := range ids {
		result = append(result, strconv.FormatInt(id, 10))
	}
	return connections, result
}

func equivalentConfig(a, b []byte) bool {
	var av, bv any
	ad, bd := json.NewDecoder(bytes.NewReader(a)), json.NewDecoder(bytes.NewReader(b))
	ad.UseNumber()
	bd.UseNumber()
	_ = ad.Decode(&av)
	_ = bd.Decode(&bv)
	x, _ := json.Marshal(av)
	y, _ := json.Marshal(bv)
	return bytes.Equal(x, y)
}

func (s *ConnectionStore) view(raw []byte) ConnectionSettings {
	c, users := configView(raw)
	running, runtimeUsers := configView(s.running)
	return ConnectionSettings{Available: true, Revision: s.revision(raw), Connections: c, SuperUsers: users, RunningConnections: running, RunningSuperUsers: runtimeUsers, RestartRequired: !equivalentConfig(raw, s.running), StartupDifferences: startupDifferences(s.running, raw)}
}

func (s *ConnectionStore) Get() (ConnectionSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := s.read()
	if err != nil {
		return ConnectionSettings{}, err
	}
	return s.view(b), nil
}

func validConnectionURL(address string) bool {
	if len(address) > 2048 || strings.IndexFunc(address, unicode.IsSpace) >= 0 {
		return false
	}
	u, err := url.Parse(address)
	if err != nil || (u.Scheme != "ws" && u.Scheme != "wss") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return false
		}
	}
	return !strings.HasSuffix(u.Host, ":")
}

func applyConnectionUpdate(raw []byte, update ConnectionUpdate) ([]byte, error) {
	root, _ := decodeConfig(raw)
	if update.RemoveConnection != nil {
		if update.Connection != nil || update.SuperUsers != nil || update.AddConnection != nil {
			return nil, errors.New("删除连接须独立保存，不能同时提交其他配置修改")
		}
		return applyConnectionRemoval(root, *update.RemoveConnection)
	}
	if update.AddConnection != nil {
		if update.Connection != nil || update.SuperUsers != nil {
			return nil, errors.New("新增连接须独立保存，请先处理现有连接与主人草稿")
		}
		return applyConnectionAddition(root, *update.AddConnection)
	}
	if update.Connection == nil && update.SuperUsers == nil {
		return nil, errors.New("没有选择要修改的配置")
	}
	if change := update.Connection; change != nil {
		if change.Kind != "ws" && change.Kind != "wss" {
			return nil, errors.New("连接类型无效")
		}
		key, rawList, _ := field(root, change.Kind)
		var list []json.RawMessage
		_ = json.Unmarshal(rawList, &list)
		if change.Index < 0 || change.Index >= len(list) {
			return nil, errors.New("连接不存在，请重新加载")
		}
		conn, _ := object(list[change.Index])
		if change.URL != nil {
			if !validConnectionURL(*change.URL) {
				return nil, errors.New("地址须为有效的 ws/wss 地址；认证信息请填写在 Token 中")
			}
			k, _, _ := field(conn, "URL")
			conn[k], _ = json.Marshal(*change.URL)
		}
		switch change.TokenAction {
		case "keep":
			if change.Token != "" {
				return nil, errors.New("保留 Token 时不能同时提交新值")
			}
		case "replace":
			if change.Token == "" || len(change.Token) > 2048 || strings.IndexFunc(change.Token, unicode.IsControl) >= 0 || strings.Trim(change.Token, "*• ") == "" {
				return nil, errors.New("请输入有效的新 Token，不能使用脱敏占位文本")
			}
			k, _, _ := field(conn, "AccessToken")
			conn[k], _ = json.Marshal(change.Token)
		case "clear":
			if change.Token != "" {
				return nil, errors.New("清除 Token 时不能同时提交新值")
			}
			k, _, _ := field(conn, "AccessToken")
			conn[k] = json.RawMessage(`""`)
		default:
			return nil, errors.New("请选择保留、替换或清除 Token")
		}
		list[change.Index], _ = json.Marshal(conn)
		root[key], _ = json.Marshal(list)
	}
	if update.SuperUsers != nil {
		if len(*update.SuperUsers) == 0 {
			return nil, errors.New("请至少保留一个主人账号")
		}
		if len(*update.SuperUsers) > 32 {
			return nil, errors.New("主人账号最多 32 个")
		}
		ids := []int64{}
		seen := map[int64]bool{}
		for _, text := range *update.SuperUsers {
			id, err := strconv.ParseInt(text, 10, 64)
			if err != nil || id <= 0 || strconv.FormatInt(id, 10) != text {
				return nil, errors.New("主人账号须为不重复的正整数")
			}
			if seen[id] {
				return nil, errors.New("主人账号不能重复")
			}
			seen[id] = true
			ids = append(ids, id)
		}
		k, z, _ := field(root, "zero")
		zero, _ := object(z)
		userKey, _, _ := field(zero, "super_users")
		zero[userKey], _ = json.Marshal(ids)
		root[k], _ = json.Marshal(zero)
	}
	b, err := json.MarshalIndent(root, "", "  ")
	return append(b, '\n'), err
}

func (s *ConnectionStore) Save(update ConnectionUpdate) (ConnectionSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.read()
	if err != nil {
		return ConnectionSettings{}, err
	}
	if !hmac.Equal([]byte(update.Revision), []byte(s.revision(old))) {
		return ConnectionSettings{}, errConfigConflict
	}
	next, err := applyConnectionUpdate(old, update)
	if err != nil {
		return ConnectionSettings{}, fmt.Errorf("%w: %s", errConfigValidation, err)
	}
	if equivalentConfig(old, next) {
		return s.view(old), nil
	}
	if len(next) > 1<<20 {
		return ConnectionSettings{}, fmt.Errorf("%w: 修改后配置超过 1 MiB，原配置保留", errConfigValidation)
	}
	if err = os.MkdirAll(s.backups, 0700); err != nil {
		return ConnectionSettings{}, errors.New("备份目录不可写，配置未保存")
	}
	info, err := os.Lstat(s.backups)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ConnectionSettings{}, errors.New("备份目录无效，配置未保存")
	}
	if err = protectPath(s.backups, true); err != nil {
		return ConnectionSettings{}, errors.New("备份权限设置失败，配置未保存")
	}
	id, err := writeFileBackup(s.backups, "startup-config", old)
	if err != nil {
		return ConnectionSettings{}, errors.New("写前备份失败，配置未保存")
	}
	// Catch an external edit made while the backup was being created.
	current, err := s.read()
	if err != nil || !bytes.Equal(old, current) {
		return ConnectionSettings{}, errConfigConflict
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".admin-config-*.tmp")
	if err != nil {
		return ConnectionSettings{}, errors.New("配置目录不可写，原配置保留")
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = protectPath(tmp, false); err == nil {
		_, err = f.Write(next)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		current, checkErr := s.read()
		if checkErr != nil || !bytes.Equal(old, current) {
			return ConnectionSettings{}, errConfigConflict
		}
		err = s.replace(tmp, s.path)
	}
	if err != nil {
		return ConnectionSettings{}, errors.New("配置写入失败，原配置保留；写前备份已保存")
	}
	result := s.view(next)
	result.BackupID = id
	return result, nil
}

func (s *Server) getConnections(w http.ResponseWriter, _ *http.Request) {
	if s.cfg.Connections == nil {
		jsonReply(w, 200, ConnectionSettings{Available: false})
		return
	}
	value, err := s.cfg.Connections.Get()
	if err != nil {
		apiError(w, 500, err.Error())
		return
	}
	jsonReply(w, 200, value)
}

func (s *Server) saveConnections(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Connections == nil {
		apiError(w, 409, "请使用 -c 启动配置文件后再管理连接")
		return
	}
	var update ConnectionUpdate
	if !readJSONLimit(w, r, &update, 16384) {
		return
	}
	if s.cfg.Logs != nil && update.Connection != nil {
		s.cfg.Logs.AddSecrets(update.Connection.Token)
	}
	if s.cfg.Logs != nil && update.AddConnection != nil {
		s.cfg.Logs.AddSecrets(update.AddConnection.Token)
	}
	if add := update.AddConnection; add != nil && add.Kind == "wss" {
		port, ok := reversePort(add.URL)
		adminPort, adminOK := reversePort("ws://" + s.cfg.Addr)
		if ok && adminOK && port == adminPort {
			apiError(w, 400, "反向监听端口与当前后台端口冲突，请选择其他端口")
			return
		}
	}
	value, err := s.cfg.Connections.Save(update)
	if err != nil {
		status := 500
		if errors.Is(err, errConfigValidation) {
			status = 400
		}
		if errors.Is(err, errConfigConflict) {
			status = 409
		}
		apiError(w, status, err.Error())
		return
	}
	jsonReply(w, 200, value)
}
