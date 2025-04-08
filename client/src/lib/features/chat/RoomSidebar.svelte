<script lang="ts">
    import RoomParticipants from "./RoomParticipants.svelte";
    import RoomControl from "./RoomControl.svelte";
    import type { Participant } from "$lib/api/rooms";

    interface Props {
        currentUserRole: number;
        participants: Participant[];
        maxParticipants: number;
        expiresAt: number;
        roomId: string;
        secretKey?: string;
    }

    let {
        currentUserRole,
        participants,
        maxParticipants,
        expiresAt,
        roomId,
        secretKey,
    }: Props = $props();
</script>

<div class="room-sidebar">
    <RoomParticipants {currentUserRole} {participants} {maxParticipants} />
    <RoomControl {expiresAt} {roomId} {secretKey} {currentUserRole} />
</div>

<style>
    .room-sidebar {
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }

    @media (max-width: 768px) {
        .room-sidebar {
            flex-direction: row;
            gap: 0.2rem;
        }

        :global(.room-sidebar > *) {
            flex: 1 1 0;
        }
    }
</style>
