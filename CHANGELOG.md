# 更新记录

## 未发布

### 变更
- 语言由 Garnet 更名为 **UniGo**，源文件后缀由 `.gn` 改为 `.ug`。`.gn` 已被 Google 的构建工具 GN 占用；UniGo 取"统一写法的 Go"之意，命名方式参照 TypeScript 之于 JavaScript
- 命令改为 `unigo` / `unigo-lsp`，输出目录改为 `unigo-out/`，Go 模块名改为 `unigo`
- GitHub 仓库改为 shumengya/unigo

### 新增
- 项目按组件重新组织：`docs/`、`compiler/`、`lsp/`、`editors/`、`examples/`、`tests/`、`stdlib/`、`reference/`、`scripts/`
- VS Code 扩展：TextMate 语法高亮、括号、注释、代码片段
- Zed 扩展与 `tree-sitter-unigo` 语法：高亮、缩进、括号、大纲、文本对象
- `unigo-lsp` 骨架
- `tests/pass/`：必须编译通过的用例；`examples/` 下的示例也会被自动检查
- 构建与测试脚本 `scripts/build.*`、`scripts/test.*`

### 修复
- 字面量传给接口类型参数（如 `fmt.Println("你好")`）时误报"参数类型不符"
