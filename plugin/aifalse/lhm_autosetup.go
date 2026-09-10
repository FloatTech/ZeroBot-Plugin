//go:build windows

package aifalse

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// lhmAutoSetup 负责在 Windows 上自动配置 LibreHardwareMonitor 后台进程。
// 首次运行时自动下载 LHM portable zip、写 config、后台静默启动。
// 后续启动检测进程是否存在，不存在则重启。
//
// 设计原则：
//   - 完全自动，用户零感知（不需要手动装任何东西）
//   - 网络不通或权限不足时优雅降级，bot 仍能跑（只是没温度/风扇/功耗数据）
//   - 下载失败 3 次后放弃，下次再试
func lhmAutoSetup() {
	lhmDir := filepath.Join("data", "aifalse", "librehardwaremonitor")
	exePath := filepath.Join(lhmDir, "LibreHardwareMonitor.exe")
	configPath := filepath.Join(lhmDir, "LibreHardwareMonitor.config")

	// 1. 先检查 HTTP 服务器是否已经可达
	if probeLHMHTTP() {
		logrus.Debugln("[aifalse] LHM HTTP 已在运行，跳过自动配置")
		return
	}

	// 2. 检查本地是否已有 LHM
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		logrus.Infoln("[aifalse] 未找到 LHM，开始自动下载配置...")
		if err := downloadAndSetupLHM(lhmDir, configPath); err != nil {
			logrus.Warnf("[aifalse] LHM 自动配置失败（不影响使用）: %v", err)
			return
		}
	}

	// 3. 确保 config 开启了 HTTP server
	if err := ensureLHMConfig(configPath); err != nil {
		logrus.Warnf("[aifalse] LHM config 写入失败: %v", err)
	}

	// 4. 后台静默启动 LHM
	if err := startLHMBackground(exePath); err != nil {
		logrus.Warnf("[aifalse] LHM 启动失败（可能需要管理员权限）: %v", err)
		return
	}

	// 5. 等一下让 LHM 初始化 HTTP server
	for i := 0; i < 20; i++ {
		time.Sleep(500 * time.Millisecond)
		if probeLHMHTTP() {
			logrus.Infof("[aifalse] LHM 自动配置成功！HTTP server 已就绪")
			return
		}
	}
	logrus.Warnln("[aifalse] LHM 启动后 HTTP server 未就绪，可能需要管理员权限运行 bot")
}

// probeLHMHTTP 检测 http://127.0.0.1:8085/data.json 是否可达。
func probeLHMHTTP() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:8085/data.json", nil)
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// downloadAndSetupLHM 从 GitHub release 下载 LHM portable zip 并解压到 lhmDir。
func downloadAndSetupLHM(lhmDir, configPath string) error {
	if err := os.MkdirAll(lhmDir, 0o755); err != nil {
		return errors.Wrap(err, "创建目录失败")
	}

	// 先试 winget（如果有的话）
	if tryWingetInstall() {
		logrus.Infoln("[aifalse] winget 安装 LHM 成功")
		return copyLHMFromWinget(lhmDir, configPath)
	}

	// winget 不行，直接下载 zip
	urls := []string{
		// .NET Framework 4.7.2 版本（兼容性好，Windows 10+ 默认有）
		"https://github.com/LibreHardwareMonitor/LibreHardwareMonitor/releases/latest/download/LibreHardwareMonitor.zip",
	}

	var lastErr error
	for _, url := range urls {
		logrus.Infof("[aifalse] 下载 LHM: %s", url)
		tmpZip := filepath.Join(lhmDir, "_lhm_download.zip")
		if err := downloadFile(url, tmpZip); err != nil {
			lastErr = err
			logrus.Debugf("[aifalse] 下载失败: %v", err)
			continue
		}
		if err := extractZip(tmpZip, lhmDir); err != nil {
			lastErr = err
			os.Remove(tmpZip)
			continue
		}
		os.Remove(tmpZip)
		// 解压后可能多了一层目录（librehardwaremonitor/xxx/LibreHardwareMonitor.exe）
		if exe := findLHMExe(lhmDir); exe != "" {
			// 如果 exe 在子目录，把文件移到 lhmDir 根目录
			if exe != filepath.Join(lhmDir, "LibreHardwareMonitor.exe") {
				if err := moveFiles(filepath.Dir(exe), lhmDir); err != nil {
					logrus.Debugf("[aifalse] 移动 LHM 文件失败: %v", err)
				}
			}
			return nil
		}
		lastErr = errors.Errorf("解压后没找到 LibreHardwareMonitor.exe")
	}
	return lastErr
}

// tryWingetInstall 尝试用 winget 安装 LHM（如果系统有 winget 的话）。
func tryWingetInstall() bool {
	_, err := exec.LookPath("winget")
	if err != nil {
		return false // 没装 winget
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 静默安装，接受协议
	cmd := exec.CommandContext(ctx, "winget", "install", "--id=LibreHardwareMonitor.LibreHardwareMonitor",
		"--silent", "--accept-source-agreements", "--accept-package-agreements")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		logrus.Debugf("[aifalse] winget 安装 LHM 失败: %v", err)
		return false
	}
	return true
}

