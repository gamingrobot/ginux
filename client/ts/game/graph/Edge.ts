/**
Modified typescript version of https://github.com/davidpiegza/Graph-Visualization
**/

class Edge {
    public id
    public source
    public target
    
    constructor(edge_id, source, target) {
        this.id = edge_id;
        this.source = source;
        this.target = target;
    }
}

export = Edge