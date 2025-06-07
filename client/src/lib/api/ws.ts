export interface IChatWebSocketClient {
    onMessage(callback: WsMessageCallback): void;
    onClose(callback: WsCloseCallback): void;
    send(data: string): void;
    close(code?: number, reason?: string): void;
}

export interface WebSocketMessage {
    type: string;
    data: Record<string, any>;
}

export type WsMessageCallback = (data: WebSocketMessage) => void;
export type WsCloseCallback = (event: CloseEvent) => void;

export class ChatWebSocketClient implements IChatWebSocketClient {
    private ws: WebSocket;
    private messageListeners: WsMessageCallback[] = [];
    private closeListeners: WsCloseCallback[] = [];

    constructor(token: string) {
        const url = `/ws/room?token=${token}`;
        this.ws = new WebSocket(url);
        this.ws.binaryType = "arraybuffer";

        this.ws.onmessage = this.handleMessage.bind(this);
        this.ws.onclose = this.handleClose.bind(this);
    }

    private sendPong() {
        const pong = new Uint8Array([1]);
        this.ws.send(pong);
    }

    private handleMessage(event: MessageEvent) {
        if (typeof event.data === "string") {
            try {
                const parsedData: WebSocketMessage = JSON.parse(event.data);
                this.messageListeners.forEach((callback) => callback(parsedData));
            } catch (error) {
                console.error("Error parsing WebSocket message:", error);
            }
        } else {
            // binary data - means we received ping
            this.sendPong();
        }
    }

    private handleClose(event: CloseEvent) {
        this.closeListeners.forEach((callback) => callback(event));
    }

    public send(data: string): void {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(data);
        } else {
            throw new Error("Cannot send message, try refreshing.");
        }
    }

    public close(code?: number, reason?: string): void {
        this.ws.close(code, reason);
    }

    public onMessage(callback: WsMessageCallback): void {
        this.messageListeners.push(callback);
    }

    public onClose(callback: WsCloseCallback): void {
        this.closeListeners.push(callback);
    }
}