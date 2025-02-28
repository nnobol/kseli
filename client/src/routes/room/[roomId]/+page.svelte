<script lang="ts">
    import ChatRoom from "$lib/chat/ChatRoom.svelte";
    import { initializeChatStore } from "$lib/chat/chatStore";

    let { data } = $props();

    initializeChatStore();
</script>

<ChatRoom {data} />
