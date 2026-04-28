# push

发送推送通知到 bark-server 设备。

## 用法

```bash
bark-cli push [body] [flags]
```

`body` 可作为位置参数传入，也可通过 `--body` 指定。

## 选项

| 选项 | 别名 | 说明 | 默认值 |
|------|------|------|--------|
| `--title` | `-t` | 通知标题 | — |
| `--subtitle` | `-s` | 通知副标题 | — |
| `--body` | `-b` | 通知内容 | — |
| `--server` | — | 指定实例名称（不传则使用默认实例） | — |
| `--sound` | — | 提示音 | `1107` |
| `--level` | — | 通知级别 (`critical`/`active`/`timeSensitive`/`passive`) | — |
| `--volume` | — | 提示音量 (0-10) | `5` |
| `--badge` | — | 角标数字 | `0` |
| `--icon` | — | 图标 URL | — |
| `--group` | — | 通知分组 | — |
| `--url` | — | 点击跳转 URL | — |
| `--copy` | — | 复制内容 | — |
| `--archive` | — | 保存通知 | `false` |
| `--call` | — | 持续响铃 30 秒 | `false` |
| `--json` | — | 从 JSON 字符串读取参数 | — |

## 示例

```bash
# 最简推送
bark-cli push "简单消息"

# 带标题和内容
bark-cli push -t "标题" -b "内容"

# 通知级别
bark-cli push -t "紧急" -b "内容" --level critical

# 带跳转链接
bark-cli push -t "链接" --url https://example.com

# 分组 + 角标 + 提示音
bark-cli push --group "提醒" --badge 1 --sound "1107"

# 持续响铃
bark-cli push --call

# JSON 参数
bark-cli push --json '{"title":"标题","body":"内容","level":"active"}'

# 推送到指定实例
bark-cli push --server ipad "发到 iPad"

# 管道输入
echo "管道消息" | bark-cli push
cat log.txt | bark-cli push -t "日志告警"
```
