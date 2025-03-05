export interface WebSocketClient {
    send: (data: string) => void;
    onMessage: (callback: (data: WebSocketMessage) => void) => void;
    close: () => void;
}

export interface WebSocketMessage {
    type: string;
    data: Record<string, any>;
}

export function connectWebSocket(token: string): WebSocketClient {
    const url = `/ws/room?token=${token}`;
    const ws = new WebSocket(url);

    let messageCallback: ((data: WebSocketMessage) => void) | null = null;

    ws.onopen = () => {
        console.log("Connected to WebSocket, readyState:", ws.readyState);
    };

    ws.onerror = (error) => {
        console.error("WebSocket error:", error, "readyState:", ws.readyState);
    };

    ws.onclose = (event) => {
        console.log("WebSocket close event details:", {
            code: event.code,
            reason: event.reason,
            wasClean: event.wasClean,
            type: event.type,
            timeStamp: event.timeStamp
        });
    };

    return {
        send: (data: string) => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(data);
            } else {
                console.warn("WebSocket not open, cannot send:", data);
            }
        },
        onMessage: (callback: (data: WebSocketMessage) => void) => {
            messageCallback = callback;
            ws.onmessage = (event) => {
                try {
                    const parsedData: WebSocketMessage = JSON.parse(event.data);
                    if (messageCallback) messageCallback(parsedData);
                } catch (error) {
                    console.error("Error parsing WebSocket message:", error);
                }
            };
        },
        close: () => ws.close(),
    };
}