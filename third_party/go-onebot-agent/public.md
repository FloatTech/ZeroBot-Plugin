# 公开 API

## `get_msg` 获取消息

### 参数

| 字段名         | 数据类型  | 说明   |
| ------------ | ----- | ------ |
| `message_id` | number (int32) | 消息 ID |

### 响应数据

| 字段名         | 数据类型    | 说明       |
| ------------ | ------- | ---------- |
| `time`       | number (int32) | 发送时间   |
| `message_type` | string | 消息类型 |
| `sender`     | *User  | 发送人信息 |
| `message`    | json.RawMessage | 消息内容   |

## `get_forward_msg` 获取合并转发消息

### 参数

| 字段名         | 数据类型   | 说明   |
| ------------ | ------ | ------ |
| `id` | string | 合并转发 ID |

### 响应数据

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `message` | json.RawMessage | 消息内容 |

## `get_stranger_info` 获取陌生人信息

### 参数

| 字段名 | 数据类型 | 默认值 | 说明 |
| ----- | ------- | ----- | --- |
| `user_id` | number | - | QQ 号 |
| `no_cache` | boolean | `false` | 是否不使用缓存（使用缓存可能更新不及时，但响应更快） |

### 响应数据

| 字段名 | 数据类型 | 说明 |
| ----- | ------- | --- |
| `user_id` | number (int64) | QQ 号 |
| `nickname` | string | 昵称 |
| `sex` | string | 性别，`male` 或 `female` 或 `unknown` |
| `age` | number (int32) | 年龄 |

## `get_friend_list` 获取好友列表

### 参数

无

### 响应数据

响应内容为 JSON 数组，每个元素如下：

| 字段名 | 数据类型 | 说明 |
| ----- | ------- | --- |
| `user_id` | number (int64) | QQ 号 |
| `nickname` | string | 昵称 |
| `remark` | string | 备注名 |

## `get_group_info` 获取群信息

### 参数

| 字段名 | 数据类型 | 默认值 | 说明 |
| ----- | ------- | ----- | --- |
| `group_id` | number | - | 群号 |
| `no_cache` | boolean | `false` | 是否不使用缓存（使用缓存可能更新不及时，但响应更快） |

### 响应数据

| 字段名 | 数据类型 | 说明 |
| ----- | ------- | --- |
| `group_id` | number (int64) | 群号 |
| `group_name` | string | 群名称 |
| `member_count` | number (int32) | 成员数 |
| `max_member_count` | number (int32) | 最大成员数（群容量） |

## `get_group_list` 获取群列表

### 参数

无

### 响应数据

响应内容为 JSON 数组，每个元素和上面的 `get_group_info` 接口相同。

## `get_group_member_info` 获取群成员信息

### 参数

| 字段名 | 数据类型 | 默认值 | 说明 |
| ----- | ------- | ----- | --- |
| `group_id` | number | - | 群号 |
| `user_id`  | number | - | QQ 号 |
| `no_cache` | boolean | `false` | 是否不使用缓存（使用缓存可能更新不及时，但响应更快） |

### 响应数据

*User

## `get_group_member_list` 获取群成员列表

### 参数

| 字段名 | 数据类型 | 默认值 | 说明 |
| ----- | ------- | ----- | --- |
| `group_id` | number (int64) | - | 群号 |

### 响应数据

响应内容为 JSON 数组，每个元素的内容和上面的 `get_group_member_info` 接口相同。
