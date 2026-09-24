# 参照语言笔记

设计 Garnet 时对照用的语法总览，不是 Garnet 的文档。

| 文件 | 内容 |
| --- | --- |
| `golang.go` | Go 语法总览。Garnet 兼容 Go 的库与生态，先弄清 Go 有什么 |
| `typescript.ts` | TypeScript 的类型语法。Garnet 对标的是"TS 之于 JS"这层关系 |

这两个文件有各自的 `go.mod`（`module garnet-reference`），和转译器无关，不会被 `go test ./...` 收录。
运行方式写在各文件头部注释里。
