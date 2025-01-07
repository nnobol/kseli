<script lang="ts">
    import Modal from "$lib/ModalCreateRoom.svelte";

    let showModal: boolean = false;

    function openModal(): void {
        showModal = true;
    }

    function closeModal(): void {
        showModal = false;
    }
</script>

<main>
    <section>
        <h1>
            <span class="highlight">Kseli</span> - private, secure and fast communication.
        </h1>
        <button on:click={openModal}>Create a Chat Room</button>
    </section>

    {#if showModal}
        <Modal {closeModal} />
    {/if}
</main>

<style>
    main {
        display: flex;
        justify-content: center;
        align-items: flex-start;
        flex: 1;
        padding: 2rem;
    }

    section {
        text-align: center;
    }

    h1 {
        font-size: 3.5rem;
        margin-top: 10rem;
        margin-bottom: 8rem;
    }

    button {
        background-color: #d26100;
        color: #bcb594;
        border: none;
        border-radius: 5px;
        cursor: pointer;
        font-size: 2rem;
        padding: 0.75rem 1.5rem;
        cursor: pointer;
        font-family: inherit;
        font-weight: bold;
        transition:
            background-color 0.3s ease,
            color 0.3s ease;
    }

    button:hover {
        background-color: #ab4f00;
        color: #ada47c;
    }

    .highlight {
        color: #d26100;
        font-size: 4rem;
    }
</style>
