//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func main() {
	testDir := "./api"

	files, err := filepath.Glob(filepath.Join(testDir, "*_test.go"))
	if err != nil {
		fmt.Printf("❌ 读取测试文件失败: %v\n", err)
		os.Exit(1)
	}

	fset := token.NewFileSet()
	hasError := false

	for _, file := range files {
		_, err := parser.ParseFile(fset, file, nil, parser.AllErrors)
		if err != nil {
			fmt.Printf("❌ 语法错误 %s: %v\n", filepath.Base(file), err)
			hasError = true
		} else {
			fmt.Printf("✅ 语法正确 %s\n", filepath.Base(file))
		}
	}

	if hasError {
		os.Exit(1)
	}

	fmt.Println("\n✅ 所有测试文件语法检查通过！")
}
