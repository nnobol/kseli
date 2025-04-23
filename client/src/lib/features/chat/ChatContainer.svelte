<script lang="ts">
    import ChatWindow from "./ChatWindow.svelte";
    import MessageInput from "./MessageInput.svelte";
</script>

<div class="chat-container">
    <ChatWindow />
    <MessageInput />
</div>

<style>
    .chat-container {
        display: flex;
        flex: 1;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 0.8rem;
        background-color: #f9f9f9;
    }
</style>
