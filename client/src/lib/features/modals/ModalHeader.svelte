<script lang="ts">
    import CloseButton from "../../components/buttons/CloseButton.svelte";

    interface Props {
        headerTitle: string;
        loading: boolean;
        closeModal: () => void;
    }

    let { headerTitle, loading, closeModal }: Props = $props();
</script>

<header>
    <h2 class="header-title {loading ? 'loading' : ''}">{headerTitle}</h2>

    <CloseButton onClick={closeModal} disabled={loading} buttonType="modal" />
</header>

<style>
    header {
        position: relative;
        justify-content: center;
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
