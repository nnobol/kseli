<script lang="ts">
    import type { GetRoomOkResponse } from "$lib/api/rooms";
    import ChatContainer from "$lib/features/chat/ChatContainer.svelte";
    import RoomSidebar from "$lib/features/chat/RoomSidebar.svelte";
    import { onMount } from "svelte";

    interface Props {
        data: GetRoomOkResponse;
    }

    let { data }: Props = $props();

    let keyboardOpen = $state(false);
    let initialHeight = window.innerHeight;
    let page: HTMLElement;

    onMount(() => {
        if (window.innerWidth >= 768) return;

        const resizeDueToKeyboard = () => {
            const visualHeight =
                window.visualViewport?.height ?? window.innerHeight;
            keyboardOpen = visualHeight < initialHeight * 0.7;
            document.documentElement.style.setProperty(
                "--h",
                `${visualHeight}px`,
            );
            if (page) {
                page.style.transform = `translateY(${window.visualViewport?.offsetTop}px)`;
            }
        };

        window.visualViewport?.addEventListener("resize", resizeDueToKeyboard);

        resizeDueToKeyboard();

        return () => {
            window.visualViewport?.removeEventListener(
                "resize",
                resizeDueToKeyboard,
            );
        };
    });
</script>

<main bind:this={page}>
    <ChatContainer />
    {#if !keyboardOpen}
        <RoomSidebar
            currentUserRole={data.userRole}
            maxParticipants={data.maxParticipants}
            expiresAt={data.expiresAt}
            inviteLink={data.inviteLink}
        />
    {/if}
</main>

<style>
    main {
        display: flex;
        margin: 0 auto;
        width: 100%;
        transition: transform 0.25s ease-in-out;
    }

    @media (min-width: 769px) {
        main {
            flex: 1;
            padding: 0.75rem;
            gap: 0.5rem;
            max-width: 1440px;
        }
    }

    @media (max-width: 768px) {
        main {
            flex-direction: column;
            padding: 0.375rem;
            gap: 0.25rem;
            height: var(--h);
        }
    }
</style>
