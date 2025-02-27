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
    try {
        const sessionId = getSessionId();
        const fingerprint = getFingerprint();

        const response = await fetch('/api/rooms', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Api-Key': import.meta.env.VITE_API_KEY,
                'X-Session-Id': sessionId,
                'X-Fingerprint': fingerprint
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
    try {
        const sessionId = getSessionId();
        const fingerprint = getFingerprint();

        const response = await fetch(`/api/rooms/${roomId}/join`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Api-Key': import.meta.env.VITE_API_KEY,
                'X-Session-Id': sessionId,
                'X-Fingerprint': fingerprint
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

export interface User {
    id: number;
    username: string;
    role: number;
}

export interface GetRoomOkResponse {
    userRole: number;
    maxParticipants: number;
    participants: User[];
    secretKey?: string;
}

export async function getRoom(roomId: string): Promise<GetRoomOkResponse> {
    try {
        const token = getTokenFromLocalStorage();
        const origin = window.location.origin;

        const response = await fetch(`/api/rooms/${roomId}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
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

export function setTokenInLocalStorage(token: string, hours: number) {
    const expiryTime = new Date().getTime() + hours * 60 * 60 * 1000;
    localStorage.setItem("roomToken", JSON.stringify({ value: token, expiry: expiryTime }));
}

export function getTokenFromLocalStorage(): string | null {
    const item = localStorage.getItem("roomToken");
    if (!item) return null;

    const { value, expiry } = JSON.parse(item);
    if (new Date().getTime() > expiry) {
        localStorage.removeItem("roomToken");
        return null;
    }
    return value;
}

function generateSessionId() {
    return crypto.randomUUID()
}

function generateFingerprint() {
    const components = [
        navigator.userAgent,
        navigator.hardwareConcurrency,
        navigator.maxTouchPoints,
        screen.width,
        screen.height,
        new Date().getTimezoneOffset()
    ];

    return hashFNV32(components.join("."));
}

function hashFNV32(input: string) {
    let hash = 2166136261;
    for (let i = 0; i < input.length; i++) {
        hash ^= input.charCodeAt(i);
        hash += (hash << 1) + (hash << 4) + (hash << 7) + (hash << 8) + (hash << 24);
    }
    return (hash >>> 0).toString(16);
}

function getFingerprint() {
    const storedFingerprint = getFromLocalStorage("fingerprint");
    if (storedFingerprint) return storedFingerprint;

    const fingerprint = generateFingerprint();
    setInLocalStorage("fingerprint", fingerprint, 30);
    return fingerprint;
}

function getSessionId() {
    const storedSessionId = getFromLocalStorage("sessionId");
    if (storedSessionId) return storedSessionId;

    const sessionId = generateSessionId();
    setInLocalStorage("sessionId", sessionId, 30);
    return sessionId;
}

function setInLocalStorage(key: string, value: string, days: number) {
    const expiryTime = new Date().getTime() + days * 24 * 60 * 60 * 1000;
    localStorage.setItem(key, JSON.stringify({ value, expiry: expiryTime }));
}

function getFromLocalStorage(key: string): string | null {
    const item = localStorage.getItem(key);
    if (!item) return null;

    const { value, expiry } = JSON.parse(item);
    if (new Date().getTime() > expiry) {
        localStorage.removeItem(key);
        return null;
    }
    return value;
}