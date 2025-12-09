package main

import (
	"fmt"
	"os"

	"github.com/cloudwego/thriftgo/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_fields_debug.go <thrift_file>")
		os.Exit(1)
	}

	thriftFile := os.Args[1]
	ast, err := parser.ParseFile(thriftFile, []string{}, true)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed file: %s\n", thriftFile)
	fmt.Printf("Number of structs: %d\n", len(ast.Structs))

	for _, s := range ast.Structs {
		fmt.Printf("\nStruct: %s\n", s.Name)
		fmt.Printf("  Annotations count: %d\n", len(s.Annotations))
		for _, anno := range s.Annotations {
			fmt.Printf("    Key: %s, Values: %v\n", anno.Key, anno.Values)
		}
		genFields := s.Annotations.Get("ts.gen_fields")
		if len(genFields) > 0 {
			fmt.Printf("  ✓ Found ts.gen_fields: %v\n", genFields)
		} else {
			fmt.Printf("  ✗ No ts.gen_fields annotation\n")
		}
		expandable := s.Annotations.Get("expandable")
		if len(expandable) > 0 {
			fmt.Printf("  ✓ Found expandable: %v\n", expandable)
		}
	}
}

