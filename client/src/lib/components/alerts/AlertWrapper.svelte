<script lang="ts">
    import { onMount } from "svelte";
    import type { Snippet } from "svelte";

    interface Props {
        content: Snippet;
    }

    let { content }: Props = $props();

    let dialogElement: HTMLDialogElement | null = null;

    onMount(() => {
        if (dialogElement) {
            dialogElement.showModal();
            dialogElement.focus();
        }
    });
</script>

<dialog bind:this={dialogElement}>
    {@render content()}
</dialog>

<style>
    dialog {
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
        z-index: 1000;
    }
</style>
