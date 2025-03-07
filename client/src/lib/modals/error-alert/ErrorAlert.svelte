<script lang="ts">
    import AlertWrapper from "./AlertWrapper.svelte";
    import CloseButton from "../buttons/CloseButton.svelte";
    import { CloseButtonTypes } from "../buttons/types";

    interface Props {
        errorMessage: string;
        clearErrorMessage: () => void;
        closable?: boolean;
    }

    let { errorMessage, clearErrorMessage, closable = true }: Props = $props();
</script>

{#snippet errorAlert()}
    <section>
        <p>{errorMessage}</p>

        {#if closable}
            <CloseButton
                onClick={clearErrorMessage}
                buttonType={CloseButtonTypes.Error}
            />
        {/if}
    </section>
{/snippet}

<AlertWrapper content={errorAlert} />

<style>
    section {
        display: flex;
        align-items: center;
        text-align: center;
        justify-content: space-between;
        gap: 1rem;

        background: #ffe4e4;
        color: #d8000c;
        border: 1px solid #d8000c;
        border-radius: 8px;
        padding: 1rem 2rem;
        max-width: 400px;
        width: 90%;
    }

    p {
        word-wrap: break-word;
        overflow-wrap: break-word;
        white-space: pre-wrap;
        word-break: break-word;
        font-weight: bold;
        font-size: 1rem;
        max-width: 300px;
    }
</style>
