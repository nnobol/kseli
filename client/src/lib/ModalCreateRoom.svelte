<script lang="ts">
    import { scale, fade } from "svelte/transition";
    import { goto } from "$app/navigation";
    import { tick, onMount } from "svelte";

    interface Props {
        closeModal: () => void;
    }

    let { closeModal }: Props = $props();
    let isModalClosed: boolean = $state(true);
    let dialogElement = $state<HTMLDialogElement>()!;
    let username: string = $state("");
    let selectedRadio: string = $state("");
    let maxParticipants: number = $state(0);
    let loading: boolean = $state(false);
    let errorMessage: string = $state("");
    let fieldErrors: Record<string, string> = $state({});

    async function handleClose(): Promise<void> {
        if (!loading && errorMessage === "") isModalClosed = true;
    }

    function handleKeyDown(event: KeyboardEvent): void {
        if (!loading && errorMessage === "" && event.key === "Escape") {
            handleClose();
        }
    }

    onMount(async () => {
        isModalClosed = false;
        await tick();
        dialogElement.focus();
    });

    interface CreateRoomPayload {
        username: string;
        maxParticipants: number;
    }

    interface CreateRoomOkResponse {
        roomId: string;
    }

    interface CreateRoomErrorResponse {
        errorMessage: string;
        fieldErrors?: Record<string, string>;
    }

    async function createRoom() {
        try {
            loading = true;
            errorMessage = "";
            fieldErrors = {};

            const payload: CreateRoomPayload = {
                username,
                maxParticipants,
            };

            const response = await fetch("/api/room", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(payload),
            });

            const responseBody = await response.json();

            if (!response.ok) {
                const data = responseBody as CreateRoomErrorResponse;

                if (data.fieldErrors) {
                    setFieldErrors(data.fieldErrors);
                } else {
                    errorMessage =
                        data.errorMessage || "An unexpected error occurred.";
                }
                return;
            }

            const data = responseBody as CreateRoomOkResponse;
            // goto(`/room/${data.roomId}`);
        } catch (err) {
            errorMessage = "Failed to connect to the server. Please try again.";
        } finally {
            loading = false;
        }
    }

    function setFieldErrors(errors: Record<string, string>) {
        fieldErrors = { ...errors };
    }

    function clearFieldError(field: string) {
        if (fieldErrors[field]) {
            delete fieldErrors[field];
            fieldErrors = { ...fieldErrors };
        }
    }

    function handleBlur(field: string) {
        if (fieldErrors[field]) {
            const input = document.getElementById(
                field,
            ) as HTMLInputElement | null;
            input?.blur();
        }
    }

    function handleSubmit(event: Event) {
        event.preventDefault();

        fieldErrors = {};

        if (!username) {
            fieldErrors.username = "Username is required";
        }

        if (!maxParticipants) {
            fieldErrors.maxParticipants = "Select one of the values";
        }

        if (Object.keys(fieldErrors).length > 0) {
            return;
        }

        createRoom();
    }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions, a11y_no_noninteractive_element_interactions --->
{#if !isModalClosed}
    <dialog
        onclick={handleClose}
        onkeydown={handleKeyDown}
        bind:this={dialogElement}
        transition:fade={{ duration: 200 }}
    >
        <section
            class="modal-content {errorMessage ? 'error-active' : ''}"
            onclick={(event) => {
                event.stopPropagation();
            }}
            in:scale={{ duration: 200 }}
            out:scale={{ duration: 200 }}
            onoutroend={closeModal}
        >
            <button
                class="close-button"
                onclick={handleClose}
                disabled={loading}
            >
                <img
                    src="/close-icon.svg"
                    alt="Close"
                    style="width: 1.6rem; height: auto;"
                />
            </button>

            {#if errorMessage}
                <div class="error-overlay">
                    <section class="error-alert">
                        <p>{errorMessage}</p>
                        <button
                            class="close-error"
                            onclick={() => (errorMessage = "")}
                        >
                            ✖
                        </button>
                    </section>
                </div>
            {/if}

            <h2 class={loading ? "loading" : ""}>Create a Chat Room</h2>
            <form class="input-form" novalidate onsubmit={handleSubmit}>
                <fieldset class="text-field" disabled={loading}>
                    <div>
                        <input
                            type="text"
                            id="username"
                            name="username"
                            required
                            bind:value={username}
                            placeholder=""
                            class:error={fieldErrors.username}
                            onblur={() => handleBlur("username")}
                            onfocus={() => clearFieldError("username")}
                        />
                        <label for="username" class:error={fieldErrors.username}
                            >Username</label
                        >
                    </div>
                    {#if fieldErrors.username}
                        <span class="error-message">{fieldErrors.username}</span
                        >
                    {/if}
                </fieldset>

                <fieldset class="radio-group" disabled={loading}>
                    <legend>Maximum Number of Participants</legend>
                    <ul class="radio-options">
                        {#each [2, 3, 4, 5] as number}
                            <li>
                                <label
                                    class:error={fieldErrors.maxParticipants}
                                >
                                    <input
                                        type="radio"
                                        name="maxParticipants"
                                        value={number}
                                        required
                                        bind:group={selectedRadio}
                                        onchange={() => {
                                            maxParticipants =
                                                parseInt(selectedRadio);
                                            clearFieldError("maxParticipants");
                                        }}
                                        disabled={loading}
                                    />
                                    <span>{number}</span>
                                </label>
                            </li>
                        {/each}
                    </ul>
                    {#if fieldErrors.maxParticipants}
                        <span class="error-message"
                            >{fieldErrors.maxParticipants}</span
                        >
                    {/if}
                </fieldset>

                <button
                    type="submit"
                    class="create-button"
                    disabled={loading || Object.keys(fieldErrors).length > 0}
                >
                    {#if loading}
                        <span class="spinner"></span>
                    {:else}
                        Create
                    {/if}
                </button>
            </form>
        </section>
    </dialog>
{/if}

<style>
    /* MODAL & LAYOUT */
    dialog {
        position: fixed;
        top: 0;
        left: 0;
        width: 100vw;
        height: 100vh;
        background: rgba(0, 0, 0, 0.4);
        backdrop-filter: blur(5px);
        -webkit-backdrop-filter: blur(5px);
        display: flex;
        justify-content: center;
        align-items: center;
        border: none;
        padding: 0;
        margin: 0;
    }

    .modal-content {
        background: #cbc6ac;
        color: #32012f;
        padding: 1rem 2rem;
        margin: 2rem;
        border-radius: 8px;
        max-width: 500px;
        width: 90%;
        text-align: center;
        position: relative;
        box-shadow: 0px 4px 12px rgba(0, 0, 0, 0.3);
    }

    .modal-content h2 {
        font-weight: bold;
        font-size: 1.5rem;
    }

    .modal-content.error-active {
        pointer-events: none;
        user-select: none;
    }

    .modal-content h2.loading {
        opacity: 0.6;
        pointer-events: none;
    }

    /* CLOSE BUTTON */
    .close-button {
        background: none;
        border: none;
        padding: 0;
        cursor: pointer;
        position: absolute;
        top: 0.5rem;
        right: 0.5rem;
    }

    .close-button:hover img {
        opacity: 0.8;
    }

    .close-button[disabled] {
        cursor: not-allowed;
        opacity: 0.6;
    }

    /* FIELDSET DISABLED OVERLAY */
    fieldset[disabled] {
        opacity: 0.6;
        pointer-events: none;
    }

    /* CREATE BUTTON & SPINNER */
    .create-button {
        background-color: #d26100;
        color: #bcb594;
        border: none;
        border-radius: 5px;
        cursor: pointer;
        font-size: 1rem;
        padding: 0.5rem 1rem;
        font-family: inherit;
        font-weight: bold;
        transition:
            background-color 0.3s ease,
            color 0.3s ease;
        margin-top: 0.5rem;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-width: 6rem;
        text-align: center;
    }

    .create-button:hover {
        background-color: #ab4f00;
        color: #ada47c;
    }

    .create-button[disabled] {
        cursor: not-allowed;
        opacity: 0.6;
    }

    .spinner {
        border: 2px solid rgba(255, 255, 255, 0.3);
        border-top: 2px solid #ffffff;
        border-radius: 50%;
        width: 1rem;
        height: 1rem;
        animation: spin 1s linear infinite;
    }

    @keyframes spin {
        0% {
            transform: rotate(0deg);
        }
        100% {
            transform: rotate(360deg);
        }
    }

    /* FORMS & TEXT FIELDS */
    .input-form {
        margin-top: 1.5rem;
    }

    .input-form fieldset {
        border: none;
        padding: 0;
        margin-bottom: 0.8rem;
        position: relative;
    }

    .text-field div {
        position: relative;
        display: flex;
        flex-direction: column;
    }

    .text-field input {
        width: 100%;
        padding: 0.5rem 0.5rem;
        font-size: 1rem;
        border: 2px solid #32012f;
        border-radius: 4px;
        outline: none;
        background: transparent;
        color: #32012f;
        transition: border-color 0.15s ease-in-out;
        font-family: inherit;
    }

    .text-field label {
        position: absolute;
        top: 50%;
        left: 0.5rem;
        transform: translateY(-50%);
        font-size: 1rem;
        color: #32012f;
        pointer-events: none;
        transition: all 0.15s ease-in-out;
        background: #cbc6ac;
        font-family: inherit;
        padding: 0 0.2rem;
    }

    /* On Focus or Non-Empty: Orange Border & Moved Label */
    .text-field input:focus,
    .text-field input:not(:placeholder-shown) {
        border-color: #d26100;
    }

    .text-field input:focus + label,
    .text-field input:not(:placeholder-shown) + label {
        top: 0;
        left: 0.4rem;
        font-size: 0.75rem;
        color: #d26100;
        border-right: 1px solid #cbc6ac;
        border-left: 1px solid #cbc6ac;
    }

    /* The Vertical "Border Bits" (Pseudo-Elements) in Orange */
    .text-field input:focus + label::before,
    .text-field input:not(:placeholder-shown) + label::before,
    .text-field input:focus + label::after,
    .text-field input:not(:placeholder-shown) + label::after {
        content: "";
        position: absolute;
        width: 2px;
        height: 0.5rem;
        background-color: #d26100;
        top: 50%;
        transform: translateY(-50%);
    }

    .text-field input:focus + label::before,
    .text-field input:not(:placeholder-shown) + label::before {
        left: -0.1rem;
    }

    .text-field input:focus + label::after,
    .text-field input:not(:placeholder-shown) + label::after {
        right: -0.1rem;
    }

    /* RADIO GROUP */
    .radio-group {
        margin-bottom: 1rem;
        text-align: center;
    }

    .radio-group legend {
        font-size: 1rem;
        margin-bottom: 0.5rem;
        color: #32012f;
    }

    /* Radio Buttons Row */
    .radio-options {
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;
        list-style: none;
        padding: 0;
        margin: 0;
        justify-content: center;
    }

    .radio-options label {
        display: flex;
        align-items: center;
        cursor: pointer;
        font-size: 1rem;
        color: #32012f;
        position: relative;
    }

    /* Circle of the Radio option */
    .radio-options input {
        appearance: none;
        -webkit-appearance: none;
        -moz-appearance: none;
        width: 1rem;
        height: 1rem;
        border: 2px solid #32012f;
        border-radius: 50%;
        position: relative;
        margin-right: 0.5rem;
        transition: border-color 0.15s;
        cursor: pointer;
    }

    .radio-options input:hover {
        border-color: #d26100;
    }

    .radio-options input:checked {
        border-color: #d26100;
    }

    .radio-options input:checked::after {
        content: "";
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        width: 0.55rem;
        height: 0.55rem;
        background-color: #d26100;
        border-radius: 50%;
    }

    .radio-options label span {
        transition: color 0.2s ease;
    }

    .radio-options label:hover span {
        color: #d26100;
    }

    .radio-options input:checked + span {
        color: #d26100;
    }

    /* ERROR STATES */
    .text-field input.error {
        border-color: red !important;
    }

    .text-field input:focus + label.error,
    .text-field input:not(:placeholder-shown) + label.error {
        color: red !important;
    }

    .text-field input.error:focus {
        border-color: red !important;
    }

    /* The Vertical "Border Bits" (Pseudo-Elements) in Red When Error is Active */
    .text-field input.error:not(:placeholder-shown) + label::before,
    .text-field input.error:not(:placeholder-shown) + label::after {
        background-color: red !important;
    }

    .radio-group label.error {
        color: red !important;
    }

    .radio-options label.error:hover span {
        color: red !important;
    }

    .radio-group label.error input {
        border-color: red !important;
    }

    .error-message {
        font-size: 0.875rem;
        color: red;
        margin-top: 0.2rem;
        display: block;
    }

    /* ERROR ALERT */
    .error-overlay {
        pointer-events: auto;
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.6);
        backdrop-filter: blur(3px);
        border-radius: 8px;
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 1000;
    }

    .error-alert {
        background-color: #ffe4e4;
        color: #d8000c;
        border: 1px solid #d8000c;
        border-radius: 8px;
        padding: 1rem 1.5rem;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 1rem;
        max-width: 400px;
        width: 90%;
        text-align: center;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
        font-size: 1rem;
        position: relative;
        pointer-events: auto;
    }

    /* ERROR MESSAGE TEXT */
    .error-alert p {
        margin: 0;
        flex-grow: 1;
        text-align: center;
        user-select: text;
    }

    /* ERROR CLOSE BUTTON */
    .close-error {
        background: none;
        border: none;
        color: #d8000c;
        font-size: 1.5rem;
        font-weight: bold;
        cursor: pointer;
        transition:
            transform 0.2s ease,
            opacity 0.2s ease;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0;
        pointer-events: auto;
    }

    .close-error:hover {
        transform: scale(1.1);
        opacity: 0.8;
    }
</style>
