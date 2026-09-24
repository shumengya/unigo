# 脚本

| 脚本 | 作用 |
| --- | --- |
| `build.sh` / `build.ps1` | 构建 `bin/garnet` 和 `bin/garnet-lsp` |
| `test.sh` / `test.ps1` | 跑转译器测试 + tree-sitter 语法测试 + VS Code 高亮测试 |

`.ps1` 文件带 UTF-8 BOM 并设置输出编码，避免 Windows PowerShell 5.1 把中文读成乱码。

编辑器那两部分测试在没装依赖时会跳过，先执行：

```sh
cd editors/tree-sitter-garnet && npm install
cd editors/vscode && npm install
```
