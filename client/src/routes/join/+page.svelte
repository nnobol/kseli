<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import ModalFormLayout from "$lib/features/modals/ModalFormLayout.svelte";
    import FloatingInputField from "$lib/components/fields/FloatingInputField.svelte";
    import ErrorAlert from "$lib/components/error-alert/ErrorAlert.svelte";
    import type { JoinRoomPayload, RoomErrorResponse } from "$lib/api/rooms";
    import { joinRoom } from "$lib/api/rooms";
    import { tokenStore } from "$lib/stores/tokenStore";

    let username: string = $state("");
    let errorMessage: string = $state("");
    let fieldErrors: Record<string, string> = $state({});
    let hasSubmitted: boolean = $state(false);
    let loading: boolean = $state(false);
    let buttonDisabled: boolean = $derived(
        Object.keys(fieldErrors).length > 0 || loading,
    );

    function validateUsername(): void {
        if (!hasSubmitted) return;

        if (!username) {
            fieldErrors.username = "Username is required.";
        } else if (/\s/.test(username)) {
            fieldErrors.username = "Username cannot contain spaces.";
        } else if (username.length < 3 || username.length > 15) {
            fieldErrors.username = "Username must be 3-15 characters long.";
        } else {
            delete fieldErrors.username;
        }
    }

    async function handleSubmit(event: Event) {
        event.preventDefault();

        hasSubmitted = true;

        validateUsername();

        if (Object.keys(fieldErrors).length > 0) {
            return;
        }

        const token = page.url.searchParams.get("invite");

        if (!token) {
            errorMessage =
                "Missing invite token. Please use a valid invite URL.";
            return;
        }

        loading = true;

        const payload: JoinRoomPayload = { username };

        try {
            const response = await joinRoom(token, payload);
            tokenStore.set(response.token);
            await goto(`/room/${response.roomId}`);
        } catch (err: any) {
            const error = err as RoomErrorResponse;

            errorMessage = error.errorMessage || "";
            fieldErrors = error.fieldErrors || {};
        } finally {
            loading = false;
        }
    }
</script>

{#snippet fields()}
    <FloatingInputField
        id="username"
        type="text"
        labelText="Username"
        disabled={loading}
        bind:value={username}
        fieldError={fieldErrors.username}
        onInput={() => validateUsername()}
    />
{/snippet}

<main>
    <ModalFormLayout
        headerTitle="JOIN A ROOM"
        buttonText="JOIN"
        {loading}
        {buttonDisabled}
        closeModal={null}
        {fields}
        onSubmit={handleSubmit}
    />
</main>

{#if errorMessage}
    <ErrorAlert {errorMessage} clearErrorMessage={() => (errorMessage = "")} />
{/if}

<style>
    main {
        display: flex;
        flex: 1 0;
        align-items: center;
        justify-content: center;
        z-index: 0;
    }

    main::before {
        content: "";
        position: absolute;
        inset: 0;
        background-image: url("/join-blob.png");
        background-repeat: no-repeat;
        background-position: center;
        background-size: contain;
        opacity: 0.5;
        z-index: -1;
        pointer-events: none;
    }

    @media (min-width: 1280px) {
        main::before {
            background-size: 1280px, auto;
        }
    }
</style>
