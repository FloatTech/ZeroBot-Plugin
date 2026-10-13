package localadmin

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func reversePort(address string) (int, bool) {
	u, err := url.Parse(address)
	var port string
	if err == nil && u.Host != "" {
		port = u.Port()
	} else {
		_, port, err = net.SplitHostPort(address)
		if err != nil {
			return 0, false
		}
	}
	n, err := strconv.Atoi(port)
	return n, err == nil && n > 0 && n <= 65535
}

func validateReverseAddition(root jsonObject, add ConnectionAddition) error {
	u, err := url.Parse(add.URL)
	n, valid := reversePort(add.URL)
	if !validConnectionURL(add.URL) || err != nil || u.Scheme != "ws" || u.Hostname() != "127.0.0.1" || !valid || strconv.Itoa(n) != u.Port() || (u.EscapedPath() != "" && u.EscapedPath() != "/") {
		return errors.New("本机反向监听须使用 ws://127.0.0.1:端口，且只能使用根路径")
	}
	if add.WithoutToken || add.Token == "" {
		return errors.New("本机反向监听必须配置新的 Token")
	}
	for _, kind := range []string{"ws", "wss"} {
		_, raw, _ := field(root, kind)
		var list []json.RawMessage
		_ = json.Unmarshal(raw, &list)
		for _, item := range list {
			conn, _ := object(item)
			_, value, _ := field(conn, "URL")
			var address string
			_ = json.Unmarshal(value, &address)
			if port, ok := reversePort(address); !ok || port != n {
				continue
			}
			if kind == "wss" {
				return errors.New("该端口已用于已保存的反向监听，请编辑已有项或选择其他端口")
			}
			other, err := url.Parse(address)
			if err == nil && (other.Hostname() == "127.0.0.1" || strings.EqualFold(other.Hostname(), "localhost") || other.Hostname() == "::1") {
				return errors.New("该端口已用于本机正向连接目标，请选择其他监听端口")
			}
		}
	}
	return nil
}
