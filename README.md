# Source code for: https://kseli.app

This is the full source code for **kseli** — a real-time chat room web app written in **Go** and **SvelteKit**.

The goal of the project was a very simple, fast and anonymous experience when chatting. Features include full client side end to end encryption and a server that does not persist any data and keeps everything in memory.

## Stack and Tools

- **Frontend:** SvelteKit + TypeScript
- **Backend:** Go (using only the standard library + [gobwas/ws](https://github.com/gobwas/ws) for WebSockets)