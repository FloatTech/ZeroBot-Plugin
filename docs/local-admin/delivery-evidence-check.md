# 本机交付证据核对

在仓库根目录的 PowerShell 执行：

```powershell
node scripts/check-local-admin-delivery.mjs --evidence-prefix log-clear-display
```

只有后端变化、界面资源保持原样时，可明确分开程序与界面证据：

```powershell
node scripts/check-local-admin-delivery.mjs --evidence-prefix wordcount-range --ui-evidence-prefix log-clear-display
```

`--evidence-prefix`始终对应当前测试程序、Go/vet/构建、整机和部署；可选的`--ui-evidence-prefix`仅指定原界面/新增界面/布局记录，默认与程序前缀相同。分开使用必须存在`.local/<界面前缀>-ui-baseline.json`，包含原界面证据文件与全部静态资源的SHA-256；文件摘要、完整资源文件集合和当前资源内容必须一致，否则退出1，不能仅凭修改时间复用。单一前缀核对成功会保存该界面基线；分开核对不会覆盖旧基线。该基线是本机证据索引，不是配置备份，不会产生新界面测试。

原记录较早、尚无基线文件时，只能从此前已通过的核对结果与保留的原测试文件核实后导入；不能根据当前新程序替旧界面测试生成“通过”结果。此次21:50:21已通过的`log-clear-display`记录用于导入，原197项/64项保留实际执行时间；先核对全部12个界面资源摘要一致和原3份证据早于原部署，原记录另存`.local/log-clear-display-evidence-before-split.json`，导入不会运行新UI测试。热词程序采用22:18:20整机/22:22:03部署证据，分开核对通过。

PowerShell的`Write-Host`就绪信息可能未被原管道捕获。若构建日志缺少该标记，必须有Go版本和模块校验成功记录，并用原升级前`.local/<程序前缀>-before.json`中的`stagedHash`核对测试程序，仍须满足实际整机、部署、时间顺序与磁盘程序一致检查；不能仅凭“all modules verified”判断构建成功。存在就绪标记但程序名错误时仍失败。

`node --test scripts/test-delivery-evidence.mjs`的8项合成CLI回归通过，覆盖单前缀自动基线、分开来源及原时间、修改资源且保持mtime、增加/删除资源、修改仍通过的UI测试记录、基线缺失/重复/错误前缀/未来时间、非法参数无写入、缺失Write-Host标记的原快照摘要，以及新程序/Go回归失败。不访问服务或真实QQ。证据`.local/delivery-split-tool-tests.txt`。

前缀对应 `.local/<前缀>-ui-tests.txt`、`-focused-ui.txt`、`-layout-result.json`、`-go-tests.txt`、`-vet.txt`、`-build.txt`、`-smoke.txt` 和 `-deployment.json`，以及 `.local/bin/zerobot-<前缀>-tested.exe`。分开核对时前三类文件取界面前缀，其他文件取程序前缀。必须先有真实回归和部署记录；脚本不运行测试，不构建或部署程序，不能用不存在的前缀创建“通过”证据。两种前缀均仅允许小写字母、数字和中间短横线。

需要 Node.js 和读取本机忽略目录的文件权限，不需要后台密码。脚本读取既有结果，确认界面测试没有失败/取消/跳过、布局为离线通过、四个相关 Go 包有通过记录且 vet 无诊断、构建记录和 Windows 整机/部署检查通过；核对磁盘日常程序与测试构建的 SHA-256、资源修改时间及构建/整机/部署顺序。另检查 `docs/local-admin` 的本地 Markdown 链接和整机脚本声明的固定 HTML/JS/CSS 入口，记录当前静态资源摘要。

通过时退出码为 0，标准输出简要结果；失败退出码为 1，并列出未满足的检查。参数不符合格式时退出码为 2。完整结果写入 `.local/current-delivery-evidence.json`，每次覆盖本机上一份核对结果；该文件不是配置备份，不用于恢复。输出不读取配置、凭据、QQ 正文或模型记录，程序和静态资源摘要只用于版本核对。

已有测试记录和修改时间不能独立证明所有业务正确。磁盘程序一致不代表进程正在运行，历史部署的账号/连接正常不代表当前仍在线；脚本不访问服务、QQ 或模型接口。它不替代新代码的回归、全新机器构建、真实页面体验、实际 QQ 业务或长期性能验收。文档维护使用已有证据时，必须保留原测试和部署时间，并说明本轮没有重新执行。

新功能开发仍先更新使用说明，执行相关回归后再核对新前缀；核对通过不能自动关闭其他人工待办。当前交付范围和执行记录见 [交付核对](release-v0.1.md)。
