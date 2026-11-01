# Go 泛型速查

> 适用范围：泛型自 Go 1.18 引入；本项目 `go.mod` 声明 `go 1.27.2`，因此可使用截至 Go 1.27 的能力。
> 权威定义见 [Go 语言规范 · Type parameter declarations](https://go.dev/ref/spec#Type_parameter_declarations)。

## 1. 核心概念

| 术语 | 含义 |
| --- | --- |
| 类型参数（type parameter） | 声明中待定的类型占位符，如 `func F[T any]` 中的 `T` |
| 类型实参（type argument） | 实例化时传入的具体类型，如 `F[int]` 中的 `int` |
| 约束（constraint） | 规定类型参数可取哪些类型（通常是接口） |
| 类型集（type set） | 约束所描述的合法类型集合 |
| 实例化（instantiation） | 用类型实参替换类型参数，得到非泛型声明 |

## 2. 基本语法

### 泛型函数

```go
func Map[T, U any](in []T, f func(T) U) []U {
	out := make([]U, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

nums := Map([]int{1, 2, 3}, func(v int) string { return strconv.Itoa(v) })
```

### 泛型类型

```go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

s := &Stack[int]{}
```

泛型类型必须实例化后才能作为变量、字段等使用；`Stack` 本身不是类型。

### 实例化与类型推断

```go
func Equal[T comparable](a, b T) bool { return a == b }

Equal(1, 2)        // 推断 T = int
Equal("a", "b")    // 推断 T = string
// Equal(1, "a")   // 编译错误：无法统一 T

func Zero[T any]() T { var z T; return z }

// z := Zero()     // 编译错误：无法从返回值推断类型
z := Zero[int]()   // 必须显式实例化
```

推断只看函数实参，**不会**根据返回值或上下文推断；可以只显式指定部分类型实参，其余交给推断。

## 3. 约束

### `any` 与 `comparable`

- `any` 是 `interface{}` 的别名，类型集为所有类型。
- `comparable` 表示可用 `==`、`!=` 比较的类型，**不**保证支持 `<`、`>`。
- Go 1.20 起，普通接口类型也可满足 `comparable`（运行期比较可能 panic）。

### 类型集、并集与 `~`

```go
interface{ int }       // 仅 int
interface{ ~int }      // 底层类型为 int 的所有类型（含 type MyInt int）
interface{ int|string } // int 或 string
interface{ int; string } // 空类型集：没有类型同时是二者
```

`~T` 表示“底层类型为 T”，是实现自定义数值类型约束的关键：

```go
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func Sum[T Number](xs []T) T {
	var s T
	for _, x := range xs {
		s += x
	}
	return s
}

type Celsius float64
Sum([]Celsius{1, 2}) // OK：Celsius 的底层类型是 float64
```

### 自定义约束与内置 `cmp.Ordered`

```go
import "cmp"

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
```

`cmp.Ordered` 覆盖所有有序内置类型（含字符串），需要 `> < <= >=` 时优先使用它，而不是 `comparable`。

### 含类型集或方法的接口只能作约束

```go
type Number interface{ ~int | ~float64 }

var n Number        // 编译错误：含类型并集的接口不能作为值类型
var x any = n       // 同上
```

普通接口（如 `any`、`io.Reader`）仍可正常作为值类型。

## 4. 标准库泛型设施（Go 1.21+）

```go
import (
	"cmp"
	"maps"
	"slices"
)

nums := []int{3, 1, 2}
slices.Sort(nums)                    // 原地排序
slices.SortFunc(people, func(a, b Person) int {
	return cmp.Compare(a.Age, b.Age) // 自定义比较
})
idx := slices.Index(nums, 2)         // 未找到返回 -1
copyNums := slices.Clone(nums)

m := map[string]int{"a": 1}
m2 := maps.Clone(m)
keys := slices.Sorted(maps.Keys(m)) // Go 1.23+：maps.Keys 返回 iter.Seq[string]

lo, hi := min(1, 2), max(1, 2)       // Go 1.21 内置函数
_ = lo
_ = hi
_ = m2
_ = idx
_ = copyNums
```

- `slices`：排序、查找、比较、克隆、`Collect`/`Sorted` 等。
- `maps`：`Clone`、`Equal`、`Keys`、`Values` 等；Go 1.23 起 `Keys`/`Values` 返回 `iter.Seq`。
- `cmp`：`cmp.Compare`、`cmp.Less` 及 `cmp.Ordered` 约束。

## 5. 常见模式

### Filter / Reduce

```go
func Filter[T any](in []T, keep func(T) bool) []T {
	var out []T
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func Reduce[T, A any](in []T, init A, f func(A, T) A) A {
	acc := init
	for _, v := range in {
		acc = f(acc, v)
	}
	return acc
}
```

### 用指针类型满足值方法约束

```go
// P 被约束为 *T，从而能对值类型调用指针接收者方法。
type Stringer[T any] interface {
	*T
	String() string
}

func PrintAll[T any, P Stringer[T]](items []P) {
	for _, it := range items {
		fmt.Println(it.String())
	}
}
```

### 安全地做类型判断

不能直接对类型参数做类型断言或 type switch，需先转成 `any`：

```go
func KindOf[T any](v T) string {
	switch any(v).(type) {
	case int, int64:
		return "integer"
	case string:
		return "string"
	default:
		return "other"
	}
}
```

## 6. 限制与坑

- **运算符**：约束里没有 `~int` 等底层类型信息时，不能对类型参数使用 `+`、`<` 等运算符。
- **`comparable` 不提供排序**：需要 `<` 时用 `cmp.Ordered`。
- **零值**：用 `var zero T`；不要假设 `nil`（`T` 可能是非指针类型）。
- **不能直接类型断言/type switch 类型参数**：先 `any(v)`。
- **不支持特化/重载**：不能为不同具体类型提供不同实现；需要时用接口、函数值或代码生成。
- **无 const 泛型、可变类型参数包、高阶类型**：类型参数只能是类型，不能是常量或“类型的类型”。
- **性能**：实现采用 GC shape stenciling + dictionaries，并非完整单态化；热路径注意接口装箱与字典开销。
- **约束过度**：约束越宽，能用的操作越少；越窄，复用性越差。按实际用到的操作取最小集合。
- **泛型方法（版本相关）**：Go 1.27 之前，方法不能声明接收者之外的类型参数；Go 1.27 起语言规范允许泛型方法，例如：

```go
type T int

func (T) M[P T]() {}
```

  官方规范在“实例化”条目中已把 method 与 function、type 并列并标注 Go 1.27；如需在项目中使用，请以 `go 1.27+` 为准并补充测试。

## 7. 版本演进（与泛型相关）

| 版本 | 变化 |
| --- | --- |
| 1.18 | 引入类型参数、类型集、`~T`、`any`、`comparable` |
| 1.20 | 普通接口类型可满足 `comparable`（运行期比较可能 panic） |
| 1.21 | 新增 `slices`、`maps`、`cmp` 包及内置 `min`/`max`/`clear` |
| 1.23 | 新增 `iter`；`maps.Keys`/`Values` 返回 `iter.Seq`，配合 `slices.Collect`/`Sorted` |
| 1.24 | 完整支持泛型类型别名 |
| 1.27 | 语言规范支持泛型方法（方法可声明自己的类型参数） |

## 8. 最佳实践

1. 先用具体类型；出现两处以上重复且逻辑一致时再抽象为泛型。
2. 能用标准库 `slices`/`maps`/`cmp` 就不自己写。
3. 约束按需最小化，自定义数值类型记得加 `~`。
4. 需要类型推断时优先用泛型函数而非泛型方法。
5. 为泛型代码补充多类型实参的测试（至少一个内置类型 + 一个命名类型）。

## 9. 参考

- [Go 语言规范](https://go.dev/ref/spec)
- [Go 1.18 泛型设计文档](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md)
- [Go 1.21 Release Notes](https://go.dev/doc/go1.21)
- [Go 1.23 Release Notes](https://go.dev/doc/go1.23)
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24)
- [slices 包文档](https://pkg.go.dev/slices)
- [maps 包文档](https://pkg.go.dev/maps)
- [cmp 包文档](https://pkg.go.dev/cmp)
