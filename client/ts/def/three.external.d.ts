declare module THREE {
    export class OrbitControls {
        object: any;
        domElement: any;
        localElement: any;
        enabled: boolean;
        target: THREE.Vector3;
        noZoom: boolean;
        zoomSpeed: number;
        minDistance: number;
        maxDistance: number;
        noRotate: boolean;
        rotateSpeed: number;
        noPan: boolean;
        keyPanSpeed: number;
        autoRotate: boolean;
        autoRotateSpeed: number;
        minPolarAngle: number;
        maxPolarAngle: number;
        noKeys: boolean;
        keys: {
            LEFT: number;
            UP: number;
            RIGHT: number;
            BOTTOM: number;
        };
        constructor(object: any, domElement: any, localElement: any);
        rotateLeft(angle: any): void;
        rotateUp(angle: any): void;
        panLeft(distance: any): void;
        panUp(distance: any): void;
        pan(delta: any): void;
        dollyIn(dollyScale: any): void;
        dollyOut(dollyScale: any): void;
        update(): void;
        getAutoRotationAngle(): number;
        getZoomScale(): number;
    }
}