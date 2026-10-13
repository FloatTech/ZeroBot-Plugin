package localadmin

import (
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// RuntimeOverview reads the current caller registry on every request. It reports
// OneBot transport availability, not an unverified claim about QQ login state.
func RuntimeOverview(started time.Time, configured, pluginCount int) func() Overview {
	version := "开发版本"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && len(setting.Value) >= 8 {
				version = setting.Value[:8]
			}
			if setting.Key == "vcs.modified" && setting.Value == "true" {
				version += " · 本地修改"
			}
		}
	}
	return func() Overview {
		accounts := []string{}
		zero.APICallers.Range(func(id int64, _ zero.APICaller) bool {
			if id > 0 {
				accounts = append(accounts, strconv.FormatInt(id, 10))
			}
			return true
		})
		sort.Strings(accounts)
		state := "waiting"
		if configured == 0 {
			state = "unconfigured"
		}
		if len(accounts) > 0 {
			state = "connected"
		}
		now := time.Now()
		return Overview{StartedAt: started, CheckedAt: now, UptimeSeconds: int64(now.Sub(started).Seconds()), Version: version, GoVersion: runtime.Version(), ConnectionState: state, ConfiguredConnections: configured, BotAccounts: accounts, PluginCount: pluginCount}
	}
}
