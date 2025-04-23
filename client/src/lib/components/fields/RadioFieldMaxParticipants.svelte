<script lang="ts">
    interface Props {
        disabled: boolean;
        fieldError: string;
        onChange: (value: number) => void;
    }

    let { disabled, fieldError, onChange }: Props = $props();

    let hasSelection: boolean = $state(false);

    function handleChange(value: number) {
        hasSelection = true;
        onChange(value);
    }
</script>

<fieldset {disabled} class:error={fieldError} class:active={hasSelection}>
    <legend class:error={fieldError} class:active={hasSelection}>
        Max Number of Participants
    </legend>
    <ul>
        {#each [2, 3, 4, 5] as option}
            <li>
                <label class:error={fieldError}>
                    <input
                        type="radio"
                        name="maxParticipants"
                        value={option}
                        onchange={() => handleChange(option)}
                        {disabled}
                    />
                    <span>{option}</span>
                </label>
            </li>
        {/each}
    </ul>
    {#if fieldError}
        <span>{fieldError}</span>
    {/if}
</fieldset>

<style>
    fieldset {
        border: 2px solid var(--color-modal-text);
        border-radius: 4px;
        padding: 0.5rem;
        transition: border-color 0.15s ease-in-out;
    }

    fieldset:disabled {
        opacity: 0.6;
        pointer-events: none;
    }

    fieldset.error {
        border-color: red;
    }

    fieldset.active {
        border-color: var(--color-button-light);
    }

    legend {
        padding: 0 0.5rem;
        color: var(--color-modal-text);
        border-right: 2px solid var(--color-modal-text);
        border-left: 2px solid var(--color-modal-text);
        line-height: 0.6rem;
        transition:
            color 0.15s ease-in-out,
            border-color 0.15s ease-in-out;
    }

    legend.error {
        color: red;
        border-color: red;
    }

    legend.active {
        color: var(--color-button-light);
        border-color: var(--color-button-light);
    }

    ul {
        display: flex;
        gap: 1rem;
        list-style: none;
        justify-content: center;
    }

    label {
        display: flex;
        align-items: center;
        cursor: pointer;
        color: var(--color-modal-text);
    }

    input {
        appearance: none;
        -webkit-appearance: none;
        -moz-appearance: none;
        width: 1rem;
        height: 1rem;
        border: 2px solid var(--color-modal-text);
        border-radius: 50%;
        position: relative;
        margin-right: 0.25rem;
        transition: border-color 0.15s;
        cursor: pointer;
    }

    input:hover {
        border-color: var(--color-button-light);
    }

    input:checked {
        border-color: var(--color-button-light);
    }

    input:checked::after {
        content: "";
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        width: 0.55rem;
        height: 0.55rem;
        background-color: var(--color-button-light);
        border-radius: 50%;
    }

    span {
        transition: color 0.2s ease;
    }

    label:hover span {
        color: var(--color-button-light);
    }

    input:checked + span {
        color: var(--color-button-light);
    }

    /* Error states */
    label.error {
        color: red;
    }

    label.error input {
        border-color: red;
    }

    label.error:hover span {
        color: red;
    }

    fieldset > span {
        font-size: 0.875rem;
        color: red;
        display: block;
        margin-top: 0.2rem;
    }
</style>
