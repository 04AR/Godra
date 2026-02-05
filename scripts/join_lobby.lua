-- join_lobby.lua
-- KEYS[1]: lobby_key (e.g. "lobby:xyz")
-- ARGV[1]: user_id

local lobby_key = KEYS[1]
local user_id = ARGV[1]
local players_key = lobby_key .. ":players"

if redis.call("EXISTS", lobby_key) == 0 then
    return redis.error_reply("Lobby does not exist")
end

local capacity = tonumber(redis.call("HGET", lobby_key, "capacity"))
local current_count = redis.call("SCARD", players_key)

-- Check if user is already in
if redis.call("SISMEMBER", players_key, user_id) == 1 then
    return "OK" -- idempotent
end

if current_count >= capacity then
    return redis.error_reply("Lobby is full")
end

redis.call("SADD", players_key, user_id)
return "OK"
