<script lang="ts">
    import CloseButton from "../../components/buttons/CloseButton.svelte";

    interface Props {
        headerTitle: string;
        loading: boolean;
        closeModal: (() => void) | null;
    }

    let { headerTitle, loading, closeModal }: Props = $props();
</script>

<div class="header">
    <h2 class="header-title {loading ? 'loading' : ''}">{headerTitle}</h2>

    {#if !!closeModal}
        <CloseButton
            onClick={closeModal}
            disabled={loading}
            buttonType="modal"
        />
    {/if}
</div>

<style>
    .header {
        position: relative;
        justify-content: center;
    }

    .header-title {
        font-weight: var(--font-weight-bold);
        font-size: 1.5rem;
    }
</style>
