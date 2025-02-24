<script lang="ts">
    import CreateChatRoomModal from "$lib/components/modals/CreateChatRoomModal.svelte";
    import JoinChatRoomModal from "$lib/components/modals/JoinChatRoomModal.svelte";

    let isCreateModalOpen: boolean = false;
    let isJoinModalOpen: boolean = false;

    function openCreateModal(): void {
        isCreateModalOpen = true;
    }

    function closeCreateModal(): void {
        isCreateModalOpen = false;
    }

    function openJoinModal(): void {
        isJoinModalOpen = true;
    }

    function closeJoinModal(): void {
        isJoinModalOpen = false;
    }
</script>

<main>
    <h1>
        <span class="highlight">Kseli</span> - private, secure and fast communication.
    </h1>

    <div class="button-container">
        <button
            class="join-chat-room"
            onclick={() => {
                if (!isCreateModalOpen) openJoinModal();
            }}
        >
            Join Chat Room
        </button>

        <button
            class="create-chat-room"
            onclick={() => {
                if (!isJoinModalOpen) openCreateModal();
            }}
        >
            Create Chat Room
        </button>
    </div>

    {#if isCreateModalOpen}
        <CreateChatRoomModal closeModal={closeCreateModal} />
    {/if}

    {#if isJoinModalOpen}
        <JoinChatRoomModal closeModal={closeJoinModal} />
    {/if}
</main>

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

    .button-container {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .create-chat-room {
        background-color: #1f011d;
        color: #bcb594;
        border: 2px solid #d26100;
        border-radius: 5px;
        font-size: 2rem;
        padding: 0.75rem 1.5rem;
        font-family: inherit;
        font-weight: bold;
        transition:
            color 0.3s ease,
            border-color 0.3s ease,
            transform 0.3s ease;
        cursor: pointer;
    }

    .create-chat-room:hover {
        transform: scale(1.05);
        border-color: #ab4f00;
        color: #ada47c;
    }

    .join-chat-room {
        background-color: #1f011d;
        color: #d26100;
        border: 2px solid #3a1f3b;
        border-radius: 5px;
        font-size: 2rem;
        padding: 0.75rem 1.5rem;
        font-family: inherit;
        font-weight: bold;
        transition:
            color 0.3s ease,
            border-color 0.3s ease,
            transform 0.3s ease;
        cursor: pointer;
    }

    .join-chat-room:hover {
        transform: scale(1.05);
        border-color: #2d182e;
        color: #ab4f00;
    }
</style>
