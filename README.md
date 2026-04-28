# bark-cli

> [Bark](https://github.com/Finb/bark-server) 推送通知的命令行客户端 — 多实例管理、AI 对话推送、MCP 集成，全部在终端完成。

English: [README.en.md](README.en.md)

![Release](https://img.shields.io/github/v/release/cicbyte/bark-cli?style=flat)
![Go Report Card](https://goreportcard.com/badge/github.com/cicbyte/bark-cli)
![License](https://img.shields.io/github/license/cicbyte/bark-cli)
![Last Commit](https://img.shields.io/github/last-commit/cicbyte/bark-cli)

## 功能特性

- **多实例管理** — 添加多个 bark-server 实例，按名称引用，支持设置默认实例
- **一键推送** — 标题、正文、分组、角标、图标、URL 跳转、持续响铃等全部参数支持
- **自动注册** — `server add` 时传入 device_token 自动完成设备注册
- **AI 对话推送** — 通过自然语言发送推送，LLM 自动解析意图并调用推送 API
- **MCP Server** — stdio 模式，让 Claude Desktop、Cherry Studio 等 AI 客户端直接发送推送
- **多格式输出** — table / JSON / JSONL，`--format` 全局切换
- **多 AI 提供商** — OpenAI、Ollama、智谱等 OpenAI 兼容 API

## 效果展示

| iOS 通知 | AI 单轮对话 | AI 多轮对话 | Cherry Studio |
|:---:|:---:|:---:|:---:|
| ![iOS 通知](images/bark-app.jpg) | ![AI 单轮对话](images/chat.png) | ![AI 多轮对话](images/loop-chat.png) | ![Cherry Studio MCP](images/cherry-claw.png) |

## 安装

### 从 Release 下载

前往 [Releases](https://github.com/cicbyte/bark-cli/releases) 下载对应平台的预编译二进制文件。

### 从源码构建

```bash
git clone https://github.com/cicbyte/bark-cli.git
cd bark-cli
go build -o bark-cli .
```

<details>
<summary>交叉编译（可选，含版本号注入和 UPX 压缩）</summary>

```bash
python scripts/build.py --local    # 当前平台
python scripts/build.py             # 全平台（Windows/Linux/macOS）
```

</details>

**环境要求：** Go >= 1.23

## 快速开始

```bash
# 添加 bark 实例（自动注册设备）
bark-cli server add iphone --device-token <your_device_token>

# 设为默认实例
bark-cli server use iphone

# 发送推送
bark-cli push "Hello from bark-cli"
bark-cli push -t "提醒" -b "会议即将开始" --group work --level timeSensitive

# 通过 AI 对话推送
bark-cli chat "帮我发一条推送提醒下午3点开会"
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `push [body]` | 发送推送通知 |
| `server add <name>` | 添加 bark 实例（自动注册设备） |
| `server list` | 列出所有实例 |
| `server use <name>` | 设置默认实例 |
| `server remove <name>` | 删除实例 |
| `server ping` | 检查服务器连通性 |
| `server info` | 查看服务器信息 |
| `device list` | 列出所有设备实例 |
| `device check <key>` | 检查设备是否已注册 |
| `chat [question]` | AI 对话推送（`-i` 多轮交互） |
| `mcp` | 启动 MCP Server |
| `config list` | 查看应用配置 |

### 推送

```bash
bark-cli push "简单消息"                                      # 最简推送
bark-cli push -t "标题" -b "内容" --level critical           # 带级别
bark-cli push -t "链接" --url https://example.com            # 带跳转
bark-cli push --group "提醒" --badge 1 --sound "1107"        # 分组+角标+提示音
bark-cli push --call                                          # 持续响铃 30 秒
bark-cli push --json '{"title":"标题","body":"内容","level":"active"}'  # JSON 参数
bark-cli push --server ipad "发到 iPad"                       # 指定实例
echo "管道消息" | bark-cli push                             # 管道输入
cat log.txt | bark-cli push -t "日志告警"                    # 管道+标题
```

### 实例管理

```bash
bark-cli server add iphone --device-token <token>              # 添加实例（自动注册）
bark-cli server add myserver --url https://bark.example.com --token <api_token>
bark-cli server list                                           # 列出实例（* 标记默认）
bark-cli server use iphone                                     # 设置默认
bark-cli server remove myserver                                # 删除实例
bark-cli server ping                                           # 连通性检测
bark-cli server info                                           # 服务器版本信息
```

### 设备检查

`device check` 支持实例名、device_key、device_token 三种输入方式：

```bash
bark-cli device check iphone              # 按实例名
bark-cli device check <device_key>        # 按 device_key
bark-cli device check <device_token>      # 按 device_token
```

## 配置

```bash
bark-cli config list                     # 列出所有配置
bark-cli config get ai.model             # 查看配置值
bark-cli config set ai.provider openai   # 设置配置值
bark-cli config set ai.api_key           # 敏感字段交互式输入
```

配置文件：`~/.cicbyte/bark-cli/config/config.yaml`（首次运行自动创建）

## MCP Server

`bark-cli mcp` 以 stdio 模式运行 MCP Server，让 AI 客户端直接发送推送通知。

**Claude Desktop：**

```json
{
  "mcpServers": {
    "bark": {
      "command": "bark-cli",
      "args": ["mcp"]
    }
  }
}
```

**Cherry Studio：** 设置 → MCP 服务器，命令 `bark-cli`，参数 `mcp`

## 全局选项

```bash
bark-cli push --format json          # JSON 缩进输出
bark-cli server list --format jsonl  # JSONL 逐行输出
```

## 数据存储

```
~/.cicbyte/bark-cli/
├── config/
│   └── config.yaml    # 应用配置（实例、AI、日志等）
└── logs/
    └── bark-cli_log_YYYYMMDD.log  # 日志文件（自动轮转）
```

## 技术栈

- Go 1.23+
- [Cobra](https://github.com/spf13/cobra) — CLI 框架
- [mcp-go](https://github.com/mark3labs/mcp-go) — MCP Server
- [go-openai](https://github.com/sashabaranov/go-openai) — OpenAI 兼容 API
- [go-pretty](https://github.com/jedib0t/go-pretty) — 终端表格
- [Glamour](https://github.com/charmbracelet/glamour) — Markdown 渲染
- [Zap](https://github.com/uber-go/zap) — 结构化日志

## 许可证

[MIT](LICENSE) © 2026 cicbyte
