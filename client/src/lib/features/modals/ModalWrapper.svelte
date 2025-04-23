<script lang="ts">
  import { onMount } from "svelte";
  import type { Snippet } from "svelte";
  import { fade } from "svelte/transition";

  interface Props {
    loading: boolean;
    closeModal: () => void;
    content: Snippet<[{ closeModal: () => void }]>;
  }

  let { loading, closeModal, content }: Props = $props();

  let modal: HTMLDivElement | null = null;

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget && !loading) {
      closeModal();
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape" && !loading) {
      closeModal();
    }
  }

  function trapFocus(e: FocusEvent) {
    if (!modal) return;
    const nextFocusedElement = e.relatedTarget as Node | null;
    if (nextFocusedElement && !modal.contains(nextFocusedElement)) {
      modal.focus();
    }
  }

  onMount(() => {
    modal?.focus();
  });
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions  - -->
<!-- svelte-ignore a11y_no_noninteractive_tabindex  - -->
<div
  class="modal-backdrop"
  bind:this={modal}
  onkeydown={handleKeyDown}
  onclick={handleBackdropClick}
  onfocusout={trapFocus}
  transition:fade={{ duration: 200 }}
  tabindex="0"
  role="dialog"
>
  {@render content({ closeModal })}
</div>

<style>
  .modal-backdrop {
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
