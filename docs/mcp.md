# mcp

以 stdio 模式启动 MCP Server，让 Claude Desktop、Claude Code、Cherry Studio 等 AI 客户端直接调用推送功能。

## 用法

```bash
bark-cli mcp
```

## 注册的 Tools

| Tool | 说明 |
|------|------|
| `push` | 发送推送通知 |
| `device_register` | 注册推送设备 |
| `device_check` | 检查设备是否已注册 |
| `server_ping` | 检查服务器连通性 |
| `server_info` | 获取服务器信息 |

## 配置

### Claude Desktop

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`：

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

### Claude Code

编辑 `.claude/settings.json` 或项目级 MCP 配置：

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

### Cherry Studio

设置 → MCP 服务器 → 添加：
- 名称：`bark`
- 命令：`bark-cli`
- 参数：`mcp`
