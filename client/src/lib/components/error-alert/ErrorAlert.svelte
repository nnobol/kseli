<script lang="ts">
    import AlertWrapper from "./AlertWrapper.svelte";
    import CloseButton from "../buttons/CloseButton.svelte";

    interface Props {
        errorMessage: string;
        clearErrorMessage: () => void;
    }

    let { errorMessage, clearErrorMessage }: Props = $props();
</script>

{#snippet errorAlert()}
    <section>
        <p>{errorMessage}</p>

        <CloseButton onClick={clearErrorMessage} buttonType="error" />
    </section>
{/snippet}

<AlertWrapper content={errorAlert} />

<style>
    section {
        display: flex;
        align-items: center;
        text-align: center;
        gap: 1.5rem;

        background: #ffe4e4;
        color: #d8000c;
        border: 1px solid #d8000c;
        border-radius: 8px;
        padding: 1rem 2rem;
        max-width: 20rem;
        width: 90%;
    }

    p {
        word-wrap: break-word;
        overflow-wrap: break-word;
        white-space: pre-wrap;
        word-break: break-word;
        text-transform: uppercase;
        font-weight: var(--font-weight-bold);
        font-size: 1rem;
        max-width: 15rem;
    }
</style>
