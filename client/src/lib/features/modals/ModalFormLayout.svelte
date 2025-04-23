<script lang="ts">
    import ModalHeader from "./ModalHeader.svelte";
    import SubmitButton from "../../components/buttons/SubmitButton.svelte";
    import { scale } from "svelte/transition";
    import type { Snippet } from "svelte";

    interface Props {
        headerTitle: string;
        buttonText: string;
        loading: boolean;
        buttonDisabled: boolean;
        closeModal: (() => void) | null;
        fields: Snippet;
        onSubmit: (event: Event) => Promise<void>;
    }

    let {
        headerTitle,
        buttonText,
        loading,
        buttonDisabled,
        closeModal,
        fields,
        onSubmit,
    }: Props = $props();
</script>

<section in:scale={{ duration: 200 }} out:scale={{ duration: 200 }}>
    <ModalHeader {headerTitle} {loading} {closeModal} />

    <form novalidate onsubmit={onSubmit}>
        <div class="fields">
            {@render fields()}
        </div>

        <SubmitButton {buttonText} {loading} {buttonDisabled} />
    </form>
</section>

<style>
    section {
        display: flex;
        flex-direction: column;
        text-align: center;
        gap: 1.5rem;

        background: var(--color-light-primary);
        color: var(--color-modal-text);
        padding: 1rem 2rem;
        border-radius: 8px;
        max-width: 25rem;
        width: 90%;
    }

    form {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
    }

    .fields {
        display: flex;
        flex-direction: column;
        width: 90%;
        gap: 1rem;
    }
</style>
