//go:build !windows

package aifalse

// lhmAutoSetup Linux/macOS 不需要自动配置 LHM——gopsutil 直接读 sysfs 拿真实温度。
func lhmAutoSetup() {}
