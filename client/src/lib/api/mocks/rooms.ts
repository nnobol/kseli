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
        roomId: '123',
        token: 'token',
    } as JoinRoomOkResponse;
}

export async function getRoom(roomId: string, token: string): Promise<GetRoomOkResponse> {
    await new Promise(resolve => setTimeout(resolve, 500));

    return {
        userRole: 1,
        maxParticipants: 5,
        participants: [
            { id: 1, username: "Nika", role: 1 },
            { id: 2, username: "Sigma", role: 2 },
            { id: 2, username: "Vano", role: 2 },
            { id: 2, username: "Otara", role: 2 },
            { id: 2, username: "Mako", role: 2 }
            
        ],
        expiresAt: Math.floor(Date.now() / 1000) + 1800,
        roomId: roomId,
        inviteLink: "http://localhost:5173/join?invite=123&sdfgeruiferuifhreipufhweriufrupfhwerufrweiupfhwerfheruipfhergfherpgf"
    } as GetRoomOkResponse
}