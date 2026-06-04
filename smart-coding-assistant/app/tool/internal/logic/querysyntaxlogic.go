package logic

import (
	"context"
	"fmt"
	"strings"

	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"
)

type QuerySyntaxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuerySyntaxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuerySyntaxLogic {
	return &QuerySyntaxLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *QuerySyntaxLogic) QuerySyntax(in *pb.QuerySyntaxRequest) (*pb.QuerySyntaxResponse, error) {
	lang := in.GetLanguage()
	query := in.GetQuery()
	ctx := in.GetContext()

	explanation, example := explainSyntax(lang, query, ctx)
	return &pb.QuerySyntaxResponse{
		Explanation: explanation,
		Example:     example,
		Success:     true,
	}, nil
}

// syntaxDB 常见编程概念的解释和示例
var syntaxDB = map[string]struct {
	explanation string
	example     func(lang string) string
}{
	"async": {"async/await 是异步编程语法糖，允许以同步风格编写异步代码。async 声明一个异步函数，返回 Promise/Future；await 暂停执行直到 Promise 完成。",
		func(l string) string { return "async function fetchData() {\n  const res = await fetch('/api/data');\n  return res.json();\n}" }},
	"await": {"await 用于等待一个 Promise/Future 完成并获取其结果。只能在 async 函数内使用。",
		func(l string) string { return "const data = await fetchData();" }},
	"goroutine": {"goroutine 是 Go 语言中的轻量级协程，由 Go 运行时调度。使用 go 关键字启动，比线程更轻量。",
		func(l string) string { return "go func() {\n  fmt.Println(\"Hello from goroutine\")\n}()" }},
	"channel": {"channel 是 Go 语言中用于 goroutine 之间通信的管道。支持同步和缓冲两种模式。",
		func(l string) string { return "ch := make(chan int)\ngo func() { ch <- 42 }()\nval := <-ch" }},
	"decorator": {"装饰器是一种设计模式/语法，用于在不修改原函数的情况下扩展其功能。Python 使用 @ 语法。",
		func(l string) string { return "@timer\ndef my_func():\n    pass\n\ndef timer(func):\n    def wrapper(*args, **kwargs):\n        start = time.time()\n        result = func(*args, **kwargs)\n        print(f'{func.__name__} took {time.time()-start}s')\n        return result\n    return wrapper" }},
	"closure": {"闭包是一个函数及其引用环境的组合——函数可以访问在其外部作用域中定义的变量。",
		func(l string) string { return "function makeCounter() {\n  let count = 0;\n  return function() { return ++count; };\n}\nconst counter = makeCounter();\ncounter(); // 1\ncounter(); // 2" }},
	"generator": {"生成器是可以暂停和恢复执行的函数。Python 中用 yield，JS 中用 function*。",
		func(l string) string { return "def fibonacci():\n    a, b = 0, 1\n    while True:\n        yield a\n        a, b = b, a + b" }},
	"interface": {"接口定义了一组方法签名，任何实现了这些方法的类型都隐式满足该接口。Go 接口是隐式实现的。",
		func(l string) string { return "type Reader interface {\n  Read(p []byte) (n int, err error)\n}\n\ntype File struct { ... }\nfunc (f *File) Read(p []byte) (n int, err error) {\n  return f.fd.Read(p)\n}" }},
	"pointer": {"指针存储了另一个变量的内存地址。通过指针可以直接操作原始数据而不拷贝。",
		func(l string) string { return "func increment(x *int) {\n  *x++  // 解引用并修改\n}\nn := 5\nincrement(&n)  // n = 6" }},
	"inheritance": {"继承是 OOP 中一个类可以派生自另一个类的机制。Go 使用组合代替继承，Python/JS 使用类继承。",
		func(l string) string { return "class Animal:\n    def speak(self):\n        return '...'\n\nclass Dog(Animal):\n    def speak(self):\n        return 'Woof!'" }},
	"polymorphism": {"多态允许不同类型的对象通过统一接口调用，各类型表现不同行为。",
		func(l string) string { return "func MakeSound(a Animal) {\n  fmt.Println(a.Speak())\n}\nMakeSound(Dog{})  // Woof!\nMakeSound(Cat{})  // Meow!" }},
	"recursion": {"递归是函数调用自身的编程技术，用于将问题分解为更小的同类子问题。",
		func(l string) string { return "func factorial(n int) int {\n  if n <= 1 { return 1 }\n  return n * factorial(n-1)\n}" }},
}

func explainSyntax(lang, query, context string) (string, string) {
	queryLower := strings.ToLower(query)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s 中的 %s\n\n", lang, query))

	if context != "" {
		sb.WriteString(fmt.Sprintf("**场景**: %s\n\n", context))
	}

	// 查找匹配的语法概念
	if entry, ok := syntaxDB[queryLower]; ok {
		sb.WriteString(entry.explanation)
		sb.WriteString("\n")
		return sb.String(), entry.example(lang)
	}

	// 部分匹配
	for key, entry := range syntaxDB {
		if strings.Contains(queryLower, key) || strings.Contains(key, queryLower) {
			sb.WriteString(fmt.Sprintf("**相关概念: %s**\n\n", key))
			sb.WriteString(entry.explanation)
			return sb.String(), entry.example(lang)
		}
	}

	// 未找到精确匹配，给通用解释
	sb.WriteString(fmt.Sprintf("%s 是 %s 中的语法概念。建议查阅官方文档或使用搜索引擎获取详细信息。\n", query, lang))
	sb.WriteString(fmt.Sprintf("\n**通用示例**:\n```%s\n// %s 在 %s 中的基本用法\n// 请参考 %s 官方文档了解更多\n```\n", lang, query, lang, lang))

	return sb.String(), fmt.Sprintf("// %s 概念示例 - 请参考 %s 文档\n", query, lang)
}
