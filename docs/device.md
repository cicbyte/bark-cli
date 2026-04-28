# device

管理推送设备。

## 子命令

### device list

列出所有设备实例，与 `server list` 类似但侧重设备信息。

```bash
bark-cli device list
```

### device check

检查设备是否已在 bark-server 注册。支持三种输入方式：

```bash
bark-cli device check <instance|device_key|device_token>
```

优先级：实例名 > device_key > device_token。未匹配到任何实例时，直接作为 device_key 使用。

```bash
# 按实例名
bark-cli device check iphone

# 按 device_key
bark-cli device check <device_key>

# 按 device_token
bark-cli device check <device_token>
```
