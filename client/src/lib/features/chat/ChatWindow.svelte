<script lang="ts">
    interface Message {
        username: string;
        content: string;
    }

    interface Props {
        messages: Message[];
    }

    let { messages }: Props = $props();
    let shouldAutoScroll = $state(true);
    let chatContainer: HTMLElement;

    function isAtBottom() {
        if (!chatContainer) return true;

        const threshold = 50;
        const position =
            chatContainer.scrollHeight -
            chatContainer.scrollTop -
            chatContainer.clientHeight;
        return position <= threshold;
    }

    function handleScroll() {
        shouldAutoScroll = isAtBottom();
    }

    $effect(() => {
        if (chatContainer && messages.length) {
            if (shouldAutoScroll) {
                chatContainer.scrollTop = chatContainer.scrollHeight;
            }
        }
    });
</script>

<div class="chat-window" bind:this={chatContainer} onscroll={handleScroll}>
    {#each messages as message}
        <p><strong>{message.username}:</strong> {message.content}</p>
    {/each}
</div>

<style>
    .chat-window {
        flex: 1 1 0;
        padding: 0.25rem 0;
        border-bottom: 2px solid #ccc;
        border-top: 2px solid #ccc;
        overflow-y: auto;
    }

    p {
        font-size: 1rem;
        color: #24292f;
        margin: 0.1rem 0;
        word-break: break-word;
    }
</style>
