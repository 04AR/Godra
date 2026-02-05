-- heartbeat.lua
-- ROLE: player
-- KEYS: None (Dynamic)
-- ARGV[1]: user_id

local user_id = ARGV[1]
if not user_id then
    return redis.error_reply("User ID required")
end

-- 1. Update Sorted Set Score (Timestamp)
local now = redis.call("TIME")[1]
redis.call("ZADD", "active_sessions:guests", now, user_id)

-- 2. Set Expiry Key (User Request for explicit TTL check)
redis.call("SETEX", "guest:" .. user_id .. ":heartbeat", 10, 1)

return "OK"
