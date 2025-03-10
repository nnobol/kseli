<script lang="ts">
    import { onMount } from "svelte";
    import { beforeNavigate } from "$app/navigation";
    import ChatRoom from "$lib/chat/ChatRoom.svelte";
    import {
        initializeChatStore,
        disconnectChatStore,
    } from "$lib/stores/chatStore";

    let { data } = $props();

    onMount(() => {
        initializeChatStore(data.roomDetails.participants, data.token);
    });

    beforeNavigate(({ type }) => {
        if (type === "popstate") {
            disconnectChatStore();
            sessionStorage.clear();
        }
    });

    function handleBeforeUnload() {
        disconnectChatStore();
    }
</script>

<!-- <svelte:window on:beforeunload={handleBeforeUnload} /> -->
<ChatRoom data={data.roomDetails} />
