# Plumego CLI Examples

This directory contains example scripts demonstrating real-world usage of the plumego CLI.

## Scripts

### create-and-test.sh

Complete workflow for creating, testing, and building a new project.

**Usage:**
```bash
./create-and-test.sh <project-name> [template]

# Examples
./create-and-test.sh myapp api
./create-and-test.sh webapp fullstack
```

**What it does:**
1. Creates a new project with specified template
2. Initializes Go dependencies
3. Runs health checks
4. Generates additional components
5. Runs tests with coverage
6. Builds the application
7. Reports summary

**Output:**
- Project directory with full structure
- `health.json` - Health check results
- `test-results.json` - Test results with coverage
- `build-info.json` - Build metadata
- `bin/<project-name>` - Compiled binary

### ci-pipeline.sh

Production-ready CI/CD pipeline for automated testing and deployment.

**Usage:**
```bash
./ci-pipeline.sh [project-dir] [output-dir]

# Examples
./ci-pipeline.sh . ./ci-output
./ci-pipeline.sh /path/to/project ./artifacts
```

**What it does:**
1. **Health Check** - Validates project structure and dependencies
2. **Security Validation** - Checks for security issues
3. **Tests** - Runs tests with race detector and coverage
4. **Build** - Compiles optimized binary
5. **Report Generation** - Creates summary report

**Output:**
```
ci-output/
├── health.json          # Health check results
├── security.json        # Security scan results
├── test-results.json    # Test results with coverage
├── build-info.json      # Build metadata
├── app                  # Compiled binary
└── summary.txt          # Pipeline summary
```

**CI Integration:**

GitHub Actions:
```yaml
- name: Run CI Pipeline
  run: |
    chmod +x cmd/plumego/examples/ci-pipeline.sh
    ./cmd/plumego/examples/ci-pipeline.sh . ./artifacts

- name: Upload Artifacts
  uses: actions/upload-artifact@v3
  with:
    name: pipeline-results
    path: artifacts/
```

GitLab CI:
```yaml
ci:
  script:
    - chmod +x cmd/plumego/examples/ci-pipeline.sh
    - ./cmd/plumego/examples/ci-pipeline.sh . ./artifacts
  artifacts:
    paths:
      - artifacts/
```

### dev-workflow.sh

Interactive menu for common development tasks.

**Usage:**
```bash
./dev-workflow.sh [project-dir]

# Examples
./dev-workflow.sh
./dev-workflow.sh /path/to/project
```

**Features:**
- Start development server with hot reload
- Run tests (with/without coverage)
- Generate components, handlers, middleware
- Build application
- Check project health
- Analyze routes

**Screenshot:**
```
=====================================
Plumego Development Workflow
=====================================
Project: myapp

1) Start dev server
2) Run tests
3) Run tests with coverage
4) Generate component
5) Generate handler
6) Generate middleware
7) Build application
8) Check project health
9) Analyze routes
0) Exit

Select option:
```

## Usage Patterns

### For Local Development

```bash
# Start interactive workflow
./dev-workflow.sh

# Or use individual commands
plumego dev &
plumego test --watch
```

### For CI/CD

```bash
# Run full pipeline
./ci-pipeline.sh . ./artifacts

# Check exit code
if [ $? -eq 0 ]; then
  echo "Pipeline passed"
  # Deploy artifacts
else
  echo "Pipeline failed"
  exit 1
fi
```

### For Project Scaffolding

```bash
# Create and setup new project
./create-and-test.sh my-new-api api

# Navigate to project
cd my-new-api

# Start development
plumego dev
```

## Exit Codes

All scripts follow standard exit code conventions:

- `0` - Success
- `1` - Failure (tests failed, build failed, etc.)
- `2` - Configuration error
- `130` - User interrupted (Ctrl+C)

## Prerequisites

- `plumego` CLI installed and in PATH
- `jq` for JSON processing
- Go 1.24+ installed

## Tips

### Parsing JSON Output

```bash
# Extract specific fields
COVERAGE=$(plumego test --cover --format json | jq -r '.data.coverage_percent')

# Filter routes
plumego routes --format json | jq '.data.routes[] | select(.method == "GET")'

# Check health status
STATUS=$(plumego check --format json | jq -r '.checks.security.status')
```

### Chaining Commands

```bash
# Build and run tests only if health check passes
plumego check && plumego test && plumego build

# Generate multiple components
for component in Auth Database Cache; do
  plumego generate component $component --with-tests
done
```

### Environment Variables

```bash
# Override defaults
APP_ADDR=:3000 plumego dev

# Custom env file for config inspection
plumego --env-file .env.prod config show
```

## Contributing

Have a useful script or workflow? Please contribute!

1. Add your script to this directory
2. Make it executable: `chmod +x your-script.sh`
3. Document it in this README
4. Submit a pull request

## License

Same as plumego core - see [LICENSE](../../../LICENSE).
