<script lang="ts">
    import { onMount } from "svelte";
    import type { Snippet } from "svelte";

    interface Props {
        content: Snippet;
    }

    let { content }: Props = $props();

    let alert: HTMLDivElement | null = null;

    function trapFocus(e: FocusEvent) {
        if (!alert) return;
        const nextFocusedElement = e.relatedTarget as Node | null;
        if (nextFocusedElement && !alert.contains(nextFocusedElement)) {
            alert.focus();
        }
    }

    onMount(() => {
        alert?.focus();
    });
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex  - -->
<div
    class="alert"
    bind:this={alert}
    onfocusout={trapFocus}
    tabindex="0"
    role="alertdialog"
>
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
