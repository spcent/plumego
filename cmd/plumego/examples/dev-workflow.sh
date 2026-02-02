#!/bin/bash
#
# dev-workflow.sh - Interactive development workflow helper
#
# Provides common development tasks in an easy-to-use menu
#

set -euo pipefail

PROJECT_DIR="${1:-.}"

cd "$PROJECT_DIR"

show_menu() {
  clear
  echo "====================================="
  echo "Plumego Development Workflow"
  echo "====================================="
  echo "Project: $(basename $(pwd))"
  echo ""
  echo "1) Start dev server"
  echo "2) Run tests"
  echo "3) Run tests with coverage"
  echo "4) Generate component"
  echo "5) Generate handler"
  echo "6) Generate middleware"
  echo "7) Build application"
  echo "8) Check project health"
  echo "9) Analyze routes"
  echo "0) Exit"
  echo ""
  echo -n "Select option: "
}

start_dev_server() {
  echo ""
  echo "Starting development server..."
  echo "Press Ctrl+C to stop"
  echo ""
  plumego dev --verbose
}

run_tests() {
  echo ""
  echo "Running tests..."
  plumego test --format text
  echo ""
  read -p "Press Enter to continue..."
}

run_tests_with_coverage() {
  echo ""
  echo "Running tests with coverage..."
  plumego test --cover --format text
  echo ""
  
  if [ -f coverage.out ]; then
    echo "Opening coverage report in browser..."
    go tool cover -html=coverage.out
  fi
  
  read -p "Press Enter to continue..."
}

generate_component() {
  echo ""
  read -p "Component name: " name
  read -p "Generate tests? (y/n): " tests
  
  if [ "$tests" = "y" ]; then
    plumego generate component "$name" --with-tests --format text
  else
    plumego generate component "$name" --format text
  fi
  
  echo ""
  read -p "Press Enter to continue..."
}

generate_handler() {
  echo ""
  read -p "Handler name: " name
  read -p "HTTP methods (comma-separated, e.g., GET,POST): " methods
  read -p "Generate tests? (y/n): " tests
  
  if [ "$tests" = "y" ]; then
    plumego generate handler "$name" --methods "$methods" --with-tests --format text
  else
    plumego generate handler "$name" --methods "$methods" --format text
  fi
  
  echo ""
  read -p "Press Enter to continue..."
}

generate_middleware() {
  echo ""
  read -p "Middleware name: " name
  read -p "Generate tests? (y/n): " tests
  
  if [ "$tests" = "y" ]; then
    plumego generate middleware "$name" --with-tests --format text
  else
    plumego generate middleware "$name" --format text
  fi
  
  echo ""
  read -p "Press Enter to continue..."
}

build_application() {
  echo ""
  echo "Building application..."
  plumego build --format text
  echo ""
  read -p "Press Enter to continue..."
}

check_health() {
  echo ""
  echo "Checking project health..."
  plumego check --format text
  echo ""
  read -p "Press Enter to continue..."
}

analyze_routes() {
  echo ""
  echo "Analyzing routes..."
  plumego routes --format text
  echo ""
  read -p "Press Enter to continue..."
}

# Main loop
while true; do
  show_menu
  read -r option
  
  case $option in
    1) start_dev_server ;;
    2) run_tests ;;
    3) run_tests_with_coverage ;;
    4) generate_component ;;
    5) generate_handler ;;
    6) generate_middleware ;;
    7) build_application ;;
    8) check_health ;;
    9) analyze_routes ;;
    0) echo "Goodbye!"; exit 0 ;;
    *) echo "Invalid option"; sleep 1 ;;
  esac
done
