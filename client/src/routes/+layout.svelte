<script lang="ts">
    import "../styles/global.css";
    import { onMount, onDestroy } from "svelte";
    import { page } from "$app/state";
    import { goto } from "$app/navigation";
    import { getItemFromLocalStorage } from "$lib/api/utils";

    let channel: BroadcastChannel;

    onMount(() => {
        const activeRoomId = getItemFromLocalStorage("roomId");
        const currentPath = page.url.pathname;
        if (activeRoomId && currentPath !== `/room/${activeRoomId}`) {
            goto(`/room/${activeRoomId}`);
        }

        channel = new BroadcastChannel("active-room");
        channel.onmessage = (event) => {
            const { roomId } = event.data;
            const currentPath = page.url.pathname;

            if (roomId && currentPath !== `/room/${roomId}`) {
                goto(`/room/${roomId}`);
            } else if (!roomId && currentPath.startsWith("/room/")) {
                goto("/");
            }
        };
    });

    onDestroy(() => {
        if (channel) channel.close();
    });
</script>

<slot />
