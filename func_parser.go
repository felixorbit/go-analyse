package main

import (
	"fmt"
	"go/ast"
	"sort"
)

type FunctionParser struct {
	funcSet  map[string]struct{}
	graph    *Graph
	ignoreFn []string
}

func NewFunctionParser(ignoreFns []string) *FunctionParser {
	return &FunctionParser{
		funcSet:  make(map[string]struct{}),
		graph:    NewGraph(),
		ignoreFn: ignoreFns,
	}
}

// FunctionCall represents a function call with its caller and callee.
type FunctionCall struct {
	Caller string
	Callee string
}

// HandleFile parses a Go file.
func (cp *FunctionParser) HandleFile(n ast.Node) bool {
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
}

func (cp *FunctionParser) GetFuncCalls() []FunctionCall {
	sort.Slice(cp.graph.Edges, func(i, j int) bool {
		if cp.graph.Edges[i].From.Ins == cp.graph.Edges[j].From.Ins {
			return cp.graph.Edges[i].From.Outs > cp.graph.Edges[j].From.Outs
		}
		return cp.graph.Edges[i].From.Ins < cp.graph.Edges[j].From.Ins
	})
	var calls []FunctionCall
	for _, edge := range cp.graph.Edges {
		calls = append(calls, FunctionCall{Caller: edge.From.Key, Callee: edge.To.Key})
	}
	ignoreFnSet := make(map[string]struct{})
	for _, fn := range cp.ignoreFn {
		ignoreFnSet[fn] = struct{}{}
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
		if _, ok := ignoreFnSet[call.Caller]; ok {
			continue
		}
		if _, ok := ignoreFnSet[call.Callee]; ok {
			continue
		}
		validCalls = append(validCalls, call)
	}
	return validCalls
}

func (cp *FunctionParser) Result() CodeParseResult {
	return &FunctionParseResult{
		calls: cp.GetFuncCalls(),
	}
}

type FunctionParseResult struct {
	calls []FunctionCall
}

func (fpr *FunctionParseResult) Mermaid() string {
	res := "```mermaid\n"
	res += "flowchart LR\n"
	for _, call := range fpr.calls {
		res += fmt.Sprintf("%s --> %s\n", call.Caller, call.Callee)
	}
	res += "```"
	return res
}
