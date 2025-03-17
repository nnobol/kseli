import { errorStore } from '$lib/stores/errorStore';
import { get } from 'svelte/store';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
    sessionStorage.clear();

    let errorMessage = "";
    let error = get(errorStore);
    errorStore.set(null);

    if (error === "kick") {
        errorMessage = "You have been kicked from the chat room."
    } else if (error === "ban") {
        errorMessage = "You have been banned from the chat room."
    } else if (error === "close-user") {
        errorMessage = "The admin has closed the chat room."
    } else if (error === "close") {
        errorMessage = "The chat room closed because the time ran out."
    } else if (error === "error") {
        errorMessage = "An unexpected server error occured."
    }

    return { errorMessage };
};