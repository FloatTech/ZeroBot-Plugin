package control

// EnableMark 启用：●，禁用：○
type EnableMark bool

// String 打印启用状态
func (em EnableMark) String() string {
	if bool(em) {
		return "●"
	}
	return "○"
}

// EnableMarkIn 打印 ● 或 ○
func (m *Control[CTX]) EnableMarkIn(grp int64) EnableMark {
	return EnableMark(m.IsEnabledIn(grp))
}
