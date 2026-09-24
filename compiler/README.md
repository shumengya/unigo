# 转译器 unigo

把 `.ug` 源码转成 Go 源码，再调用 `go build` 得到二进制。

```
.ug ──lexer──▶ 记号 ──parser──▶ 语法树 ──check──▶ 带类型信息的语法树 ──gogen──▶ .go ──go build──▶ 二进制
```

## 目录

| 目录 | 职责 |
| --- | --- |
| `token/` | 记号种类与位置 |
| `lexer/` | 词法分析。这一层就拦下 `:=`、`/* */`、字符字面量等禁止写法 |
| `ast/` | 语法树节点 |
| `parser/` | 语法分析。生成 Go 时用 Go 自己的语法树和打印器，但解析 `.ug` 必须用自己的 parser |
| `check/` | 类型检查与语义规则：冒号空格、名字长度、枚举穷尽、`else return`、newtype 混算 |
| `gogen/` | 生成 Go 代码 |
| `compile/` | 驱动：串起上面几步，写 `unigo-out/` 并调用 `go build` |
| `diag/` | 报错信息 |
| `names/` | 名字规则（长度限制等） |
| `cmd/unigo/` | 命令行入口 |

## 用法

```sh
go build -o bin/unigo ./compiler/cmd/unigo
./bin/unigo build examples/tour/tour.ug
```

## 测试

```sh
go test ./compiler/...
```

- `tests/fail/` 下每个 `.ug` 的第一行写 `// error: 期望的错误信息`，必须编译失败且报错包含该信息。
- `tests/pass/` 和 `examples/` 下的 `.ug` 必须编译通过。

## 落点

输出写在源文件同级的 `unigo-out/`：生成的 `.go`、一个临时 `go.mod`、以及最终二进制。
生成 Go 时用 Go 的语法树和打印器，不做字符串拼接。详细规则见 [docs/spec](../docs/spec/)。
