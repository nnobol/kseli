<script lang="ts">
    let { isErrorPage }: { isErrorPage: boolean } = $props();
</script>

<footer class:error={isErrorPage}>
    <a href="https://github.com/NikolozOboladze" target="_blank" rel="noopener">
        <img
            src={isErrorPage ? "/github-icon-error.svg" : "/github-icon.svg"}
            alt="GitHub logo"
        />
    </a>
</footer>

<style>
    footer {
        background-color: #1f011d;
        color: #cbc6ac;
        font-size: 1rem;
        border-top: 2px solid #3a1f3b;
        padding: 0.5rem 1rem;
        display: flex;
        justify-content: flex-end;
        align-items: center;
        user-select: none;
    }

    footer.error {
        background-color: #ffe4e4;
        border-top: 2px solid #ffa9a9;
    }

    footer a img {
        width: 1.8rem;
        height: auto;
        transition: transform 0.3s ease;
    }

    footer a:hover img {
        transform: scale(1.1);
    }
</style>
