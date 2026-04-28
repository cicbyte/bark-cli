# version

显示 bark-cli 版本信息。

## 用法

```bash
bark-cli version
```

## 输出示例

```
bark-cli v0.1.0
  commit: 07d55c7
  built:  2026-04-28T14:22:00
```

版本号通过 `VERSION` 文件管理，构建时通过 `-ldflags -X` 注入到二进制文件中。Release 构建会自动包含版本号、Git commit 和构建时间。
