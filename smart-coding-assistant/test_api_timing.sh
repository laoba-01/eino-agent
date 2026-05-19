#!/bin/bash

# API 接口调用时长测试脚本
BASE_URL="http://localhost:8080"
RESULTS_FILE="api_timing_results_$(date +%Y%m%d_%H%M%S).txt"
ITERATIONS=3

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# 浮点数 * 1000 转换用 awk
to_ms() { awk "BEGIN { printf \"%.0f\", $1 * 1000 }"; }
cmp_float() { awk "BEGIN { exit(!($1)) }"; }

TOTAL_START=$(date +%s%3N)

echo "==============================================" | tee "$RESULTS_FILE"
echo "  API 接口调用时长测试报告" | tee -a "$RESULTS_FILE"
echo "  测试时间: $(date '+%Y-%m-%d %H:%M:%S')" | tee -a "$RESULTS_FILE"
echo "  目标服务: $BASE_URL" | tee -a "$RESULTS_FILE"
echo "  每组测试次数: $ITERATIONS" | tee -a "$RESULTS_FILE"
echo "==============================================" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

test_endpoint() {
    local name="$1"
    local method="$2"
    local path="$3"
    local body="$4"
    local token="$5"

    local total_time=0
    local min_time=999999
    local max_time=0
    local success_count=0
    local fail_count=0

    local curl_cmd="curl -s -o /dev/null -w '%{http_code} %{time_total}'"
    if [ "$method" = "POST" ]; then
        curl_cmd="$curl_cmd -X POST -H 'Content-Type: application/json'"
    fi
    if [ -n "$token" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $token'"
    fi
    if [ -n "$body" ]; then
        curl_cmd="$curl_cmd -d '$body'"
    fi
    curl_cmd="$curl_cmd $BASE_URL$path"

    echo "----------------------------------------------" | tee -a "$RESULTS_FILE"
    printf "${CYAN}[测试]${NC} %-35s ${YELLOW}%s${NC} %s\n" "$name" "$method" "$path" | tee -a "$RESULTS_FILE"
    echo "" | tee -a "$RESULTS_FILE"

    for i in $(seq 1 $ITERATIONS); do
        local result=$(eval $curl_cmd 2>/dev/null)
        local http_code=$(echo "$result" | awk '{print $1}')
        local time_sec=$(echo "$result" | awk '{print $2}')
        local time_ms=$(to_ms "$time_sec")

        if [[ "$http_code" -ge 200 && "$http_code" -lt 300 ]]; then
            success_count=$((success_count + 1))
            printf "  ${GREEN}[通过]${NC} 第%d次: HTTP %s | 耗时: %sms\n" "$i" "$http_code" "$time_ms" | tee -a "$RESULTS_FILE"
            total_time=$(awk "BEGIN { printf \"%.3f\", $total_time + $time_sec }")
            if awk "BEGIN { exit(!($time_ms < $min_time)) }"; then
                min_time=$time_ms
            fi
            if awk "BEGIN { exit(!($time_ms > $max_time)) }"; then
                max_time=$time_ms
            fi
        else
            fail_count=$((fail_count + 1))
            printf "  ${RED}[失败]${NC} 第%d次: HTTP %s | 耗时: %sms\n" "$i" "$http_code" "$time_ms" | tee -a "$RESULTS_FILE"
        fi

        sleep 0.3
    done

    if [ "$success_count" -gt 0 ]; then
        local avg_ms=$(awk "BEGIN { printf \"%.2f\", ($total_time / $success_count) * 1000 }")
    else
        local avg_ms="N/A"
        min_time="N/A"
        max_time="N/A"
    fi

    echo "" | tee -a "$RESULTS_FILE"
    echo "  ┌─────────────────────────────────┐" | tee -a "$RESULTS_FILE"
    printf "  │ 平均耗时: %-8s ms             │\n" "$avg_ms" | tee -a "$RESULTS_FILE"
    printf "  │ 最小耗时: %-8s ms             │\n" "$min_time" | tee -a "$RESULTS_FILE"
    printf "  │ 最大耗时: %-8s ms             │\n" "$max_time" | tee -a "$RESULTS_FILE"
    printf "  │ 成功/总数: %d/%d                  │\n" "$success_count" "$ITERATIONS" | tee -a "$RESULTS_FILE"
    echo "  └─────────────────────────────────┘" | tee -a "$RESULTS_FILE"
    echo "" | tee -a "$RESULTS_FILE"

    echo "$avg_ms"
}

echo ">>> 第一阶段: 健康检查与基础连通性测试" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

HEALTH_AVG=$(test_endpoint "健康检查" "GET" "/health")

echo ">>> 第二阶段: 认证接口测试" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# 先生成唯一用户名，确保注册测试不带缓存干扰
REGISTER_USER="testuser_$(date +%s)"
REGISTER_AVG=$(test_endpoint "用户注册" "POST" "/api/auth/register" \
    "{\"username\":\"$REGISTER_USER\",\"password\":\"test123456\",\"email\":\"$REGISTER_USER@example.com\"}")

# 注册一个固定基准用户，用于后续测试
BENCH_USER="bench_$(date +%s)"
curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"$BENCH_USER\",\"password\":\"test123456\",\"email\":\"$BENCH_USER@example.com\"}" \
    "$BASE_URL/api/auth/register" > /dev/null 2>&1

# 登录获取token
LOGIN_RESP=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"$BENCH_USER\",\"password\":\"test123456\"}" \
    "$BASE_URL/api/auth/login")
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))" 2>/dev/null || echo "$LOGIN_RESP" | grep -o '"token":"[^"]*"' | head -1 | sed 's/"token":"//;s/"//')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${YELLOW}[警告]${NC} 无法获取认证token，受保护接口将使用mock token" | tee -a "$RESULTS_FILE"
    TOKEN="mock-jwt-token"
else
    echo -e "${GREEN}[信息]${NC} 已获取认证token" | tee -a "$RESULTS_FILE"
fi

echo "" | tee -a "$RESULTS_FILE"

# 登录性能测试
LOGIN_AVG=$(test_endpoint "用户登录" "POST" "/api/auth/login" \
    "{\"username\":\"$BENCH_USER\",\"password\":\"test123456\"}")

echo ">>> 第三阶段: 受保护接口测试 (需要认证)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

CHAT_AVG=$(test_endpoint "AI对话(Chat)" "POST" "/api/chat" \
    '{"message":"你好，请介绍一下Go语言的特点","context":{}}' \
    "$TOKEN")

REPORT_AVG=$(test_endpoint "学习报告" "GET" "/api/learning/report" "" "$TOKEN")

POINTS_AVG=$(test_endpoint "知识点列表" "GET" "/api/learning/points?sort_by=last_seen" "" "$TOKEN")

LOGOUT_AVG=$(test_endpoint "用户登出" "POST" "/api/auth/logout" "" "$TOKEN")

TOTAL_END=$(date +%s%3N)
TOTAL_ELAPSED=$((TOTAL_END - TOTAL_START))

echo "==============================================" | tee -a "$RESULTS_FILE"
echo "  测试汇总" | tee -a "$RESULTS_FILE"
echo "==============================================" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"
echo "  ┌─────────────────────────────────────────────┐" | tee -a "$RESULTS_FILE"
printf "  │ %-35s %8s ms │\n" "GET  /health"              "$HEALTH_AVG"     | tee -a "$RESULTS_FILE"
printf "  │ %-35s %8s ms │\n" "POST /api/auth/register"    "$REGISTER_AVG"   | tee -a "$RESULTS_FILE"
printf "  │ %-35s %8s ms │\n" "POST /api/auth/login"       "$LOGIN_AVG"      | tee -a "$RESULTS_FILE"
printf "  │ %-35s %8s ms │\n" "POST /api/chat"             "$CHAT_AVG"       | tee -a "$RESULTS_FILE"
printf "  │ %-35s %8s ms │\n" "GET  /api/learning/report" "$REPORT_AVG"     | tee -a "$RESULTS_FILE"
printf "  │ %-35s %8s ms │\n" "GET  /api/learning/points" "$POINTS_AVG"     | tee -a "$RESULTS_FILE"
printf "  │ %-35s %8s ms │\n" "POST /api/auth/logout"      "$LOGOUT_AVG"     | tee -a "$RESULTS_FILE"
echo "  ├─────────────────────────────────────────────┤" | tee -a "$RESULTS_FILE"
printf "  │ %-35s %8d ms │\n" "总测试耗时"                "$TOTAL_ELAPSED"  | tee -a "$RESULTS_FILE"
echo "  └─────────────────────────────────────────────┘" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"
echo "详细结果已保存至: $RESULTS_FILE" | tee -a "$RESULTS_FILE"
