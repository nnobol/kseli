<script lang="ts">
    import CloseButton from "../common/buttons/CloseButton.svelte";
    import { CloseButtonTypes } from "../common/buttons/types";

    interface Props {
        headerTitle: string;
        loading: boolean;
        closeModal: () => void;
    }

    let { headerTitle, loading, closeModal }: Props = $props();
</script>

<header>
    <h2 class="header-title {loading ? 'loading' : ''}">{headerTitle}</h2>

    <CloseButton
        onClick={closeModal}
        disabled={loading}
        buttonType={CloseButtonTypes.Modal}
    />
</header>

<style>
    header {
        position: relative;
        display: flex;
        justify-content: center;
        align-items: center;
        margin-bottom: 2rem;
    }

    .header-title {
        font-weight: bold;
        font-size: 1.5rem;
    }

    .header-title.loading {
        opacity: 0.6;
    }
</style>
