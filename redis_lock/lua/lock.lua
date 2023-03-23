-- 抢到的锁的值与上一次相同表示加锁成功，不是返回加锁失败
local val = redis.call("get", KEYS[1])
if not val then
    -- 表示锁不存在，表示没有人拿到锁
    return redis.call("set", KEYS[1], ARGV[1], "EX", ARGV[2])
elseif val == ARGV[1] then
    -- 上一次抢到了锁
    redis.call("expire", KEYS[1], ARGV[2])
    return "OK"
else
    -- 锁被其他人持有
    return ""
end