<script lang="ts">
    import { onMount } from "svelte";
    import { beforeNavigate } from "$app/navigation";
    import ChatRoom from "$lib/features/chat/ChatRoom.svelte";
    import { initChatSession, resetChatSession, endChatSession } from "$lib/stores/chatStore";
    import { closeRoom } from "$lib/api/rooms";

    let { data } = $props();

    onMount(() => {
        resetChatSession();
        initChatSession(data.roomDetails.participants, data.token);
    });

    beforeNavigate(({ type }) => {
        if (type === "popstate") {
            if (data.roomDetails.userRole === 1) {
                closeRoom();
            } else {
                endChatSession();
            }
        }
    });
</script>

<ChatRoom data={data.roomDetails} />
