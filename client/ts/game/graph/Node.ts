/**
Modified typescript version of https://github.com/davidpiegza/Graph-Visualization
**/
class Node {
    public id
    public position
    public draw_object
    private _nodesTo
    private _nodesFrom
    
    constructor(node_id) {
        this.id = node_id;
        this._nodesTo = [];
        this._nodesFrom = [];
        this.position = {}; 
    }
    public addConnectedTo(node):boolean {
        if(this.connectedTo(node) === false) {
            this._nodesTo.push(node);
            return true;
        }
        return false;
    }
    public connectedTo(node):boolean {
        for(var i=0; i < this._nodesTo.length; i++) {
            var connectedNode = this._nodesTo[i];
            if(connectedNode.id == node.id) {
                return true;
            }
        }
        return false;
    }

}

export = Node