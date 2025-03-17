<script lang="ts">
    import { onMount } from "svelte";

    interface Props {
        message: string;
        targetElement: HTMLElement;
        clearError: () => void;
    }

    let { message, targetElement, clearError }: Props = $props();

    let x = $state(0);
    let y = $state(0);
    let visible = $state(false);
    let toastRef: HTMLDivElement | undefined = $state();
    let timeout = 0;

    function updatePosition() {
        if (!targetElement || !toastRef) return;

        const rect = targetElement.getBoundingClientRect();
        const toastWidth = toastRef.offsetWidth;
        const toastHeight = toastRef.offsetHeight;
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;
        const viewportMargin = 10;

        // Default position: centered above the target element
        let newX = rect.left + rect.width / 2;
        let newY = rect.top - toastHeight - 10;

        // Flip to below if it would overflow the top (considering the margin)
        if (newY < viewportMargin) {
            newY = rect.bottom + 10;
        }

        // Clamp Y to ensure it stays within viewport bounds with margin
        newY = Math.max(viewportMargin, newY);
        newY = Math.min(viewportHeight - toastHeight - viewportMargin, newY);

        // Clamp X to ensure it stays within viewport bounds with margin (considering translateX(-50%))
        const minX = toastWidth / 2 + viewportMargin;
        const maxX = viewportWidth - toastWidth / 2 - viewportMargin;
        newX = Math.max(minX, Math.min(maxX, newX));

        x = newX;
        y = newY;
    }

    onMount(() => {
        updatePosition();
        visible = true;
        timeout = setTimeout(() => {
            visible = false;
            clearError();
        }, 2000);
    });

    $effect(() => {
        if (visible && toastRef) {
            updatePosition();
        }
    });

    function closeManually() {
        visible = false;
        clearTimeout(timeout);
        clearError();
    }
</script>

{#if visible}
    <div
        bind:this={toastRef}
        class="error-toast"
        style="top: {y}px; left: {x}px;"
    >
        <p>{message}</p>
        <button class="close-btn" onclick={closeManually}>✖</button>
    </div>
{/if}

<style>
    .error-toast {
        position: fixed;
        background: rgba(255, 0, 0, 0.85);
        color: #f9f9f9;
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 0.8rem;
        display: flex;
        align-items: center;
        gap: 8px;
        white-space: nowrap;
        z-index: 1000;
        transform: translateX(-50%);
    }

    .close-btn {
        background: none;
        border: none;
        color: #f9f9f9;
        font-size: 1rem;
        cursor: pointer;
    }

    .close-btn:hover {
        opacity: 0.8;
    }
</style>
