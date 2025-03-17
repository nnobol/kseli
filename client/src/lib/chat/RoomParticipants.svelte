<script lang="ts">
    import { kickUser, banUser, type Participant } from "$lib/api/rooms";
    import TooltipWrapper from "$lib/common/TooltipWrapper.svelte";
    import InlineErrorToast from "$lib/common/InlineErrorToast.svelte";

    interface Props {
        currentUserRole: number;
        participants: Participant[];
        maxParticipants: number;
    }

    let { currentUserRole, participants, maxParticipants }: Props = $props();

    let errorMessage: string | null = $state(null);
    let errorElement: HTMLElement | null = $state(null);

    function getRoleIcon(role: number) {
        return role === 1 ? "/admin-icon.svg" : "/user-icon.svg";
    }

    function getRoleTitle(role: number) {
        return role === 1 ? "Admin" : "User";
    }

    async function handleKick(userId: number, event: MouseEvent) {
        const targetElement = event.currentTarget as HTMLElement;
        try {
            await kickUser({ userId });
            errorMessage = null;
        } catch (err) {
            errorMessage = "Failed to kick user.";
            errorElement = targetElement;
        }
    }

    // handle errors and display them somehow
    async function handleBan(userId: number, event: MouseEvent) {
        const targetElement = event.currentTarget as HTMLElement;
        try {
            await banUser({ userId });
            errorMessage = null;
        } catch (err) {
            errorMessage = "Failed to ban user.";
            errorElement = targetElement;
        }
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
                                onclick={(e) => handleKick(participant.id, e)}
                            >
                                <img src="/kick-icon.svg" alt="Kick" />
                            </button>
                        </TooltipWrapper>
                        <TooltipWrapper content="Ban User">
                            <button
                                onclick={(e) => handleBan(participant.id, e)}
                            >
                                <img src="/ban-icon.svg" alt="Ban" />
                            </button>
                        </TooltipWrapper>
                    </div>
                {/if}
            </li>
        {/each}
    </ul>

    {#if errorMessage && errorElement}
        <InlineErrorToast
            message={errorMessage}
            targetElement={errorElement}
            clearError={() => {
                errorMessage = null;
                errorElement = null;
            }}
        />
    {/if}
</section>

<style>
    section {
        display: flex;
        flex: 1;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 1rem;
        background-color: #fff;
    }

    h2 {
        color: #24292f;
        padding-bottom: 0.25rem;
        margin-bottom: 0.25rem;
        border-bottom: 1px solid #ddd;
        font-size: 1.25rem;
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
        padding: 0.3rem;
    }

    .participant-info {
        display: flex;
        gap: 0.2rem;
    }

    .role-icon {
        width: 1rem;
        height: 1rem;
        display: block;
    }

    span {
        line-height: 1rem;
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
        width: 1rem;
        height: 1rem;
    }
</style>
