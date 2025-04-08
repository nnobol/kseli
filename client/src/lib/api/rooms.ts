import { getItemFromLocalStorage, getItemFromSessionStorage, setItemInLocalStorage } from "./utils";
import { API_KEY, useMocks } from "$lib/env";

export interface RoomErrorResponse {
    statusCode?: number;
    errorMessage?: string;
    fieldErrors?: Record<string, string>;
}

export interface CreateRoomPayload {
    username: string;
    maxParticipants: number;
}

export interface CreateRoomOkResponse {
    roomId: string;
    token: string;
}

export async function createRoom(payload: CreateRoomPayload): Promise<CreateRoomOkResponse> {
    if (useMocks) {
        const { createRoom: createRoomMock } = await import('./mocks/rooms');
        return createRoomMock(payload);
    }

    try {
        const sessionId = getUserSessionId();

        const response = await fetch('/api/rooms', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY,
                'X-Participant-Session-Id': sessionId
            },
            body: JSON.stringify(payload),
        });

        const responseBody = await response.json();

        if (!response.ok) {
            throw responseBody as RoomErrorResponse;
        }

        return responseBody as CreateRoomOkResponse;
    } catch (err) {
        if (err instanceof Error) {
            throw {
                errorMessage: "An unexpected server error occurred. Please try again later.",
            } as RoomErrorResponse;
        }

        throw err as RoomErrorResponse;
    }
}

export interface JoinRoomPayload {
    username: string;
    roomSecretKey: string;
}

export interface JoinRoomOkResponse {
    token: string;
}

export async function joinRoom(roomId: string, payload: JoinRoomPayload): Promise<JoinRoomOkResponse> {
    if (useMocks) {
        const { joinRoom: joinRoomMock } = await import('./mocks/rooms');
        return joinRoomMock(payload);
    }

    try {
        const sessionId = getUserSessionId();

        const response = await fetch(`/api/rooms/${roomId}/join`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY,
                'X-Participant-Session-Id': sessionId
            },
            body: JSON.stringify(payload),
        });

        const responseBody = await response.json();

        if (!response.ok) {
            throw responseBody as RoomErrorResponse;
        }

        return responseBody as JoinRoomOkResponse;
    } catch (err) {
        if (err instanceof Error) {
            throw {
                errorMessage: "An unexpected server error occurred. Please try again later.",
            } as RoomErrorResponse;
        }

        throw err as RoomErrorResponse;
    }
}

export interface Participant {
    id: number;
    username: string;
    role: number;
}

export interface GetRoomOkResponse {
    userRole: number;
    maxParticipants: number;
    participants: Participant[];
    expiresAt: number;
    roomId: string;
    secretKey?: string;
}

export async function getRoom(roomId: string, token: string): Promise<GetRoomOkResponse> {
    if (useMocks) {
        const { getRoom: getRoomMock } = await import('./mocks/rooms');
        return getRoomMock(roomId, token);
    }

    try {
        const origin = window.location.origin;

        const response = await fetch(`/api/rooms/${roomId}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token,
                'X-Origin': origin
            }
        });

        const responseBody = await response.json();

        if (!response.ok) {
            throw {
                ...responseBody,
                statusCode: response.status,
            } as RoomErrorResponse;
        }

        return responseBody as GetRoomOkResponse;
    } catch (err) {
        if (err instanceof Error) {
            throw {
                errorMessage: "An unexpected server error occurred. Please try again later.",
                statusCode: 500,
            } as RoomErrorResponse;
        }

        throw err as RoomErrorResponse;
    }
}

export async function closeRoom(): Promise<void> {
    try {
        const token = getItemFromSessionStorage("token");
        const roomId = getItemFromSessionStorage("activeRoomId")

        const response = await fetch(`/api/rooms/${roomId!}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token!
            }
        });

        if (response.status === 204) {
            return;
        }

        const responseBody = await response.json();
        throw responseBody as RoomErrorResponse;
    } catch (err) {
        if (err instanceof Error) {
            throw {
                errorMessage: "An unexpected server error occurred. Please try again later.",
            } as RoomErrorResponse;
        }

        throw err as RoomErrorResponse;
    }
}

export interface UserPayload {
    userId: number;
}

export async function kickUser(payload: UserPayload): Promise<void> {
    try {
        const token = getItemFromSessionStorage("token");
        const roomId = getItemFromSessionStorage("activeRoomId")

        const response = await fetch(`/api/rooms/${roomId!}/kick`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token!
            },
            body: JSON.stringify(payload),
        });

        if (response.status === 204) {
            return;
        }

        const responseBody = await response.json();
        throw responseBody as RoomErrorResponse;
    } catch (err) {
        if (err instanceof Error) {
            throw {
                errorMessage: "An unexpected server error occurred. Please try again later.",
            } as RoomErrorResponse;
        }

        throw err as RoomErrorResponse;
    }
}

export async function banUser(payload: UserPayload): Promise<void> {
    try {
        const token = getItemFromSessionStorage("token");
        const roomId = getItemFromSessionStorage("activeRoomId")

        const response = await fetch(`/api/rooms/${roomId!}/ban`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token!
            },
            body: JSON.stringify(payload),
        });

        if (response.status === 204) {
            return;
        }

        const responseBody = await response.json();
        throw responseBody as RoomErrorResponse;
    } catch (err) {
        if (err instanceof Error) {
            throw {
                errorMessage: "An unexpected server error occurred. Please try again later.",
            } as RoomErrorResponse;
        }

        throw err as RoomErrorResponse;
    }
}

function getUserSessionId() {
    const storedSessionId = getItemFromLocalStorage("userSessionId");
    if (storedSessionId) return storedSessionId;

    const sessionId = generateUserSessionId();
    setItemInLocalStorage("userSessionId", sessionId, 720);
    return sessionId;
}

function generateUserSessionId() {
    const components = [
        navigator.userAgent,
        crypto.randomUUID(),
        new Date().getTime()
    ];

    return hashFNV32(components.join("."))
}

function hashFNV32(input: string) {
    let hash = 2166136261;
    for (let i = 0; i < input.length; i++) {
        hash ^= input.charCodeAt(i);
        hash += (hash << 1) + (hash << 4) + (hash << 7) + (hash << 8) + (hash << 24);
    }
    return (hash >>> 0).toString(16);
}