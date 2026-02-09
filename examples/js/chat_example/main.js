import { GodraClient } from '../../../sdk/js/src/index.js';

// Configuration
const API_URL = 'http://localhost:8080';
const client = new GodraClient(API_URL);

// State
let token = null;
let currentUser = null;
let currentLobby = null;

// UI Elements
const app = {
    screens: {
        login: document.getElementById('login-section'),
        lobby: document.getElementById('lobby-section'),
        chat: document.getElementById('chat-section')
    },
    status: document.getElementById('connection-status'),
    inputs: {
        username: document.getElementById('username'),
        password: document.getElementById('password'),
        lobbyId: document.getElementById('lobby-id-input'),
        message: document.getElementById('message-input')
    },
    buttons: {
        login: document.getElementById('btn-login'),
        register: document.getElementById('btn-register'),
        guest: document.getElementById('btn-guest'),
        createLobby: document.getElementById('btn-create-lobby'),
        joinLobby: document.getElementById('btn-join-lobby'),
        logout: document.getElementById('btn-logout'),
        leaveRoom: document.getElementById('btn-leave-room'),
        send: document.getElementById('btn-send')
    },
    displays: {
        user: document.getElementById('user-display'),
        room: document.getElementById('room-display'),
        messages: document.getElementById('messages-container')
    }
};

// UI Helpers
function showScreen(screenId) {
    Object.values(app.screens).forEach(el => el.classList.add('hidden'));
    app.screens[screenId].classList.remove('hidden');
}

function updateStatus(connected) {
    if (connected) {
        app.status.textContent = 'Connected';
        app.status.className = 'status connected';
    } else {
        app.status.textContent = 'Disconnected';
        app.status.className = 'status disconnected';
    }
}

function addMessage(user, text, isSelf) {
    const msgDiv = document.createElement('div');
    msgDiv.className = `message ${isSelf ? 'self' : 'other'}`;

    const metaDiv = document.createElement('div');
    metaDiv.className = 'message-meta';
    metaDiv.textContent = user;

    const contentDiv = document.createElement('div');
    contentDiv.textContent = text;

    msgDiv.appendChild(metaDiv);
    msgDiv.appendChild(contentDiv);

    app.displays.messages.appendChild(msgDiv);
    app.displays.messages.scrollTop = app.displays.messages.scrollHeight;
}

// Event Listeners
app.buttons.guest.addEventListener('click', async () => {
    try {
        const data = await client.auth.guestLogin();
        token = data.token;

        currentUser = { username: "Guest" };
        if (data.user_id) currentUser.username = data.user_id;

        app.displays.user.textContent = currentUser.username;
        showScreen('lobby');
    } catch (err) {
        alert('Guest login failed: ' + err.message);
    }
});

app.buttons.login.addEventListener('click', async () => {
    const username = app.inputs.username.value;
    const password = app.inputs.password.value;
    if (!username || !password) return alert('Enter username and password');

    try {
        const data = await client.auth.login(username, password);
        token = data.token;

        currentUser = { username };
        app.displays.user.textContent = username;
        showScreen('lobby');
    } catch (err) {
        alert('Login failed: ' + err.message);
    }
});

app.buttons.register.addEventListener('click', async () => {
    const username = app.inputs.username.value;
    const password = app.inputs.password.value;
    if (!username || !password) return alert('Enter username and password');

    try {
        await client.auth.register(username, password);
        alert('Registered! You can now login.');
    } catch (err) {
        alert('Registration failed: ' + err.message);
    }
});

app.buttons.logout.addEventListener('click', () => {
    // client.auth.logout(); // SDK doesn't have logout
    token = null;
    currentUser = null;
    showScreen('login');
});

