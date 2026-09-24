# 语言规范

这里放 Garnet 的正式规范，是对白皮书的精确化：白皮书说"应该怎样"，规范说"语法树的形状、类型规则、求值顺序、生成 Go 的对应规则"。

## 计划内容

| 文件 | 内容 |
| --- | --- |
| `lexical.md` | 字符集、标识符、字面量、注释、分号 |
| `grammar.md` | 完整语法（EBNF 或 railroad 图） |
| `types.md` | 类型系统、赋值相容、newtype 与 alias 的区别 |
| `errors.md` | `else return` 的语义与错误传递规则 |
| `codegen.md` | `.gn` 的每个构造对应生成什么 Go 代码 |

## 已有的语法定义

`editors/tree-sitter-garnet/grammar.js` 已经是一份可执行的语法，能解析全部示例并带语料测试。
写 `grammar.md` 时以它为准。
