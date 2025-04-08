import type { CreateRoomPayload, CreateRoomOkResponse, JoinRoomPayload, JoinRoomOkResponse, GetRoomOkResponse } from "../rooms";

export async function createRoom(payload: CreateRoomPayload): Promise<CreateRoomOkResponse> {
    await new Promise(resolve => setTimeout(resolve, 500));

    return {
        roomId: 'LJsVVlFK',
        token: 'token'
    } as CreateRoomOkResponse;
}

export async function joinRoom(payload: JoinRoomPayload): Promise<JoinRoomOkResponse> {
    await new Promise(resolve => setTimeout(resolve, 500));

    return {
        token: 'token'
    } as JoinRoomOkResponse;
}

export async function getRoom(roomId: string, token: string): Promise<GetRoomOkResponse> {
    await new Promise(resolve => setTimeout(resolve, 500));

    return {
        userRole: 1,
        maxParticipants: 3,
        participants: [
            { id: 1, username: "Nika", role: 1 },
            { id: 2, username: "Lado", role: 2 }
        ],
        expiresAt: Math.floor(Date.now() / 1000) + 1800,
        roomId: roomId,
        secretKey: "Awf5NzKL6Y6Wug"
    } as GetRoomOkResponse
}