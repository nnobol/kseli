export interface CreateRoomPayload {
    username: string;
    maxParticipants: number;
}

export interface CreateRoomOkResponse {
    roomId: string;
}

export interface CreateRoomErrorResponse {
    errorMessage: string;
    fieldErrors?: Record<string, string>;
}

export async function createRoom(payload: CreateRoomPayload): Promise<CreateRoomOkResponse> {
    try {
        const response = await fetch('/api/room', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });

        const responseBody = await response.json();

        if (!response.ok) {
            throw responseBody as CreateRoomErrorResponse;
        }

        return responseBody as CreateRoomOkResponse;
    } catch (err) {
        if (err instanceof Error) {
            throw {
                errorMessage: "An unexpected server error occurred. Please try again later.",
            } as CreateRoomErrorResponse;
        }

        throw err as CreateRoomErrorResponse;
    }
}

export async function joinRoom(): Promise<void> {
    console.log("joining room...")
}