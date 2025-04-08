import type { IChatWebSocketClient, WsMessageCallback, WsCloseCallback, WebSocketMessage } from "../ws";

export class ChatWebSocketClient implements IChatWebSocketClient {
    private messageListeners: WsMessageCallback[] = [];
    private closeListeners: WsCloseCallback[] = [];
    private token: string;

    constructor(token: string) {
        this.token = token;
        setTimeout(() => {
            this.simulateMessage({ type: 'msg', data: { username: 'system', content: 'Mock connection established.' } });
        }, 200);
    }

    public send(data: string): void {
        setTimeout(() => {
            this.simulateMessage({ type: 'msg', data: { username: 'mock-server', content: data } });
        }, 300);
    }

    public close(code?: number, reason?: string): void {
        const event = new CloseEvent("close", {
            code: code || 1000,
            reason: reason || '',
            wasClean: true
        });
        this.closeListeners.forEach(callback => callback(event));
    }

    public onMessage(callback: WsMessageCallback): void {
        this.messageListeners.push(callback);
    }

    public onClose(callback: WsCloseCallback): void {
        this.closeListeners.push(callback);
    }

    private simulateMessage(message: WebSocketMessage) {
        this.messageListeners.forEach(callback => callback(message));
    }
}