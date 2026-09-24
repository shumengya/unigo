# Garnet for Zed

为 Garnet（`.gn`）提供 Zed 编辑器支持，基于 tree-sitter：

- 语法高亮（关键字、类型、内置类型、结构体/接口/枚举、字段、函数与方法调用、中文标识符）
- 括号匹配、自动缩进、`//` 注释切换
- 大纲（Outline）：package、import、func、struct、interface、enum、newtype、alias、顶层 var/const

## 目录

```
editors/
├── zed/                      本扩展
│   ├── extension.toml        Zed 扩展清单
│   └── languages/garnet/     语言配置与查询（highlights / brackets / indents / outline / locals / textobjects）
└── tree-sitter-garnet/       Garnet 的 tree-sitter 语法（Zed 通过 extension.toml 按提交号引用）
    ├── grammar.js
    ├── src/                  tree-sitter generate 生成的解析器
    └── queries/              与 zed/languages/garnet 同步的查询
```

## 本地安装

1. Zed 中执行命令 `zed: install dev extension`
2. 选择本目录 `editors/zed`
3. 打开任意 `.gn` 文件

Zed 构建扩展时会下载 wasi-sdk（约 600MB），网络不好时先在 Zed 设置里配置代理。

`extension.toml` 里的 grammar 指向 GitHub 上的主仓库（`path = "editors/tree-sitter-garnet"`）和一个提交号。
Zed 构建时按这个提交号去拉语法，所以**改完语法必须先推送，再更新 `rev`**：

```sh
cd ../tree-sitter-garnet
npx tree-sitter generate
npx tree-sitter test                              # 语料与高亮测试
npx tree-sitter parse ../../examples/tour/tour.gn # 确认没有 ERROR
git commit -am "..." && git push
git rev-parse HEAD                                # 把这个提交号填回 extension.toml 的 rev，再提交一次
```

查询文件改在 `tree-sitter-garnet/queries/`，再同步复制到 `languages/garnet/`。
然后在 Zed 的扩展页对 Garnet 点 “Rebuild”。

## 发布

仓库已公开，按 Zed 官方流程把本扩展提交到 zed-industries/extensions 即可（扩展路径为 `editors/zed`）。
