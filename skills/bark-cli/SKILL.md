---
name: bark-cli
description: 通过 bark-cli 发送推送通知。当用户要求发送通知、推送消息、提醒、告警到 iOS 设备时使用。触发词：发送推送、推送通知、发通知、提醒我、bark、iOS 通知。
---

# bark-cli Push Skill

通过 bark-cli 向 iOS 设备发送推送通知。前提：用户已配置好 bark 实例（`bark-cli server add`），本 skill 只负责推送，不涉及实例管理。

## 前置检查

推送前执行 `bark-cli server list` 确认已有可用实例。若无实例，提示用户先运行 `bark-cli server add <name> --device-token <token>` 配置。

## 基本推送

```bash
bark-cli push "消息内容"
```

## 完整参数

```bash
bark-cli push [body] \
  -t "标题" \
  -s "副标题" \
  -b "正文" \
  --level critical \
  --group "分组名" \
  --sound "1107" \
  --volume 7 \
  --badge 1 \
  --url "https://example.com" \
  --icon "https://example.com/icon.png" \
  --copy "复制内容" \
  --archive \
  --call \
  --server <实例名>
```

| 参数 | 说明 | 示例 |
|------|------|------|
| `body`（位置参数） | 通知正文 | `"消息内容"` |
| `-t` / `--title` | 标题 | `"会议提醒"` |
| `-s` / `--subtitle` | 副标题 | `"3楼会议室"` |
| `-b` / `--body` | 正文（与位置参数二选一） | `"内容"` |
| `--level` | 通知级别：`critical`（紧急）、`active`（活跃）、`timeSensitive`（时效）、`passive`（静默） | `critical` |
| `--sound` | 提示音名称 | `"1107"`、`"alarm"`、`""`（静音） |
| `--volume` | 音量 0-10 | `7` |
| `--badge` | App 角标数字 | `1` |
| `--group` | 通知分组（相同分组可折叠） | `"work"` |
| `--url` | 点击通知跳转的 URL | `"https://..."` |
| `--icon` | 通知图标 URL | `"https://..."` |
| `--copy` | 复制到剪贴板的内容 | `"订单号: 12345"` |
| `--archive` | 保存到 Bark 通知历史 | — |
| `--call` | 持续响铃 30 秒 | — |
| `--server` | 指定目标实例 | `ipad` |
| `--json` | JSON 格式传参 | 见下方 |

## JSON 参数

需要复杂参数时使用 JSON：

```bash
bark-cli push --json '{"title":"告警","body":"CPU 使用率 95%","level":"critical","group":"monitor","sound":"alarm"}'
```

## 指定实例

多实例时用 `--server` 指定目标：

```bash
bark-cli push --server iphone "发到 iPhone"
bark-cli push --server ipad "发到 iPad"
```

不指定时使用默认实例。

## 输出格式

```bash
bark-cli push --format json "消息"     # JSON 输出
bark-cli push --format jsonl "消息"    # JSONL 输出
```

## 使用规范

- 推送前确认有可用实例（`bark-cli server list`），无实例时提示用户配置，不要自行执行 `server add`
- 不要执行 `server add`、`server remove`、`server use`、`chat`、`mcp` 等管理命令
- `--level critical` 仅用于真正紧急的场景（告警、紧急通知）
- 使用 `--group` 对同类通知分组，避免通知刷屏
- 消息内容简洁明确，标题控制在 20 字以内
