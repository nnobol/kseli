<script lang="ts">
    import { page } from "$app/state";
    import { goto } from "$app/navigation";
</script>

{#if page.status === 404}
    <h1>404 - Page Not Found</h1>
    <p>Page '{page.url.pathname}' doesn’t exist.</p>
{:else}
    <h1>Error {page.status || 500}</h1>
    <p>{page.error?.message || "Something went wrong"}</p>
{/if}
<button on:click={() => goto("/")}>Go Home</button>
