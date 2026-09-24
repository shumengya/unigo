package main

// ============================================================================
// Go 语法总览（学习示例）
// ============================================================================
//
// 覆盖：日常写 Go 会用到的语法和语法糖（到 Go 1.22+ 的常用部分）。
// 运行：go run golang.go
//
// 和 JS / Python 不同的几件大事：
// - 编译期就检查类型；未使用的局部变量、导入会直接编译失败
// - 没有 class / 继承，用 struct + 方法 + 接口
// - 错误是返回值，不是异常
// - 并发是语法级的：go / chan / select

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
)

// ============================================================================
// 1. 注释
// ============================================================================

// 单行注释

/*
多行注释
*/

// 文档注释写在 package / 函数 / 类型正上方，godoc 会收集。
// 习惯：完整句子，以被注释对象的名字开头。

// ============================================================================
// 2. 变量、常量、iota
// ============================================================================

// 包级变量：var 声明，零值可用（不用先「初始化再使用」）
var packageVar = 1

const Pi = 3.14 // 常量必须编译期能算出来
const (
	StatusOK = iota // 0，往下自动 +1
	StatusFail
	StatusBusy
)

func vars() {
	var a int        // 零值 0
	var s string     // 零值 ""
	var ok bool      // 零值 false
	var p *int       // 零值 nil
	b := 2           // 短声明：只能在函数里，自动推断类型
	a, b = b, a      // 交换
	c, d := 3, "go"  // 一次多个
	c, e := 4, 5     // 至少一个是新变量时，才能再用 :=
	const local = 10 // 函数里也可以 const
	_, _ = ok, p
	_, _, _, _, _ = s, c, d, e, local
}

// ============================================================================
// 3. 基本类型与转换
// ============================================================================

func types() map[string]any {
	var (
		i   int     = 42
		i8  int8    = 127
		u   uint    = 1
		f   float64 = 3.14 // 小数默认 float64
		b   byte    = 'A'  // byte 就是 uint8
		r   rune    = '语'  // rune 就是 int32，表示一个 Unicode 码点
		ok  bool    = true
		str string  = "hello"
	)
	// 没有隐式转换，必须显式：int(f)、float64(i)
	sum := float64(i) + f
	_ = i8
	_ = u
	return map[string]any{"i": i, "sum": sum, "b": b, "r": r, "ok": ok, "str": str}
}

// 类型定义 vs 别名
type Celsius float64  // 新类型，和 float64 不能直接混算
type Kelvin = float64 // 别名，完全是同一个类型

func (c Celsius) String() string { // 可以给新类型加方法
	return fmt.Sprintf("%.1f°C", c)
}

// ============================================================================
// 4. 指针
// ============================================================================

func inc(n *int) {
	*n++ // 解引用后赋值
}

func pointers() int {
	x := 10
	p := &x // 取地址
	inc(p)
	inc(&x) // 也可以直接取地址传进去
	return x
}

// Go 没有指针运算。结构体用指针接收者才能改原值，见第 9 节。

// ============================================================================
// 5. 数组、切片、map
// ============================================================================

func collections() map[string]any {
	// 数组：长度是类型的一部分，[3]int 和 [4]int 不是同一类型。日常几乎都用切片。
	arr := [3]int{1, 2, 3}
	arr2 := [...]int{4, 5} // 编译器数长度

	// 切片：指向底层数组的视图（ptr, len, cap）
	s := []int{1, 2, 3}
	s = append(s, 4, 5)        // 可能换底层数组
	s = append(s, []int{6}...) // 展开另一个切片
	sub := s[1:4]              // 半开区间 [1, 4)
	head := s[:2]
	tail := s[2:]
	cloned := slices.Clone(s) // 独立拷贝，改 cloned 不影响 s
	cloned[0] = 99

	copied := make([]int, len(s))
	copy(copied, s)

	s = slices.Delete(s, 1, 2) // 删掉下标 1
	has := slices.Contains(s, 3)
	slices.Sort(s)

	// 预分配：知道大概长度时用 make，少搬迁
	buf := make([]int, 0, 8) // len=0, cap=8
	buf = append(buf, 1, 2)

	// map：无序。零值 nil，读返回零值；写入前必须 make 或字面量
	m := map[string]int{"a": 1}
	m["b"] = 2
	v, ok := m["a"] // 二值形式：有没有这个键
	_, missing := m["z"]
	delete(m, "b")
	clear(m) // 清空所有键（Go 1.21+）；对切片则是把元素置零、长度不变

	_ = arr
	_ = arr2
	_ = head
	_ = tail
	_ = copied
	_ = buf
	return map[string]any{
		"sub":     sub,
		"cloned0": cloned[0],
		"has3":    has,
		"v":       v,
		"ok":      ok,
		"missing": missing,
		"s":       s,
	}
}

