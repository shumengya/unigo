# Garnet for VS Code

为 Garnet（`.gn`）提供编辑器支持。

## 功能

- 语法高亮：关键字、类型注解 `名字 :类型`、`map[键, 值]`、切片 `T[]`、指针 `*os.File`、
  结构体 / 接口 / 枚举 / `newtype` / `alias` 声明、`else return`、枚举 `case Status.ready`、
  结构体字面量字段、函数与方法调用、中文标识符
- 禁止的写法直接标红：`:=`、`/* */`、`type`、`class`、`new`、`extends`、`implements`、
  `fallthrough`、`iota`、字符字面量 `'a'`、不支持的转义
- `//` 行注释切换（Ctrl+/）、括号匹配与自动闭合、缩进
- 常用代码片段：`func`、`main`、`var`、`varm`、`if`、`ife`、`switch`、`struct`、`interface`、`enum` 等
- 默认使用 Tab 缩进

## 安装

```sh
npm install
npm run package          # 生成 garnet-lang-0.1.0.vsix
code --install-extension garnet-lang-0.1.0.vsix
```

## 开发

语法由 `scripts/build-grammar.js` 生成 `syntaxes/garnet.tmLanguage.json`，不要直接改 JSON。

```sh
npm test                 # 重新生成语法并用 vscode-textmate 跑分词断言
node test/tokenize.js -v # 打印 test/sample.gn 每个 token 的 scope
```

在 VS Code 里按 F5（扩展开发宿主）打开任意 `.gn` 文件即可预览。
