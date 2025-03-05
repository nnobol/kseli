<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import ChatRoom from "$lib/chat/ChatRoom.svelte";
    import {
        initializeChatStore,
        disconnectChatStore,
    } from "$lib/stores/chatStore";
    import { setItemInLocalStorage } from "$lib/api/utils.js";

    let { data } = $props();
    let channel: BroadcastChannel;

    onMount(() => {
        initializeChatStore(data.roomDetails.participants, data.token);
        setItemInLocalStorage("roomToken", data.token, 1);
        setItemInLocalStorage("roomId", data.roomId, 1);

        channel = new BroadcastChannel("active-room");
        channel.postMessage({ roomId: data.roomId });
    });

    onDestroy(() => {
        disconnectChatStore();
    });
</script>

<ChatRoom data={data.roomDetails} />
