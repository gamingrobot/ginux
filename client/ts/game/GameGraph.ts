import Graph = require("game/graph/Graph");

class GameGraph {

    private _graph: Graph = null;
    private _gotGraph = false;

    constructor() {
        this._graph = new Graph()
        $.getJSON( "/graph", this.loadGraph)
    }

    public loadGraph(data) {
        console.log(data)
    }

    public update():void {
    }
}

export = GameGraph;
