import { GodraClient } from '../../../sdk/js/src/index.js';

const API_URL = 'http://localhost:8080';

async function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function createLobby(client, token) {
    const gameId = Math.floor(Math.random() * 100000).toString();
    const gameKey = `game:${gameId}`;
    const playersKey = `game:${gameId}:players`;

    const res = await fetch(`${API_URL}/api/rpc`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            script: 'create_lobby',
            args: ["4"],
            keys: [gameKey, playersKey]
        })
    });

    if (!res.ok) throw new Error("Create Lobby Failed: " + await res.text());

    // We already know the ID because we generated it
    return gameId;
}

async function sendMessage(token, text, lobbyId) {
    await fetch(`${API_URL}/api/rpc`, {
        method: "POST",
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            script: 'send_chat',
            args: [text, lobbyId]
        })
    });
}

async function runTest() {
    console.log("Starting Chat Isolation Test...");

    const clients = {
        A: new GodraClient(API_URL),
        B: new GodraClient(API_URL),
        C: new GodraClient(API_URL),
        D: new GodraClient(API_URL)
    };

    const tokens = {};
    const received = { A: [], B: [], C: [], D: [] };

    try {
        // 1. Login Everyone
        console.log("Logging in 4 clients...");
        for (const [name, client] of Object.entries(clients)) {
            const auth = await client.auth.guestLogin();
            tokens[name] = auth.token;
        }

        // 2. Create Two Lobbies
        console.log("Creating Lobby 1 (A & B)...");
        const lobby1 = await createLobby(clients.A, tokens.A);
        console.log(`Lobby 1 ID: ${lobby1}`);

        console.log("Creating Lobby 2 (C & D)...");
        const lobby2 = await createLobby(clients.C, tokens.C);
        console.log(`Lobby 2 ID: ${lobby2}`);

        // 3. Connect Clients
        // Group 1
        await clients.A.realtime.connect(tokens.A, lobby1);
        await clients.B.realtime.connect(tokens.B, lobby1);

        // Group 2
        await clients.C.realtime.connect(tokens.C, lobby2);
        await clients.D.realtime.connect(tokens.D, lobby2);

        // 4. Setup Listeners
        Object.keys(clients).forEach(name => {
            clients[name].realtime.on('chat', (payload) => {
                console.log(`[${name}] Received: ${payload.message}`);
                received[name].push(payload);
            });
        });

        // Wait for sockets to be ready
        await sleep(1000);

        // 5. Test Case 1: A sends to Lobby 1
        console.log("\n--- Sending Message in Lobby 1 (A -> B) ---");
        await sendMessage(tokens.A, "Hello form Lobby 1", lobby1);

        await sleep(1000); // Wait for delivery

        // Verify Group 1 received it
        if (received.B.find(m => m.message === "Hello form Lobby 1")) {
            console.log("✅ Client B received message.");
        } else {
            throw new Error("❌ Client B DID NOT receive message.");
        }

        // Verify Group 2 did NOT receive it
        if (received.C.length > 0 || received.D.length > 0) {
            throw new Error("❌ Leakage detected! Client C or D received message from Lobby 1.");
        } else {
            console.log("✅ Isolation Confirmed (C & D received nothing).");
        }

        // 6. Test Case 2: C sends to Lobby 2
        console.log("\n--- Sending Message in Lobby 2 (C -> D) ---");
        // Clear buffers for cleaner verification (optional, but easier)
        received.A = []; received.B = []; received.C = []; received.D = [];

        await sendMessage(tokens.C, "Hello from Lobby 2", lobby2);

        await sleep(1000);

        // Verify Group 2 received it
        if (received.D.find(m => m.message === "Hello from Lobby 2")) {
            console.log("✅ Client D received message.");
        } else {
            throw new Error("❌ Client D DID NOT receive message.");
        }

        // Verify Group 1 did NOT receive it
        if (received.A.length > 0 || received.B.length > 0) {
            throw new Error("❌ Leakage detected! Client A or B received message from Lobby 2.");
        } else {
            console.log("✅ Isolation Confirmed (A & B received nothing).");
        }

        console.log("\nSUCCESS: Lobbies are isolated.");
        process.exit(0);

    } catch (err) {
        console.error("\nTEST FAILED:", err);
        process.exit(1);
    } finally {
        Object.values(clients).forEach(c => c.realtime.socket && c.realtime.socket.close());
    }
}

runTest();
