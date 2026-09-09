//go:build windows

package aifalse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	wmiGPUInfo = func() []gpuInfo {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// PowerShell: Get-CimInstance Win32_VideoController 返回名称/显存/驱动
		// 过滤掉虚拟显卡（远程桌面、虚拟机模拟器、虚拟显示适配器等）
		filter := `'virtual|indirect|basic|spacedesk|mumu|mirabox|gameviewer|idd|remote|display adapter|Microsoft Basic'`
		psCmd := `Get-CimInstance Win32_VideoController | ` +
			fmt.Sprintf(`Where-Object { $_.Name -notmatch %s } | `, filter) +
			`Select-Object Name, AdapterRAM, DriverVersion | ` +
			`ConvertTo-Json -Compress`
		cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psCmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Warnf("[aifalse] PowerShell Get-CimInstance 失败: %v, 输出: %s", err, string(out))
			return pnpFallbackGPU()
		}

		return parsePowerShellGPU(out)
	}
}

// parsePowerShellGPU 解析 PowerShell ConvertTo-Json 的输出。
// PowerShell 在只有 1 个结果时返回单个 JSON 对象，多个时返回数组，需要统一处理。
func parsePowerShellGPU(out []byte) []gpuInfo {
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "null" {
		return nil
	}

	var entries []struct {
		Name          string      `json:"Name"`
		AdapterRAM    interface{} `json:"AdapterRAM"`
		DriverVersion string      `json:"DriverVersion"`
	}

	// 先尝试解析为数组
	if err := json.Unmarshal(out, &entries); err != nil {
		// 可能是单个对象，再试一次
		var single struct {
			Name          string      `json:"Name"`
			AdapterRAM    interface{} `json:"AdapterRAM"`
			DriverVersion string      `json:"DriverVersion"`
		}
		if err2 := json.Unmarshal(out, &single); err2 != nil {
			logrus.Warnf("[aifalse] PowerShell GPU JSON 解析失败: %v / %v, raw=%s", err, err2, trimmed)
			return nil
		}
		if single.Name == "" {
			return nil
		}
		entries = []struct {
			Name          string      `json:"Name"`
			AdapterRAM    interface{} `json:"AdapterRAM"`
			DriverVersion string      `json:"DriverVersion"`
		}{single}
	}

	result := make([]gpuInfo, 0, len(entries))
	for _, e := range entries {
		if e.Name == "" {
			continue
		}
		vendor := detectVendor(e.Name)
		memMiB := parseRAM(e.AdapterRAM)
		result = append(result, gpuInfo{
			Name:     e.Name,
			Vendor:   vendor,
			MemTotal: memMiB,
		})
	}
	return result
}

// pnpFallbackGPU 在 Get-CimInstance 也失败时，用 Get-PnpDevice -Class Display 兜底。
// PnP 设备没有显存信息，但至少能列出显卡名称。
func pnpFallbackGPU() []gpuInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := `'virtual|indirect|basic|spacedesk|mumu|mirabox|gameviewer|idd|remote|display adapter|Microsoft Basic'`
	psCmd := `Get-PnpDevice -Class Display | ` +
		fmt.Sprintf(`Where-Object { $_.FriendlyName -and $_.Status -eq 'OK' -and $_.FriendlyName -notmatch %s } | `, filter) +
		`Select-Object FriendlyName | ` +
		`ConvertTo-Json -Compress`
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Warnf("[aifalse] PowerShell Get-PnpDevice 也失败: %v", err)
		return nil
	}

	var items []struct {
		Name string `json:"FriendlyName"`
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "null" {
		return nil
	}
	if err := json.Unmarshal(out, &items); err != nil {
		var single struct {
			Name string `json:"FriendlyName"`
		}
		if err2 := json.Unmarshal(out, &single); err2 != nil {
			logrus.Warnf("[aifalse] PnP JSON 解析失败: %v / %v", err, err2)
			return nil
		}
		if single.Name == "" {
			return nil
		}
		items = []struct {
			Name string `json:"FriendlyName"`
		}{single}
	}

	result := make([]gpuInfo, 0, len(items))
	for _, item := range items {
		result = append(result, gpuInfo{
			Name:   item.Name,
			Vendor: detectVendor(item.Name),
		})
	}
	return result
}

// parseRAM 解析 PowerShell 返回的 AdapterRAM。
// CimInstance 的 AdapterRAM 是 uint64，但 JSON 序列化后可能是 number 或 null。
func parseRAM(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val / (1024 * 1024)
	case int64:
		return float64(val) / (1024 * 1024)
	case int:
		return float64(val) / (1024 * 1024)
	case nil:
		return 0
	}
	return 0
}

// hwiSensor 对应 HWiNFO64 注册表 HKLM\SOFTWARE\HWiNFO64\VSB 里的传感器条目。
// HWiNFO 在后台运行时（HWInfoSever 进程）会持续把传感器数据写入注册表，
// 这是最简单最稳定的读取方式 —— 不需要额外装 LibreHardwareMonitor。
type hwiSensor struct {
	Name  string  `json:"Name"`  // 如 "CPU Package"
	Label string  `json:"Label"` // 分组标签，如 "CPU [#0]: Intel Core i7-4170"
	Value float64 `json:"Value"` // 裸数值
	Unit  string  `json:"Unit"`  // "°C" / "RPM" / "W"
}

