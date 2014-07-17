//tools
import RenderContainer = require("tools/RenderContainer");
import TickProvider = require("tools/TickProvider");
import IWindow = require("tools/IWindow");
declare var window: IWindow;

import Stats = require("stats");

class Ginux {

    private _tickProvider: TickProvider = null;
    private _renderContainer: RenderContainer = null;

    public initialise():void {
        var scene = new THREE.Scene();
        // CAMERA
        var SCREEN_WIDTH = window.innerWidth, SCREEN_HEIGHT = window.innerHeight;
        var VIEW_ANGLE = 45, ASPECT = SCREEN_WIDTH / SCREEN_HEIGHT, NEAR = 0.1, FAR = 200000;
        var camera = new THREE.PerspectiveCamera( VIEW_ANGLE, ASPECT, NEAR, FAR);
        var skyboxCamera = new THREE.PerspectiveCamera( VIEW_ANGLE, ASPECT, NEAR, FAR);
        scene.add(camera);
        camera.position.set(0,10,-100);
        camera.lookAt(scene.position);
        var renderer = new THREE.WebGLRenderer( {antialias:true} );
        renderer.setSize(SCREEN_WIDTH, SCREEN_HEIGHT);
        var DPR = (window.devicePixelRatio) ? window.devicePixelRatio : 1;
        renderer.setViewport( 0, 0, SCREEN_WIDTH*DPR, SCREEN_HEIGHT*DPR );
        renderer.setClearColorHex( 0x6495ED, 1 );
        window.container.appendChild( renderer.domElement );
        this._renderContainer =  new RenderContainer(renderer, scene, camera);
        // STATS
        var stats = new Stats();
        window.container.appendChild( stats.domElement );
        //CAMERA CONTROLS
        var camControls = new THREE.OrbitControls( camera, renderer.domElement, renderer.domElement );
        camControls.minDistance = 20;
        camControls.maxDistance = 1000;
        camControls.noKeys = true;

        // AXIS
        var axes = new THREE.AxisHelper(100);
        scene.add( axes );

        this._tickProvider = new TickProvider();
        this._tickProvider.add(stats.update, stats);
        this._tickProvider.add(camControls.update, camControls);
        this._tickProvider.add(this.render, this, 0); //render as the last step

    }

    public render():void {
        this._renderContainer.renderer.render( this._renderContainer.scene, this._renderContainer.camera );
    }

    public start():void {
        this._tickProvider.start();
    }
}

export = Ginux;
