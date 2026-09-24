# 更新记录

## 未发布

### 新增
- 项目按组件重新组织：`docs/`、`compiler/`、`lsp/`、`editors/`、`examples/`、`tests/`、`stdlib/`、`reference/`、`scripts/`
- VS Code 扩展：TextMate 语法高亮、括号、注释、代码片段
- Zed 扩展与 `tree-sitter-garnet` 语法：高亮、缩进、括号、大纲、文本对象
- `garnet-lsp` 骨架
- `tests/pass/`：必须编译通过的用例；`examples/` 下的示例也会被自动检查
- 构建与测试脚本 `scripts/build.*`、`scripts/test.*`

### 修复
- 字面量传给接口类型参数（如 `fmt.Println("你好")`）时误报"参数类型不符"
