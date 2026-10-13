# 本地机器人管理后台：使用与开发文档

本地管理后台v0.1，Windows为首要验证环境；管理自己的单个ZeroBot-Plugin进程及已有多条OneBot连接。

| 入口 | 用途 |
| --- | --- |
| [使用指南](user-guide.md) | 安装与启动、各后台页面、QQ命令、Agent和排障 |
| [第一版说明](release-v0.1.md) | 功能、验证、限制及升级方式 |
| [运行维护](operations.md) | 服务启停、配置生效、恢复与数据边界 |
| [Windows OneBot](onebot-windows.md) | 本机NapCat安装、账号、连接与密钥配对 |
| [AI配置检查](ai-config-check.md) | 接口类型/地址/密钥/模型与实际业务检查 |
| [AI角色权限](ai-role-permissions.md) | 主人、群主、管理员和普通用户权限 |
| [Agent验证](real-agent-acceptance.md) | 实际动作、回执、持久记忆与重启验证 |
| [Agent档案](agent-history-2026-10-03.md) | 对话留存、重置、备份及模型请求容量 |
| [群语音](qq-ai-voice.md) | QQ内置群AI语音与私聊限制 |
| [识图](deepseek-vision.md) | 现有Agent图片描述流程与模型配置 |
| [GPT引用](gpt-quoted-message.md) | 引用文字与本次要求分别处理 |
| [响应与插件](response-status.md) | 全局/群响应规则与插件开关区别 |
| [配置备份](backup-restore-2026-10-02.md) | 预览、确认、覆盖范围与恢复后重启 |
| [统计与观察](combined-observation.md) | 模型、回复和资源资料独立与联合核对 |
| [需求](requirements.md) | 范围、验收标准与设计 |
| [开发进度](progress.md) | 第一版实现及已验证范围 |
| [自测](self-test.md) | 复验入口与真实/隔离边界 |
| [GitHub交付](github-delivery.md) | 中文提交/PR、文件筛选与中间记录排除 |

功能细节由使用指南链接到对应说明。每次功能更新同步维护使用文档；没有对应说明时先创建。根目录[开发约定](../../AGENTS.md)同时适用。

公开文档不包含本机QQ身份、凭据、截图、聊天明细或私有验收文件；中间交流与原始开发记录集中在被忽略的docs/local-notes，运行产物继续保存在.local。
