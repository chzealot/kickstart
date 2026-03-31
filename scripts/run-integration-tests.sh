#!/bin/bash

# 集成测试脚本
# 用于本地手动执行和 CI 集成

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$WORKSPACE"

echo "======================================"
echo "Running Integration Tests"
echo "======================================"
echo "Working directory: $WORKSPACE"
echo ""

# 步骤 1: 安装依赖
echo "Installing dependencies..."
go mod download
go mod verify
echo "✅ Dependencies installed"
echo ""

# 步骤 2: 运行集成测试
echo "Running integration tests..."

if [ ! -d "./tests/integration" ]; then
  echo "❌ tests/integration directory not found"
  exit 1
fi

echo "Running tests with race detector..."
set +e
go test -v -race -coverprofile=coverage-integration.out -covermode=atomic ./tests/integration/... > test-output.log 2>&1
TEST_EXIT_CODE=$?
set -e

# 显示测试输出
echo "=== Test Output ==="
cat test-output.log
echo "==================="
echo ""

# 生成测试统计
if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "✅ Tests passed"
  TOTAL_TESTS=$(grep -c "^=== RUN" test-output.log || echo "0")
  PASSED_TESTS=$(grep -c "^--- PASS:" test-output.log || echo "0")
  echo "Total tests: $TOTAL_TESTS"
  echo "Passed tests: $PASSED_TESTS"
else
  echo "❌ Tests failed with exit code $TEST_EXIT_CODE"
  FAILED_TESTS=$(grep -c "^--- FAIL:" test-output.log || echo "0")
  echo "Failed tests: $FAILED_TESTS"

  echo ""
  echo "=== Failed Tests ==="
  grep -A 10 "^--- FAIL:" test-output.log || echo "No FAIL markers found"
  echo "===================="
fi

# 计算覆盖率
if [ -f coverage-integration.out ]; then
  COVERAGE=$(go tool cover -func=coverage-integration.out | grep total | awk '{print $3}' || echo "0%")
  echo "Coverage: $COVERAGE"
  echo ""
  echo "To view detailed coverage report, run:"
  echo "  go tool cover -html=coverage-integration.out"
else
  echo "Coverage: 0%"
fi

echo ""
echo "======================================"
if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "✅ Integration Tests Completed Successfully"
else
  echo "❌ Integration Tests Failed"
fi
echo "======================================"

exit $TEST_EXIT_CODE
