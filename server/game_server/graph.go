package main

import (
	"fmt"
)

type NodeId int
type EdgeId int

func (id NodeId) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf("%d", id)), nil
}

type Graph struct {
	Nodes map[NodeId]Node
	Edges map[EdgeId]Edge
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[NodeId]Node),
		Edges: make(map[EdgeId]Edge),
	}
}

func (g *Graph) AddNode(node Node) bool {
	if _, exits := g.Nodes[node.Id]; exits {
		return false
	} else {
		g.Nodes[node.Id] = node
		return true
	}
	return false
}

func (g *Graph) GetNode(nodeId NodeId) Node {
	return g.Nodes[nodeId]
}

func (g *Graph) AddEdge(edge Edge) bool {
	if _, exits := g.Edges[edge.Id]; exits {
		return false
	} else {
		g.Edges[edge.Id] = edge
		return true
	}
	return false
}

func (g *Graph) String() string {
	return fmt.Sprintf("%+v\n %+v", g.Nodes, g.Edges)
}

type Node struct {
	Id NodeId
}

type Edge struct {
	Id   EdgeId
	Head NodeId
	Tail NodeId
}
