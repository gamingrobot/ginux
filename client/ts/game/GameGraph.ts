import Graph = require("game/graph/Graph");
import Node = require("game/graph/Node");
import ForceDirected = require("game/graph/ForceDirected");


class GameGraph {

    private _graph: Graph = null;
    private _gotGraph = false;
    private _geometries;
    private _layout;
    private _render = null;

    constructor(render) {
        this._graph = new Graph()
        this._geometries = [];
        this._render = render
        $.ajax({
            url : "/graph",
            dataType : 'json',
            context : this,
            success : this.loadGraph
        })
    }

    public loadGraph(data, status, jqXHR) {
        console.log(data);
        for (var key in data.Nodes) {
            var node = new Node(data.Nodes[key].Id);
            if(this._graph.addNode(node)){
                this.drawNode(node);
            }
        }
        for (var key in data.Edges) {
            var source = this._graph.getNode(data.Edges[key].Head);
            var target = this._graph.getNode(data.Edges[key].Tail);
            this._graph.addEdge(data.Edges[key].Id, source, target);
            this.drawEdge(source, target);
        }
        this._gotGraph = true;
        var layout_options = {width: 2000, height: 2000, iterations: 100000 }
        this._layout = new ForceDirected(this._graph, layout_options);
        this._layout.init();
    }

    public drawNode(node) {
        var draw_object = new THREE.Mesh( new THREE.CubeGeometry( 25, 25, 25 ), new THREE.MeshBasicMaterial( {  color: Math.random() * 0xffffff, opacity: 0.5 } ) );

        var area = 5000;
        draw_object.position.x = Math.floor(Math.random() * (area + area + 1) - area);
        draw_object.position.y = Math.floor(Math.random() * (area + area + 1) - area);
        draw_object.position.z = Math.floor(Math.random() * (area + area + 1) - area);

        draw_object.id = node.id;
        node.draw_object = draw_object;
        node.position = draw_object.position;
        this._render.scene.add( node.draw_object );
    }

    public drawEdge(source, target) {
        var material = new THREE.LineBasicMaterial({ color: 0xff0000, opacity: 1, linewidth: 0.5 });

        var tmp_geo = new THREE.Geometry();
        tmp_geo.vertices.push(source.draw_object.position);
        tmp_geo.vertices.push(target.draw_object.position);

        var line = new THREE.Line( tmp_geo, material, THREE.LinePieces );
        line.scale.x = line.scale.y = line.scale.z = 1;

        this._geometries.push(tmp_geo);

        this._render.scene.add( line );
    }

    public update():void {
        if(!this._gotGraph){
            return
        }
        // Generate layout if not finished
        if(!this._layout.finished) {
            this._layout.generate();
        }

        // Update position of lines (edges)
        for(var i=0; i<this._geometries.length; i++) {
            this._geometries[i].verticesNeedUpdate = true;
        }
    }  
}

export = GameGraph;
