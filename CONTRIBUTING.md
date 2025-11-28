# Contributing to ProtoDiff

Thank you for your interest in contributing to ProtoDiff! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Guidelines](#coding-guidelines)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and constructive in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/uzdada/protodiff.git
   cd protodiff
   ```
3. **Add the upstream repository**:
   ```bash
   git remote add upstream https://github.com/uzdada/protodiff.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker
- kubectl
- Access to a Kubernetes cluster (minikube, kind, or cloud provider)
- golangci-lint (optional, for linting)

### Install Dependencies

```bash
make deps
```

### Build the Project

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Locally

```bash
# Ensure your kubeconfig is configured
make run
```

## How to Contribute

### Types of Contributions

We welcome various types of contributions:

- **Bug fixes**: Fix issues reported in GitHub Issues
- **Features**: Add new functionality (please discuss in an issue first)
- **Documentation**: Improve README, code comments, or add guides
- **Tests**: Add or improve test coverage
- **Examples**: Add example configurations or use cases

### Workflow

1. **Create a new branch** for your work:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding guidelines

3. **Commit your changes** with clear, descriptive messages:
   ```bash
   git commit -m "Add feature: description of what you did"
   ```

4. **Keep your fork synchronized** with upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request** on GitHub

## Coding Guidelines

### Go Code Style

- Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code (automatically done by `make fmt`)
- Use meaningful variable and function names
- Keep functions focused and concise
- Add comments for exported functions and complex logic

### Project Structure

Follow the existing project structure:

```
internal/
├── core/           # Business logic and domain models
├── adapters/       # External integrations (k8s, grpc, web, bsr)
└── scanner/        # Orchestration logic
```

### Comments and Documentation

- **All code comments must be in English** to maintain consistency for the global community
- Use godoc-style comments for exported types and functions:
  ```go
  // FetchSchema retrieves the schema definition from BSR for a given module.
  // It returns an error if the module is not found or network issues occur.
  func FetchSchema(ctx context.Context, module string) (*Schema, error) {
      // implementation
  }
  ```

- Add inline comments for complex logic:
  ```go
  // Calculate checksum using FNV-1a hash for performance
  hash := fnv.New32a()
  ```

### Error Handling

- Always handle errors appropriately
- Wrap errors with context using `fmt.Errorf`:
  ```go
  if err != nil {
      return fmt.Errorf("failed to fetch schema from %s: %w", address, err)
  }
  ```

### Logging

- Use structured logging where possible
- Log at appropriate levels:
  - `log.Printf()` for informational messages
  - `log.Fatalf()` for fatal errors (use sparingly)
- Include context in log messages:
  ```go
  log.Printf("Validated %s/%s: %s", pod.Namespace, pod.Name, result.Status)
  ```

## Testing

### Writing Tests

- Write unit tests for new functions
- Place test files alongside source files: `foo.go` → `foo_test.go`
- Use table-driven tests for multiple test cases:
  ```go
  func TestFetchSchema(t *testing.T) {
      tests := []struct {
          name    string
          module  string
          wantErr bool
      }{
          {"valid module", "buf.build/acme/user", false},
          {"invalid module", "invalid", true},
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              // test implementation
          })
      }
  }
  ```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/core/store/
```

## Pull Request Process

1. **Ensure all tests pass** before submitting
2. **Update documentation** if you've changed functionality
3. **Add tests** for new features
4. **Keep PRs focused**: One feature/fix per PR
5. **Describe your changes** in the PR description:
   - What problem does this solve?
   - How did you solve it?
   - Any breaking changes?
   - Screenshots (if UI changes)

### PR Title Format

Use conventional commit format:

- `feat: add support for custom gRPC ports`
- `fix: resolve race condition in store`
- `docs: update installation guide`
- `test: add tests for scanner package`
- `refactor: simplify schema comparison logic`

### Review Process

- At least one maintainer will review your PR
- Address review comments by pushing new commits
- Once approved, a maintainer will merge your PR

## Reporting Bugs

When reporting bugs, please include:

1. **ProtoDiff version**: `protodiff --version` (or Docker image tag)
2. **Kubernetes version**: `kubectl version`
3. **Go version** (if building from source): `go version`
4. **Steps to reproduce** the issue
5. **Expected behavior** vs **actual behavior**
6. **Logs** (if applicable):
   ```bash
   kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff
   ```
7. **Configuration** (ConfigMap, environment variables)

## Feature Requests

We welcome feature requests! Please:

1. **Search existing issues** to avoid duplicates
2. **Describe the feature** clearly
3. **Explain the use case** and why it's valuable
4. **Provide examples** of how it would work

## Questions?

If you have questions:

- Check the [README.md](README.md)
- Search [existing issues](https://github.com/uzdada/protodiff/issues)
- Open a new issue with the "question" label

## License

By contributing to ProtoDiff, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to ProtoDiff! 🎉
