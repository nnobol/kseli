<script lang="ts">
    import { goto } from "$app/navigation";
    import ModalWrapper from "./ModalWrapper.svelte";
    import ModalFormLayout from "./ModalFormLayout.svelte";
    import FloatingInputField from "../../components/fields/FloatingInputField.svelte";
    import RadioFieldMaxParticipants from "../../components/fields/RadioFieldMaxParticipants.svelte";
    import ErrorAlert from "../../components/error-alert/ErrorAlert.svelte";
    import { createRoom } from "$lib/api/rooms";
    import { tokenStore } from "$lib/stores/tokenStore";
    import { keyStore, generateEncryptionKey } from "$lib/stores/keyStore";
    import type { CreateRoomPayload, RoomErrorResponse } from "$lib/api/rooms";

    interface Props {
        closeModal: () => void;
    }

    let { closeModal }: Props = $props();

    let username: string = $state("");
    let maxParticipants: number = $state(0);

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
        validateMaxParticipants();
    }

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

    function validateMaxParticipants(): void {
        if (!hasSubmitted) return;

        if (!maxParticipants) {
            fieldErrors.maxParticipants = "Select one of the values.";
        } else {
            delete fieldErrors.maxParticipants;
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

        const payload: CreateRoomPayload = { username, maxParticipants };

        try {
            const response = await createRoom(payload);
            const key = await generateEncryptionKey();
            tokenStore.set(response.token);
            keyStore.set(key);
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

    <RadioFieldMaxParticipants
        disabled={loading}
        fieldError={fieldErrors.maxParticipants}
        onChange={(value) => {
            maxParticipants = value;
            validateMaxParticipants();
        }}
    />
{/snippet}

{#snippet modalContent({ closeModal }: { closeModal: () => void })}
    <ModalFormLayout
        headerTitle="CREATE A ROOM"
        buttonText="CREATE"
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
