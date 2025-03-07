<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { beforeNavigate } from "$app/navigation";
    import ChatRoom from "$lib/chat/ChatRoom.svelte";
    import {
        initializeChatStore,
        disconnectChatStore,
    } from "$lib/stores/chatStore";
    import { setItemInLocalStorage } from "$lib/api/utils.js";
    import { broadcastRoomInfo } from "$lib/broadcast/broadcast.js";

    let { data } = $props();

    function cleanupRoom() {
        disconnectChatStore();
        localStorage.removeItem("token");
        localStorage.removeItem("activeRoomId");
        broadcastRoomInfo(null, null);
    }

    onMount(() => {
        initializeChatStore(data.roomDetails.participants, data.token);
        setItemInLocalStorage("token", data.token, 1);
        setItemInLocalStorage("activeRoomId", data.roomId, 1);
        broadcastRoomInfo(data.roomId, null);

        window.addEventListener("beforeunload", cleanupRoom);
    });

    beforeNavigate(() => {
        cleanupRoom();
    });

    onDestroy(() => {
        window.removeEventListener("beforeunload", cleanupRoom);
        cleanupRoom();
    });
</script>

<ChatRoom data={data.roomDetails} />
