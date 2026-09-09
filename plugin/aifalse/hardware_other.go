//go:build !windows

package aifalse

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	wmiGPUInfo = func() []gpuInfo {
		// Linux/macOS: 用 lspci 列出 VGA/3D 设备
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		out, err := exec.CommandContext(ctx, "lspci").CombinedOutput()
		if err != nil {
			logrus.Debugf("[aifalse] lspci 失败: %v", err)
			return nil
		}

		vgaRe := regexp.MustCompile(`VGA compatible controller:\s*(.*)`)
		displayRe := regexp.MustCompile(`Display controller:\s*(.*)`)

		var result []gpuInfo
		for _, line := range strings.Split(string(out), "\n") {
			var name string
			if m := vgaRe.FindStringSubmatch(line); len(m) > 1 {
				name = m[1]
			} else if m := displayRe.FindStringSubmatch(line); len(m) > 1 {
				name = m[1]
			}
			if name == "" {
				continue
			}
			name = strings.TrimSpace(name)
			result = append(result, gpuInfo{
				Name:   name,
				Vendor: detectVendor(name),
			})
		}
		if len(result) > 0 {
			logrus.Debugf("[aifalse] lspci 识别到 %d 个 GPU", len(result))
		}
		return result
	}
}

// queryHWiNFO 非 Windows 平台无 HWiNFO。
func queryHWiNFO() (temps, fans, powers []*status) { return nil, nil, nil }

// lhmSensor 非 Windows 平台不采集（LibreHardwareMonitor 仅支持 Windows），
// 但字段需与 hardware_windows.go 保持一致，供跨平台代码（tempstate）编译通过。
type lhmSensor struct {
	Name       string
	SensorType string // Temperature / Fan / Power / Load / Clock / Voltage / Control
	Parent     string
	Value      float64
}

// queryLibreHardwareMonitor 非 Windows 平台永远返回 nil。
func queryLibreHardwareMonitor() []lhmSensor { return nil }

// queryLibreHardwareMonitorHTTP 非 Windows 平台永远返回 nil。
func queryLibreHardwareMonitorHTTP() []lhmSensor { return nil }

// mergeLHMSensors 非 Windows 平台占位。
func mergeLHMSensors(wmi, http []lhmSensor) []lhmSensor { return nil }

// lhmSensorDisplayName 非 Windows 平台占位。
func lhmSensorDisplayName(s lhmSensor) string { return "" }
