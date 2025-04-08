import { writable } from "svelte/store";
import type { IChatWebSocketClient } from "$lib/api/ws";
import { goto } from "$app/navigation";
import { errorStore } from "./errorStore";
import { useMocks } from "$lib/env";

interface Message {
    username: string;
    content: string;
}

interface Participant {
    id: number;
    username: string;
    role: number;
}

async function createChatWebSocketClient(token: string): Promise<IChatWebSocketClient> {
    let ChatWS;
    if (useMocks) {
        const wsModule = await import('$lib/api/mocks/ws');
        ChatWS = wsModule.ChatWebSocketClient;
    } else {
        const wsModule = await import('$lib/api/ws');
        ChatWS = wsModule.ChatWebSocketClient;
    }
    return new ChatWS(token);
}

let chatConnection: IChatWebSocketClient | null = null;
export const messages = writable<Message[]>([]);
export const participants = writable<Participant[]>([]);

export async function initChatSession(initialParticipants: Participant[], token: string) {
    participants.set(initialParticipants);

    if (!chatConnection) {
        chatConnection = await createChatWebSocketClient(token);
        chatConnection.onMessage((message) => {
            switch (message.type) {
                case "msg":
                    const msgData = message.data as Message;
                    messages.update((msgs) => {
                        msgs.push(msgData);
                        return msgs.slice();
                    });
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
            if (event.code === 1000) {
                errorStore.set(event.reason);
            } else if (event.code !== 1000) {
                errorStore.set("error")
            }
            endChatSession();
        })
    }
}

export function sendMessage(content: string) {
    if (chatConnection) {
        chatConnection.send(content);
    } else {
        throw new Error("Cannot send message, try refreshing.");
    }
}

export function endChatSession() {
    if (chatConnection) {
        chatConnection.close(1000, "leave");
        chatConnection = null;
        messages.set([]);
        participants.set([]);
        goto("/");
    }
}