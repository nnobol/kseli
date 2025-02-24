import type { PageLoad } from './$types'
import { getRoom } from '../../../api/rooms';
import type { GetRoomOkResponse, RoomErrorResponse } from '../../../api/rooms';

interface RoomPageData {
    roomDetails: GetRoomOkResponse;
    error: null;
}

interface RoomPageError {
    roomDetails: null;
    error: RoomErrorResponse
}

export type LoadReturn = RoomPageData | RoomPageError;

export const load: PageLoad = async ({ params }) => {
    const roomId = params.roomId;
    try {
        const roomDetails = await getRoom(roomId);
        return { roomDetails, error: null } as RoomPageData;
    } catch (err) {
        const error = err as RoomErrorResponse;
        return {
            roomDetails: null,
            error: {
                errorMessage: error.errorMessage || "Failed to load room details",
                statusCode: error.statusCode || 500,
            },
        } as RoomPageError;;
    }
};