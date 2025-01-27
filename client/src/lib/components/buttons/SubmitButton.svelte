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
        background-color: #d26100;
        color: #bcb594;
        border: none;
        border-radius: 5px;
        cursor: pointer;
        font-size: 1rem;
        padding: 0.5rem 1rem;
        font-family: inherit;
        font-weight: bold;
        transition: opacity 0.2s ease;
        margin-top: 1.5rem;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-width: 6rem;
        min-height: 2.2rem;
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
        border-top: 2px solid #ffffff;
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
