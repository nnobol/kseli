<script lang="ts">
    import { onMount } from "svelte";
    import type { Snippet } from "svelte";

    interface Props {
        content: Snippet;
    }

    let { content }: Props = $props();

    let alert: HTMLDivElement | null = null;

    function trapFocus() {
        if (!alert) return;

        const focusedElement = document.activeElement;
        if (!alert.contains(focusedElement)) {
            alert.focus();
        }
    }

    onMount(() => {
        alert?.focus();

        document.addEventListener("focusin", trapFocus);
        return () => {
            document.removeEventListener("focusin", trapFocus);
        };
    });
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex  - -->
<div class="alert" bind:this={alert} tabindex="0" role="alert">
    {@render content()}
</div>

<style>
    .alert {
        all: unset;
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: rgba(0, 0, 0, 0.4);
        display: flex;
        justify-content: center;
        align-items: center;
        border: none;
        backdrop-filter: blur(3px);
        z-index: 2000;
    }
</style>
