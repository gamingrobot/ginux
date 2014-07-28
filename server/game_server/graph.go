package main

import (
	"fmt"
	"strconv"
)

type NodeId int
type EdgeId int

type Graph struct {
	Nodes map[string]Node
	Edges map[string]Edge
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]Node),
		Edges: make(map[string]Edge),
	}
}

func (g *Graph) AddNode(node Node) bool {
	if _, exits := g.Nodes[strconv.Itoa(int(node.Id))]; exits {
		return false
	} else {
		g.Nodes[strconv.Itoa(int(node.Id))] = node
		return true
	}
	return false
}

func (g *Graph) GetNode(nodeId NodeId) Node {
	return g.Nodes[strconv.Itoa(int(nodeId))]
}

func (g *Graph) AddEdge(edge Edge) bool {
	if _, exits := g.Edges[strconv.Itoa(int(edge.Id))]; exits {
		return false
	} else {
		g.Edges[strconv.Itoa(int(edge.Id))] = edge
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
