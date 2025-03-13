export interface WebSocketMessage {
    type: string;
    data: Record<string, any>;
}

type WsMessageCallback = (data: WebSocketMessage) => void;
type WsEventCallback = () => void;
type WsErrorCallback = (error: Event) => void;
type WsCloseCallback = (event: CloseEvent) => void;

export class ChatWebSocketClient {
    private ws: WebSocket;
    private messageListeners: WsMessageCallback[] = [];
    private openListeners: WsEventCallback[] = [];
    private errorListeners: WsErrorCallback[] = [];
    private closeListeners: WsCloseCallback[] = [];

    constructor(token: string) {
        const url = `/ws/room?token=${token}`;
        this.ws = new WebSocket(url);
        this.ws.binaryType = "arraybuffer";

        this.ws.onopen = this.handleOpen.bind(this);
        this.ws.onmessage = this.handleMessage.bind(this);
        this.ws.onerror = this.handleError.bind(this);
        this.ws.onclose = this.handleClose.bind(this);
    }

    private handleOpen() {
        this.openListeners.forEach((callback) => callback());
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
            // means we received ping
            this.sendPong();
        }
    }

    private handleError(event: Event) {
        this.errorListeners.forEach((callback) => callback(event));
    }

    private handleClose(event: CloseEvent) {
        this.closeListeners.forEach((callback) => callback(event));
    }

    public send(data: string): void {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(data);
        } else {
            console.warn("WebSocket not open, cannot send:", data);
        }
    }

    public close(code?: number, reason?: string): void {
        this.ws.close(code, reason);
    }

    public onMessage(callback: WsMessageCallback): void {
        this.messageListeners.push(callback);
    }

    public onOpen(callback: WsEventCallback): void {
        this.openListeners.push(callback);
    }

    public onError(callback: WsErrorCallback): void {
        this.errorListeners.push(callback);
    }

    public onClose(callback: WsCloseCallback): void {
        this.closeListeners.push(callback);
    }
}