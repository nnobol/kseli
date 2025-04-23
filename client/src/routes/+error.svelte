<script lang="ts">
    import { page } from "$app/state";
    import { goto } from "$app/navigation";
    import { onMount } from "svelte";
    import { getItemFromSessionStorage } from "$lib/api/utils";

    let isSameSession = false;

    onMount(() => {
        isSameSession = !!getItemFromSessionStorage("activeRoomId");
        sessionStorage.clear();
    });
</script>

<main>
    <div class="err-text">
        {#if isSameSession}
            <h1>Error 500</h1>
            <p>
                You tried navigating to a different page which closed the room.
            </p>
        {:else if page.status === 404}
            <h1>404 - Page Not Found</h1>
            <p>Route '{page.url.pathname}' doesn’t exist.</p>
        {:else}
            <h1>Error {page.status || 500}</h1>
            <p>{page.error?.message || "Something went wrong"}</p>
        {/if}
    </div>
    <button onclick={() => goto("/")}>Go Home</button>
</main>

<style>
    main {
        display: flex;
        flex: 1;
        background-color: #ffe4e4;
        color: #d8000c;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        text-align: center;
        padding: 2rem;
        gap: 5rem;
    }

    .err-text {
        display: flex;
        flex-direction: column;
        gap: 2rem;
    }

    h1 {
        font-size: 3.5rem;
        font-weight: var(--font-weight-black);
        text-transform: uppercase;
    }

    p {
        font-size: 2rem;
        font-weight: var(--font-weight-bold);
    }

    button {
        background-color: #24292f;
        color: #ffe4e4;
        border: none;
        border-radius: 5px;
        padding: 0.5rem 1rem;
        font-size: 1.5rem;
        cursor: pointer;
        font-family: inherit;
        font-weight: var(--font-weight-bold);
        transition: opacity 0.25s ease;
    }

    button:hover {
        opacity: 0.8;
    }
</style>
