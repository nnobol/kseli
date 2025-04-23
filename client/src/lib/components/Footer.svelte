<script lang="ts">
    import TooltipWrapper from "./TooltipWrapper.svelte";

    let { isErrorPage }: { isErrorPage: boolean } = $props();
</script>

<footer class:error={isErrorPage}>
    <TooltipWrapper content="Check Me Out">
        <a
            href="https://github.com/NikolozOboladze"
            target="_blank"
            rel="noopener"
        >
            <img
                src={isErrorPage
                    ? "/github-icon-error.svg"
                    : "/github-icon.svg"}
                alt="GitHub logo"
            />
        </a>
    </TooltipWrapper>
</footer>

<style>
    footer {
        background-color: var(--colog-bg);
        color: #cbc6ac;
        border-top: 2px solid #3a1f3b;
        padding: 0.5rem;
        text-align: right;
        user-select: none;
    }

    footer.error {
        background-color: #ffe4e4;
        border-top: 2px solid #ffa9a9;
    }

    footer a img {
        width: 1.8rem;
        height: auto;
        transition: transform 0.25s ease;
    }

    footer a:hover img {
        transform: scale(1.1);
    }
</style>
