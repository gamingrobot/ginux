/**
Modified typescript version of https://github.com/davidpiegza/Graph-Visualization
**/

import Graph = require("game/graph/Graph");

class ForceDirected {

    public attraction_multiplier
    public repulsion_multiplier
    public max_iterations
    public graph: Graph
    public width
    public height
    public finished

    private _callback_positionUpdated
    private _EPSILON
    private _attraction_constant
    private _repulsion_constant
    private _forceConstant
    private _layout_iterations
    private _temperature
    private _nodes_length
    private _edges_length

    // performance test
    private _mean_time
    
    constructor(graph: Graph, options) {
        this.attraction_multiplier = options.attraction || 5;
        this.repulsion_multiplier = options.repulsion || 0.75;
        this.max_iterations = options.iterations || 1000;
        this.graph = graph;
        this.width = options.width || 200;
        this.height = options.height || 200;
        this.finished = false;

        this._callback_positionUpdated = options.positionUpdated;
        this._EPSILON = 0.000001;
        this._layout_iterations = 0;
        this._temperature = 0;
        this._mean_time = 0;
    }
    public init(){
        this.finished = false;
        this._temperature = this.width / 10.0;
        this._nodes_length = this.graph.nodes.length;
        this._edges_length = this.graph.edges.length;
        this._forceConstant = Math.sqrt(this.height * this.width / this._nodes_length);
        this._attraction_constant = this.attraction_multiplier * this._forceConstant;
        this._repulsion_constant = this.repulsion_multiplier * this._forceConstant;
    }
    public generate():boolean {
        if(this._layout_iterations < this.max_iterations && this._temperature > 0.000001) {
            var start = new Date().getTime();
          
            // calculate repulsion
            for(var i=0; i < this._nodes_length; i++) {
                var node_v = this.graph.nodes[i];
                node_v.layout = node_v.layout || {};
                if(i==0) {
                    node_v.layout.offset_x = 0;
                    node_v.layout.offset_y = 0;
                    node_v.layout.offset_z = 0;
                }

                node_v.layout.force = 0;
                node_v.layout.tmp_pos_x = node_v.layout.tmp_pos_x || node_v.position.x;
                node_v.layout.tmp_pos_y = node_v.layout.tmp_pos_y || node_v.position.y;
                node_v.layout.tmp_pos_z = node_v.layout.tmp_pos_z || node_v.position.z;

                for(var j=i+1; j < this._nodes_length; j++) {
                    var node_u = this.graph.nodes[j];
                    if(i != j) {
                        node_u.layout = node_u.layout || {};
                        node_u.layout.tmp_pos_x = node_u.layout.tmp_pos_x || node_u.position.x;
                        node_u.layout.tmp_pos_y = node_u.layout.tmp_pos_y || node_u.position.y;
                        node_u.layout.tmp_pos_z = node_u.layout.tmp_pos_z || node_u.position.z;

                        var delta_x = node_v.layout.tmp_pos_x - node_u.layout.tmp_pos_x;
                        var delta_y = node_v.layout.tmp_pos_y - node_u.layout.tmp_pos_y;
                        var delta_z = node_v.layout.tmp_pos_z - node_u.layout.tmp_pos_z;

                        var delta_length = Math.max(this._EPSILON, Math.sqrt((delta_x * delta_x) + (delta_y * delta_y)));
                        var delta_length_z = Math.max(this._EPSILON, Math.sqrt((delta_z * delta_z) + (delta_y * delta_y)));

                        var force = (this._repulsion_constant * this._repulsion_constant) / delta_length;
                        var force_z = (this._repulsion_constant * this._repulsion_constant) / delta_length_z;
            

                        node_v.layout.force += force;
                        node_u.layout.force += force;

                        node_v.layout.offset_x += (delta_x / delta_length) * force;
                        node_v.layout.offset_y += (delta_y / delta_length) * force;

                        if(i==0) {
                            node_u.layout.offset_x = 0;
                            node_u.layout.offset_y = 0;
                            node_u.layout.offset_z = 0;
                        }
                        node_u.layout.offset_x -= (delta_x / delta_length) * force;
                        node_u.layout.offset_y -= (delta_y / delta_length) * force;

                        node_v.layout.offset_z += (delta_z / delta_length_z) * force_z;
                        node_u.layout.offset_z -= (delta_z / delta_length_z) * force_z;
                    }
                }
            }
          
            // calculate attraction
            for(var i=0; i < this._edges_length; i++) {
                var edge = this.graph.edges[i];
                var delta_x = edge.source.layout.tmp_pos_x - edge.target.layout.tmp_pos_x;
                var delta_y = edge.source.layout.tmp_pos_y - edge.target.layout.tmp_pos_y;
                var delta_z = edge.source.layout.tmp_pos_z - edge.target.layout.tmp_pos_z;

                var delta_length = Math.max(this._EPSILON, Math.sqrt((delta_x * delta_x) + (delta_y * delta_y)));
                var delta_length_z = Math.max(this._EPSILON, Math.sqrt((delta_z * delta_z) + (delta_y * delta_y)));
                
                var force = (delta_length * delta_length) / this._attraction_constant;
                var force_z = (delta_length_z * delta_length_z) / this._attraction_constant;
                
                edge.source.layout.force -= force;
                edge.target.layout.force += force;

                edge.source.layout.offset_x -= (delta_x / delta_length) * force;
                edge.source.layout.offset_y -= (delta_y / delta_length) * force;
                edge.source.layout.offset_z -= (delta_z / delta_length_z) * force_z;
            
                edge.target.layout.offset_x += (delta_x / delta_length) * force;
                edge.target.layout.offset_y += (delta_y / delta_length) * force;
                edge.target.layout.offset_z += (delta_z / delta_length_z) * force_z;
                    
            }
          
            // calculate positions
            for(var i=0; i < this._nodes_length; i++) {
                var node = this.graph.nodes[i];
                var delta_length = Math.max(this._EPSILON, Math.sqrt(node.layout.offset_x * node.layout.offset_x + node.layout.offset_y * node.layout.offset_y));
                var delta_length_z = Math.max(this._EPSILON, Math.sqrt(node.layout.offset_z * node.layout.offset_z + node.layout.offset_y * node.layout.offset_y));

                node.layout.tmp_pos_x += (node.layout.offset_x / delta_length) * Math.min(delta_length, this._temperature);
                node.layout.tmp_pos_y += (node.layout.offset_y / delta_length) * Math.min(delta_length, this._temperature);
                node.layout.tmp_pos_z += (node.layout.offset_z / delta_length_z) * Math.min(delta_length_z, this._temperature);
            
                var updated = true;
                node.position.x -=  (node.position.x-node.layout.tmp_pos_x)/10;
                node.position.y -=  (node.position.y-node.layout.tmp_pos_y)/10;
                node.position.z -=  (node.position.z-node.layout.tmp_pos_z)/10;
            
                // execute callback function if positions has been updated
                if(updated && typeof this._callback_positionUpdated === 'function') {
                    this._callback_positionUpdated(node);
                }
            }
            this._temperature *= (1 - (this._layout_iterations / this.max_iterations));
            this._layout_iterations++;

            var end = new Date().getTime();
            this._mean_time += end - start;
        } else {
            if(!this.finished) {        
                console.log("Average time: " + (this._mean_time/this._layout_iterations) + " ms");
            }
            this.finished = true;
            return false;
        }
        return true;  
    }
    public stopCalculating() {
        this._layout_iterations = this.max_iterations;
    }

}

export = ForceDirected