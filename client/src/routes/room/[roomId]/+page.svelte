<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import ChatRoom from "$lib/chat/ChatRoom.svelte";
    import {
        initializeChatStore,
        disconnectChatStore,
    } from "$lib/chat/chatStore";

    let { data } = $props();

    onMount(() => {
        initializeChatStore(data.participants);
    });

    onDestroy(() => {
        disconnectChatStore();
    });
</script>

<ChatRoom {data} />
