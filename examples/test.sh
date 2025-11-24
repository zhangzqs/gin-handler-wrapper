#!/bin/bash

# 测试脚本 - Gin Handler Wrapper Complete Example

BASE_URL="http://localhost:8080"

echo "=========================================="
echo "🧪 Testing Gin Handler Wrapper API"
echo "=========================================="
echo ""

# 检查服务是否运行
echo "📡 Checking if server is running..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo "❌ Server is not running at $BASE_URL"
    echo "Please start the server with: go run main.go"
    exit 1
fi
echo "✅ Server is running"
echo ""

# 测试健康检查
echo "=========================================="
echo "1️⃣  Testing GET /health (WrapGetter)"
echo "=========================================="
curl -s "$BASE_URL/health" | jq '.'
echo ""

# 创建用户
echo "=========================================="
echo "2️⃣  Testing POST /users (WrapHandler)"
echo "=========================================="
echo "Creating user: Charlie"
curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}' | jq '.'
echo ""

# 获取用户列表
echo "=========================================="
echo "3️⃣  Testing GET /users (WrapHandler)"
echo "=========================================="
curl -s "$BASE_URL/users?page=1&page_size=10" | jq '.'
echo ""

# 获取用户详情
echo "=========================================="
echo "4️⃣  Testing GET /users/:id (WrapHandler)"
echo "=========================================="
echo "Getting user with ID=1"
curl -s "$BASE_URL/users/1" | jq '.'
echo ""

# 更新用户
echo "=========================================="
echo "5️⃣  Testing PUT /users/:id (WrapConsumer)"
echo "=========================================="
echo "Updating user with ID=1"
curl -s -X PUT "$BASE_URL/users/1" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.updated@example.com"}' | jq '.'
echo ""

# 验证更新
echo "Verifying update..."
curl -s "$BASE_URL/users/1" | jq '.'
echo ""

# 搜索文章
echo "=========================================="
echo "6️⃣  Testing GET /articles/search (WrapHandler)"
echo "=========================================="
echo "Searching for keyword: Go"
curl -s "$BASE_URL/articles/search?keyword=Go&page=1&page_size=10" | jq '.'
echo ""

# 清除缓存
echo "=========================================="
echo "7️⃣  Testing POST /cache/clear (WrapAction)"
echo "=========================================="
curl -s -X POST "$BASE_URL/cache/clear" | jq '.'
echo ""

# 同步数据
echo "=========================================="
echo "8️⃣  Testing POST /data/sync (WrapAction)"
echo "=========================================="
curl -s -X POST "$BASE_URL/data/sync" | jq '.'
echo ""

# 删除用户
echo "=========================================="
echo "9️⃣  Testing DELETE /users/:id (WrapConsumer)"
echo "=========================================="
echo "Deleting user with ID=1"
curl -s -X DELETE "$BASE_URL/users/1" | jq '.'
echo ""

# 验证删除
echo "Verifying deletion (should return 404)..."
curl -s "$BASE_URL/users/1" | jq '.'
echo ""

# 测试错误处理
echo "=========================================="
echo "🔟 Testing Error Handling"
echo "=========================================="
echo "Getting non-existent user (ID=999):"
curl -s "$BASE_URL/users/999" | jq '.'
echo ""

echo "Creating duplicate user:"
curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob","email":"bob@example.com"}' | jq '.'
echo ""

echo "Invalid request (missing required fields):"
curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test"}' | jq '.'
echo ""

echo "=========================================="
echo "✅ All tests completed!"
echo "=========================================="
