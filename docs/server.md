# server

管理 bark-server 实例。支持添加、列出、切换、删除多个实例。

## 子命令

### server add

添加 bark 实例，传入 `--device-token` 时自动完成设备注册。

```bash
bark-cli server add <name> [flags]
```

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `--url` | bark-server 地址 | `https://api.day.app` |
| `--device-token` | 设备推送 token | — |
| `--token` | API 认证 token（自建服务器需要） | — |
| `--timeout` | 请求超时（秒） | `10` |

```bash
# 使用官方服务器
bark-cli server add iphone --device-token <token>

# 使用自建服务器
bark-cli server add myserver --url https://bark.example.com --token <api_token>
```

### server list

列出所有实例，`*` 标记默认实例。

```bash
bark-cli server list
```

### server use

设置默认实例，后续 push/chat 等命令不指定 `--server` 时使用该实例。

```bash
bark-cli server use <name>
```

### server remove

删除实例。

```bash
bark-cli server remove <name>
```

### server ping

检查服务器连通性。

```bash
bark-cli server ping              # 默认实例
bark-cli server ping --server myserver
```

### server info

查看服务器版本信息。

```bash
bark-cli server info              # 默认实例
bark-cli server info --server myserver
```
