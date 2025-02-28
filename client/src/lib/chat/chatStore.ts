import { writable } from "svelte/store";
import { connectWebSocket } from "$lib/api/ws";

interface Message {
    username: string;
    content: string;
}

interface Participant {
    id: number;
    username: string;
    role: number;
}

// Store state
let wsClient: ReturnType<typeof connectWebSocket> | null = null;
export const messages = writable<Message[]>([]);
export const participants = writable<Participant[]>([
    { id: 1, username: "nika", role: 1 }, // Dummy data for testing
    { id: 2, username: "lado", role: 2 },
]);

// Initialize store and connect WebSocket
export function initializeChatStore() {
    if (!wsClient) {
        wsClient = connectWebSocket();
        wsClient.onMessage((data) => {
            // PieSocket echoes raw text, so we'll fake a message format for testing
            messages.update((msgs) => [...msgs, { username: "test-user", content: data }]);
        });
    }
}

// Send a chat message
export function sendMessage(content: string) {
    if (wsClient) {
        wsClient.send(content); // PieSocket echoes raw strings
    } else {
        console.warn("WebSocket not initialized, cannot send message");
    }
}

// Cleanup (optional)
export function disconnectChatStore() {
    if (wsClient) {
        wsClient.close();
        wsClient = null;
        messages.set([]);
        participants.set([]);
    }
}