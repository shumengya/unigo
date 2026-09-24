# 示例

| 文件 | 说明 |
| --- | --- |
| `tour/tour.ug` | 语法总览，覆盖白皮书里已经定下的全部写法 |
| `hello/hello.ug` | 最小可运行程序，含结构体 |
| `helloworld/helloworld.ug` | 最简单的打印 |

每个示例都必须能编译通过，这由 `compiler/compile` 的 `TestPassFixtures` 保证。
新增示例后不用登记，测试会自动扫到。

编译并运行：

```sh
./bin/unigo build examples/hello/hello.ug
./examples/hello/unigo-out/hello.exe
```
