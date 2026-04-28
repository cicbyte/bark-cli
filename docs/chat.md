# chat

AI 对话推送，通过自然语言发送推送通知。LLM 自动解析意图并调用推送 API。

## 用法

```bash
bark-cli chat [question] [flags]
```

## 选项

| 选项 | 别名 | 说明 | 默认值 |
|------|------|------|--------|
| `--interactive` | `-i` | 多轮对话模式 | `false` |
| `--non-stream` | — | 非流式输出 | `false` |

## 示例

```bash
# 单轮对话
bark-cli chat "帮我发一条推送提醒下午3点开会"

# 多轮交互
bark-cli chat -i
```

## 多轮对话

使用 `-i` 进入交互模式：

```
  user > 帮我发一条推送标题是测试
  ▸ push
  ✓ 推送成功

推送已发送...

  user > 再发一条到iPad
  ▸ push
  ✓ 推送成功

推送已发送...

  user > /quit
  再见!
```

交互命令：

| 命令 | 说明 |
|------|------|
| `/quit` / `/exit` / `/q` | 退出对话 |
| `/clear` | 清除对话上下文 |

## 可用工具

AI 在对话中可调用以下工具：

| 工具 | 说明 |
|------|------|
| `push` | 发送推送通知 |
| `device_register` | 注册推送设备 |
| `device_check` | 检查设备是否已注册 |
| `server_ping` | 检查服务器连通性 |
| `server_info` | 获取服务器信息 |
