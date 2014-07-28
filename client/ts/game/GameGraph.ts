import Graph = require("game/graph/Graph");
import Node = require("game/graph/Node");

class GameGraph {

    private _graph: Graph = null;
    private _gotGraph = false;

    constructor() {
        this._graph = new Graph()
        $.ajax({
            url : "/graph",
            dataType : 'json',
            context : this,
            success : this.loadGraph
        })
    }

    public loadGraph(data, status, jqXHR) {
        console.log(data)
        for (var key in data.Nodes) {
            this._graph.addNode(new Node(data.Nodes[key].Id))
        }
        for (var key in data.Edges) {
            this._graph.addEdge(data.Edges[key].Id, this._graph.getNode(data.Edges[key].Head),this._graph.getNode(data.Edges[key].Tail))
        }
    }

    public update():void {
    }
}

export = GameGraph;
