interface IWindow extends Window {
    requestAnimFrame(callback: any, element?: any): any;
    cancelRequestAnimFrame(callback: any, element?: any): any;
    container: any;
    DEBUG: boolean;
}

export = IWindow;