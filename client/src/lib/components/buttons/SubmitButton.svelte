<script lang="ts">
    interface Props {
        buttonText: string;
        loading: boolean;
        buttonDisabled: boolean;
    }

    let { buttonText, loading, buttonDisabled }: Props = $props();
</script>

<button class="submit-button" type="submit" disabled={buttonDisabled}>
    {#if loading}
        <span class="spinner"></span>
    {:else}
        {buttonText}
    {/if}
</button>

<style>
    .submit-button {
        background-color: var(--color-modal-text);
        color: var(--color-text);
        border: none;
        border-radius: 50px;
        cursor: pointer;
        font-size: 1rem;
        padding: 0.8rem 1.5rem;
        font-family: inherit;
        font-weight: var(--font-weight-bold);
        transition: opacity 0.2s ease;

        /* for the spinner */
        min-width: 6.75rem;
        min-height: 2.75rem;
        justify-content: center;
        display: inline-flex;
    }

    .submit-button:hover {
        opacity: 0.8;
    }

    .submit-button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .spinner {
        border: 2px solid rgba(255, 255, 255, 0.3);
        border-top: 2px solid var(--color-text);
        border-radius: 50%;
        width: 1rem;
        height: 1rem;
        animation: spin 1s linear infinite;
    }

    @keyframes spin {
        0% {
            transform: rotate(0deg);
        }
        100% {
            transform: rotate(360deg);
        }
    }
</style>
