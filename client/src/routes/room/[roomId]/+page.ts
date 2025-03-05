import type { PageLoad } from './$types'
import { getRoom } from '$lib/api/rooms';
import { error } from '@sveltejs/kit';
import { tokenStore } from '$lib/stores/tokenStore';
import { get } from 'svelte/store';
import { getItemFromLocalStorage } from '$lib/api/utils';

export const load: PageLoad = async ({ params }) => {
    const roomId = params.roomId;
    let token = get(tokenStore);
    tokenStore.set(null);

    if (!token) {
        token = getItemFromLocalStorage("roomToken")
    }

    if (!token) {
        throw error(500, 'Something went wrong. Please try again.');
    }

    try {
        const roomDetails = await getRoom(roomId, token);
        return { roomDetails, token, roomId };
    } catch (err: any) {
        throw error(err.statusCode || 500, err.errorMessage || 'An unexpected error occurred.');
    }
};