<script lang="ts">
    import { endChatSession } from "$lib/stores/chatStore";
    import { closeRoom } from "$lib/api/rooms";
    import InlineToast from "$lib/components/InlineToast.svelte";
    import TooltipWrapper from "$lib/components/TooltipWrapper.svelte";
    import { onMount, tick } from "svelte";

    interface Props {
        expiresAt: number;
        inviteLink?: string;
        currentUserRole: number;
        adminUsername: string;
    }

    let { expiresAt, inviteLink, currentUserRole, adminUsername }: Props =
        $props();

    let toasts = $state<
        {
            id: number;
            typeError: boolean;
            message: string;
            element: HTMLElement;
            visible: boolean;
            timeout: number;
        }[]
    >([]);

    let remainingTime: string = $state("Loading...");
    let intervalId: number | undefined = $state();
    let truncatedInviteLink = inviteLink?.substring(0, 25) + "...";

    function formatTime(totalSeconds: number): string {
        const minutes = Math.floor(totalSeconds / 60);
        const seconds = totalSeconds % 60;
        return `${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
    }

    function updateCountdown() {
        const now = Math.floor(Date.now() / 1000);
        const diff = expiresAt - now;

        if (diff <= 0) {
            remainingTime = "Closing...";
            clearInterval(intervalId);
        } else {
            remainingTime = formatTime(diff);
        }
    }

    onMount(() => {
        updateCountdown();
        intervalId = setInterval(updateCountdown, 1000);

        return () => {
            if (intervalId) clearInterval(intervalId);
        };
    });

    async function showToast(
        message: string,
        targetElement: HTMLElement,
        typeError: boolean,
        delay: number,
    ) {
        const toastId = Date.now();
        toasts = [
            ...toasts,
            {
                id: toastId,
                typeError,
                message,
                element: targetElement,
                visible: false,
                timeout: 0,
            },
        ];

        await tick();
        toasts = toasts.map((toast) =>
            toast.id === toastId ? { ...toast, visible: true } : toast,
        );

        setTimeout(() => {
            toasts = toasts.map((toast) =>
                toast.id === toastId ? { ...toast, visible: false } : toast,
            );
        }, delay);
    }

    function clearToast(index: number) {
        const toast = toasts[index];
        if (toast.timeout) clearTimeout(toast.timeout);
        toasts = toasts.filter((_, i) => i !== index);
    }

    async function handleClose(event: MouseEvent) {
        const targetElement = event.currentTarget as HTMLElement;
        if (toasts.some((t) => t.element === targetElement)) return;

        try {
            await closeRoom();
        } catch (err) {
            await showToast("Failed to close room.", targetElement, true, 1500);
        }
    }

    async function copyToClipboard(event: MouseEvent, value: string) {
        await navigator.clipboard.writeText(value);
        const targetElement = event.target as HTMLElement;
        if (toasts.some((t) => t.element === targetElement)) return;

        await showToast("Copied!", targetElement, false, 750);
    }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<section>
    <div class="room-info">
        <div class="room-info-block">
            <h3>Room Info</h3>
            <div class="info-line">
                <span class="username-container">
                    <span class="username-truncated">{adminUsername}</span>
                    <span>'s Room</span>
                </span>
                <span>|</span>
                <TooltipWrapper content="Remaining Time">
                    <span>{remainingTime}</span>
                </TooltipWrapper>
            </div>
        </div>

        {#if currentUserRole === 1 && inviteLink}
            <div class="room-info-block">
                <h3>Invite Link</h3>
                <div class="info-line invite-container">
                    <span
                        class="copyable invite-truncated"
                        onclick={(e) => copyToClipboard(e, inviteLink)}
                    >
                        {inviteLink}
                    </span>
                </div>
            </div>
        {/if}
    </div>
    <div class="button-wrapper">
        {#if currentUserRole === 1}
            <button onclick={(e) => handleClose(e)}>Close Room</button>
        {:else}
            <button onclick={() => endChatSession()}>Leave Room</button>
        {/if}
    </div>

    {#each toasts as toast, index (toast.id)}
        <InlineToast
            typeError={toast.typeError}
            message={toast.message}
            targetElement={toast.element}
            visible={toast.visible}
            removeToast={() => clearToast(index)}
        />
    {/each}
</section>

<style>
    section {
        display: flex;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 0.3rem;
        color: #24292f;
        background-color: #fff;
        gap: 0.75rem;
    }

    .room-info {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .room-info-block {
        align-items: center;
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
    }

    .room-info-block h3 {
        font-size: 0.8rem;
        text-transform: uppercase;
        letter-spacing: 0.05rem;
        color: #24292f;
    }

    .info-line {
        display: flex;
        border-bottom: 1px solid #ccc;
        padding-bottom: 0.25rem;
        gap: 0.5rem;
        font-size: 0.8rem;
        width: 100%;
        justify-content: center;
    }

    .username-container {
        display: flex;
        min-width: 0;
        white-space: nowrap;
    }

    .username-truncated {
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .invite-container {
        max-width: 200px;
        overflow: hidden;
    }

    .invite-truncated {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .copyable {
        cursor: pointer;
    }

    .copyable:hover {
        background-color: #e0e0e0;
    }

    .copyable:active {
        background-color: #b0b0b0;
    }

    .button-wrapper {
        display: flex;
        justify-content: center;
        gap: 0.35rem;
    }

    button {
        font-size: 0.8rem;
        font-family: inherit;
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
