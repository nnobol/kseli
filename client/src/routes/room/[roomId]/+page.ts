export const ssr = false;

import type { PageLoad } from './$types'
import { getRoom } from '$lib/api/rooms';
import { error } from '@sveltejs/kit';
import { tokenStore } from '$lib/stores/tokenStore';
import { keyStore } from '$lib/stores/keyStore';
import { get } from 'svelte/store';
import { getItemFromSessionStorage, setItemInSessionStorage } from '$lib/api/utils';
import { useMocks } from '$lib/env';

export const load: PageLoad = async ({ params }) => {
    const roomId = params.roomId;
    let token = get(tokenStore);
    let key = get(keyStore);
    tokenStore.set(null);
    keyStore.set(null);

    const sessionToken = getItemFromSessionStorage("token");
    const sessionKey = getItemFromSessionStorage("encryptionKey");
    const activeRoomId = getItemFromSessionStorage("activeRoomId");

    // If there are session items and stored + param room ids match, we assume this is a refresh
    if (sessionToken && sessionKey && activeRoomId === roomId) {
        // Same tab, we count this a page reload
        try {
            const roomDetails = await getRoom(activeRoomId, sessionToken);
            return { roomDetails, token: sessionToken };
        } catch (err: any) {
            throw error(err.statusCode || 500, err.errorMessage || 'An unexpected error occurred.');
        }
    }

    // Navigation to a different room from the same tab
    if (sessionToken && sessionKey && activeRoomId !== roomId) {
        throw error(500, "You tried navigating to a different chat room which closed the room.");
    }

    // If no token from tokenStore, we assume this is a fresh room creation or join
    if (!token) {
        if (useMocks) {
            token = "mock-token";
        } else {
            throw error(403, "You need to join or create a room.");
        }
    }
    if (!key) {
        throw error(403, "You need to join or create a room.");
    }

    try {
        const roomDetails = await getRoom(roomId, token);
        setItemInSessionStorage("token", token, 0.5);
        setItemInSessionStorage("activeRoomId", roomId, 0.5);
        setItemInSessionStorage("encryptionKey", key, 0.5);
        return { roomDetails, token };
    } catch (err: any) {
        throw error(err.statusCode || 500, err.errorMessage || 'An unexpected error occurred.');
    }
};