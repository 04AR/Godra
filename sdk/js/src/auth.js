export class AuthService {
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
    }

    async login(username, password) {
        const response = await fetch(`${this.baseUrl}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        if (!response.ok) throw new Error('Login failed');
        return await response.json(); // Returns { token: "..." }
    }

    async register(username, password) {
        const response = await fetch(`${this.baseUrl}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        if (!response.ok) throw new Error('Registration failed');
        return await response.text();
    }

    async guestLogin() {
        const response = await fetch(`${this.baseUrl}/guest-login`, {
            method: 'POST'
        });
        if (!response.ok) throw new Error('Guest login failed');
        return await response.json(); // Returns { token, user_id, role }
    }
}
