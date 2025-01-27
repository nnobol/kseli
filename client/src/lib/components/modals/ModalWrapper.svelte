<script lang="ts">
  import { onMount } from "svelte";
  import type { Snippet } from "svelte";
  import { fade } from "svelte/transition";

  interface Props {
    closeModal: () => void;
    content: Snippet<[{ closeModal: () => void }]>;
  }

  let { closeModal, content }: Props = $props();

  let dialogElement: HTMLDialogElement | null = null;

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      closeModal();
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      closeModal();
    }
  }

  onMount(() => {
    if (dialogElement) {
      dialogElement.showModal();
      dialogElement.focus();
    }
  });
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions  - -->
<dialog
  bind:this={dialogElement}
  onkeydown={handleKeyDown}
  onclick={handleBackdropClick}
  transition:fade={{ duration: 200 }}
>
  {@render content({ closeModal })}
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
