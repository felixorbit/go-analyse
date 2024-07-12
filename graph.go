package main

import "fmt"

type Graph struct {
	Nodes   []*GraphNode
	Edges   []*GraphEdge
	edgeSet map[string]*GraphEdge
	nodeSet map[string]*GraphNode
}

type GraphNode struct {
	Key  string
	Ins  int
	Outs int
}

type GraphEdge struct {
	From *GraphNode
	To   *GraphNode
}

func (g *Graph) AddNode(key string) *GraphNode {
	node := &GraphNode{Key: key}
	g.nodeSet[key] = node
	g.Nodes = append(g.Nodes, node)
	return node
}

func (g *Graph) AddEdge(fromNodeKey, toNodeKey string) {
	from, ok := g.nodeSet[fromNodeKey]
	if !ok {
		from = g.AddNode(fromNodeKey)
	}
	to, ok := g.nodeSet[toNodeKey]
	if !ok {
		to = g.AddNode(toNodeKey)
	}
	edge := &GraphEdge{from, to}
	edgeKey := edge.Key()
	if _, ok := g.edgeSet[edgeKey]; ok {
		return
	}
	g.edgeSet[edgeKey] = edge
	g.Edges = append(g.Edges, edge)
	from.Outs += 1
	to.Ins += 1
}

func (g *Graph) RemoveNode(key string) {
	delete(g.nodeSet, key)
	for i := range g.Nodes {
		if g.Nodes[i].Key == key {
			g.Nodes = append(g.Nodes[:i], g.Nodes[i:]...)
		}
	}
	for i := range g.Edges {
		edge := g.Edges[i]
		if edge.From.Key == key {
			delete(g.edgeSet, edge.Key())
			g.Edges = append(g.Edges[:i], g.Edges[i:]...)
			edge.To.Ins -= 1
		} else if edge.To.Key == key {
			delete(g.edgeSet, edge.Key())
			g.Edges = append(g.Edges[:i], g.Edges[i:]...)
			edge.From.Outs -= 1
		}
	}
}

func (fc GraphEdge) Key() string {
	return fmt.Sprintf("%s-%s", fc.From.Key, fc.To.Key)
}

func NewGraph() *Graph {
	return &Graph{
		edgeSet: make(map[string]*GraphEdge),
		nodeSet: map[string]*GraphNode{},
	}
}
