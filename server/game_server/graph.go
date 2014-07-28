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

func (g *Graph) AddNode(node Node) bool{
	if _, exits := g.nodes[node.Id]; exits {
		return false
	} else {
		g.nodes[node.Id] = node
		return true
	}
	return false
}

func (g *Graph) GetNode(nodeId NodeId) Node {
	return g.nodes[nodeId]
}

func (g *Graph) AddEdge(edge Edge) bool{
	if _, exits := g.edges[edge.Id]; exits {
		return false
	} else {
		g.edges[edge.Id] = edge
		return true
	}
	return false
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
