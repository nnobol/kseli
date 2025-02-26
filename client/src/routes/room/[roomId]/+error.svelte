<script lang="ts">
    import { page } from "$app/state";
    import { goto } from "$app/navigation";
    import Footer from "$lib/common/Footer.svelte";
</script>

<main>
    <h1>Error {page.status}</h1>
    <p>{page.error?.message || "Something went wrong"}</p>
    <button on:click={() => goto("/")}>Go Home</button>
</main>

<Footer isErrorPage={true} />

<style>
    main {
        background-color: #ffe4e4;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        padding: 2rem;
    }

    h1 {
        color: #d8000c;
        text-align: center;
        font-size: 3.5rem;
        margin-bottom: 0.5rem;
    }

    p {
        color: #d8000c;
        text-align: center;
        font-size: 2rem;
        margin-bottom: 3rem;
    }

    button {
        background-color: #24292f;
        color: #ffe4e4;
        border: none;
        border-radius: 5px;
        cursor: pointer;
        font-size: 1.5rem;
        padding: 0.5rem 1rem;
        font-family: inherit;
        font-weight: bold;
        transition: opacity 0.2s ease;
        margin-top: 1.5rem;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-width: 6rem;
        min-height: 2.2rem;
    }

    button:hover {
        opacity: 0.8;
    }
</style>
