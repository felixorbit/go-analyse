package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Graph struct {
	Nodes   []*GraphNode
	Edges   []*GraphEdge
	edgeSet map[string]*GraphEdge
	nodeSet map[string]*GraphNode
}

type GraphNode struct {
	Name string
	Ins  int
	Outs int
}

type GraphEdge struct {
	From *GraphNode
	To   *GraphNode
}

func (g *Graph) AddNode(id string) *GraphNode {
	node := &GraphNode{Name: id}
	g.nodeSet[id] = node
	g.Nodes = append(g.Nodes, node)
	return node
}

func (g *Graph) AddEdge(fromID, toID string) {
	from, ok := g.nodeSet[fromID]
	if !ok {
		from = g.AddNode(fromID)
	}
	to, ok := g.nodeSet[toID]
	if !ok {
		to = g.AddNode(toID)
	}
	edge := &GraphEdge{from, to}
	edgeKey := edge.ID()
	if _, ok := g.edgeSet[edgeKey]; ok {
		return
	}
	g.edgeSet[edgeKey] = edge
	g.Edges = append(g.Edges, edge)
	from.Outs += 1
	to.Ins += 1
}

func (g *Graph) RemoveNode(id string) {
	delete(g.nodeSet, id)
	for i := range g.Nodes {
		if g.Nodes[i].Name == id {
			g.Nodes = append(g.Nodes[:i], g.Nodes[i:]...)
		}
	}
	for i := range g.Edges {
		edge := g.Edges[i]
		if edge.From.Name == id {
			delete(g.edgeSet, edge.ID())
			g.Edges = append(g.Edges[:i], g.Edges[i:]...)
			edge.To.Ins -= 1
		} else if edge.To.Name == id {
			delete(g.edgeSet, edge.ID())
			g.Edges = append(g.Edges[:i], g.Edges[i:]...)
			edge.From.Outs -= 1
		}
	}
}

func (fc GraphEdge) ID() string {
	return fmt.Sprintf("%s-%s", fc.From.Name, fc.To.Name)
}

func NewGraph() *Graph {
	return &Graph{
		edgeSet: make(map[string]*GraphEdge),
		nodeSet: map[string]*GraphNode{},
	}
}

type CodeParser struct {
	funcSet map[string]struct{}
	graph   *Graph
}

func NewCodeParser() *CodeParser {
	return &CodeParser{
		funcSet: make(map[string]struct{}),
		graph:   NewGraph(),
	}
}

// FunctionCall represents a function call with its caller and callee.
type FunctionCall struct {
	Caller string
	Callee string
}

// parseGoFile parses a Go file and returns the function calls within it.
func (cp *CodeParser) parseGoFile(filename string, fset *token.FileSet) error {
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl: // 函数声明
			caller := fn.Name.Name
			cp.funcSet[caller] = struct{}{}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr) // 函数调用
				if !ok {
					return true
				}
				switch inst := call.Fun.(type) {
				case *ast.Ident: // 直接调用函数
					callee := inst.Name
					cp.graph.AddEdge(caller, callee)
				case *ast.SelectorExpr: // 调用对象方法
					callee := inst.Sel.Name
					cp.graph.AddEdge(caller, callee)
					// 可以解析 inst.X 做进一步判断
				default:
					// fmt.Printf("%T %+v \n", inst, call.Fun)
				}
				return true
			})
		}
		return true
	})
	return nil
}

// parseDirectory parses all Go files in the specified directory and returns the function calls.
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

func (cp *CodeParser) GetFuncCalls() []FunctionCall {
	sort.Slice(cp.graph.Edges, func(i, j int) bool {
		if cp.graph.Edges[i].From.Ins == cp.graph.Edges[j].From.Ins {
			return cp.graph.Edges[i].From.Outs > cp.graph.Edges[j].From.Outs
		}
		return cp.graph.Edges[i].From.Ins < cp.graph.Edges[j].From.Ins
	})
	var calls []FunctionCall
	for _, edge := range cp.graph.Edges {
		calls = append(calls, FunctionCall{Caller: edge.From.Name, Callee: edge.To.Name})
	}
	// 只保留文件中声明的函数
	var validCalls []FunctionCall
	for _, call := range calls {
		if _, ok := cp.funcSet[call.Caller]; !ok {
			continue
		}
		if _, ok := cp.funcSet[call.Callee]; !ok {
			continue
		}
		validCalls = append(validCalls, call)
	}
	return validCalls
}

func main() {
	var fileDir, files, outDir string
	flag.StringVar(&fileDir, "dir", "", "文件目录")
	flag.StringVar(&files, "files", "", "文件名，用逗号分隔")
	flag.StringVar(&outDir, "out_dir", "./files", "保存目录")
	flag.Parse()

	var err error
	cp := NewCodeParser()
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
	calls := cp.GetFuncCalls()

	res := "```mermaid\n"
	res += "flowchart LR\n"
	for _, call := range calls {
		res += fmt.Sprintf("%s --> %s\n", call.Caller, call.Callee)
	}
	res += "```"
	fmt.Println(res)
	ioutil.WriteFile(filepath.Join(outDir, "flowchart.md"), []byte(res), os.ModePerm)
}
