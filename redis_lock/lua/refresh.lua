-- 检查一下是不是自己的锁
-- 是的话就刷新一下
if redis.call("get", KEYS[1]) == ARGV[1] then
    -- 刷新一下
    return redis.call("expire", KEYS[1], ARGV[2])
else
    return 0
end
