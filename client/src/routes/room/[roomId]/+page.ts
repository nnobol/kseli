import type { PageLoad } from './$types'
import { getRoom } from '$lib/api/rooms';
import { error, redirect } from '@sveltejs/kit';
import { tokenStore } from '$lib/stores/tokenStore';
import { errorStore } from '$lib/stores/errorStore';
import { get } from 'svelte/store';
import { getItemFromLocalStorage } from '$lib/api/utils';

export const load: PageLoad = async ({ params }) => {
    const roomId = params.roomId;
    let token = get(tokenStore);
    tokenStore.set(null);

    const activeRoomId = getItemFromLocalStorage("activeRoomId");
    if (activeRoomId && activeRoomId !== roomId) {
        errorStore.set("You are already in a chat room in another tab. Please leave that room first.");
        throw redirect(303, '/');
    }

    if (activeRoomId && activeRoomId === roomId) {
        // handle this differently - reconnect to the room
        errorStore.set("You're already in this room in another tab.");
        throw redirect(303, '/');
    }

    if (!token) {
        token = getItemFromLocalStorage("token");
    }

    if (!token) {
        throw error(403, "No token provided. Please join a room first.");
    }

    try {
        const roomDetails = await getRoom(roomId, token);
        return { roomDetails, token, roomId };
    } catch (err: any) {
        // clear the local storage
        throw error(err.statusCode || 500, err.errorMessage || 'An unexpected error occurred.');
    }
};