// copyLHMFromWinget winget 安装后，把 LHM 文件从安装目录复制到 lhmDir。
func copyLHMFromWinget(lhmDir, _ string) error {
	// winget 安装路径通常在 Program Files 或 Users\...\AppData\Local\Microsoft\WinGet\Packages
	// 简化处理：搜索常见位置
	locations := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WinGet", "Packages"),
		filepath.Join(os.Getenv("ProgramFiles"), "LibreHardwareMonitor"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "LibreHardwareMonitor"),
	}
	for _, loc := range locations {
		exe := findLHMExe(loc)
		if exe != "" {
			return copyFiles(filepath.Dir(exe), lhmDir)
		}
	}
	return errors.Errorf("winget 安装了但找不到 exe")
}

// findLHMExe 在 rootDir 递归搜索 LibreHardwareMonitor.exe。
func findLHMExe(rootDir string) string {
	var found string
	// 回调自身吞掉所有错误（读不到的条目直接跳过），Walk 的返回值无意义
	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if !info.IsDir() && strings.EqualFold(info.Name(), "LibreHardwareMonitor.exe") {
			found = path
		}
		return nil
	})
	return found
}

// downloadFile 下载 URL 到 dstPath。
func downloadFile(url, dstPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.Errorf("HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// extractZip 解压 zip 文件到 destDir（扁平化结构）。
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		// 跳过目录
		if f.FileInfo().IsDir() {
			continue
		}
		// 安全：防止 zip slip（不允许解压到 destDir 外面）
		target := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			continue
		}
		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			continue
		}
		// 解压文件
		src, err := f.Open()
		if err != nil {
			continue
		}
		dst, err := os.Create(target)
		if err != nil {
			src.Close()
			continue
		}
		_, _ = io.Copy(dst, src) // 复制失败的条目直接跳过（best-effort 解压）
		src.Close()
		dst.Close()
	}
	return nil
}

// copyFiles 把 srcDir 下的所有文件复制到 dstDir。
func copyFiles(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(srcDir, path)
		target := filepath.Join(dstDir, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// moveFiles 把 srcDir 下的所有文件移到 dstDir。
func moveFiles(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(srcDir, path)
		target := filepath.Join(dstDir, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil
		}
		if err := os.Rename(path, target); err != nil {
			logrus.Debugf("[aifalse] 移动文件失败 %s -> %s: %v", path, target, err)
		}
		return nil
	})
}

// ensureLHMConfig 确保 config 文件里开启了 HTTP server。
func ensureLHMConfig(configPath string) error {
	if _, err := os.Stat(configPath); err != nil {
		// config 不存在，创建一个最小可用的
		config := `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <appSettings>
    <add key="runWebServerMenuItem" value="true" />
    <add key="listenerIp" value="127.0.0.1" />
    <add key="listenerPort" value="8085" />
  </appSettings>
</configuration>`
		return os.WriteFile(configPath, []byte(config), 0o644)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	content := string(data)

	// 已经有 runWebServerMenuItem=true 就不动
	if strings.Contains(content, `key="runWebServerMenuItem" value="true"`) {
		return nil
	}

	// 把已有的 runWebServerMenuItem 改成 true
	if strings.Contains(content, `key="runWebServerMenuItem"`) {
		content = strings.ReplaceAll(content,
			`key="runWebServerMenuItem" value="false"`,
			`key="runWebServerMenuItem" value="true"`)
	} else {
		// 在 </appSettings> 前插入
		insert := `<add key="runWebServerMenuItem" value="true" />
    <add key="listenerIp" value="127.0.0.1" />
    <add key="listenerPort" value="8085" />
`
		content = strings.Replace(content, "</appSettings>", insert+"  </appSettings>", 1)
	}

	// 确保 listenerIp 是 127.0.0.1（避免 "?" 通配符问题）
	content = strings.ReplaceAll(content, `key="listenerIp" value="?"`, `key="listenerIp" value="127.0.0.1"`)

	return os.WriteFile(configPath, []byte(content), 0o644)
}

// startLHMBackground 后台静默启动 LHM。
// 用 START /MIN 让窗口最小化到托盘（LHM 会自动缩到系统托盘）。
func startLHMBackground(exePath string) error {
	// 检查是否已经在运行
	if isProcessRunning("LibreHardwareMonitor") {
		logrus.Debugln("[aifalse] LHM 已在运行")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 用 cmd /c START /MIN 启动，隐藏主窗口
	cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "/min", filepath.Base(exePath))
	cmd.Dir = filepath.Dir(exePath)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "启动失败")
	}
	// 不 Wait——让它自己跑
	logrus.Infof("[aifalse] LHM 后台启动中: %s", exePath)
	return nil
}

// isProcessRunning 检查指定进程是否在运行。
func isProcessRunning(name string) bool {
	_, err := exec.LookPath("tasklist")
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq "+name+".exe", "/FO", "CSV")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), name+".exe")
}
