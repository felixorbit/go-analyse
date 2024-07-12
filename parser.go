package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type CodeParser struct {
	parser ElementParser
}

type CodeParseResult interface {
	Mermaid() string
}

func NewCodeParser(p ElementParser) *CodeParser {
	return &CodeParser{
		parser: p,
	}
}

func (cp *CodeParser) parseGoFile(filename string, fset *token.FileSet) error {
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}
	ast.Inspect(file, cp.parser.HandleFile)
	return nil
}

// ParseDirectory parses all Go files in the specified directory and returns the function calls.
func (cp *CodeParser) ParseDirectory(dir string) error {
	fset := token.NewFileSet()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			err := cp.parseGoFile(path, fset)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (cp *CodeParser) ParseFiles(files []string) error {
	fset := token.NewFileSet()
	for _, path := range files {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			err := cp.parseGoFile(path, fset)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cp *CodeParser) Result() CodeParseResult {
	return cp.parser.Result()
}

type ElementParser interface {
	HandleFile(n ast.Node) bool
	Result() CodeParseResult
}
