package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"tinyb/b"
)

func main() {
	input := flag.String("i", "", "输入 .b 文件")
	output := flag.String("o", "out.tvm", "输出文件")
	flag.Parse()

	if *input == "" {
		fmt.Println("TinyB - TinyVM B编译器 v0.1")
		fmt.Println("用法: tinyb -i input.b [-o out.tvm]")
		return
	}

	src, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}

	tokens := b.NewLexer(string(src)).Lex()
	ast := b.NewParser(tokens).Parse()
	gen := b.NewCodeGen()
	asm := gen.Generate(ast)

	outFile := *output
	if outFile == "" {
		base := strings.TrimSuffix(*input, ".b")
		outFile = base + ".asm"
	}

	if err := os.WriteFile(outFile, []byte(asm), 0644); err != nil {
		panic(err)
	}
	fmt.Printf("✅ 汇编已写入: %s\n", outFile)
}
