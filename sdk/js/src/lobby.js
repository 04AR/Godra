export class LobbyService {
    constructor(baseUrl, authService) {
        this.baseUrl = baseUrl;
    }

    async createLobby(token, maxPlayers) {
        const response = await fetch(`${this.baseUrl}/api/rpc`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                script: 'create_lobby',
                args: [maxPlayers]
            })
        });
        if (!response.ok) throw new Error('Failed to create lobby');
        const data = await response.json();
        return data.result; // Returns lobby_id
    }

    async joinLobby(token, lobbyId) {
        // Technically join is handled on WS connect for verification, 
        // but we can check if it exists via RPC if needed.
        // For now, this is a placeholder or can perform a "check_lobby" RPC.
        return true;
    }
}
