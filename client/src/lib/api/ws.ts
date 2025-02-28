export interface WebSocketClient {
    send: (data: string) => void;
    onMessage: (callback: (data: string) => void) => void;
    close: () => void;
}

export function connectWebSocket(): WebSocketClient {
    // public echo server for testing
    const url = "wss://echo.websocket.org";
    const ws = new WebSocket(url);

    let messageCallback: ((data: string) => void) | null = null;

    ws.onopen = () => {
        console.log("Connected to WebSocket");
    };

    ws.onerror = (error) => {
        console.error("WebSocket error:", error);
    };

    ws.onclose = () => {
        console.log("WebSocket closed");
    };

    return {
        send: (data: string) => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(data);
            } else {
                console.warn("WebSocket not open, cannot send:", data);
            }
        },
        onMessage: (callback: (data: string) => void) => {
            messageCallback = callback;
            ws.onmessage = (event) => callback(event.data);
        },
        close: () => ws.close(),
    };
}