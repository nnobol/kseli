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

let wsClient: ReturnType<typeof connectWebSocket> | null = null;
export const messages = writable<Message[]>([]);
export const participants = writable<Participant[]>([]);

export function initializeChatStore(initialParticipants: Participant[], token: string) {
    participants.set(initialParticipants);

    if (!wsClient) {
        wsClient = connectWebSocket(token);
        wsClient.onMessage((message) => {
            if (message.type === "msg") {
                const msgData = message.data as Message;
                messages.update((msgs) => [...msgs, msgData]);
            } else if (message.type === "join") {
                const userData = message.data as Participant;

                participants.update((users) => {
                    const alreadyExists = users.some((user) => user.id === userData.id);
                    return alreadyExists ? users : [...users, userData];
                });
            } else if (message.type === "leave") {
                const userData = message.data as { id: number };
                participants.update((users) => users.filter(user => user.id !== userData.id));
            }
        });
    }
}

export function sendMessage(content: string) {
    if (wsClient) {
        wsClient.send(content);
    } else {
        console.warn("WebSocket not initialized, cannot send message");
    }
}

export function disconnectChatStore() {
    if (wsClient) {
        wsClient.close();
        wsClient = null;
        messages.set([]);
        participants.set([]);
    }
}