# 语言一致性测试

直接放 `.ug` 源码，由 `compiler/compile/compile_test.go` 自动扫到。

## 两个目录

| 目录 | 要求 |
| --- | --- |
| `pass/` | 必须**编译通过**。用来锁住"合法写法一直合法" |
| `fail/` | 必须**编译失败**。第一行写 `// error: 期望的错误信息`，报错必须包含这段文字 |

## 例子

```unigo
// error: 参数类型不符
package main;

func double(num :int) :int {
	return num * 2;
}

func main() {
	var result :int = double("520");
}
```

## 加用例的原则

每种禁止的写法、每条从白皮书读出来的规则，都应该在 `fail/` 里有一个用例；
每个修好的 bug，都应该在 `pass/` 里补一个能复现它的用例。跑测试：

```sh
go test ./compiler/...
```
