import { AuthService } from './auth.js';
import { LobbyService } from './lobby.js';
import { RealtimeService } from './realtime.js';

export class GodraClient {
    constructor(baseUrl) {
        this.auth = new AuthService(baseUrl);
        this.lobby = new LobbyService(baseUrl, this.auth);
        this.realtime = new RealtimeService(baseUrl);
    }
}
