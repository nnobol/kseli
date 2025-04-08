<script lang="ts">
    import CreateChatRoomModal from "$lib/features/modals/CreateChatRoomModal.svelte";
    import JoinChatRoomModal from "$lib/features/modals/JoinChatRoomModal.svelte";
    import Footer from "$lib/components/Footer.svelte";
    import ErrorAlert from "$lib/components/error-alert/ErrorAlert.svelte";

    let { data } = $props();

    let errorMessage = $state(data.errorMessage);
    let activeModal = $state<"create" | "join" | null>(null);

    const modalComponents = {
        create: CreateChatRoomModal,
        join: JoinChatRoomModal,
    };

    function toggleModal(modal: "create" | "join" | null): void {
        if (activeModal && modal !== null) return;
        activeModal = modal;
    }
</script>

<main>
    <h1>
        <span class="highlight">Kseli</span> - anonymous, temporary chat rooms.
    </h1>

    <div class="btns">
        <button class="join-btn" onclick={() => toggleModal("join")}>
            Join Chat Room
        </button>

        <button class="create-btn" onclick={() => toggleModal("create")}>
            Create Chat Room
        </button>
    </div>

    {#if activeModal}
        {@const Modal = modalComponents[activeModal]}
        <Modal closeModal={() => toggleModal(null)} />
    {/if}
</main>

<Footer isErrorPage={false} />

{#if errorMessage}
    <ErrorAlert {errorMessage} clearErrorMessage={() => (errorMessage = "")} />
{/if}

<style>
    main {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        padding: 2rem;
        gap: 6rem;
    }

    h1 {
        text-align: center;
        font-size: 3.5rem;
    }

    .highlight {
        color: #d26100;
        font-size: 4rem;
    }

    .btns {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    button {
        background-color: #1f011d;
        border-radius: 5px;
        font-size: 2rem;
        padding: 0.75rem 1.25rem;
        font-family: inherit;
        font-weight: bold;
        cursor: pointer;
        transition:
            color 0.25s ease,
            border-color 0.25s ease,
            transform 0.25s ease;
    }

    button:hover {
        transform: scale(1.05);
    }

    .create-btn {
        color: #bcb594;
        border: 2px solid #d26100;
    }

    .create-btn:hover {
        border-color: #ab4f00;
        color: #ada47c;
    }

    .join-btn {
        color: #d26100;
        border: 2px solid #3a1f3b;
    }

    .join-btn:hover {
        border-color: #2d182e;
        color: #ab4f00;
    }
</style>
