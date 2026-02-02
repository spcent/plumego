#!/bin/bash
#
# create-and-test.sh - Create a new plumego project and run tests
#
# Usage: ./create-and-test.sh <project-name> [template]
#

set -euo pipefail

PROJECT_NAME="${1:-myapp}"
TEMPLATE="${2:-api}"

echo "Creating project: $PROJECT_NAME (template: $TEMPLATE)"

# Create project
OUTPUT=$(plumego new "$PROJECT_NAME" \
  --template "$TEMPLATE" \
  --format json)

# Extract project path
PROJECT_PATH=$(echo "$OUTPUT" | jq -r '.data.path')
echo "✓ Project created at: $PROJECT_PATH"

# Navigate to project
cd "$PROJECT_PATH"

# Initialize dependencies
echo "Initializing dependencies..."
go mod tidy

# Run health check
echo "Running health check..."
if plumego check --format json > health.json; then
  echo "✓ Health check passed"
else
  echo "✗ Health check failed"
  jq '.data.checks' health.json
  exit 1
fi

# Generate additional components
echo "Generating components..."
plumego generate component Database --with-tests --format json
plumego generate middleware Logger --format json
plumego generate handler Health --methods GET --with-tests --format json

# Run tests
echo "Running tests..."
if plumego test --cover --format json > test-results.json; then
  PASSED=$(jq -r '.data.passed' test-results.json)
  TOTAL=$(jq -r '.data.tests' test-results.json)
  COVERAGE=$(jq -r '.data.coverage_percent // 0' test-results.json)
  
  echo "✓ Tests passed: $PASSED/$TOTAL"
  echo "✓ Coverage: ${COVERAGE}%"
else
  echo "✗ Tests failed"
  jq '.data.failures' test-results.json
  exit 1
fi

# Build
echo "Building application..."
if plumego build --output "./bin/$PROJECT_NAME" --format json > build-info.json; then
  SIZE=$(jq -r '.data.size_mb' build-info.json)
  BUILD_TIME=$(jq -r '.data.build_time_ms' build-info.json)
  
  echo "✓ Build complete"
  echo "  Binary: ./bin/$PROJECT_NAME"
  echo "  Size: ${SIZE} MB"
  echo "  Build time: ${BUILD_TIME}ms"
else
  echo "✗ Build failed"
  exit 1
fi

echo ""
echo "Project $PROJECT_NAME is ready!"
echo "Next steps:"
echo "  cd $PROJECT_PATH"
echo "  plumego dev"
