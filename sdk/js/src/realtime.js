export class RealtimeService {
    constructor(baseUrl) {
        this.baseUrl = baseUrl.replace('http', 'ws');
        this.socket = null;
        this.callbacks = {};
        this.state = {};
        this.sendInterval = 50; // 50ms throttling
        this.pendingInputs = [];
        this.heartbeatInterval = null;
    }

    connect(token, lobbyId) {
        return new Promise((resolve, reject) => {
            const url = `${this.baseUrl}/ws?token=${token}&game_id=${lobbyId}`;
            this.socket = new WebSocket(url);

            this.socket.onopen = () => {
                console.log('Connected to Godra Server');
                this.startHeartbeat();
                this.startInputLoop();
                resolve();
            };

            this.socket.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                this.handleMessage(msg);
            };

            this.socket.onerror = (err) => reject(err);
            this.socket.onclose = () => this.stopHeartbeat();
        });
    }

    on(event, callback) {
        this.callbacks[event] = callback;
    }

    handleMessage(msg) {
        if (msg.type === 'batch') {
            msg.events.forEach(e => this.handleSingleEvent(e));
        } else {
            this.handleSingleEvent(msg);
        }
    }

    handleSingleEvent(event) {
        if (this.callbacks[event.type]) {
            this.callbacks[event.type](event.payload);
        }
        // Also trigger generic 'state_update'
        if (this.callbacks['state_update']) {
            this.callbacks['state_update'](event);
        }
    }

    sendInput(action, payload) {
        // Just push to pending, loop handles sending
        this.pendingInputs.push({ action, payload });
    }

    startInputLoop() {
        setInterval(() => {
            if (this.pendingInputs.length > 0 && this.socket.readyState === WebSocket.OPEN) {
                // Batching not fully implemented on client send yet in this snippet, 
                // but we send the latest or all. 
                // Simple version: send one by one or batch if server supports it.
                // Assuming server accepts raw JSON frames.
                const inputs = [...this.pendingInputs];
                this.pendingInputs = [];

                inputs.forEach(input => {
                    this.socket.send(JSON.stringify(input));
                });
            }
        }, this.sendInterval);
    }

    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            if (this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({ action: "heartbeat" }));
            }
        }, 5000);
    }

    stopHeartbeat() {
        if (this.heartbeatInterval) clearInterval(this.heartbeatInterval);
    }
}
