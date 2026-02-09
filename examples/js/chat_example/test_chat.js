import { GodraClient } from '../../../sdk/js/src/index.js';
import { spawn } from 'child_process';
import assert from 'assert';

const API_URL = 'http://localhost:8080';

async function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function runTest() {
    console.log("Starting Chat Verification Test...");

    const clientA = new GodraClient(API_URL);
    const clientB = new GodraClient(API_URL);

    try {
        // 1. Login
        console.log("Logging in Client A...");
        const loginA = await clientA.auth.guestLogin();
        const tokenA = loginA.token;
        const userA = loginA.user_id || "GuestA";

        console.log("Logging in Client B...");
        const loginB = await clientB.auth.guestLogin();
        const tokenB = loginB.token;

        // 2. Create Lobby (Client A)
        console.log("Client A creating lobby...");

        const gameId = Math.floor(Math.random() * 100000).toString();
        const gameKey = `game:${gameId}`;
        const playersKey = `game:${gameId}:players`;

        const createRes = await fetch(`${API_URL}/api/rpc`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${tokenA}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                script: 'create_lobby',
                args: ["4"],
                keys: [gameKey, playersKey]
            })
        });

        if (!createRes.ok) {
            const text = await createRes.text();
            throw new Error(`Create Lobby Failed: ${createRes.status} ${createRes.statusText} - ${text}`);
        }

        const createData = await createRes.json();
        let lobbyId = createData.result;
        try { lobbyId = JSON.parse(lobbyId); } catch (e) { }

        // Remove quotes if present
        if (typeof lobbyId === 'string' && lobbyId.startsWith('"') && lobbyId.endsWith('"')) {
            lobbyId = lobbyId.slice(1, -1);
        }

        if (typeof lobbyId === 'string' && lobbyId.startsWith('game:')) {
            lobbyId = lobbyId.substring(5);
        }

        console.log(`Lobby Created: ${lobbyId}`);

        // 3. Connect Realtime
        console.log("Client A connecting to WS...");
        await clientA.realtime.connect(tokenA, lobbyId);

        console.log("Client B connecting to WS...");
        await clientB.realtime.connect(tokenB, lobbyId);

        // 4. Setup Chat Listener on Client B
        const messageReceived = new Promise((resolve, reject) => {
            const timeout = setTimeout(() => reject(new Error("Timeout waiting for message")), 5000);

            clientB.realtime.on('chat', (payload) => {
                console.log("Client B received chat:", payload);
                if (payload.message === "Hello from A") {
                    clearTimeout(timeout);
                    resolve(true);
                }
            });
        });

        // 5. Client A sends message
        // Wait a bit for connections to stabilize
        await sleep(1000);

        console.log("Client A sending message...");
        const chatScriptArgs = ["Hello from A", lobbyId];
        await fetch(`${API_URL}/api/rpc`, {
            method: "POST",
            headers: {
                'Authorization': `Bearer ${tokenA}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                script: 'send_chat',
                args: chatScriptArgs
            })
        });

        // 6. Verify
        await messageReceived;
        console.log("SUCCESS: Client B received the message.");

        // Cleanup
        clientA.realtime.socket.close();
        clientB.realtime.socket.close();
        process.exit(0);

    } catch (err) {
        console.error("TEST FAILED:", err);
        if (clientA.realtime.socket) clientA.realtime.socket.close();
        if (clientB.realtime.socket) clientB.realtime.socket.close();
        process.exit(1);
    }
}

runTest();
