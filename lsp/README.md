# 语言服务器 garnet-lsp

为编辑器提供诊断、悬停、跳转、补全。**目前只是骨架**，`Run` 直接返回"还没有实现"。

## 目录

| 路径 | 职责 |
| --- | --- |
| `server/` | 协议实现 |
| `cmd/garnet-lsp/` | 命令行入口，走标准输入输出 |

## 用法

```sh
go build -o bin/garnet-lsp ./lsp/cmd/garnet-lsp
```

## 实现顺序

1. **诊断**：打开和修改 `.gn` 文件时跑 `compiler/parser` + `compiler/check`，把错误推给编辑器。这一步成本最低、收益最大，因为规则检查已经写好了。
2. **悬停**：读 `compiler/check.Info` 里的类型信息显示标识符类型。
3. **跳转定义 / 查找引用**：需要把 `check.Info` 里的位置信息整理成符号索引。
4. **补全**：关键字、局部变量、结构体字段、导入的 Go 包成员（包成员要靠 `go/importer` 列出来）。
5. **格式化**：接 `garnet fmt`。

## 注意

- 诊断需要"不落盘也能检查"的入口。现在 `compile.Build` 会写文件并调 `go build`，LSP 只需要 `parser.Parse` + `check.Check`，不要走完整编译。
- 编辑器插件（`editors/vscode`、`editors/zed`）后续接上这个服务器。
