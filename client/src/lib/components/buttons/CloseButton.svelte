<script lang="ts">
    interface Props {
        onClick: () => void;
        disabled?: boolean;
        buttonType: "modal" | "error";
    }

    let { onClick, disabled = false, buttonType }: Props = $props();
</script>

<button onclick={onClick} {disabled} class={`${buttonType}`} aria-label="Close">
    <svg
        viewBox="-0.04 -0.04 41.85 41.84"
        fill="currentColor"
        style="width: 1.2rem;"
    >
        <path
            d="M27.948 20.878 40.291 8.536a5 5 0 1 0-7.07-7.071L20.878 13.809 8.535 1.465a5 5 0 0 0-7.07 7.071l12.342 12.342L1.465 33.22a5 5 0 1 0 7.07 7.071l12.343-12.342 12.343 12.343c.976.977 2.256 1.464 3.535 1.464s2.56-.487 3.535-1.464a5 5 0 0 0 0-7.071z"
        />
    </svg>
</button>

<style>
    button {
        background: none;
        border: none;
        cursor: pointer;
        transition: opacity 0.2s ease;
    }

    button:hover {
        opacity: 0.8;
    }

    button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    /* modal close button */
    .modal {
        position: absolute;
        top: -0.25rem;
        right: -0.85rem;
        color: var(--color-modal-text);
    }

    /* error alert close button */
    .error {
        color: #d8000c;
    }
</style>
