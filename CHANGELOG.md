# ZeroBot-Plugin v1.10.25 更新日志

**发布日期：** 2026-09-09

---

## 🎯 本次更新重点

- **服务菜单插件**（新插件）：液态玻璃 UI 服务列表，两列大卡片布局，不放大图片即可看清
- Windows 硬件监控 **零配置**：首次运行自动下载 LibreHardwareMonitor
- 自检图 CPU 实时采样（1秒间隔，不再是旧值）
- GPU 检测全覆盖（NVIDIA / Intel / AMD）
- 温度卡片精简（只保留 CPU Core，去掉冗余项）

---

## 🍮 服务菜单插件（新增）

全新插件 `plugin/servicemenu`，iOS 液态玻璃风格 UI：

- **服务列表**：全部插件两列大卡片（~514px 宽）展示，插件名 28px + 简介 20px 双层描边白字，右侧绿色「开」/ 红色「关」胶囊徽章，**不放大图片即可看清**
- **长简介省略**：按实测像素宽度二分截断（rune 安全），占满可用宽度才省略
- **主题切换**：暗夜玻璃 / 柔光玻璃 / 霓虹赛博 / 极简暖灰 4 套液态玻璃主题（`menu <编号>` 或 `主题 <name>`）
- **菜单用法**：`菜单用法 <插件名>` 渲染单插件用法卡
- **重载**：`重载全部` / `重载 <插件名>`（超级用户）优雅重启进程
- **随机背景**：复用 `data/aifalse/` 目录图片，毛玻璃与背景实时联动
- 命令无需 @机器人，直接触发

**涉及文件：** `plugin/servicemenu/main.go`、`plugin/servicemenu/render_test.go`

### 渲染管线亮点

- SDF 距离场 + 数值梯度法线实现边缘折射与轮廓光（垂直边缘、随卡片尺寸缩放）
- 双线性插值位移采样，杜绝最近邻采样的彩色颗粒磨损感
- 预乘 alpha 输出，无暗边光晕；高斯模糊落影式阴影（内缩下移），无角落暗环

---

## 🆕 新功能

### 1. Windows 自动配置 LibreHardwareMonitor（LHM）

之前需要手动安装 LHM/HWiNFO 才能拿到真实的 CPU 温度/风扇/功耗，现在：

- **第一次发自检**时 bot 会：
  1. 检测 `http://127.0.0.1:8085/data.json` 是否可访问
  2. 没装 LHM → 自动从 GitHub 下载 portable zip（1.8 MB）
  3. 解压到 `data/aifalse/librehardwaremonitor/`
  4. 自动写入 config 开启 HTTP server
  5. 后台静默启动
- **用户零感知**：不需要手动装任何东西
- **降级保护**：没网/没权限时优雅降级，只显示能拿到的数据

**涉及文件：** `plugin/aifalse/lhm_autosetup.go`、`plugin/aifalse/lhm_autosetup_other.go`

### 2. LHM HTTP REST API 备选采集

LHM v0.9.4+ 的 .NET 版本不再默认推 WMI namespace，新增 HTTP REST API 作为备选：

- WMI (`root/LibreHardwareMonitor/Sensor`) 和 HTTP REST (`http://127.0.0.1:8085/data.json`) **并行跑**
- 谁有数据用谁，两个都有就自动去重合并
- 超时保护：单路 3s 超时，不阻塞主流程

**涉及文件：** `plugin/aifalse/hardware_windows.go`（`queryLibreHardwareMonitorHTTP`、`mergeLHMSensors`）

### 3. 风扇转速 + CPU 功耗采集

LHM/HWiNFO 提供了真实传感器数据时：

- 自动显示风扇转速卡片（如 `Fan #2: 2700 RPM`）
- 自动显示 CPU 功耗卡片（如 `CPU Package: 23.4 W`）
- **无数据自动隐藏**：不装 LHM/HWiNFO 时不渲染这两个卡片

**涉及文件：** `plugin/aifalse/main.go`（`tempstate()`）

---

## ⚡ 性能优化

### 4. CPU 使用率后台采样

之前 CPU 使用率只取一次快照，可能显示旧值或 0%。现在：

- bot 启动时后台 goroutine 每 **1 秒** 采样一次 CPU
- 自检时直接读采样好的值，**零阻塞**

### 5. 并行化硬件采集

自检图数据采集全部改为并行（3s 总超时）：

```
并行启动：
├─ CPU / RAM / SWAP（gopsutil）
├─ 磁盘（WMI / gopsutil）
├─ GPU（PowerShell / nvidia-smi）
├─ 温度/风扇/功耗（3 路并行）
│   ├─ HWiNFO 注册表
│   ├─ LHM WMI
│   ├─ LHM HTTP REST
│   └─ gopsutil ACPI 热区（兜底）
```

### 6. 网络请求超时保护

- nvidia-smi / PowerShell GPU 查询加 3s 超时
- 温度采集 3s 总超时
- HTTP 请求全部带 context timeout

### 7. 跳过原始文件加速启动

`file.SkipOriginal = true`，不加载原始文件内容，减少启动时的内存和 CPU 开销。

---

## 🎨 自检图改进

### 8. GPU 卡片样式统一

之前 GPU 固定用横向进度条，Intel/AMD 核显没利用率时显示假的 0% 绿色条。现在：

