# Godra

Godra is a high-performance, concurrent multiplayer game backend built in **Go**. It leverages **WebSockets** for real-time communication, **Redis/Dragonfly/Valkey** for state synchronization (using Lua scripts), and supports both **SQLite** and **PostgreSQL** for persistence.

## Features

- **Real-Time Communication**: WebSocket-based architecture for low-latency game inputs.
- **Distributed State**: Game state is synchronized via Redis Pub/Sub, allowing for potential scaling.
- **Atomic Updates**: Game logic uses Lua scripts running potentially on Redis to ensure atomicity.
- **Hot-Reloading**: Lua scripts in the `scripts/` directory are watched and reloaded automatically on change.
- **Pluggable Database**: Supports SQLite (default) and PostgreSQL via GORM.
- **Authentication**: JWT-based authentication with Role-Based Access Control (RBAC).
- **Guest Access**: Temporary guest accounts with automatic cleanup.
- **Observability**: Structured logging (`slog`) and metrics endpoint (`/metrics`).

## Architecture

1.  **Client (Godot/SDK)**: Connects via WebSocket, sends inputs, and receives state patches.
2.  **Server (Go)**:
    *   Authenticates users.
    *   Validates and batches inputs.
    *   Executes Lua scripts on Redis to update state.
    *   Broadcasts state updates from Redis back to connected clients.
3.  **Redis**: Acts as the "source of truth" for hot game state and coordinates Pub/Sub.
4.  **Database**: Persists user accounts and long-term data.

## Getting Started

### Prerequisites

- **Go** (1.21 or higher)
- **Redis** (running on `localhost:6379`)

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/example/godra.git
    cd godra_server
    ```
2.  Install dependencies:
    ```bash
    go mod download
    ```

### Running the Server

**Note**: The server requires the `scripts/` folder to be present in the working directory to load game logic.

**Option 1: Default (SQLite)**
```bash
go run main.go
```

**Option 2: With PostgreSQL**
```bash
go run main.go -db-type postgres -db-dsn "host=localhost user=postgres password=secret dbname=godra port=5432 sslmode=disable"
```

The server will start on port `8080`.

## API Endpoints

### HTTP

-   `POST /register`: Create a new account (`username`, `password`).
-   `POST /login`: Authenticate (`username`, `password`) -> Returns JWT.
-   `POST /guest-login`: Get a temporary session.
-   `GET /metrics`: Prometheus-formatted metrics.
-   `POST /api/rpc`: Execute a Lua script (requires Auth header).

### WebSocket

-   `WS /ws?token=<JWT>&game_id=<LOBBY_ID>`: Connect to a game instance.

## SDKs

### C# (Godot)
Located in `sdk/csharp/GodotSDK`.
- **GodotClient**: Main entry point.

### JavaScript (Web/Node)
Located in `sdk/js`.
- **GodraClient**: Entry point. ES Module based.
- Usage: `import { GodraClient } from './sdk/js/src/index.js';`

### Dart (Flutter)
Located in `sdk/dart`.
- **GodraClient**: Entry point.
- Usage: `final client = GodraClient('http://localhost:8080');`

## Logic & Scripting

Game logic is defined in `scripts/*.lua`. You can modify these files while the server is running.

-   **`update_state.lua`**: Handles generic state updates.
-   **`move_player.lua`**: Validates movement and updates position.
-   **`create_lobby.lua`**: Sets up new game rooms.

## License

MIT

