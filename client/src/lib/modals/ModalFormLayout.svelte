<script lang="ts">
    import ModalHeader from "./ModalHeader.svelte";
    import SubmitButton from "./buttons/SubmitButton.svelte";
    import { scale } from "svelte/transition";
    import type { Snippet } from "svelte";

    interface Props {
        headerTitle: string;
        buttonText: string;
        loading: boolean;
        buttonDisabled: boolean;
        closeModal: () => void;
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
        {@render fields()}

        <SubmitButton {buttonText} {loading} {buttonDisabled} />
    </form>
</section>

<style>
    section {
        text-align: center;

        background: #cbc6ac;
        color: #32012f;
        padding: 1rem 2rem;
        border-radius: 8px;
        max-width: 500px;
        width: 90%;
    }
</style>
