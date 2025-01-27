<script lang="ts">
    interface Props {
        id: string;
        type: string;
        labelText: string;
        disabled: boolean;
        value: string;
        fieldError: string;
        onInput: () => void;
    }

    let {
        id,
        type,
        labelText,
        disabled,
        value = $bindable(""),
        fieldError,
        onInput,
    }: Props = $props();
</script>

<fieldset {disabled}>
    <div>
        <input
            bind:value
            {id}
            {type}
            placeholder=""
            oninput={onInput}
            class:error={fieldError}
        />
        <label for={id} class:error={fieldError}>{labelText}</label>
    </div>
    {#if fieldError}
        <span>{fieldError}</span>
    {/if}
</fieldset>

<style>
    fieldset {
        border: none;
        padding: 0;
        margin-bottom: 0.8rem;
        position: relative;
    }

    fieldset:disabled {
        opacity: 0.6;
        pointer-events: none;
    }

    div {
        position: relative;
        display: flex;
        flex-direction: column;
    }

    input {
        width: 100%;
        padding: 0.5rem 0.5rem;
        font-size: 1rem;
        border: 2px solid #32012f;
        border-radius: 4px;
        outline: none;
        background: transparent;
        color: #32012f;
        font-family: inherit;
        transition: border-color 0.15s ease-in-out;
    }

    label {
        /* Position the label over the input initially */
        position: absolute;
        top: 50%;
        left: 0.5rem;
        transform: translateY(-50%);

        pointer-events: none;
        font-size: 1rem;
        color: #32012f;
        background: #cbc6ac;
        padding: 0 0.2rem;
        transition:
            top 0.15s ease-in-out,
            left 0.15s ease-in-out,
            font-size 0.15s ease-in-out,
            color 0.15s ease-in-out;
    }

    span {
        font-size: 0.875rem;
        color: red;
        display: block;
    }

    /* NORMAL STATE TRANSFORMATIONS */
    /* Change input border color when focused or has content (not in error state) */
    input:focus:not(.error),
    input:not(:placeholder-shown):not(.error) {
        border-color: #d26100;
    }

    /* Transform label when input is focused or has content (not in error state) */
    input:focus + label:not(.error),
    input:not(:placeholder-shown) + label:not(.error) {
        top: 0;
        left: 0.4rem;
        font-size: 0.75rem;
        color: #d26100;
        border-right: 1px solid #cbc6ac;
        border-left: 1px solid #cbc6ac;
    }

    /* ERROR STATE TRANSFORMATIONS */
    /* Basic error states */
    input.error {
        border-color: red;
    }
    input.error + label {
        color: red;
    }

    /* Transform label when input is focused or has content (error state) */
    input.error:focus + label,
    input.error:not(:placeholder-shown) + label {
        top: 0;
        left: 0.4rem;
        font-size: 0.75rem;
        border-right: 1px solid #cbc6ac;
        border-left: 1px solid #cbc6ac;
    }

    /* BORDER BITS - BASE STYLES */
    input:focus + label::before,
    input:focus + label::after,
    input:not(:placeholder-shown) + label::before,
    input:not(:placeholder-shown) + label::after {
        content: "";
        position: absolute;
        width: 2px;
        height: 0.5rem;
        top: 50%;
        transform: translateY(-50%);
    }

    /* BORDER BITS POSITIONING */
    /* Before element (left side) */
    input:focus + label::before,
    input:not(:placeholder-shown) + label::before {
        left: -0.1rem;
    }
    /* After element (right side) */
    input:focus + label::after,
    input:not(:placeholder-shown) + label::after {
        right: -0.1rem;
    }

    /* BORDER BITS - NORMAL STATE */
    input:focus + label:not(.error)::before,
    input:not(:placeholder-shown) + label:not(.error)::before,
    input:focus + label:not(.error)::after,
    input:not(:placeholder-shown) + label:not(.error)::after {
        background-color: #d26100;
    }

    /* BORDER BITS - ERROR STATE */
    input.error:focus + label::before,
    input.error:not(:placeholder-shown) + label::before,
    input.error:focus + label::after,
    input.error:not(:placeholder-shown) + label::after {
        background-color: red;
    }
</style>
