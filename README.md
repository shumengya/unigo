# UniGo

> 书同文，车同轨。

UniGo 是一门基于 Go 生态的编程语言，源文件后缀 `.ug`。它和 Go 的关系类似 TypeScript 之于 JavaScript、
Kotlin 之于 Java：语法重新设计，运行时和库沿用 Go。

```
.ug 源码 ──unigo──▶ .go 源码 ──go build──▶ 各平台二进制
```

设计原则：**每种意思只保留一种写法**。用强制统一的语法减少心智成本，在人和机器之间取得平衡。

```unigo
package main;

import fmt from "fmt";

struct Student {
	name :string;
	age :int;
}

func main() {
	var item :Student = Student{name = "树萌芽", age = 18};
	var written :int = fmt.Println(item.name) else return;
}
```

## 目录

| 目录 | 内容 |
| --- | --- |
| [`docs/`](docs/) | 白皮书、语言规范、入门教程、设计决策、路线图 |
| [`compiler/`](compiler/) | 转译器 `unigo`：词法 → 语法 → 类型检查 → 生成 Go |
| [`lsp/`](lsp/) | 语言服务器 `unigo-lsp`（骨架） |
| [`editors/`](editors/) | 编辑器支持：VS Code、Zed、tree-sitter 语法 |
| [`examples/`](examples/) | 示例程序 |
| [`tests/`](tests/) | 语言一致性测试（应当通过 / 应当报错的源码） |
| [`stdlib/`](stdlib/) | 标准库（规划中） |
| [`reference/`](reference/) | 参照语言的语法笔记（Go、TypeScript），设计时对照用 |
| [`scripts/`](scripts/) | 构建、测试脚本 |

## 快速开始

需要 Go 1.22+。

```powershell
# Windows
./scripts/build.ps1                     # 生成 bin/unigo.exe 和 bin/unigo-lsp.exe
./bin/unigo.exe build examples/tour/tour.ug
./examples/tour/unigo-out/tour.exe
```

```sh
# Linux / macOS / Git Bash
./scripts/build.sh
./bin/unigo build examples/tour/tour.ug
```

运行全部测试：`./scripts/test.ps1` 或 `./scripts/test.sh`。

## 现状

处于早期原型阶段：单文件转译已可用，能调用 Go 标准库。
缺失的能力和计划见 [路线图](docs/ROADMAP.md)。

## 许可

[MIT](LICENSE)
