<script lang="ts">
    interface Props {
        secretKey?: string;
        currentUserRole: number;
    }

    let { secretKey, currentUserRole }: Props = $props();
</script>

<div class="room-control">
    {#if currentUserRole == 1}
        <p>Key: {secretKey}</p>
    {/if}
    <div class="button-wrapper">
        <button>Leave Room</button>
        {#if currentUserRole == 1}
            <button>Close Room</button>
        {/if}
    </div>
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
