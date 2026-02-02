-- create_lobby.lua
-- ROLE: guest
-- KEYS[1]: lobby_key (e.g. "game:123")
-- KEYS[2]: players_key (e.g. "game:123:players")
-- ARGV[1]: user_id
-- ARGV[2]: capacity

local lobby_key = KEYS[1]
local players_key = KEYS[2]
local user_id = ARGV[1]
local capacity = tonumber(ARGV[2]) or 4

if redis.call("EXISTS", lobby_key) == 1 then
    return redis.error_reply("Lobby already exists")
end

redis.call("HSET", lobby_key, "owner", user_id)
redis.call("HSET", lobby_key, "capacity", capacity)
redis.call("HSET", lobby_key, "created_at", redis.call("TIME")[1])
redis.call("HSET", lobby_key, "status", "open")

redis.call("SADD", players_key, user_id)

-- Return the ID part (strip "lobby:") for convenience, or just return the key
return lobby_key
