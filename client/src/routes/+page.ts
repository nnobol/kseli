import { errorStore } from '$lib/stores/errorStore';
import { get } from 'svelte/store';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
    sessionStorage.clear();

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
        "error": "An unexpected server error occurred.",
    };

    return {
        errorMessage: errorReason ? errorMessageMap[errorReason] ?? "Unknown error occurred." : ""
    };
};