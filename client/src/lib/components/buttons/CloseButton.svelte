<script lang="ts">
    interface Props {
        onClick: () => void;
        disabled?: boolean;
        buttonType: "modal" | "error";
    }

    let { onClick, disabled = false, buttonType }: Props = $props();
</script>

<button onclick={onClick} {disabled} class={`${buttonType}`}> ✖ </button>

<style>
    button {
        background: none;
        border: none;
        cursor: pointer;
        color: #32012f;
        font-size: 1.5rem;
        font-weight: bold;
        transition: opacity 0.2s ease;
    }

    button:hover {
        opacity: 0.8;
    }

    button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    /* Styles for modal close button */
    .modal {
        position: absolute;
        top: 50%;
        right: 0rem;
        transform: translateY(-50%);
        color: #32012f;
    }

    /* Styles for error alert close button */
    .error {
        color: #d8000c;
    }
</style>
