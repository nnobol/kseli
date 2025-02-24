<script lang="ts">
    import type { LoadReturn } from "./+page";

    export let data: LoadReturn;
    const { roomDetails, error } = data;
</script>

{#if error}
    <h1>Error {error.statusCode}</h1>
    <p>{error.errorMessage}</p>
{:else if roomDetails}
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
{:else}
    <p>Loading...</p>
{/if}
