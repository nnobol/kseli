import type { PageLoad } from './$types'
import { getRoom } from '../../../api/rooms';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
    try {
        const roomId = params.roomId;
        const roomDetails = await getRoom(roomId);
        return { roomDetails };
    } catch (err: any) {
        throw error(err.statusCode || 500, err.errorMessage || 'An unexpected error occurred.');
    }
};