<script lang="ts">
    import ModalWrapper from "./ModalWrapper.svelte";
    import ModalFormLayout from "./ModalFormLayout.svelte";
    import FloatingInputField from "../fields/FloatingInputField.svelte";
    import ErrorAlert from "../alerts/ErrorAlert.svelte";
    import { joinRoom } from "../../../api/rooms";
    import type {
        JoinRoomPayload,
        RoomErrorResponse,
    } from "../../../api/rooms";

    interface Props {
        closeModal: () => void;
    }

    let { closeModal }: Props = $props();

    let username: string = $state("");
    let roomId: string = $state("");
    let roomSecretKey: string = $state("");

    let errorMessage: string = $state("");
    let fieldErrors: Record<string, string> = $state({});
    let hasSubmitted: boolean = $state(false);

    let loading: boolean = $state(false);
    let buttonDisabled: boolean = $derived(
        Object.keys(fieldErrors).length > 0 || loading,
    );

    function clearErrorMessage(): void {
        errorMessage = "";
    }

    function validateAllFields(): void {
        if (!hasSubmitted) return;

        validateUsername();
        validateRoomId();
        validateRoomSecretKey();
    }

    function validateUsername(): void {
        if (!hasSubmitted) return;

        if (!username) {
            fieldErrors.username = "Username is required.";
        } else if (/\s/.test(username)) {
            fieldErrors.username = "Username cannot contain spaces.";
        } else if (username.length < 5 || username.length > 20) {
            fieldErrors.username = "Username must be 5-20 characters long.";
        } else {
            delete fieldErrors.username;
        }
    }

    function validateRoomId(): void {
        if (!hasSubmitted) return;

        if (!roomId) {
            fieldErrors.roomId = "Chat Room Id is required.";
        } else if (/\s/.test(roomId)) {
            fieldErrors.roomId = "Chat Room Id cannot contain spaces.";
        } else {
            delete fieldErrors.roomId;
        }
    }

    function validateRoomSecretKey(): void {
        if (!hasSubmitted) return;

        if (!roomSecretKey) {
            fieldErrors.roomSecretKey = "Chat Room Secret Key is required.";
        } else if (/\s/.test(roomSecretKey)) {
            fieldErrors.roomSecretKey =
                "Chat Room Secret Key cannot contain spaces.";
        } else {
            delete fieldErrors.roomSecretKey;
        }
    }

    async function handleSubmit(event: Event) {
        event.preventDefault();

        hasSubmitted = true;

        validateAllFields();

        if (Object.keys(fieldErrors).length > 0) {
            return;
        }

        loading = true;

        const payload: JoinRoomPayload = { username, roomSecretKey };

        try {
            const response = await joinRoom(roomId, payload);
            // Handle success (e.g., redirect or update the UI)
            console.log("Room joined successfully");
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

    <FloatingInputField
        id="roomId"
        type="text"
        labelText="Room Id"
        disabled={loading}
        bind:value={roomId}
        fieldError={fieldErrors.roomId}
        onInput={() => validateRoomId()}
    />

    <FloatingInputField
        id="roomSecretKey"
        type="password"
        labelText="Room Secret Key"
        disabled={loading}
        bind:value={roomSecretKey}
        fieldError={fieldErrors.roomSecretKey}
        onInput={() => validateRoomSecretKey()}
    />
{/snippet}

{#snippet modalContent({ closeModal }: { closeModal: () => void })}
    <ModalFormLayout
        headerTitle="Join a Chat Room"
        buttonText="Join"
        {loading}
        {buttonDisabled}
        {closeModal}
        {fields}
        onSubmit={handleSubmit}
    />
{/snippet}

<ModalWrapper {loading} {closeModal} content={modalContent} />

{#if errorMessage}
    <ErrorAlert {errorMessage} {clearErrorMessage} />
{/if}
