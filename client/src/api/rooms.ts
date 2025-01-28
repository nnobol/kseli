export interface CreateRoomPayload {
    username: string;
    maxParticipants: number;
}

export interface CreateRoomOkResponse {
    roomId: string;
    token: string;
}

export interface RoomErrorResponse {
    errorMessage: string;
    fieldErrors?: Record<string, string>;
}

export async function createRoom(payload: CreateRoomPayload): Promise<CreateRoomOkResponse> {
    try {
        const response = await fetch('/api/rooms', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-KEY': import.meta.env.VITE_API_KEY
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
        const response = await fetch(`/api/rooms/${roomId}/users`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-KEY': import.meta.env.VITE_API_KEY
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