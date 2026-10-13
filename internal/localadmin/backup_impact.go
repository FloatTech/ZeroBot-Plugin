package localadmin

import "encoding/json"

// Fixed labels and aggregate counts only: no config values, unknown keys,
// database names, record identifiers, credential hashes or rows leave the API.
type RestoreImpact struct {
	Label        string `json:"label"`
	Changed      bool   `json:"changed"`
	CurrentCount *int   `json:"currentCount,omitempty"`
	BackupCount  *int   `json:"backupCount,omitempty"`
	Unit         string `json:"unit,omitempty"`
}

func impactComparison(label, unit string, current, backup []byte, currentCount, backupCount int) RestoreImpact {
	v := RestoreImpact{Label: label, Changed: !equivalentConfig(current, backup), Unit: unit}
	if currentCount >= 0 && backupCount >= 0 {
		v.CurrentCount, v.BackupCount = &currentCount, &backupCount
	}
	return v
}

func takeImpactField(obj jsonObject, name string) (json.RawMessage, error) {
	key, raw, err := field(obj, name)
	if err == nil && raw != nil {
		delete(obj, key)
	}
	return raw, err
}

func arrayImpactCount(raw json.RawMessage) int {
	var items []json.RawMessage
	_ = json.Unmarshal(raw, &items) // Inputs have already passed decodeConfig.
	return len(items)
}

func fileRestoreImpacts(scope string, current, backup []byte) ([]RestoreImpact, error) {
	if scope == "admin-options" {
		a, err := object(current)
		if err != nil {
			return nil, errBackupInvalid
		}
		b, err := object(backup)
		if err != nil {
			return nil, errBackupInvalid
		}
		return []RestoreImpact{
			impactComparison("后台启用状态", "", a["enabled"], b["enabled"], -1, -1),
			impactComparison("后台访问地址", "", a["address"], b["address"], -1, -1),
		}, nil
	}
	if scope != "startup-config" {
		return nil, errBackupInvalid
	}
	a, err := decodeConfig(current)
	if err != nil {
		return nil, errBackupInvalid
	}
	b, err := decodeConfig(backup)
	if err != nil {
		return nil, errBackupInvalid
	}
	impacts := make([]RestoreImpact, 0, 6)
	for _, item := range []struct{ key, label string }{{"ws", "正向 OneBot 连接"}, {"wss", "反向 OneBot 连接"}} {
		x, err := takeImpactField(a, item.key)
		if err != nil {
			return nil, errBackupInvalid
		}
		y, err := takeImpactField(b, item.key)
		if err != nil {
			return nil, errBackupInvalid
		}
		impacts = append(impacts, impactComparison(item.label, "条连接", x, y, arrayImpactCount(x), arrayImpactCount(y)))
	}
	x, err := takeImpactField(a, "zero")
	if err != nil {
		return nil, errBackupInvalid
	}
	y, err := takeImpactField(b, "zero")
	if err != nil {
		return nil, errBackupInvalid
	}
	az, err := object(x)
	if err != nil {
		return nil, errBackupInvalid
	}
	bz, err := object(y)
	if err != nil {
		return nil, errBackupInvalid
	}
	x, err = takeImpactField(az, "super_users")
	if err != nil {
		return nil, errBackupInvalid
	}
	y, err = takeImpactField(bz, "super_users")
	if err != nil {
		return nil, errBackupInvalid
	}
	impacts = append(impacts, impactComparison("机器人主人", "个账号", x, y, arrayImpactCount(x), arrayImpactCount(y)))
	x, _ = json.Marshal(az)
	y, _ = json.Marshal(bz)
	impacts = append(impacts, impactComparison("其他机器人参数", "项设置", x, y, len(az), len(bz)))
	x, _ = json.Marshal(a)
	y, _ = json.Marshal(b)
	impacts = append(impacts, impactComparison("其他启动设置", "项设置", x, y, len(a), len(b)))
	changed := false
	for _, item := range impacts {
		changed = changed || item.Changed
	}
	if !changed && !equivalentConfig(current, backup) {
		// Case aliases / absent versus null still follow the existing full-JSON comparison.
		impacts = append(impacts, RestoreImpact{Label: "字段名称或表示差异", Changed: true})
	}
	return impacts, nil
}

func databaseImpactBucket(name string) int {
	switch name {
	case "__resp":
		return 0
	case "__block":
		return 1
	case "__switch_inherit":
		return 2
	default:
		return 3
	}
}

func newDatabaseImpacts() []RestoreImpact {
	labels := []string{"响应规则与AI扩展", "全局用户封禁", "插件开关继承标记", "插件与其他库内数据"}
	items := make([]RestoreImpact, len(labels))
	for i, label := range labels {
		items[i] = RestoreImpact{Label: label, Unit: "条记录", CurrentCount: new(int), BackupCount: new(int)}
	}
	return items
}
