# GroupManager
一个点开即用的管理群聊插件

![Badge](https://img.shields.io/badge/OneBot-v11-black)
[![License](https://img.shields.io/github/license/Yiwen-Chan/GroupManagerBot.svg)](https://raw.githubusercontent.com/Yiwen-Chan/GroupManagerBot/master/LICENSE)
[![QQ 群](https://img.shields.io/badge/qq%E7%BE%A4-1048452984-green.svg)](https://jq.qq.com/?_wv=1027&k=QMb7x1mM)

本项目符合 [OneBot](https://github.com/howmanybots/onebot) 标准，下面内容为旧版本移植，不确保是否有误

可基于以下项目与机器人框架/平台进行交互
| 项目地址 | 平台 | 核心作者 | 备注 |
| --- | --- | --- | --- |
| [Yiwen-Chan/OneBot-YaYa](https://github.com/Yiwen-Chan/OneBot-YaYa) | [先驱](https://www.xianqubot.com/) | kanri |  |
| [richardchien/coolq-http-api](https://github.com/richardchien/coolq-http-api) | CKYU | richardchien | 可在 Mirai 平台使用 [mirai-native](https://github.com/iTXTech/mirai-native) 加载 |
| [Mrs4s/go-cqhttp](https://github.com/Mrs4s/go-cqhttp) | [MiraiGo](https://github.com/Mrs4s/MiraiGo) | Mrs4s |  |
| [yyuueexxiinngg/cqhttp-mirai](https://github.com/yyuueexxiinngg/cqhttp-mirai) | [Mirai](https://github.com/mamoe/mirai) | yyuueexxiinngg |  |
| [takayama-lily/onebot](https://github.com/takayama-lily/onebot) | [OICQ](https://github.com/takayama-lily/oicq) | takayama |  |
| [ProtobufBot](https://github.com/ProtobufBot) | [Mirai](https://github.com/mamoe/mirai) | lz1998 | 事件和 API 数据内容和 OneBot 一致，通信方式不兼容 |

## 开始使用

注意：本插件使用websocket与cqhttp项目进行交互，非反向ws

1.建议选择go-cqhttp，下载releases并按照下面配置设置

2.下载GroupManager的releases，可直接运行

3.第一次运行自动产生config.json，修改后再次运行

4.发送“群管系统”呼出菜单（还没写，详情可以看生成的config.json）

5.在config.json可自定义各种命令以及回复内容

## 配置相关

GroupManager部分设置(GroupManager\config.json)
```json
...
    "插件版本": "1",
	"监听地址": "127.0.0.1",
	"监听端口": "8080",
	"Token": "",
	"主人QQ": 
		- "66666666",
		- "88888888",
...
```
若使用OneBot-YaYa，部分配置如下(XQ\onebot\config.yml)
```yaml
...
  bots:
    # bot的qq号
    - bot: 87654321
      # 正向Websocket服务器
      websocket:
        - name: GroupManager
          enable: true
          host: 127.0.0.1
          port: 8080
          access_token: ""
          post_message_format: string
...
```
若使用cqhttp-mirai，部分配置如下(miraiOK\plugins\CQHTTPMirai\setting.yml)
```yaml
...
  # 正向Websocket服务器
  ws:
    # 可选，是否启用正向Websocket服务器，默认不启用
    enable: true
    # 可选，上报消息格式，string 为字符串格式，array 为数组格式, 默认为string
    postMessageFormat: string
    # 可选，访问口令, 默认为空, 即不设置Token
    accessToken: ""
    # 监听主机
    wsHost: "127.0.0.1"
    # 监听端口
    wsPort: 8080
...
```
若使用go-cqhttp，部分配置如下(go-cqhttp/config.json)
```json
...
	"ws_config": {
		"enabled": true,
		"host": "127.0.0.1",
		"port": 2333
	},
...
```
注意:以上仅列出了与ws相关设置，其他配置可自行摸索或找我配置
## 功能列表
### 禁言类
- [x] 禁言
- [x] 解除禁言
- [x] 全员禁言
- [x] 解除全员禁言
### 权限类
- [x] 升为管理 *
- [x] 取消管理 *
### 设置类
- [x] 修改群名片
- [x] 设置群头衔
### 操作类
- [ ] 撤回
- [ ] 警告
- [x] 踢出
- [x] 退出群聊
- [ ] 敏感词
- [ ] 黑白名单
### 通知类
- [x] 入群欢迎 *
- [x] 退群通知 *
- [x] 入群申请通知 *
- [x] 添加好友通知 *
### 同意类
- [ ] 同意