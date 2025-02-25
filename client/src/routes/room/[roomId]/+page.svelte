<script lang="ts">
    import type { GetRoomOkResponse } from "../../../api/rooms";
    export let data: { roomDetails: GetRoomOkResponse };
    const { roomDetails } = data;
</script>

<h1>Room</h1>
<p>Max Participants: {roomDetails.maxParticipants}</p>
<h2>Participants</h2>
<ul>
    {#each roomDetails.participants as participant}
        <li>{participant.username} (Role: {participant.role})</li>
    {/each}
</ul>
{#if roomDetails.secretKey}
    <p>Secret Key: {roomDetails.secretKey}</p>
{/if}
