<script lang="ts">
    import { errorStore } from "$lib/stores/errorStore";
    import ErrorAlert from "$lib/modals/error-alert/ErrorAlert.svelte";
</script>

{#if $errorStore}
    <ErrorAlert
        errorMessage={$errorStore}
        clearErrorMessage={() => {}}
        closable={false}
    />
{/if}
