# 编辑器支持

| 目录 | 说明 |
| --- | --- |
| `tree-sitter-unigo/` | UniGo 的 tree-sitter 语法，被 Zed 扩展按提交号引用 |
| `zed/` | Zed 扩展：高亮、缩进、括号、大纲 |
| `vscode/` | VS Code 扩展：TextMate 语法、括号、注释、代码片段 |

两个编辑器用的语法定义是各自独立的：

- **Zed** 用 `tree-sitter-unigo/grammar.js` 生成的解析器。改动语法后要 `npx tree-sitter generate`，跑 `npx tree-sitter test`，提交并推送后，把新提交号写回 `zed/extension.toml` 的 `rev`。
- **VS Code** 没有 tree-sitter，用 TextMate 语法，由 `vscode/scripts/build-grammar.js` 生成 `vscode/syntaxes/unigo.tmLanguage.json`，不要直接改那个 JSON。

两边的改动都要跑 `scripts/test.ps1`（或 `test.sh`）里的语法测试。

后续两者都接 `lsp/` 提供的语言服务器，就能有诊断和补全。
