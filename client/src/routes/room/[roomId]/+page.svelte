<script lang="ts">
    import { onMount } from "svelte";
    import { page } from "$app/state";
    import { beforeNavigate } from "$app/navigation";
    import ChatRoom from "$lib/chat/ChatRoom.svelte";
    import { initChatSession, endChatSession } from "$lib/stores/chatStore";

    let { data } = $props();

    onMount(() => {
        initChatSession(data.roomDetails.participants, data.token);
    });

    beforeNavigate(({ type }) => {
        if (type === "popstate") {
            endChatSession();
        }
    });
</script>

<ChatRoom data={data.roomDetails} />
