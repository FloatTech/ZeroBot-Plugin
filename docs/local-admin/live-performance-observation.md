# 日常服务只读性能观察

后续DEV-73部署新增[模型接口请求统计](model-request-statistics.md)。以下180样本/15分钟属于原0CF3DE程序，不作为当前6109EA重复采样证明；此观察工具尚不登录后台或读取模型统计接口，真实数量/传输耗时和资源要分别记录，并核对窗口与统计起点。

用于剩余性能待办：在正常使用机器人期间观察当前日常ZeroBot的Windows资源和新追加日志统计，不启动模拟机器人、不发送QQ消息、不调用模型、不登录或轮询后台，也不重启服务。结果只能说明记录窗口内的资源与异常情况，不能证明所有真实业务、回复延迟或长期无泄漏。

入口为`scripts/observe-live-services.ps1`，从本项目托管进程记录核对程序路径、PID和启动时间，并兼容已有C/D目录联接。默认观察15分钟，每5秒采样CPU累计时间、私有内存、工作集、线程与句柄；从开始时的日志文件长度处读取新追加字节，输出换行数及警告/错误统计，不保存正文、账号、消息、地址或口令。进程重启或身份变化时结束本轮，不能把新进程拼接进原窗口。

产物写入本项目`.local/live-observation/`独立目录。短窗口、提前停止、采样间隔异常或身份变化必须明确标记，不能算完整15分钟观察。资源阈值与业务验收分别记录；即使资源正常，没有对应业务证据也不能关闭真实QQ/模型负载待办。停止采样只影响观察工具，不停止ZeroBot或NapCat。

## 使用步骤与结果

在有权查询本机进程并读取本项目服务日志的PowerShell中，从仓库执行：

```powershell
Set-Location <项目目录>
./scripts/observe-live-services.ps1
# 可选：观察半小时，仍每5秒采样。
./scripts/observe-live-services.ps1 -DurationSeconds 1800 -IntervalSeconds 5
```

工具只观察已按本项目启停脚本托管的日常ZeroBot；未运行、记录不匹配、路径或启动时间不可核验时拒绝采样，不自行启动或登录。输出`ARTIFACTS=...`给出本次独立目录。`samples.jsonl`每次追加一个资源样本，`progress.json`每分钟及完成时更新，`result.json`记录最终窗口、阈值、日志覆盖情况和停止原因。

需要提前结束时，在另一PowerShell中对本次目录创建停止请求文件，使用输出中的实际路径：

```powershell
New-Item -ItemType File -Path '<项目目录>
```

下次采样前结束（默认不超过约5秒），最终`stopReason=stop-request`、`completed=false`、`passed=false`；ZeroBot和NapCat继续运行。不要创建到其他观察目录。直接关闭终端或强制结束观察进程可能只留下已写样本和进度文件，缺少最终结果，不能算完成。

时长允许10秒至24小时，间隔允许1～30秒。小于900秒仅用于诊断，`fullTiming=false`；即使按请求采满10秒，也不算正式通过。正式窗口要求实际首末样本跨度至少达到请求时长、样本量达到预计数量的80%并含首样本、最大间隔不超过`2×间隔+0.5秒`；不足样本和异常间隔均如实记录。

默认资源检查为CPU平均占全机逻辑处理器容量不超过20%、私有内存相对首样本的最大增长不超过64MiB；可用`-MaxCpuPercent`与`-MaxPrivateGrowthMiB`调整并在结果中留档。CPU峰值是相邻样本区间平均值的最大值，并非瞬时峰值。没有无后台对照或请求延迟样本，不能据此推断后台开销或回复延迟。

`passed`只表示完整窗口、身份/程序保持、采样节奏、资源阈值及日志覆盖/新错误计数这些观察检查通过。`businessLoadMeasured=false`和`businessAcceptancePassed=false`始终保留，因为工具不统计真实QQ/模型请求数量或延迟。`realQQMessagesSent=false`和`modelCalls=0`仅指观察工具自身没有发起这些动作，不表示日常服务没有自然业务。

日志每次最多读取256KiB，末尾再做有限补读；积压、未换行片段或日志截短会令覆盖不完整。行首logrus的`level=warning/warn/error/fatal/panic`及行首`panic:`用于级别计数，其余行只计“其他”；正文中的`level=error`不会误算错误。开始时已存在的内容不读取，因此原有未完行的后续片段可能无法识别级别；未知格式不推断错误。日志为零或没有对应业务证据时只报告静默/未分类资源观察。

## 边界验证（DEV-72 / REQ-69 / TC-92）

16项合成日志与资源回归通过：历史不计、分块和CRLF、正文伪级别/秘密不输出、长行内存有界、积压/截短、缺文件不创建、全机CPU口径、自定义阈值、最大内存增长和小数精度、短窗口/提前停止/稀疏采样、倒退计数/时钟及样本不足。三份脚本语法通过，观察入口不包含服务启停或网络API命令。既有8个任务/4个进程身份场景及5脚本语法通过。证据`.local/live-observation-tests.txt`、`-service-paths.txt`及`.local/live-observation-tests-734046385ba54956ada2257c4ac649e2/result.json`。

10秒日常诊断目录`run-20261013-084939-45d4c387`正确标记`fullTiming=false`、`passed=false`；提前停止目录`run-20261013-085319-d1a3dd3c`32个样本/32.244秒，停止原因正确且不算完成。非法零时长在采样与目录创建前被拒绝，证据`.local/live-observation-args-result.json`。这些诊断和停止演练不计作完整15分钟观察。

## 2026-10-13 日常观察结果

最初汇总的内存增长因PowerShell调用`Math.Max`选择整数重载被取整为0，原始样本和最大/初始字节值正确。跨语言重算发现差异后保留原`result.json`及`.local/live-observation-formal-initial.json`，修正为显式浮点重载并新增8KiB及小数阈值回归。用原180个样本重新统计，`.local/live-observation-resource-resummary.json`才是本轮内存增长的正确汇总；其他指标与原结果一致，修正后检查仍通过。没有改写原始样本或声称重新采样15分钟。日志读取长度同时明确使用长整数重载，避免大文件剩余长度被转换为32位整数。

最终脚本还执行了10秒诊断`run-20261013-090756-fdab3201`，完成且`fullTiming=false`、`passed=false`，不计正式窗口。16项回归、短测/提前停止/参数保护、源码摘要和180项实际样本独立重算通过，证据`.local/live-observation-evidence-check.json`、`-tool-hashes.json`、`-resource-resummary.json`及正式目录`samples.jsonl`；修正前源码摘要保留于`-tool-hashes-before-fix.json`。
