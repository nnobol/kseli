<script lang="ts">
    interface Message {
        username: string;
        content: string;
    }

    interface Props {
        messages: Message[];
    }

    let { messages }: Props = $props();
</script>

<div class="chat-window">
    {#each messages as message}
        <p><strong>{message.username}:</strong> {message.content}</p>
    {/each}
</div>

<style>
    .chat-window {
        flex: 1;
        padding: 0.25rem 0;
        border-bottom: 2px solid #ccc;
        border-top: 2px solid #ccc;
    }

    p {
        color: #24292f;
        margin: 0.1rem 0;
    }
</style>
