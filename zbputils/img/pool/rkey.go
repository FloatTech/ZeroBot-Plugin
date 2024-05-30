package pool

import "time"

const rkeykey = "__latest_rkey__"

var rs rkeystorage

func init() {
	var err error
	rs.item, err = newItem(rkeykey, "")
	if err != nil {
		panic(err)
	}
}

type rkeystorage struct {
	*item
	lastrefresh time.Time
}

func (rs *rkeystorage) rkey(timeout time.Duration) (string, error) {
	if time.Since(rs.lastrefresh) < timeout {
		return rs.u, nil
	}
	err := rs.item.update()
	if err != nil {
		return "", err
	}
	rs.lastrefresh = time.Now()
	return rs.u, nil
}

func (rs *rkeystorage) set(timeout time.Duration, rkey string) error {
	if time.Since(rs.lastrefresh) < timeout { // 降低设置频次
		return nil
	}
	rs.item.u = rkey
	rs.lastrefresh = time.Now()
	return rs.item.push("minamoto")
}
