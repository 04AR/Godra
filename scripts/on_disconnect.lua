-- on_disconnect.lua
-- ARGV[1]: user_id

local user_id = ARGV[1]

-- If it's a guest, delete the key
-- Note: Lua doesn't have "startswith" easily, but we trust the caller (Go server) only calls this for guests
-- Or we try to delete regardless. 
-- Since UserIDs for real users are ints (as strings "1", "2"), and guests are "guest:xyz",
-- deleting "1" might be bad if "1" is a key for something else.
-- But we don't use "1" as a key at root level for users in this schema (User data is in SQL).
-- So it is safe to try DEL user_id.

redis.call("DEL", user_id)
return "OK"
