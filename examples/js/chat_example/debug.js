const API_URL = 'http://localhost:8080';

async function run() {
    try {
        console.log("1. Guest Login...");
        const loginRes = await fetch(API_URL + '/guest-login', { method: 'POST' });
        if (!loginRes.ok) throw new Error("Login failed: " + await loginRes.text());
        const loginData = await loginRes.json();
        const token = loginData.token;
        console.log("Token obtained:", token?.substring(0, 10) + "...");

        console.log("2. Create Lobby...");
        const res = await fetch(API_URL + '/api/rpc', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                script: 'create_lobby',
                args: ["4"],
                keys: ["global:lobby_id"]
            })
        });

        console.log("Status:", res.status, res.statusText);
        const text = await res.text();
        console.log("Body:", text);

    } catch (e) {
        console.error("Error:", e);
    }
}

run();
