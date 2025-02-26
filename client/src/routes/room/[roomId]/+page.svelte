<script lang="ts">
    let { data } = $props();

    function getRoleIcon(role: number) {
        return role === 1 ? "/admin-icon.svg" : "/user-icon.svg";
    }
</script>

<main>
    <!-- Flex container to create two sections -->
    <div class="container">
        <!-- Left Side: Chat Window & Input -->
        <section class="chat">
            <div class="chat-window">
                <!-- Simulated Chat Messages -->
                <p>User1: Hello!</p>
                <p>User2: Hey, what's up?</p>
                <p>User3: Let's start the meeting.</p>
                <!-- More messages here -->
            </div>

            <!-- Message Input -->
            <div class="message-input">
                <input type="text" placeholder="Type a message..." />
                <button>Send</button>
            </div>
        </section>

        <!-- Right Side: Participants List -->
        <aside class="participants">
            <h2>
                Participants ({data.participants.length}/{data.maxParticipants})
            </h2>
            <ul>
                {#each data.participants as participant}
                    <li class="participant-item">
                        <img
                            class="role-icon"
                            src={getRoleIcon(participant.role)}
                            alt="Role Icon"
                        />
                        <span>{participant.username}</span>

                        {#if data.userRole === 1 && participant.role !== 1}
                            <div class="admin-buttons">
                                <button class="admin-btn kick">
                                    <img src="/kick-icon.svg" alt="Kick" />
                                </button>
                                <button class="admin-btn ban">
                                    <img src="/ban-icon.svg" alt="Ban" />
                                </button>
                            </div>
                        {/if}
                    </li>
                {/each}
            </ul>
        </aside>
    </div>
</main>

<style>
    /* Global Page Styles */
    main {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 1rem;
        /* max-width: 1200px; */
        /* margin: auto; */
    }

    /* Two-column Layout */
    .container {
        display: flex;
        width: 100%;
        height: 100%;
        /* max-width: 1200px; */
        gap: 0.5rem;
    }

    /* Left Side: Chat */
    .chat {
        flex: 3; /* Takes up more space */
        display: flex;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 1rem;
        background-color: #f9f9f9;
    }

    .chat-window {
        flex: 1;
        overflow-y: auto;
        padding-bottom: 1rem;
        padding-top: 0.25rem;
        border-bottom: 2px solid #ccc;
        border-top: 2px solid #ccc;
        height: 100%;
    }

    .chat-window p {
        color: #24292f;
    }

    .message-input {
        display: flex;
        gap: 0.5rem;
        padding-top: 0.5rem;
    }

    .message-input input {
        flex: 1;
        padding: 0.5rem;
        border: 1px solid #ccc;
        border-radius: 4px;
    }

    .message-input button {
        padding: 0.5rem 1rem;
        background-color: #24292f;
        color: #f9f9f9;
        border: none;
        border-radius: 4px;
        cursor: pointer;
    }

    /* Right Side: Participants */
    .participants {
        display: flex;
        flex-direction: column;
        border: 2px solid #ccc;
        border-radius: 8px;
        padding: 1rem;
        background-color: #fff;
    }

    .participants h2 {
        color: #24292f;
        padding-bottom: 0.25rem;
        margin-bottom: 0.25rem;
        border-bottom: 1px solid #ddd;
        font-size: 1.25rem;
        text-align: center;
    }

    .participants ul {
        list-style: none;
        padding: 0;
        margin: 0;
    }

    .participant-item {
        color: #24292f;
        display: flex;
        align-items: center;
        gap: 0.3rem;
        padding: 0.2rem;
    }

    .role-icon {
        width: 1rem;
        height: 1rem;
        vertical-align: middle;
    }

    .admin-buttons {
        margin-left: 0.25rem;
        display: flex;
        gap: 0.3rem;
    }

    .admin-btn {
        background: none;
        border: none;
        cursor: pointer;
        padding: 0;
    }

    .admin-btn img {
        width: 1rem;
        height: 1rem;
        vertical-align: middle;
    }

    /* Responsive Design */
    @media (max-width: 768px) {
        .container {
            flex-direction: column;
            height: 100%;
        }

        .participants {
            display: none;
        }

        .chat {
            flex: 1;
            width: 100%;
            display: flex;
            flex-direction: column;
        }

        .chat-window {
            flex: 1;
            overflow-y: auto;
        }
    }
</style>
