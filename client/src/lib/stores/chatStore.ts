import { writable } from "svelte/store";
import { ChatWebSocketClient } from "$lib/api/ws";

interface Message {
    username: string;
    content: string;
}

interface Participant {
    id: number;
    username: string;
    role: number;
}

let chatConnection: ChatWebSocketClient | null = null;
export const messages = writable<Message[]>([]);
export const participants = writable<Participant[]>([]);

export function initChatSession(initialParticipants: Participant[], token: string) {
    participants.set(initialParticipants);

    if (!chatConnection) {
        chatConnection = new ChatWebSocketClient(token)
        chatConnection.onMessage((message) => {
            switch (message.type) {
                case "msg":
                    const msgData = message.data as Message;
                    messages.update((msgs) => [...msgs, msgData]);
                    break;
                case "join":
                    const joinData = message.data as Participant;
                    participants.update((users) => {
                        const alreadyExists = users.some((user) => user.id === joinData.id);
                        return alreadyExists ? users : [...users, joinData];
                    });
                    break;
                case "leave":
                    const leaveData = message.data as { id: number };
                    participants.update((users) => users.filter(user => user.id !== leaveData.id));
                    break;
            }
        });
        chatConnection.onClose((event) => {
            console.log(event.code, event.reason)
        })
    }
}

export function sendMessage(content: string) {
    if (chatConnection) {
        chatConnection.send(content);
    } else {
        console.warn("WebSocket not initialized, cannot send message");
    }
}

export function endChatSession() {
    if (chatConnection) {
        chatConnection.close(1000, "leave");
        chatConnection = null;
        messages.set([]);
        participants.set([]);
    }
}