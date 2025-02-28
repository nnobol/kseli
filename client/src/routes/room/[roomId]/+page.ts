import type { PageLoad } from './$types'
import { getRoom } from '$lib/api/rooms';
import type { GetRoomOkResponse } from '$lib/api/rooms';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
    try {
        const roomId = params.roomId;
        // const roomDetails = await getRoom(roomId);
        const roomDetails: GetRoomOkResponse = {
            "userRole": 1,
            "maxParticipants": 4,
            "participants": [
                {
                    "id": 1,
                    "username": "nika",
                    "role": 1
                },
                {
                    "id": 2,
                    "username": "lado",
                    "role": 2
                }
            ],
            "secretKey": "6xL4sVYuJW32UCtwpmJFrg"
        }
        return roomDetails;
    } catch (err: any) {
        throw error(err.statusCode || 500, err.errorMessage || 'An unexpected error occurred.');
    }
};