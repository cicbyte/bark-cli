# config

管理应用配置。

## 子命令

### config list

列出所有配置项（敏感信息自动脱敏），按分组表格显示。

```bash
bark-cli config list
```

### config get

查看指定配置项的值。敏感字段默认脱敏，`--show` 查看明文。

```bash
bark-cli config get <key>
bark-cli config get ai.api_key --show
```

### config set

设置配置项的值。敏感字段（`ai.api_key`）不提供 value 时交互式输入（不回显）。

```bash
bark-cli config set <key> [value]
```

```bash
bark-cli config set ai.provider openai
bark-cli config set ai.model gpt-4o
bark-cli config set ai.api_key sk-xxx
bark-cli config set ai.api_key                # 交互式输入（不回显）
bark-cli config set log.level debug
bark-cli config set output.format json
```

## 可用配置项

| Key | 说明 | 类型 |
|-----|------|------|
| `ai.provider` | AI 提供商 | string |
| `ai.base_url` | API 地址 | string |
| `ai.model` | 模型名称 | string |
| `ai.api_key` | API 密钥（敏感） | string |
| `ai.max_tokens` | 最大 token 数 | int |
| `ai.temperature` | 温度参数 | float |
| `ai.timeout` | 请求超时（秒） | int |
| `output.format` | 输出格式 (`table`/`json`/`jsonl`) | string |
| `log.level` | 日志级别 | string |
| `log.max_size` | 日志最大大小 (MB) | int |
| `log.max_backups` | 日志最大备份数 | int |
| `log.max_age` | 日志最大保留天数 | int |
| `log.compress` | 是否压缩日志 | bool |

## 配置文件

配置文件路径：`~/.cicbyte/bark-cli/config/config.yaml`（首次运行自动创建）。

```yaml
version: "1.0"
servers:
  default: iphone
  instances:
    iphone:
      url: "https://api.day.app"
      device_token: "your_device_token"
      device_key: "auto_registered_key"
      token: ""
      timeout: 10

ai:
  provider: openai
  base_url: https://api.openai.com/v1
  api_key: sk-xxx
  model: gpt-4o

output:
  format: table

log:
  level: info
  max_size: 10
  max_backups: 30
  max_age: 30
  compress: true
```
