<script lang="ts">
    import "../styles/global.css";
    import GlobalErrorHandler from "$lib/modals/error-alert/GlobalErrorHandler.svelte";

    let { children } = $props();
</script>

<GlobalErrorHandler />

{@render children()}
