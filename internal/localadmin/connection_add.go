package localadmin

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

// Existing clients omit Kind for forward drivers. New reverse listeners are
// constrained to local authenticated plain WebSockets supported by this driver.
type ConnectionAddition struct {
	Kind         string `json:"kind,omitempty"`
	URL          string `json:"url"`
	Token        string `json:"token"`
	WithoutToken bool   `json:"withoutToken"`
}

func connectionEndpoint(address string) string {
	u, err := url.Parse(address)
	if err != nil {
		return address
	}
	port := u.Port()
	if port == "" {
		if u.Scheme == "wss" {
			port = "443"
		} else {
			port = "80"
		}
	}
	if n, err := strconv.Atoi(port); err == nil {
		port = strconv.Itoa(n)
	}
	p := u.EscapedPath()
	if p == "/" {
		p = ""
	}
	return u.Scheme + "|" + strings.ToLower(u.Hostname()) + "|" + port + "|" + p
}

func applyConnectionAddition(root jsonObject, add ConnectionAddition) ([]byte, error) {
	kind := add.Kind
	if kind == "" {
		kind = "ws"
	}
	if kind != "ws" && kind != "wss" {
		return nil, errors.New("新增连接类型无效")
	}
	if kind == "wss" {
		if err := validateReverseAddition(root, add); err != nil {
			return nil, err
		}
	}
	if !validConnectionURL(add.URL) {
		return nil, errors.New("地址须为有效的 ws/wss 地址；认证信息请填写在 Token 中")
	}
	if add.WithoutToken {
		if add.Token != "" {
			return nil, errors.New("不使用 Token 时不能同时提交新值")
		}
	} else if add.Token == "" || len(add.Token) > 2048 || strings.IndexFunc(add.Token, unicode.IsControl) >= 0 || strings.Trim(add.Token, "*• ") == "" {
		return nil, errors.New("请输入有效的新 Token，或明确选择此连接不使用 Token")
	}
	key, raw, _ := field(root, kind)
	var list []json.RawMessage
	_ = json.Unmarshal(raw, &list)
	if len(list) >= 64 {
		return nil, errors.New("所选类型已达到64条上限，不能继续新增")
	}
	for _, item := range list {
		conn, _ := object(item)
		_, rawURL, _ := field(conn, "URL")
		var address string
		_ = json.Unmarshal(rawURL, &address)
		if validConnectionURL(address) && connectionEndpoint(address) == connectionEndpoint(add.URL) {
			return nil, errors.New("已存在相同地址的正向连接，请编辑已有连接")
		}
	}
	conn, _ := json.Marshal(map[string]string{"URL": add.URL, "AccessToken": add.Token})
	list = append(list, conn)
	root[key], _ = json.Marshal(list)
	b, err := json.MarshalIndent(root, "", "  ")
	return append(b, '\n'), err
}
