# Contributing to Forge

Thank you for your interest in contributing to Forge! This document provides guidelines and instructions for contributing.

## Project Status

**Forge is in early development.** The architecture and core concepts are still being established. To maintain consistency and avoid architectural drift during this critical phase:

- **Discuss before implementing:** For features or significant changes, please open an issue or discussion first
- **Small contributions welcome:** Bug fixes, documentation improvements, and tests are always appreciated
- **Architectural changes selective:** Major architectural decisions will be carefully reviewed to ensure they align with the project's vision
- **Patience appreciated:** We may defer some contributions until the architecture stabilizes

This approach ensures Forge develops a solid foundation. As the project matures, contribution guidelines will become more relaxed.

## Code of Conduct

Be respectful and constructive. We're all here to build something useful together.

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Git
- Make (optional, but recommended)

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/forge.git
   cd forge
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/andre-koe/forge.git
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```

## Development Workflow

### Building

```bash
make build
```

The binary will be in `bin/forge`.

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific tests
go test ./cmd -v -run TestRunRun
```

### Code Quality

Before submitting a PR, ensure your code passes all checks:

```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Run all quality checks
make check
```

### Running the Application

```bash
# Build and run
make run

# Or run directly
./bin/forge --help
```

## Making Changes

### Branching Strategy

- Create a feature branch from `main`:
  ```bash
  git checkout -b feature/your-feature-name
  ```
- Use descriptive branch names:
  - `feature/` for new features
  - `fix/` for bug fixes
  - `docs/` for documentation
  - `test/` for test improvements

### Commit Messages

Write clear, concise commit messages:

- Use the imperative mood ("Add feature" not "Added feature")
- Keep the first line under 50 characters
- Add a blank line, then a detailed description if needed

Example:
```
Add workflow validation command

Implements a new 'validate' command that checks workflow
YAML syntax and validates the DAG structure before execution.
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` and `goimports` (run `make fmt`)
- Write meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions small and focused

### Testing Requirements

- Write tests for all new features
- Maintain or improve test coverage
- Use table-driven tests where appropriate
- Mock external dependencies using the functional options pattern

Example test structure:
```go
func TestYourFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr error
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Submitting a Pull Request

**Before submitting a PR for new features, please open an issue first to discuss the approach.**

1. Ensure all tests pass and code is formatted:
   ```bash
   make check test
   ```

2. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

3. Open a Pull Request on GitHub with:
   - Clear title describing the change
   - Description of what changed and why
   - Link to the related issue (required for features)
   - Screenshots/examples if relevant
   - Explanation of how it fits the project architecture

4. Wait for review and address any feedback

### PR Review Process

- A maintainer will review your PR
- Address requested changes by pushing new commits
- Once approved, a maintainer will merge your PR

## Project Structure

```
forge/
├── cmd/               # CLI commands
│   ├── forge/        # Main entry point
│   ├── run.go        # Run command
│   ├── dry_run.go    # Dry-run command
│   ├── init.go       # Init command
│   └── version.go    # Version command
├── internal/
│   ├── dsl/          # Workflow DSL definitions
│   └── runner/       # Workflow execution engine
├── config/           # Configuration handling
└── workflows/        # Example workflows
```

## Areas for Contribution

### Encouraged Contributions (No Prior Discussion Needed)

- Documentation improvements and typo fixes
- Test coverage improvements
- Error message enhancements
- Example workflows
- Bug fixes for existing functionality
- Code comments and clarifications

### Feature Contributions (Discuss First)

**Please open an issue before working on these:**

- Additional workflow step types
- Parallel stage execution
- Workflow validation command
- Configuration file support
- Shell completion scripts
- Changes to the DSL structure
- New CLI commands
- Changes to the runner architecture

Opening an issue first allows us to:
- Ensure the feature aligns with the project vision
- Discuss implementation approaches
- Avoid duplicate work
- Save your time if the feature isn't ready to be implemented

### Bug Reports

Found a bug? Please open an issue with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Go version)
- Relevant logs or error messages

## Questions?

- Open a GitHub Discussion for questions
- Check existing issues and PRs
- Read the README for basic usage

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

Thank you for contributing to Forge! 🔨
