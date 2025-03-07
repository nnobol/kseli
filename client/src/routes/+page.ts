import type { PageLoad } from './$types';
import { getItemFromLocalStorage } from "$lib/api/utils";
import { onRoomBroadcast } from "$lib/broadcast/broadcast";
import { errorStore } from '$lib/stores/errorStore';

export const load: PageLoad = async ({ url }) => {
    const activeRoomId = getItemFromLocalStorage("activeRoomId");

    // Set up BroadcastChannel listener
    onRoomBroadcast((roomId, error) => {
        if (roomId && url.pathname !== `/room/${roomId}`) {
            errorStore.set("You are already in a chat room in another tab. Please leave that room first.");
        } else if (error === null) {
            errorStore.set(null);
        }
    });

    // If in a room and not on that room's page, throw error
    if (activeRoomId && url.pathname !== `/room/${activeRoomId}`) {
        errorStore.set("You are already in a chat room in another tab. Please leave that room first.");
        return {};
    }

    return {};
};