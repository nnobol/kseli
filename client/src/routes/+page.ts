export const ssr = false;

import { errorStore } from '$lib/stores/errorStore';
import { get } from 'svelte/store';
import type { PageLoad } from './$types';
import { getItemFromSessionStorage } from "$lib/api/utils.js";

export const load: PageLoad = async () => {
    const isSameSession =
        !!getItemFromSessionStorage("activeRoomId") ||
        !!getItemFromSessionStorage("token") ||
        !!getItemFromSessionStorage("encryptionKey");

    if (isSameSession) {
        sessionStorage.clear();
        return { errorMessage: "You navigated to the home page from the chat room. The chat room is now closed." }
    }

    let errorReason = get(errorStore);
    errorStore.set(null);

    const errorMessageMap: Record<string, string> = {
        "kick": "You have been kicked from the chat room.",
        "ban": "You have been banned from the chat room.",
        "close-admin": "",
        "close-user": "The admin has closed the chat room.",
        "close": "The chat room closed because the time ran out.",
        "token-missing": "Authentication failed: token was missing.",
        "token-invalid": "Authentication failed: token was invalid or expired.",
        "room-not-exists": "Unexpected error: this chat room does not exist.",
        "user-not-exists": "Unexpected error: you were not found in this room.",
        "message-too-large": "Unexpected error: the size of the message you sent was too large.",
        "error": "An unexpected server error occurred.",
        "invalid-session": "Unexpected error: encryption key is missing, please try again."
    };

    return {
        errorMessage: errorReason ? errorMessageMap[errorReason] ?? "Unknown error occurred." : ""
    };
};