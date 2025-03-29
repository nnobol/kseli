<script lang="ts">
    import { scale } from "svelte/transition";

    interface Props {
        typeError?: boolean;
        message: string;
        targetElement: HTMLElement;
        visible: boolean;
        removeToast: () => void;
    }

    let { typeError, message, targetElement, visible, removeToast }: Props =
        $props();

    let x = $state(0);
    let y = $state(0);
    let toastRef: HTMLDivElement | undefined = $state();

    function calculateToastPosition() {
        if (!targetElement || !toastRef) return;

        const rect = targetElement.getBoundingClientRect();
        const toastWidth = toastRef.offsetWidth;
        const toastHeight = toastRef.offsetHeight;
        const viewportWidth = window.innerWidth;
        const viewportMargin = 10;

        // Default position: centered and a bit above the target element
        let newX = rect.x + rect.width / 2;
        let newY = rect.y - toastHeight - 5;

        // Make sure X does not pass the right side of the viewport
        // Taking in translateX(-50%) into account with (toastWidth / 2)
        const rightEdge = newX + toastWidth / 2;
        if (rightEdge > viewportWidth - viewportMargin) {
            newX = viewportWidth - viewportMargin - toastWidth / 2;
        }

        x = newX;
        y = newY;
    }

    $effect(() => {
        if (visible && toastRef) calculateToastPosition();
    });
</script>

{#if visible}
    <div
        class:error={typeError}
        class="toast"
        in:scale={{ duration: 100 }}
        out:scale={{ duration: 100 }}
        onoutroend={() => removeToast()}
        bind:this={toastRef}
        style="top: {y}px; left: {x}px;"
    >
        <p>{message}</p>
    </div>
{/if}

<style>
    .toast {
        position: fixed;
        background: rgba(36, 41, 47, 0.85);
        color: #f9f9f9;
        padding: 4px 8px;
        border-radius: 5px;
        font-size: 0.8rem;
        z-index: 1000;
        transform: translateX(-50%);
    }

    .toast.error {
        background: rgba(216, 0, 12, 0.85);
        color: #ffe4e4;
    }

    p {
        white-space: nowrap;
    }
</style>
