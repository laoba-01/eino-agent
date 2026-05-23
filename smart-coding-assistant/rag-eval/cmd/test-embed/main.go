package main

import (
	"fmt"
	"rag-eval/embedding"
)

func main() {
	vec, err := embedding.Embed("Go语言数组和切片的区别")
	if err != nil {
		fmt.Printf("Embedding 失败: %v\n", err)
		return
	}
	fmt.Printf("Embedding 成功! 维度: %d\n", len(vec))
	fmt.Printf("前5个值: %v\n", vec[:5])
}
