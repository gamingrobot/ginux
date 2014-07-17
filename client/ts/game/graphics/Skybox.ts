import RenderContainer = require("tools/RenderContainer");

//Could be a view but because it creates a scene it needs to be rendered in order
class Skybox {

    private _render: RenderContainer = null;
    private _scene: THREE.Scene = null;
    private _camera: THREE.PerspectiveCamera = null;

    constructor(render: RenderContainer, camera: THREE.PerspectiveCamera){
        this._render = render;
        this._camera = camera;
        this._scene = new THREE.Scene();
        this._scene.add( camera );
        // SKYBOX
        var imagePrefix = "assets/skybox/purple-nebula-complex/1024/";
        var directions  = ["right1", "left2", "top3", "bottom4", "front5", "back6"];
        var imageSuffix = ".png";
        var skyGeometry = new THREE.BoxGeometry( 1000, 1000, 1000 );   

        var materialArray = [];
        for (var i = 0; i < 6; i++)
            materialArray.push( new THREE.MeshBasicMaterial({
                //TODO: Use LoadingManager
                color: 0xFF00FF,
                map: THREE.ImageUtils.loadTexture( imagePrefix + directions[i] + imageSuffix ),
                side: THREE.BackSide,
                depthWrite: false
            }));
        var skyMaterial = new THREE.MeshFaceMaterial( materialArray );
        var object = new THREE.Mesh( skyGeometry, skyMaterial ); // skybox
        this._scene.add( object );
    }

    public update():void {
        this._camera.rotation.setFromRotationMatrix( new THREE.Matrix4().extractRotation( this._render.camera.matrixWorld ), this._camera.rotation.order);
        this._render.renderer.render(this._scene, this._camera);
    }
}

export = Skybox;