// queryHWiNFO 从 HWiNFO64 注册表读取温度/风扇/功耗传感器数据。
// 返回的三个切片按类型分好类：温度 / 风扇 / 功耗。
// HWiNFO 没装或没在运行时返回 nil, nil, nil（正常情况，不报错）。
func queryHWiNFO() (temps, fans, powers []*status) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 尝试三个可能的注册表路径：
	// 1. HKLM\SOFTWARE\HWiNFO64\VSB      （64 位 HWiNFO 在 64 位系统）
	// 2. HKLM\SOFTWARE\WOW6432Node\HWiNFO64\VSB  （32 位 HWiNFO 在 64 位系统）
	// 3. HKLM\SOFTWARE\HWiNFO32\VSB      （32 位 HWiNFO 在 32 位系统）
	psCmd := `$keys = @('HKLM:\SOFTWARE\HWiNFO64\VSB','HKLM:\SOFTWARE\WOW6432Node\HWiNFO64\VSB','HKLM:\SOFTWARE\HWiNFO32\VSB'); ` +
		`foreach($k in $keys){ if(Test-Path $k){ $p = Get-ItemProperty $k; ` +
		`$sensors = $p.PSObject.Properties.Name | Where-Object { $_ -match '^Sensor\d+$' }; ` +
		`$result = @(); foreach($sn in $sensors){ $idx = $sn -replace '^Sensor',''; ` +
		`$v = $p."Value$idx"; if($v -match '°C|RPM|W'){ ` +
		`if($v -match '([\d.]+)\s*(°C|RPM|W)'){ $result += [PSCustomObject]@{ ` +
		`Name=$p."Sensor$idx"; Label=$p."Label$idx"; Value=[double]$Matches[1]; Unit=$Matches[2] } } } }; ` +
		`if($result.Count -gt 0){ $result | ConvertTo-Json -Compress; break } } }`

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil, nil
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "null" {
		return nil, nil, nil
	}

	var sensors []hwiSensor
	if err := json.Unmarshal(out, &sensors); err != nil {
		var single hwiSensor
		if err2 := json.Unmarshal(out, &single); err2 != nil {
			logrus.Debugf("[aifalse] HWiNFO 注册表 JSON 解析失败: %v / %v", err, err2)
			return nil, nil, nil
		}
		sensors = []hwiSensor{single}
	}

	for _, s := range sensors {
		if s.Name == "" || s.Value <= 0 {
			continue
		}
		name := s.Name
		if s.Label != "" {
			// Label 格式如 "CPU [#0]: Intel Core i7-4170"，去掉 CPU 名称冗余
			name = s.Label + " · " + s.Name
		}
		switch s.Unit {
		case "°C":
			temps = append(temps, &status{
				name:    name,
				text:    []string{fmt.Sprintf("%.1f°C", s.Value)},
				precent: 0,
			})
		case "RPM":
			fans = append(fans, &status{
				name:    name,
				text:    []string{fmt.Sprintf("%.0f RPM", s.Value)},
				precent: 0,
			})
		case "W":
			powers = append(powers, &status{
				name:    name + " 功耗",
				text:    []string{fmt.Sprintf("%.1f W", s.Value)},
				precent: 0,
			})
		}
	}
	if len(temps)+len(fans)+len(powers) > 0 {
		logrus.Infof("[aifalse] HWiNFO 注册表: 温度 %d, 风扇 %d, 功耗 %d",
			len(temps), len(fans), len(powers))
	}
	return temps, fans, powers
}

// lhmSensor 对应 LibreHardwareMonitor WMI 中 root/LibreHardwareMonitor/Sensor 的字段。
// LHM 是 HWiNFO 的开源免费替代方案，如果你没装 HWiNFO 也可以装 LHM。
type lhmSensor struct {
	Name       string  `json:"Name"`
	SensorType string  `json:"SensorType"` // Temperature / Fan / Power / Load / Clock / Voltage / Control
	Parent     string  `json:"Parent"`     // 所属硬件（如 "Supermicro X11" / "Intel Core i7-8700K"）
	Value      float64 `json:"Value"`
}

// queryLibreHardwareMonitor 通过 WMI 查询 LibreHardwareMonitor 暴露的传感器数据。
// 如果服务器上没装 LHM 或没在运行，返回空切片（不报错）。
// LHM 需要后台运行（可以关闭 UI 但不能退出进程），WMI namespace 才会注册。
func queryLibreHardwareMonitor() []lhmSensor {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Get-CimInstance -ClassName Sensor -Namespace root/LibreHardwareMonitor
	// 只取我们关心的类型，减少数据量
	psCmd := `Get-CimInstance -ClassName Sensor -Namespace root/LibreHardwareMonitor -ErrorAction SilentlyContinue | ` +
		`Where-Object { $_.SensorType -in 'Temperature','Fan','Power' -and $_.Value -ne $null } | ` +
		`Select-Object Name, SensorType, Parent, Value | ` +
		`ConvertTo-Json -Compress`
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// LHM 没装或没运行 —— 正常情况，不打警告
		return nil
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "null" {
		return nil
	}

	var sensors []lhmSensor
	if err := json.Unmarshal(out, &sensors); err != nil {
		// 可能是单个对象
		var single lhmSensor
		if err2 := json.Unmarshal(out, &single); err2 != nil {
			logrus.Debugf("[aifalse] LHM WMI JSON 解析失败: %v / %v", err, err2)
			return nil
		}
		sensors = []lhmSensor{single}
	}

	// 过滤无效值
	var result []lhmSensor
	for _, s := range sensors {
		if s.Name == "" || s.Value <= 0 {
			continue
		}
		result = append(result, s)
	}
	return result
}

