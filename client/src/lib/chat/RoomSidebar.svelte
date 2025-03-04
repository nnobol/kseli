<script lang="ts">
    import RoomParticipants from "./RoomParticipants.svelte";
    import RoomControl from "./RoomControl.svelte";
    import type { Participant } from "$lib/api/rooms";

    interface Props {
        currentUserRole: number;
        participants: Participant[];
        maxParticipants: number;
        secretKey?: string;
    }

    let { currentUserRole, participants, maxParticipants, secretKey }: Props =
        $props();
</script>

<div class="room-sidebar">
    <RoomParticipants {currentUserRole} {participants} {maxParticipants} />
    <RoomControl {secretKey} {currentUserRole} />
</div>

<style>
    .room-sidebar {
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }

    @media (max-width: 768px) {
        .room-sidebar {
            gap: 0.2rem;
        }
    }
</style>
