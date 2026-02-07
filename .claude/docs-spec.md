# Plumego Documentation Specification

## Documentation Structure

### 1. README.md (Root)
- Project overview
- Quick start (< 5 minutes)
- Installation
- Basic usage example
- Link to detailed docs

### 2. docs/ Directory Structure
```
docs/
├── getting-started.md
├── core-concepts.md
├── api-reference/
│   ├── app.md
│   ├── router.md
│   ├── context.md
│   ├── middleware.md
│   └── response.md
├── guides/
│   ├── routing.md
│   ├── middleware.md
│   └── error-handling.md
└── advanced/
    └── [list advanced topics]
```

## Documentation Style Guide

### Tone & Voice
- Clear and concise
- Beginner-friendly but not condescending
- Show code examples, not just descriptions
- Assume reader knows Go basics

### Code Examples
- Must be runnable
- Include imports
- Show full context (not fragments)
- Use realistic variable names
- Add comments only for non-obvious parts

### Structure for Each Module
1. **Purpose**: What problem does it solve? (2-3 sentences)
2. **Quick Example**: Working code (< 20 lines)
3. **API Reference**: Key types and functions
4. **Advanced Usage**: Optional features
5. **Common Patterns**: Real-world use cases

### Formatting Rules
- Use `code` for symbols
- Use ```go blocks for examples
- Keep paragraphs short (< 4 lines)
- Use bullet points for lists
- Add "See also" links to related docs

## Module Documentation Priority

### Must Document (Core Modules)
1. Router - routing and path matching
2. Context - request/response handling  
3. Middleware - middleware chain
4. App - main application setup

### Should Document (Supporting Modules)
5. Response - response helpers
6. Group - route grouping
7. Error handling
8. Utilities

### Nice to Have
- Performance tips
- Migration guides
- Comparison with other frameworks