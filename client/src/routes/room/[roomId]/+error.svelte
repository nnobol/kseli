<script lang="ts">
    import { onMount } from "svelte";
    import { page } from "$app/state";
    import { goto } from "$app/navigation";
    import Footer from "$lib/components/Footer.svelte";
    import { getItemFromSessionStorage } from "$lib/api/utils";

    let isSameSession = false;

    onMount(() => {
        isSameSession = !!getItemFromSessionStorage("activeRoomId");
        sessionStorage.clear();
    });
</script>

<main>
    <div class="err-text">
        <h1>Error {page.status}</h1>
        <p>{page.error?.message || "Something went wrong"}</p>
    </div>
    <button onclick={() => goto("/")}>Go Home</button>
</main>

<Footer isErrorPage={true} />

<style>
    main {
        display: flex;
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
        gap: 0.5rem;
    }

    h1 {
        font-size: 3.5rem;
    }

    p {
        font-size: 2rem;
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
        font-weight: bold;
        transition: opacity 0.25s ease;
    }

    button:hover {
        opacity: 0.8;
    }
</style>