// ============================================================================
// 6. 字符串、字节、rune
// ============================================================================

func stringsDemo() map[string]any {
	s := "Go语言"
	// string 不可变，本质是只读字节序列（UTF-8）
	nBytes := len(s)         // 字节数，不是字符数
	nRunes := len([]rune(s)) // 字符数（粗算）
	sub := s[0:2]            // 按字节切，切到汉字中间会乱；汉字用 rune 切片
	joined := strings.Join([]string{"a", "b"}, "-")
	has := strings.Contains(s, "语")

	bs := []byte("abc") // 可改
	bs[0] = 'A'
	back := string(bs)

	var b strings.Builder
	b.WriteString("hi")
	b.WriteByte('!')

	runes := make([]rune, 0, len(s))
	for _, r := range s { // range 字符串得到的是 rune，i 是字节下标
		runes = append(runes, r)
	}
	return map[string]any{
		"bytes": nBytes, "runes": nRunes, "sub": sub,
		"joined": joined, "has": has, "back": back, "builder": b.String(),
		"firstRune": string(runes[0]),
	}
}

// ============================================================================
// 7. 结构体、嵌入、标签
// ============================================================================

type Address struct {
	City string
}

type User struct {
	Name    string `json:"name"` // 结构体标签：encoding/json 等会读
	Age     int    `json:"age"`
	Address        // 匿名字段：嵌入，User 直接有 City
}

func structs() map[string]any {
	u := User{Name: "Ada", Age: 36, Address: Address{City: "London"}}
	u2 := User{"Bob", 20, Address{"Paris"}} // 按字段顺序，少用，加字段会坏
	u.City = "Cambridge"                    // 提升字段
	p := &User{Name: "Cam"}                 // 指针
	p.Age = 1                               // 自动解引用，不用写 (*p).Age
	return map[string]any{"city": u.City, "u2": u2.Name, "p": p.Name}
}

// ============================================================================
// 8. 函数
// ============================================================================

func add(a, b int) int { // 同类型参数可合并写
	return a + b
}

