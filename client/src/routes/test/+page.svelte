<script lang="ts">
    import RoomParticipants from "$lib/chat/RoomParticipants.svelte";
    import RoomControl from "$lib/chat/RoomControl.svelte";

    const roomDetails = {
        userRole: 1,
        maxParticipants: 4,
        participants: [
            {
                id: 1,
                username: "nika",
                role: 1,
            },
            {
                id: 2,
                username: "lado",
                role: 2,
            },
        ],
        secretKey: "super-secret-key",
    };
</script>

<main>
    <div>
        <!-- "wrapper for RoomParticipants and RoomControls" -->
        <RoomParticipants
            currentUserRole={roomDetails.userRole}
            participants={roomDetails.participants}
            maxParticipants={roomDetails.maxParticipants}
        />
        <RoomControl
            secretKey={roomDetails.secretKey}
            currentUserRole={roomDetails.userRole}
        />
    </div>
</main>

<style>
    main {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 1rem;
    }

    div {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        height: 100%;
    }
</style>
