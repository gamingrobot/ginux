class RenderContainer {
    public renderer: THREE.WebGLRenderer;
    public scene: THREE.Scene;
    public camera: THREE.PerspectiveCamera;

    constructor(renderer: THREE.WebGLRenderer, scene: THREE.Scene, camera: THREE.PerspectiveCamera) {
        this.renderer = renderer;
        this.scene = scene;
        this.camera = camera;
    }
}

export = RenderContainer;