func div(a, b int) (int, error) { // 多返回值；error 放最后
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func split(n int) (x, y int) { // 命名返回值：return 时自动带上；别滥用
	x, y = n/2, n-n/2
	return
}

func sumAll(nums ...int) int { // 可变参数，函数里是切片
	s := 0
	for _, n := range nums {
		s += n
	}
	return s
}

func apply(n int, f func(int) int) int { // 函数是一等值
	return f(n)
}

func closures() int {
	n := 0
	inc := func() { n++ } // 闭包捕获外层变量
	inc()
	inc()
	return n
}

func deferNamed() (s string) {
	// defer 在函数返回前执行，后写的先跑（栈）
	defer func() { s += "!" }()
	s = "ok"
	return
}

// ============================================================================
// 9. 方法：值接收者 vs 指针接收者
// ============================================================================

type Counter struct{ N int }

func (c Counter) Value() int { // 值接收者：改的是副本
	return c.N
}

func (c *Counter) Inc() { // 指针接收者：能改原值；有任一指针方法，通常全用指针
	c.N++
}

func methods() int {
	c := &Counter{}
	c.Inc()
	c.Inc()
	return c.Value()
}

// ============================================================================
// 10. 接口（隐式实现：不用 implements 关键字）
// ============================================================================

type Speaker interface {
	Speak() string
}

type Dog struct{ Name string }

func (d Dog) Speak() string { return d.Name + ": woof" }

type Cat struct{ Name string }

func (c Cat) Speak() string { return c.Name + ": meow" }

func say(s Speaker) string {
	return s.Speak()
}

func interfaces() map[string]any {
	var s Speaker = Dog{"Fido"}
	// any 就是 interface{}：空接口，能装任何值
	var box any = 42
	box = "now a string"

	// 类型断言
	str, ok := box.(string)

	// 类型 switch
	kind := ""
	switch box.(type) {
	case int:
		kind = "int"
	case string:
		kind = "string"
	default:
		kind = "other"
	}

	// 接口值 = (动态类型, 动态值)。动态值是 nil 指针时，接口本身不是 nil
	var d *Dog
	var sp Speaker = d
	isNilIface := sp == nil // false，这是 Go 里最常见的坑之一

	return map[string]any{
		"dog": say(s), "cat": say(Cat{"Mimi"}),
		"str": str, "ok": ok, "kind": kind, "nilIface": isNilIface,
	}
}

// ============================================================================
// 11. 错误处理
// ============================================================================

var ErrNotFound = errors.New("not found")

func find(id int) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("id=%d: %w", id, ErrNotFound) // %w 包一层，便于 errors.Is
	}
	return "ok", nil
}

func errs() map[string]any {
	_, err := find(0)
	is := errors.Is(err, ErrNotFound)

	q, err := div(10, 2)
	if err != nil {
		return map[string]any{"divErr": err.Error()}
	}
	_, errZero := div(1, 0)
	return map[string]any{"isNotFound": is, "quot": q, "zero": errZero.Error()}
}

// panic 会炸掉当前 goroutine；recover 只能写在 defer 里。
// 日常错误用 error 返回值，不要拿 panic 当 try/catch。
func safe(fn func()) (recovered any) {
	defer func() { recovered = recover() }()
	fn()
	return
}

// ============================================================================
// 12. 控制流
// ============================================================================

func flow(x int, m map[string]int) string {
	// if 可以带短语句，变量只在 if/else 里可见
	if v, ok := m["a"]; ok && v > 0 {
		_ = v
	}

	switch x { // 自动 break，不会穿透
	case 0, 1:
		return "small"
	case 2:
		return "two"
	default:
		return "other"
	}
}

func flow2(score int) string {
	switch { // 没表达式 = switch true，用来替代一长串 if-else
	case score >= 90:
		return "A"
	case score >= 60:
		return "B"
	default:
		return "C"
	}
}

func loops() int {
	sum := 0
	for i := 0; i < 3; i++ { // 经典三段
		sum += i
	}
	n := 3
	for n > 0 { // 当 while 用
		n--
	}
	for range 2 { // Go 1.22+：range 整数，只要次数
		sum++
	}
	for i := range 3 { // i = 0, 1, 2
		sum += i
	}
	xs := []int{10, 20}
	for i, v := range xs {
		sum += i + v
	}
	for k, v := range map[string]int{"a": 1} { // map 遍历顺序是随机的
		sum += v + len(k)
	}
	for {
		break // 无限循环 + break
	}
	return sum
}

// ============================================================================
// 13. 并发：goroutine、channel、select
// ============================================================================

