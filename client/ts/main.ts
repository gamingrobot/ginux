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

//each character is 10px tall * 1.2 line-height, 7px wide
console.log(Math.floor(window.innerHeight / 12));
console.log(window.innerHeight);
var term = new Terminal({
    cols: 70,
    rows: Math.floor(window.innerHeight / 12),
    screenKeys: true
});
term.open(document.getElementById("term"));
var server = "ws://" + document.location.host + "/ws";
var websocket = new WebSocket(server);
websocket.onopen = function () {
    websocket.onmessage = function (msg) {
        term.write(msg.data);
    };

    term.on('data', function (data) {
        websocket.send(data);
    });
};
websocket.onclose = function () {
    console.log(term.x, term.y);
    $('#error').empty().append("Lost Connection to WebSocket");
    $('#error').show();
};
websocket.onerror = function (err) {
    $('#error').empty().append("Error: " + err);
};

export = Main;