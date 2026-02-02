#!/bin/bash
#
# ci-pipeline.sh - Complete CI/CD pipeline using plumego CLI
#
# This script demonstrates a full CI/CD workflow:
# 1. Health checks
# 2. Security validation
# 3. Tests with coverage
# 4. Build
# 5. Artifact generation
#

set -euo pipefail

PROJECT_DIR="${1:-.}"
OUTPUT_DIR="${2:-./ci-output}"

echo "====================================="
echo "Plumego CI/CD Pipeline"
echo "====================================="
echo "Project: $PROJECT_DIR"
echo "Output: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

cd "$PROJECT_DIR"

# Step 1: Health Check
echo "[1/5] Running health check..."
if plumego check --format json > "$OUTPUT_DIR/health.json"; then
  STATUS=$(jq -r '.checks | to_entries | map(select(.value.status == "failed")) | length' "$OUTPUT_DIR/health.json")
  
  if [ "$STATUS" -eq 0 ]; then
    echo "✓ Health check passed"
  else
    echo "✗ Health check found issues"
    jq '.checks | to_entries[] | select(.value.status == "failed")' "$OUTPUT_DIR/health.json"
    exit 1
  fi
else
  echo "✗ Health check failed"
  exit 1
fi

# Step 2: Security Validation
echo ""
echo "[2/5] Running security validation..."
if plumego check --security --format json > "$OUTPUT_DIR/security.json"; then
  WARNINGS=$(jq -r '.checks.security.issues | length' "$OUTPUT_DIR/security.json")
  
  if [ "$WARNINGS" -eq 0 ]; then
    echo "✓ No security issues found"
  else
    echo "⚠ Found $WARNINGS security warnings"
    jq '.checks.security.issues' "$OUTPUT_DIR/security.json"
  fi
else
  echo "✗ Security validation failed"
  exit 1
fi

# Step 3: Run Tests
echo ""
echo "[3/5] Running tests with coverage..."
if plumego test --race --cover --format json > "$OUTPUT_DIR/test-results.json"; then
  PASSED=$(jq -r '.data.passed' "$OUTPUT_DIR/test-results.json")
  TOTAL=$(jq -r '.data.tests' "$OUTPUT_DIR/test-results.json")
  COVERAGE=$(jq -r '.data.coverage_percent // 0' "$OUTPUT_DIR/test-results.json")
  DURATION=$(jq -r '.data.duration_ms' "$OUTPUT_DIR/test-results.json")
  
  echo "✓ Tests passed: $PASSED/$TOTAL"
  echo "  Coverage: ${COVERAGE}%"
  echo "  Duration: ${DURATION}ms"
  
  if [ "$PASSED" != "$TOTAL" ]; then
    echo "✗ Some tests failed"
    jq '.data.failures' "$OUTPUT_DIR/test-results.json"
    exit 1
  fi
else
  echo "✗ Tests failed"
  jq '.data' "$OUTPUT_DIR/test-results.json"
  exit 1
fi

# Step 4: Build
echo ""
echo "[4/5] Building application..."
if plumego build \
  --output "$OUTPUT_DIR/app" \
  --trimpath \
  --ldflags "-s -w" \
  --format json > "$OUTPUT_DIR/build-info.json"; then
  
  SIZE=$(jq -r '.data.size_mb' "$OUTPUT_DIR/build-info.json")
  BUILD_TIME=$(jq -r '.data.build_time_ms' "$OUTPUT_DIR/build-info.json")
  GIT_COMMIT=$(jq -r '.data.git_commit' "$OUTPUT_DIR/build-info.json")
  
  echo "✓ Build complete"
  echo "  Binary: $OUTPUT_DIR/app"
  echo "  Size: ${SIZE} MB"
  echo "  Build time: ${BUILD_TIME}ms"
  echo "  Git commit: $GIT_COMMIT"
else
  echo "✗ Build failed"
  exit 1
fi

# Step 5: Generate Reports
echo ""
echo "[5/5] Generating reports..."

# Create summary
cat > "$OUTPUT_DIR/summary.txt" << SUMMARY
Plumego CI/CD Pipeline Summary
==============================

Project: $PROJECT_DIR
Date: $(date)

Health Check: PASSED
Security: $([ "$WARNINGS" -eq 0 ] && echo "PASSED" || echo "WARNINGS")
Tests: $PASSED/$TOTAL passed (${COVERAGE}% coverage)
Build: SUCCESS (${SIZE} MB, ${BUILD_TIME}ms)
Git Commit: $GIT_COMMIT

Artifacts:
- health.json
- security.json
- test-results.json
- build-info.json
- app (binary)
SUMMARY

cat "$OUTPUT_DIR/summary.txt"

echo ""
echo "====================================="
echo "Pipeline completed successfully!"
echo "====================================="
echo "All artifacts saved to: $OUTPUT_DIR"
