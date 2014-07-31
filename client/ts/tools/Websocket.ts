import IWindow = require("tools/IWindow");
declare var window: IWindow;

import Signal = require("tools/Signal");

class Websocket {

        signal = new Signal();
        socket = null;

        constructor() {
        }

        public connect() {
            var server = "ws://" + document.location.host + "/ws";
            this.socket = new WebSocket(server);
            this.socket.onmessage = this._onmessage;
            this.socket.onclose = this._onclose;
            this.socket.onerror = this._onerror;
        }

        public disconnect() {
            this.socket.close();
        }

        public add(listener, context, priority: number = 0) {
            this.signal.add(listener, context, priority);
        }

        public remove(listener, context) {
            this.signal.remove(listener, context);
        }

        public send(data) {
            this.socket.send(data);
        }

        private _onmessage = (data) => {
            this.signal.dispatch(data);
        }

        private _onclose = () => {
            $('#error').empty().append("Lost Connection to WebSocket");
            $('#error').show();
        }

        private _onerror = (err) => {
            $('#error').empty().append("Error: " + err);
            $('#error').show();
        }

}

export = Websocket;