// lhmHTTPNode 是 LibreHardwareMonitor HTTP REST API /data.json 的节点结构。
// 整个 JSON 是一棵递归树，传感器节点的 Value 字段是带单位的字符串，需要解析。
// 例: {"Text":"CPU Package","Type":"Temperature","Value":"53.0 °C","Children":[]}
type lhmHTTPNode struct {
	ID       int           `json:"id"`
	Text     string        `json:"Text"`
	Type     string        `json:"Type"`     // Temperature / Fan / Power / Load / Clock / Voltage / Data / Control / SmallData / Factor / Level / Throughput
	SensorID string        `json:"SensorId"` // 只有传感器节点才有
	Value    string        `json:"Value"`    // 字符串，带单位："53.0 °C" / "2700 RPM" / "23.4 W"
	RawValue string        `json:"RawValue"` // 字符串，数值+单位
	Children []lhmHTTPNode `json:"Children"`
}

// queryLibreHardwareMonitorHTTP 通过 LHM 的 HTTP REST API (默认 http://127.0.0.1:8085/data.json) 获取传感器数据。
// 与 WMI 查询并行执行，谁有数据用谁。
func queryLibreHardwareMonitorHTTP() []lhmSensor {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:8085/data.json", nil)
	if err != nil {
		return nil
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil // HTTP server 没开 —— 正常情况
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var root lhmHTTPNode
	if err := json.Unmarshal(body, &root); err != nil {
		logrus.Debugf("[aifalse] LHM HTTP JSON 解析失败: %v", err)
		return nil
	}

	// 递归收集传感器
	var result []lhmSensor
	var walk func(nodes []lhmHTTPNode, parent string)
	walk = func(nodes []lhmHTTPNode, parent string) {
		for _, n := range nodes {
			// 传感器节点的判断：有 SensorId 且有 Type
			if n.SensorID != "" && n.Type != "" {
				// 只收集我们关心的类型
				if n.Type != "Temperature" && n.Type != "Fan" && n.Type != "Power" {
					continue
				}
				// 解析 Value 字符串 "53.0 °C" / "2700 RPM" / "23.4 W" -> float64
				val := parseLHMHTTPValue(n.Value)
				if val <= 0 {
					continue // 过滤无效值
				}
				result = append(result, lhmSensor{
					Name:       n.Text,
					SensorType: n.Type,
					Parent:     parent,
					Value:      val,
				})
			} else if n.Text != "" {
				// 硬件节点（分组节点），更新 parent 继续递归
				walk(n.Children, n.Text)
				continue
			}
			walk(n.Children, parent)
		}
	}
	walk(root.Children, "")

	if len(result) == 0 {
		return nil
	}
	logrus.Infof("[aifalse] LHM HTTP REST 返回 %d 个传感器", len(result))
	return result
}

// parseLHMHTTPValue 从 LHM HTTP /data.json 的 Value 字符串中提取数字。
// 格式: "53.0 °C" / "2700 RPM" / "23.4 W" / "NaN %" 等
func parseLHMHTTPValue(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || strings.Contains(s, "NaN") || strings.Contains(s, "-") {
		return 0
	}
	// 去掉单位部分：取空格前的数字
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return 0
	}
	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	return val
}

// mergeLHMSensors 合并 WMI 和 HTTP REST 两个数据源的结果。
// 用 Name+SensorType+Parent 做 key 去重，保留第一个非空值。
func mergeLHMSensors(wmi, http []lhmSensor) []lhmSensor {
	if len(wmi) == 0 && len(http) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var merged []lhmSensor
	add := func(s lhmSensor) {
		k := s.Name + "|" + s.SensorType + "|" + s.Parent
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		merged = append(merged, s)
	}
	for _, s := range wmi {
		add(s)
	}
	for _, s := range http {
		add(s)
	}
	return merged
}

// lhmSensorDisplayName 从 LHM 传感器对象生成对用户友好的显示名。
// LHM 的 Name 类似 "CPU Package" / "CPU Core #1" / "Fan #1" / "CPU Package" (Power)
// Parent 类似 "Intel(R) Core(TM) i7-8700K" 或主板名
func lhmSensorDisplayName(s lhmSensor) string {
	name := strings.TrimSpace(s.Name)
	// 对于 Power 类型，标注清楚
	if s.SensorType == "Power" {
		return name + " 功耗"
	}
	return name
}
