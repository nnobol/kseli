<script lang="ts">
    import type { GetRoomOkResponse } from "$lib/api/rooms";
    import ChatContainer from "$lib/features/chat/ChatContainer.svelte";
    import RoomSidebar from "$lib/features/chat/RoomSidebar.svelte";
    import { messages, participants } from "$lib/stores/chatStore";

    interface Props {
        data: GetRoomOkResponse;
    }

    let { data }: Props = $props();
</script>

<main>
    <ChatContainer messages={$messages} />
    <RoomSidebar
        currentUserRole={data.userRole}
        participants={$participants}
        maxParticipants={data.maxParticipants}
        expiresAt={data.expiresAt}
        roomId={data.roomId}
        secretKey={data.secretKey}
    />
</main>

<style>
    main {
        display: flex;
        padding: 1rem;
        gap: 0.5rem;
    }

    @media (max-width: 768px) {
        main {
            flex-direction: column;
            padding: 0.5rem;
            gap: 0.25rem;
        }
    }
</style>
