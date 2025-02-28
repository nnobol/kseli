<script lang="ts">
    import type { Snippet } from "svelte";

    interface Props {
        content?: string;
        offset?: number;
        children: Snippet;
    }

    let { content = "", offset = 10, children }: Props = $props();

    let x = $state(0);
    let y = $state(0);
    let visible = $state(false);
    let tooltipRef: HTMLDivElement | undefined = $state();

    function handleMouseEnter(event: MouseEvent) {
        visible = true;
        updatePosition(event);
    }

    function handleMouseMove(event: MouseEvent) {
        updatePosition(event);
    }

    function handleMouseLeave() {
        visible = false;
    }

    function updatePosition(event: MouseEvent) {
        if (!tooltipRef) return;

        const tooltipWidth = tooltipRef.offsetWidth;
        const tooltipHeight = tooltipRef.offsetHeight;
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;

        // Default position: to the right and below the cursor
        let newX = event.clientX + offset;
        let newY = event.clientY + offset;

        // Flip to left if it would overflow right
        if (newX + tooltipWidth > viewportWidth) {
            newX = event.clientX - tooltipWidth - offset;
        }

        // Flip to above if it would overflow bottom
        if (newY + tooltipHeight > viewportHeight) {
            newY = event.clientY - tooltipHeight - offset;
        }

        newX = Math.max(0, newX);
        newY = Math.max(0, newY);

        x = newX;
        y = newY;
    }
</script>

<div
    class="wrapper"
    onmouseenter={handleMouseEnter}
    onmousemove={handleMouseMove}
    onmouseleave={handleMouseLeave}
    role="tooltip"
>
    {@render children()}

    {#if visible}
        <div
            bind:this={tooltipRef}
            class="tooltip"
            style="top: {y}px; left: {x}px;"
        >
            {content}
        </div>
    {/if}
</div>

<style>
    .wrapper {
        position: relative;
        display: inline-block;
    }

    .tooltip {
        position: fixed;
        background: rgba(0, 0, 0, 0.8);
        color: #f9f9f9;
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 0.8rem;
        white-space: nowrap;
        pointer-events: none;
        z-index: 1000;
    }
</style>