app.buttons.createLobby.addEventListener('click', async () => {
    try {
        const randomId = Math.floor(Math.random() * 10000);
        const gameKey = `game:${randomId}`;
        const playersKey = `game:${randomId}:players`;

        // Create Lobby via proper service if available, or RPC
        // Inspecting lobby.js usually reveals methods. Assuming create() or RPC call.
        // Based on walkthrough: POST /api/rpc {"script": "create_lobby", "args": [4]}
        // Using Generic RPC for flexibility as lobby.js might be thin.
        const response = await fetch(`${API_URL}/api/rpc`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                script: 'create_lobby',
                args: ["4"], // max players
                keys: [gameKey, playersKey]
            })
        });

        if (!response.ok) throw new Error('Failed to create lobby');
        const data = await response.json();

        // usage: key or id. Script returns key "game:123"
        let lobbyId = data.result;

        // If result is JSON string, parse it.
        try { lobbyId = JSON.parse(data.result); } catch (e) { }
        // Remove quotes if present
        if (typeof lobbyId === 'string' && lobbyId.startsWith('"') && lobbyId.endsWith('"')) {
            lobbyId = lobbyId.slice(1, -1);
        }

        // Extract raw ID if needed, or use key. 
        // The rest of app expects "lobby-id" usually to match what "connect" needs.
        // Realtime connect: `game_id=${lobbyId}` -> `update_state` uses `KEYS[1] = "game:" + gameID`?
        // Wait, update_state uses `KEYS[1]: game_id`.
        // If I pass "lobby:123", update_state will use "game:lobby:123"?
        // Mismatch!

        // Correction: create_lobby should use "game:<id>" or client.go should change.
        // Existing convention seemed to be "game_id" (RPC) vs "game:<id>" (Redis).
        // Let's stick to "game:<id>" everywhere for simplicity.

        // Let's assume generated ID is "123". Key is "game:123".
        // client.go -> "game:123". Perfect.

        if (lobbyId.startsWith("game:")) {
            lobbyId = lobbyId.substring(5);
        }

        enterLobby(lobbyId);
    } catch (err) {
        console.error(err);
        alert('Error creating lobby: ' + err.message);
    }
});

app.buttons.joinLobby.addEventListener('click', () => {
    const id = app.inputs.lobbyId.value;
    if (id) enterLobby(id);
});

async function enterLobby(lobbyId) {
    try {
        currentLobby = lobbyId;
        app.displays.room.textContent = lobbyId;

        // Connect Realtime
        await client.realtime.connect(token, lobbyId);
        updateStatus(true);

        // Subscribe to chat events
        client.realtime.on('chat', (payload) => {
            console.log("Chat received:", payload);
            addMessage(payload.user_id, payload.message, payload.user_id === currentUser.username);
        });

        // Also listen for connection errors/close
        client.realtime.socket.onclose = () => updateStatus(false);

        showScreen('chat');
        app.displays.messages.innerHTML = '<div class="system-message">Joined lobby ' + lobbyId + '</div>';
    } catch (err) {
        alert('Failed to join lobby: ' + err.message);
        currentLobby = null;
    }
}

app.buttons.leaveRoom.addEventListener('click', () => {
    if (client.realtime.socket) {
        client.realtime.socket.close();
    }
    currentLobby = null;
    updateStatus(false);
    showScreen('lobby');
});

app.buttons.send.addEventListener('click', sendMessage);
app.inputs.message.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') sendMessage();
});

async function sendMessage() {
    const text = app.inputs.message.value;
    if (!text || !currentLobby) return;

    try {
        // Use RPC to send chat since WS goes to update_state
        // script: send_chat.lua
        // args: [user_id, message, game_id] (based on my reading of send_chat.lua)
        // send_chat.lua: local user_id = ARGV[1], message = ARGV[2], game_id_raw = ARGV[3]

        // We need to resolve user_id safely.
        // For guests, user_id is in the token claims.
        // Ideally the script should get user_id from KEYS or context, but send_chat.lua takes it as ARGV[1].
        // This is insecure if not validated, but for demo it's fine.
        // Better: Pass "self" and let script resolve? or just pass what we know.

        // Wait, send_chat.lua line 7: `local game_id = KEYS[1] or ("game:" .. ARGV[3])`
        // RPC handler passes empty KEYS. So we must pass game_id in ARGV[3].

        const userId = currentUser.username || "Unknown";

        await fetch(`${API_URL}/api/rpc`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                script: 'send_chat',
                args: [text, currentLobby]
            })
        });

        app.inputs.message.value = '';
        // Message will be received via WS 'chat' event, so we don't add it manually here to avoid duplicates
        // unless we want optimistic UI.
    } catch (err) {
        console.error(err);
        alert('Failed to send message');
    }
}
