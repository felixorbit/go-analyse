package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var fileDir, files, outDir, ignoresVar string
	flag.StringVar(&fileDir, "dir", "./files", "文件目录")
	flag.StringVar(&files, "files", "", "文件名，用逗号分隔")
	flag.StringVar(&outDir, "out_dir", "./files", "保存目录")
	flag.StringVar(&ignoresVar, "ignores", "", "忽略名，用逗号分隔")
	flag.Parse()

	var ignores []string
	if len(ignoresVar) > 0 {
		ignores = strings.Split(ignoresVar, ",")
	}

	var err error
	cp := NewCodeParser(NewFunctionParser(ignores))
	if len(fileDir) > 0 {
		err = cp.ParseDirectory(fileDir)
	} else if len(files) > 0 {
		fileList := strings.Split(files, ",")
		err = cp.ParseFiles(fileList)
	} else {
		err = fmt.Errorf("文件为空")
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	res := cp.Result()
	fmt.Println(res.Mermaid())
	_ = os.WriteFile(filepath.Join(outDir, "flowchart.md"), []byte(res.Mermaid()), os.ModePerm)
}
