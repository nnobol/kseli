import type { PageLoad } from './$types'
import { getRoom } from '$lib/api/rooms';
import { error } from '@sveltejs/kit';
import { tokenStore } from '$lib/stores/tokenStore';
import { get } from 'svelte/store';
import { getItemFromSessionStorage, setItemInSessionStorage } from '$lib/api/utils';

export const load: PageLoad = async ({ params }) => {
    const roomId = params.roomId;
    let token = get(tokenStore);
    tokenStore.set(null);

    const sessionToken = getItemFromSessionStorage("token");
    const activeRoomId = getItemFromSessionStorage("activeRoomId");

    // If there is are session items and stored and param room ids match, we assume this is a refresh
    if (sessionToken && (activeRoomId === roomId)) {
        // Same tab, we count this a page reload
        try {
            const roomDetails = await getRoom(activeRoomId, sessionToken);
            return { roomDetails, token: sessionToken, roomId: activeRoomId };
        } catch (err: any) {
            throw error(err.statusCode || 500, err.errorMessage || 'An unexpected error occurred.');
        }
    }

    // Navigation to a different room from the same tab
    if (activeRoomId && (activeRoomId !== roomId)) {
        throw error(500, "You tried navigating to a different chat room which closed the room.");
    }

    // If no token from tokenStore, we assume this is a fresh room creation or join
    if (!token) {
        throw error(403, "You need to join or create a room.");
    }

    try {
        const roomDetails = await getRoom(roomId, token);
        setItemInSessionStorage("token", token, 1);
        setItemInSessionStorage("activeRoomId", roomId, 1);
        return { roomDetails, token, roomId };
    } catch (err: any) {
        throw error(err.statusCode || 500, err.errorMessage || 'An unexpected error occurred.');
    }
};