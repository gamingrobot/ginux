package main

import "fmt"

type NodeId int
type EdgeId int

type Graph struct {
	nodes map[NodeId]Node
	edges map[EdgeId]Edge
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[NodeId]Node),
		edges: make(map[EdgeId]Edge),
	}
}

func (g *Graph) AddNode(node Node) {
	g.nodes[node.Id] = node
}

func (g *Graph) GetNode(nodeId NodeId) Node {
	return g.nodes[nodeId]
}

func (g *Graph) AddEdge(edge Edge) {
	g.edges[edge.Id] = edge
}

func (g *Graph) String() string {
	return fmt.Sprintf("%+v\n %+v", g.nodes, g.edges)
}

type Node struct {
	Id NodeId
}

type Edge struct {
	Id   EdgeId
	Head NodeId
	Tail NodeId
}
