-- send_chat.lua
-- ROLE: player
-- KEYS[1]: game_id (e.g., "game:101")
-- ARGV[1]: user_id
-- ARGV[2]: message

local game_id = KEYS[1] or ("game:" .. ARGV[3]) -- Handling RPC vs WS calling convention might differ, let's normalize. 
-- Wait, RPC Generic Handler passes keys as empty list usually unless specified.
-- Our generic RPC handler passes NO KEYS. 
-- So we should rely on ARGV for logic usually or the handler needs update to pass keys.
-- The RPC Handler in rpc.go: `gamestate.ExecuteScript(r.Context(), req.Script, []string{}, finalArgs...)`
-- So KEYS are empty.
-- We must determine Keys from ARGV if running via RPC.
-- But if running via WS `UpdateState`, we might pass keys?
-- Let's stick to Dynamic Keys from ARGV for flexibility.

local user_id = ARGV[1]
local message = ARGV[2]
local game_id_raw = ARGV[3] -- Expect Game ID passed as arg if not in KEY

if not game_id_raw then
    return redis.error_reply("Game ID required")
end

local channel = "game_updates:game:" .. game_id_raw

local payload = cjson.encode({
    type = "chat",
    payload = {
        user_id = user_id,
        message = message,
        timestamp = redis.call("TIME")[1]
    }
})

redis.call("PUBLISH", channel, payload)
return payload
