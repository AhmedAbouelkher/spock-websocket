class EnhancedWebSocket {
    constructor(url, options = {}) {
        this.url = url;
        this.nativeSocket = new WebSocket(url);
        this.pingInterval = null;
        this.pingPongInterval = (options.pingPongInterval || 30) * 1000

        this._setupPingPong();
    }

    _setupPingPong() {
        this.nativeSocket.onopen = (event) => {
            this._startPingPong();
            if (this.onopen) this.onopen(event);
        };

        this.nativeSocket.onmessage = (event) => {
            if (event.data === 'ping') {
                this.send('pong');
                return;
            }
            if (event.data === 'pong') {
                return;
            }
            if (this.onmessage) this.onmessage(event);
        };

        this.nativeSocket.onclose = (event) => {
            this._stopPingPong();
            if (this.onclose) this.onclose(event);
        };
    }

    _startPingPong() {
        this.pingInterval = setInterval(() => {
            if (this.nativeSocket.readyState === WebSocket.OPEN) {                        
                this.send('ping');
            }
        }, this.pingPongInterval);
    }

    _stopPingPong() {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
        }
    }

    send(data) {
        this.nativeSocket.send(data);
    }

    close(code, reason) {
        this._stopPingPong();
        this.nativeSocket.close(code, reason);
    }

    get readyState() {
        return this.nativeSocket.readyState;
    }
}