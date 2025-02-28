<script lang="ts">
    import ChatWindow from "./ChatWindow.svelte";
    import MessageInput from "./MessageInput.svelte";

    interface Message {
        username: string;
        content: string;
    }

    interface Props {
        messages: Message[];
    }

    let { messages }: Props = $props();
</script>

<div class="chat-container">
    <ChatWindow {messages} />
    <MessageInput />
</div>

<style>
    .chat-container {
        display: flex;
        flex: 1;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 1rem;
        background-color: #f9f9f9;
    }
</style>