func conc() map[string]any {
	ch := make(chan int)    // 无缓冲：发送和接收必须同时准备好
	go func() { ch <- 7 }() // 新 goroutine
	got := <-ch

	buf := make(chan string, 2) // 有缓冲：没满就能发
	buf <- "a"
	buf <- "b"
	close(buf) // 关闭后还能读剩余值；再读就是零值 + ok=false
	var fromBuf []string
	for v := range buf { // range 通道，关了才结束
		fromBuf = append(fromBuf, v)
	}

	out := make(chan int, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	work := func(n int) {
		defer wg.Done()
		out <- n * n
	}
	go work(2)
	go work(3)
	wg.Wait()
	close(out)
	a, b := <-out, <-out

	// select：多个通道谁先就绪用谁；都没就绪且无 default 就等待
	ch1, ch2 := make(chan int, 1), make(chan int, 1)
	ch1 <- 1
	sel := 0
	select {
	case sel = <-ch1:
	case sel = <-ch2:
	default:
		sel = -1
	}

	return map[string]any{"got": got, "fromBuf": fromBuf, "sq": a + b, "sel": sel}
}

// 只发 / 只收：func produce(ch chan<- int)  func consume(ch <-chan int)
// 实际项目里还会用 context 做取消，那是标准库，不是新语法。

// ============================================================================
// 14. 泛型（Go 1.18+）
// ============================================================================

func Ptr[T any](v T) *T { // any 约束 = 任何类型
	return &v
}

func Keys[K comparable, V any](m map[K]V) []K { // comparable 才能做 map 键 / ==
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

type Pair[T any] struct {
	Left, Right T
}

func generics() map[string]any {
	p := Ptr(42)
	ks := Keys(map[string]int{"a": 1, "b": 2})
	slices.Sort(ks)
	pair := Pair[string]{"L", "R"}
	return map[string]any{"ptr": *p, "keys": ks, "pair": pair}
}

// ============================================================================
// 15. 内建函数
// ============================================================================

func builtins() map[string]any {
	s := []int{3, 1, 2}
	mn := min(s[0], s[1], s[2]) // Go 1.21+ 内建 min/max
	mx := max(1, 2, 3)
	_ = cap(s)
	_ = new(int) // 分配零值并返回指针；日常更常写 &T{} 或 make
	return map[string]any{"min": mn, "max": mx, "len": len(s)}
}

// ============================================================================
// main：把关键结果打出来
// ============================================================================

func main() {
	vars()

	fmt.Println("Go 语法示例运行结果：")
	printMap("types", types())
	fmt.Printf("  celsius: %s\n", Celsius(36.5))
	fmt.Printf("  pointers: %d\n", pointers())
	printMap("collections", collections())
	printMap("strings", stringsDemo())
	printMap("structs", structs())
	fmt.Printf("  add: %d\n", add(2, 3))
	x, y := split(9)
	fmt.Printf("  split: %d %d\n", x, y)
	fmt.Printf("  sumAll: %d\n", sumAll(1, 2, 3, 4))
	fmt.Printf("  apply: %d\n", apply(3, func(n int) int { return n * n }))
	fmt.Printf("  closures: %d\n", closures())
	fmt.Printf("  deferNamed: %s\n", deferNamed())
	fmt.Printf("  methods: %d\n", methods())
	printMap("interfaces", interfaces())
	printMap("errors", errs())
	fmt.Printf("  recover: %v\n", safe(func() { panic("boom") }))
	fmt.Printf("  flow: %s %s\n", flow(2, map[string]int{"a": 1}), flow2(95))
	fmt.Printf("  loops: %d\n", loops())
	printMap("conc", conc())
	printMap("generics", generics())
	printMap("builtins", builtins())
	_ = packageVar
	_ = StatusBusy
	_ = Pi
	_ = StatusOK
	_ = StatusFail
}

func printMap(title string, m map[string]any) {
	fmt.Printf("  %s: %v\n", title, m)
}

// ============================================================================
// 速查
// ============================================================================
//
//  短声明            x := 1
//  多返回值          v, err := f()     _, err := f()
//  可变参数          func f(xs ...T)    f(s...)
//  切片              s[i:j]  append  copy  slices.Clone/Delete/Contains
//  map 取值          v, ok := m[k]
//  结构体嵌入        type Dog struct { Animal }
//  指针方法          func (p *T) M()
//  接口              type I interface { M() }   隐式实现
//  类型断言          x.(T)    switch x.(type)
//  错误              fmt.Errorf("%w", err)    errors.Is
//  defer             返回前执行，后写先跑
//  循环              for / for cond / for range / for range n
//  并发              go f()    ch <- v    <-ch    select
//  泛型              func F[T any](v T) T
//  内建              min max clear len cap make append copy
//
//  没有：class、继承、三元 ?:、try/catch、隐式类型转换
