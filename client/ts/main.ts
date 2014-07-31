import Detector = require("detector");
import Terminal = require("term");
import IWindow = require("tools/IWindow");
declare var window: IWindow;

import Ginux = require("game/Ginux");
window.DEBUG = false;

var AppStarted = false;
window.container = document.getElementById( 'threejs' );

class Main {

    start() {
        if (!AppStarted) {
            console.log("ginux started!");

            if(!Detector.webgl){
                Detector.addGetWebGLMessage(); 
                return;
            }

            var ginux = new Ginux();
            ginux.initialise();
            ginux.start();

            AppStarted = true;
        } else {
            console.log("ginux is singleton and is already started!");
        }
    }
}

window.requestAnimFrame = (function () {
    return window.requestAnimationFrame ||
        function (/* function */ callback, /* DOMElement */ element?) {
            return window.setTimeout(callback, 1000 / 60);
        };
})();

window.cancelRequestAnimFrame = (function () {
    return window.cancelAnimationFrame ||
        clearTimeout
})();

export = Main;