- 所有 GPU 统一用 **info 卡样式**（左名称、右详情）
- NVIDIA：`RTX 4090 | 15% · 61/4096 MiB · 45°C · 12.3 W`
- Intel/AMD：`Intel HD Graphics 4400 | 显存 1024 MiB`

### 9. 温度卡片精简

只保留 **CPU Core 温度**，去掉冗余项：

| 之前显示（5 行） | 现在（2 行，2C4T CPU） |
|---|---|
| CPU Core #1: 54.0°C | CPU Core #1: 54.0°C ✅ |
| CPU Core #2: 54.0°C | CPU Core #2: 54.0°C ✅ |
| Core Average: 54.0°C | ❌ 去掉 |
| Core Max: 54.0°C | ❌ 去掉 |
| CPU Package: 53.0°C | ❌ 去掉 |

**降级保护：** 没有 Core 温度时退而保留 Package 温度，不会空卡片。

### 10. GPU 检测全覆盖

PowerShell `Get-CimInstance Win32_VideoController` 替代之前的简化检测：

| GPU 品牌 | 之前 | 现在 |
|---|---|---|
| NVIDIA | ✅ nvidia-smi | ✅ nvidia-smi + PowerShell 兜底 |
| Intel 核显 | ❌ 检测不到 | ✅ PowerShell 直接读显存 |
| AMD | ❌ 检测不到 | ✅ PowerShell 直接读显存 |
| 虚拟显卡（远程桌面等） | ❌ 误显示 | ✅ 自动过滤 |

### 11. 本地背景图兜底

- 自检图背景可放在 `data/aifalse/` 目录下
- 找不到背景图时用纯色兜底，不显示空白
- 之前 URL 图片加载失败会导致整个自检图渲染失败

### 12. CommandRule 修复

`CommandRule` 的 `AtBot` 模式在 @机器人 + 命令时无法匹配到正确规则，已修复。

### 13. 超级用户从文件读取

- 从 `data/superusers.txt` 读取超级用户列表
- 每行一个 QQ 号，方便多用户部署时管理

---

## 🔧 跨平台行为

| 平台 | CPU 温度 | 风扇 | 功耗 | 策略 |
|---|---|---|---|---|
| **Linux** | ✅ gopsutil 原生读 `/sys/class/hwmon` | ✅ 同上 | ✅ 同上 | 零依赖 |
| **macOS** | ✅ gopsutil 原生 | ❌ | ❌ | 零依赖 |
| **Windows 10/11/Server** | ✅ **自动下载 LHM** | ✅ 自动 | ✅ 自动 | 首次自动配置 |
| **Windows Server（没装 LHM）** | ⚠️ 降级到 ACPI 热区 | ❌ | ❌ | 优雅降级 |

---

## 📁 新增/修改文件

| 文件 | 状态 | 说明 |
|---|---|---|
| `plugin/servicemenu/` | **新增** | 服务菜单插件（液态玻璃 UI、两列布局、主题切换） |
| `plugin/aifalse/main.go` | 修改 | CPU 后台采样、并行采集、温度精简、CommandRule 修复 |
| `plugin/aifalse/hardware_windows.go` | 修改 | LHM HTTP REST API 解析、merge 函数、`parseLHMHTTPValue` |
| `plugin/aifalse/hardware_other.go` | 修改 | 新增 HTTP/merge 非 Windows 占位 |
| `plugin/aifalse/lhm_autosetup.go` | **新增** | Windows 自动下载 + 配置 + 启动 LHM |
| `plugin/aifalse/lhm_autosetup_other.go` | **新增** | Linux/macOS 空占位 |
| `plugin/aichat/main.go` | 修改 | 聊天匹配优先级降为 10，不再拦截管理命令 |
| `plugin/thesaurus/chat.go` | 修改 | 词库回复同上 |
| `plugin/score/sign_in.go` | 修改 | 签到背景图源更换 |
| `main.go` | 修改 | 注册 servicemenu、superusers.txt 读取、@机器人命令匹配修复 |
| `kanban/banner/banner.go` | 修改 | 版本号 v1.10.24 |
| `go.mod` | 修改 | gg 依赖指向带 GPU kernel 的 fork 版本 |
| `.github/workflows/release.yml` | 修改 | tag 推送触发 Release 构建 |

---

## 📦 打包说明

Release 由 GitHub Actions（GoReleaser）按 tag 自动构建，提供各平台开箱即用的二进制：

```
zbp_windows_amd64.zip      ← Windows 双击运行
zbp_linux_amd64.tar.gz     ← Linux 服务器
zbp_linux_arm64.tar.gz     ← ARM 设备（N1 / 树莓派等）
zbp_linux_386 / arm / armv6 / armv7
zbp_checksums.txt          ← 校验和
zbp_linux_amd64.deb / .rpm ← 系统包管理器安装
```

用户拿到后：解压 → 配置 `config.json`（或首次运行 `-c` 生成）→ 启动 → 发自检

### 源码编译

```powershell
go mod tidy
go build -o ZeroBot-Plugin.exe .   # Windows
go build -o ZeroBot-Plugin .       # Linux/macOS
```

---

## 🐛 已知限制

1. **Windows 需要管理员权限**（LHM 需要管理员权限才能读取 MSR 寄存器获取真实传感器数据）
2. **Linux 需要 `lm-sensors`**（某些发行版）才能读到风扇/功耗
3. **Windows Server 的 ACPI 热区**（没装 LHM 时）值不太准确，建议安装 LHM
