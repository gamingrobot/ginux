/**
Modified typescript version of https://github.com/davidpiegza/Graph-Visualization
**/
import Edge = require("game/graph/Edge");

class Graph {
    private _nodeSet
    public nodes
    public edges
    
    constructor() {
        this._nodeSet = {};
        this.nodes = [];
        this.edges = []; 
    }
    public addNode(node):boolean{
        if(this._nodeSet[node.id] == undefined) {
            this._nodeSet[node.id] = node;
            this.nodes.push(node);
            return true;
        }
        return false;
    } 
    public getNode(node_id):any {
        return this._nodeSet[node_id];
    }
    public addEdge(edge_id, source, target):boolean{
        if(source.addConnectedTo(target) === true) {
            var edge = new Edge(edge_id, source, target);
            this.edges.push(edge);
            return true;
        }
        return false;
    }

}

export = Graph