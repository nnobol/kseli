<script lang="ts">
    import { sendMessage } from "$lib/chat/chatStore";

    let message = $state("");

    function handleSend() {
        if (message.trim()) {
            sendMessage(message);
            message = "";
        }
    }

    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === "Enter" && !event.shiftKey) {
            handleSend();
        }
    }
</script>

<div class="message-input">
    <input
        type="text"
        placeholder="Type a message..."
        bind:value={message}
        onkeypress={handleKeyPress}
    />
    <button onclick={handleSend}>Send</button>
</div>

<style>
    .message-input {
        display: flex;
        gap: 0.5rem;
        padding-top: 0.5rem;
    }

    input {
        flex: 1;
        padding: 0.5rem;
        border: 1px solid #ccc;
        border-radius: 4px;
    }

    button {
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
