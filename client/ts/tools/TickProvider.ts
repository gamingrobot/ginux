import IWindow = require("tools/IWindow");
declare var window: IWindow;

import Signal = require("tools/Signal");

// Module
class TickProvider {

        previousTime = 0;
        ticked = new Signal();
        request = null;

        constructor() {
        }

        start() {
            this.previousTime = Date.now();
            this.request = window.requestAnimFrame(this.tick.bind(this));
        }

        stop() {
            window.cancelRequestAnimFrame(this.request);
        }

        add(listener, context, priority: number = 0) {
            this.ticked.add(listener, context, priority);
        }

        remove(listener, context) {
            this.ticked.remove(listener, context);
        }

        tick(timestamp) {
            timestamp = timestamp || Date.now();
            var tmp = this.previousTime;
            this.previousTime = timestamp;
            var delta = (timestamp - tmp) * 0.001;
            this.ticked.dispatch(delta);
            requestAnimationFrame(this.tick.bind(this));
        }

}

export = TickProvider;
