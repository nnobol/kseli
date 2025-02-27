<script lang="ts">
    import TooltipWrapper from "$lib/common/TooltipWrapper.svelte";

    interface Props {
        secretKey: string;
        currentUserRole: number;
    }

    let { secretKey, currentUserRole }: Props = $props();
</script>

<section>
    {#if currentUserRole == 1}
        <TooltipWrapper content="Secret Key for joining the Room">
            <p>Key: {secretKey}</p>
        </TooltipWrapper>
    {/if}
    <div class="button-wrapper">
        <button>Leave Room</button>
        {#if currentUserRole == 1}
            <button>Close Room</button>
        {/if}
    </div>
</section>

<style>
    section {
        display: flex;
        flex-grow: 0;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 0.5rem;
        color: #24292f;
        background-color: #fff;
    }

    p {
        text-align: center;
    }

    .button-wrapper {
        display: flex;
        justify-content: center;
        gap: 0.35rem;
        margin-top: 0.5rem;
    }

    button {
        padding: 0.5rem 1rem;
        background-color: #24292f;
        color: #f9f9f9;
        border: none;
        border-radius: 4px;
        cursor: pointer;
    }

    button:hover {
        opacity: 0.8;
    }
</style>
