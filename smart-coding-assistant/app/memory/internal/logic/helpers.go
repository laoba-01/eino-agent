package logic

import (
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

// =============================================================================
// Lua 原子脚本 —— 单次网络往返完成多步操作 + 乐观锁
// =============================================================================

// atomicSaveScript 原子写入: HSET + EXPIRE + _version 自增
// KEYS[1] = context key
// ARGV[1] = ttl (0 = 不过期)
// ARGV[2..N] = 交替的 field-value 对
// 返回: {1, new_version} 成功 / {0, error_msg} 失败
const atomicSaveScript = `
local key = KEYS[1]
local ttl = tonumber(ARGV[1])
local current = redis.call('HGET', key, '_version')
local v = (current and tonumber(current) or 0) + 1

-- 先写版本号
redis.call('HSET', key, '_version', v)

-- 写入用户数据 (ARGV[2..N] = field-value pairs)
for i = 2, #ARGV, 2 do
    redis.call('HSET', key, ARGV[i], ARGV[i+1])
end

if ttl > 0 then
    redis.call('EXPIRE', key, ttl)
end

return {1, v}
`

// atomicUpdateScript 带乐观锁的原子更新 (CAS)
// KEYS[1] = context key
// ARGV[1] = expected_version (0 = 无条件写入)
// ARGV[2] = ttl
// ARGV[3..N] = field-value pairs
// 返回: {1, new_version} 成功 / {0, current_version} 冲突
const atomicUpdateScript = `
local key = KEYS[1]
local expected = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])

local current_raw = redis.call('HGET', key, '_version')
local current = current_raw and tonumber(current_raw) or 0

if expected > 0 and expected ~= current then
    return {0, current}  -- 版本冲突
end

local new_v = current + 1
redis.call('HSET', key, '_version', new_v)

for i = 3, #ARGV, 2 do
    redis.call('HSET', key, ARGV[i], ARGV[i+1])
end

if ttl > 0 then
    redis.call('EXPIRE', key, ttl)
end

return {1, new_v}
`

// atomicDeleteScript 带乐观锁的原子删除
// KEYS[1] = context key
// ARGV 可选: ARGV[1] = expected_version, ARGV[2..N] = keys to delete (空=删除整个key)
// 返回: {1, deleted_count} 成功 / {0, current_version} 冲突
const atomicDeleteScript = `
local key = KEYS[1]
local expected = tonumber(ARGV[1]) or 0

if expected > 0 then
    local current_raw = redis.call('HGET', key, '_version')
    local current = current_raw and tonumber(current_raw) or 0
    if expected ~= current then
        return {0, current}
    end
end

-- 删除指定字段或整个 key
if #ARGV > 1 then
    return {1, redis.call('HDEL', key, unpack(ARGV, 2))}
else
    return {1, redis.call('DEL', key)}
end
`

// safeReleaseLockScript 安全释放分布式锁 — 只删自己持有的锁
// KEYS[1] = lock key
// ARGV[1] = lock value
// 返回: 1 (成功释放) / 0 (锁已过期或不属于自己)
const safeReleaseLockScript = `
if redis.call('GET', KEYS[1]) == ARGV[1] then
    return redis.call('DEL', KEYS[1])
end
return 0
`

// =============================================================================
// 脚本缓存 — 避免每次调用时重新哈希脚本
// =============================================================================

var (
	saveScript        = redis.NewScript(atomicSaveScript)
	updateScript      = redis.NewScript(atomicUpdateScript)
	deleteScript      = redis.NewScript(atomicDeleteScript)
	releaseLockScript = redis.NewScript(safeReleaseLockScript)
)

// =============================================================================
// 工具函数
// =============================================================================

// mapToSlice 将 map[string]string 转为 key-value 交错的 []string（零堆分配优化版）
// 用于 Lua 脚本的 ARGV 参数传递，避免 interface{} 装箱
func mapToSlice(m map[string]string) []string {
	result := make([]string, 0, len(m)*2)
	for k, v := range m {
		result = append(result, k, v)
	}
	return result
}

// mapToInterface 保留（savevector 仍需要）。
// TODO: 后续 Milvus 操作也改为零分配版本
func mapToInterface(m map[string]string) []interface{} {
	result := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		result = append(result, k, v)
	}
	return result
}

func buildFilterExpr(filter map[string]string) string {
	if len(filter) == 0 {
		return ""
	}
	exprs := make([]string, 0, len(filter))
	for k, v := range filter {
		exprs = append(exprs, fmt.Sprintf(`metadata["%s"] == "%s"`, k, v))
	}
	var sb strings.Builder
	for i, e := range exprs {
		if i > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(e)
	}
	return sb.String()
}

func joinInt64s(ids []int64) string {
	var sb strings.Builder
	for i, id := range ids {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%d", id))
	}
	return sb.String()
}
