<script lang="ts">
    import { endChatSession } from "$lib/stores/chatStore";
    import { closeRoom } from "$lib/api/rooms";
    import InlineErrorToast from "$lib/common/InlineErrorToast.svelte";
    import { onMount, onDestroy } from "svelte";

    interface Props {
        expiresAt: number;
        roomId: string;
        secretKey?: string;
        currentUserRole: number;
    }

    let { expiresAt, roomId, secretKey, currentUserRole }: Props = $props();

    let errorMessage: string | null = $state(null);
    let errorElement: HTMLElement | null = $state(null);
    let remainingTime: string = $state("Loading...");
    let intervalId: number;

    function formatTime(totalSeconds: number): string {
        if (totalSeconds <= 0) {
            return "Expired";
        }

        const minutes = Math.floor(totalSeconds / 60);
        const seconds = totalSeconds % 60;
        return `${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
    }

    function updateCountdown() {
        const now = Math.floor(Date.now() / 1000);
        const diff = expiresAt - now;

        if (diff <= 0) {
            remainingTime = "Expired";
            clearInterval(intervalId);
        } else {
            remainingTime = formatTime(diff);
        }
    }

    onMount(() => {
        updateCountdown();

        intervalId = setInterval(updateCountdown, 1000);
    });

    onDestroy(() => {
        clearInterval(intervalId);
    });

    async function handleClose(event: MouseEvent) {
        const targetElement = event.currentTarget as HTMLElement;
        try {
            await closeRoom();
        } catch (err) {
            errorMessage = "Failed to close room.";
            errorElement = targetElement;
        }
    }
</script>

<div class="room-control">
    <p>Time Remaining: {remainingTime}</p>
    <p>Room Id: {roomId}</p>
    {#if currentUserRole === 1}
        <p>Key: {secretKey}</p>
    {/if}
    <div class="button-wrapper">
        {#if currentUserRole === 1}
            <button onclick={(e) => handleClose(e)}>Close Room</button>
        {:else}
            <button onclick={() => endChatSession()}>Leave Room</button>
        {/if}
    </div>

    {#if errorMessage && errorElement}
        <InlineErrorToast
            message={errorMessage}
            targetElement={errorElement}
            clearError={() => {
                errorMessage = null;
                errorElement = null;
            }}
        />
    {/if}
</div>

<style>
    .room-control {
        display: flex;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 1rem;
        color: #24292f;
        background-color: #fff;
        gap: 1rem;
    }

    p {
        font-size: 0.8rem;
        text-align: center;
    }

    .button-wrapper {
        display: flex;
        justify-content: center;
        gap: 0.35rem;
    }

    button {
        padding: 0.5rem 1rem;
        background-color: #24292f;
        color: #f9f9f9;
        border: none;
        border-radius: 4px;
        cursor: pointer;
    }

    button:hover {
        opacity: 0.8;
    }
</style>
