<script lang="ts">
    import { tick } from "svelte";
    import { sendMessage } from "$lib/stores/chatStore";
    import InlineToast from "$lib/components/InlineToast.svelte";

    const MAX_MESSAGE_LENGTH = 150;

    let message = $state("");
    let inputEl: HTMLInputElement;

    let errors = $state<
        {
            id: number;
            message: string;
            element: HTMLElement;
            visible: boolean;
            timeout: number;
        }[]
    >([]);

    async function handleSend() {
        if (message.length > 0 && message.length <= MAX_MESSAGE_LENGTH) {
            try {
                await sendMessage(message);
            } catch (err) {
                const errorId = Date.now();
                const messageText =
                    err instanceof Error
                        ? err.message
                        : "Failed to send message.";

                errors = [
                    ...errors,
                    {
                        id: errorId,
                        message: messageText,
                        element: inputEl,
                        visible: false,
                        timeout: 0,
                    },
                ];

                await tick();

                errors = errors.map((error) =>
                    error.id === errorId ? { ...error, visible: true } : error,
                );

                setTimeout(() => {
                    errors = errors.map((error) =>
                        error.id === errorId
                            ? { ...error, visible: false }
                            : error,
                    );
                }, 1500);
            } finally {
                message = "";
            }
        }
    }

    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === "Enter" && !event.shiftKey) {
            handleSend();
        }
    }

    function clearError(index: number) {
        const error = errors[index];
        if (error.timeout) clearTimeout(error.timeout);
        errors = errors.filter((_, i) => i !== index);
    }
</script>

<div class="message-input">
    <div class="msg-len">{message.length}/{MAX_MESSAGE_LENGTH}</div>
    <div class="input-wrapper">
        <input
            bind:this={inputEl}
            type="text"
            placeholder="Type a message..."
            bind:value={message}
            onkeypress={handleKeyPress}
            maxlength={MAX_MESSAGE_LENGTH}
        />
        <button onclick={handleSend}>Send</button>
    </div>

    {#each errors as error, index (error.id)}
        <InlineToast
            typeError={true}
            message={error.message}
            targetElement={error.element}
            visible={error.visible}
            removeToast={() => clearError(index)}
        />
    {/each}
</div>

<style>
    .message-input {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        padding-top: 0.5rem;
        flex-wrap: wrap;
    }

    .msg-len {
        font-size: 0.75rem;
        margin-left: 0.25rem;
        color: #24292f;
        width: 100%;
    }

    .input-wrapper {
        display: flex;
        width: 100%;
        gap: 0.5rem;
    }

    input {
        flex: 1;
        padding: 0.5rem;
        border: 1px solid #ccc;
        border-radius: 4px;
        font-size: 0.8rem;
    }

    button {
        font-size: 0.8rem;
        font-family: inherit;
        padding: 0.5rem 1rem;
        background-color: #24292f;
        color: #f9f9f9;
        border: none;
        border-radius: 4px;
        cursor: pointer;
    }

    button:hover {
        opacity: 0.8;
    }
</style>
