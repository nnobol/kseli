<script lang="ts">
    import { kickUser, banUser, type Participant } from "$lib/api/rooms";
    import TooltipWrapper from "$lib/components/TooltipWrapper.svelte";
    import InlineToast from "$lib/components/InlineToast.svelte";
    import { tick } from "svelte";

    interface Props {
        currentUserRole: number;
        participants: Participant[];
        maxParticipants: number;
    }

    let { currentUserRole, participants, maxParticipants }: Props = $props();

    let errors = $state<
        {
            id: number;
            message: string;
            element: HTMLElement;
            visible: boolean;
            timeout: number;
        }[]
    >([]);

    function getRoleIcon(role: number) {
        return role === 1 ? "/admin-icon.svg" : "/user-icon.svg";
    }

    function getRoleTitle(role: number) {
        return role === 1 ? "Admin" : "User";
    }

    async function handleAction(
        userId: number,
        event: MouseEvent,
        action: "kick" | "ban",
    ) {
        const targetElement = event.currentTarget as HTMLElement;
        if (errors.some((e) => e.element === targetElement)) return;

        try {
            if (action === "kick") {
                await kickUser({ userId });
            } else if (action === "ban") {
                await banUser({ userId });
            }
        } catch (err) {
            const errorId = Date.now();

            errors = [
                ...errors,
                {
                    id: errorId,
                    message:
                        action === "kick"
                            ? "Failed to kick user."
                            : "Failed to ban user.",
                    element: targetElement,
                    visible: false,
                    timeout: 0,
                },
            ];

            await tick();

            errors = errors.map((error) =>
                error.id === errorId ? { ...error, visible: true } : error,
            );

            setTimeout(() => {
                errors = errors.map((error) =>
                    error.id === errorId ? { ...error, visible: false } : error,
                );
            }, 1500);
        }
    }

    function clearError(index: number) {
        const error = errors[index];
        if (error.timeout) clearTimeout(error.timeout);
        errors = errors.filter((_, i) => i !== index);
    }
</script>

<section>
    <h2>
        Participants ({participants.length}/{maxParticipants})
    </h2>
    <ul>
        {#each participants as participant}
            <li>
                <div class="participant-info">
                    <TooltipWrapper content={getRoleTitle(participant.role)}>
                        <img
                            class="role-icon"
                            src={getRoleIcon(participant.role)}
                            alt={getRoleTitle(participant.role)}
                        />
                    </TooltipWrapper>
                    <span>{participant.username}</span>
                </div>

                {#if currentUserRole === 1 && participant.role !== 1}
                    <div class="admin-buttons">
                        <TooltipWrapper content="Kick User">
                            <button
                                onclick={(e) =>
                                    handleAction(participant.id, e, "kick")}
                            >
                                <img src="/kick-icon.svg" alt="Kick" />
                            </button>
                        </TooltipWrapper>
                        <TooltipWrapper content="Ban User">
                            <button
                                onclick={(e) =>
                                    handleAction(participant.id, e, "ban")}
                            >
                                <img src="/ban-icon.svg" alt="Ban" />
                            </button>
                        </TooltipWrapper>
                    </div>
                {/if}
            </li>
        {/each}
    </ul>

    {#each errors as error, index (error.id)}
        <InlineToast
            typeError={true}
            message={error.message}
            targetElement={error.element}
            visible={error.visible}
            removeToast={() => clearError(index)}
        />
    {/each}
</section>

<style>
    section {
        display: flex;
        flex: 1;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 0.5rem;
        background-color: #fff;
    }

    h2 {
        color: #24292f;
        padding-bottom: 0.25rem;
        margin-bottom: 0.25rem;
        border-bottom: 1px solid #ddd;
        font-size: 1rem;
        text-align: center;
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;
    }

    li {
        color: #24292f;
        display: flex;
        justify-content: space-between;
        gap: 1rem;
        padding: 0.2rem;
    }

    .participant-info {
        display: flex;
        gap: 0.2rem;
    }

    .role-icon {
        width: 0.8rem;
        height: 0.8rem;
        display: block;
    }

    span {
        font-size: 0.8rem;
        line-height: 0.8rem;
    }

    .admin-buttons {
        display: flex;
        gap: 0.25rem;
    }

    button {
        display: flex;
        background: none;
        border: none;
        cursor: pointer;
        padding: 0;
    }

    button:hover {
        transform: scale(1.1);
    }

    button img {
        width: 0.8rem;
        height: 0.8rem;
    }
</style>
