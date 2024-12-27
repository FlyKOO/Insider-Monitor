# Contributing to Insider-Monitor

Thank you for your interest in contributing to Insider-Monitor! This document provides guidelines and workflows for contributing to this project.

## Development Workflow

### 1. Branching Strategy
- `main`: Production-ready code
- Feature branches: `feature/your-feature-name`
- Bug fix branches: `fix/bug-description`
- Hotfix branches: `hotfix/urgent-fix`

### 2. Making Changes
1. Fork the repository
2. Create a new branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Make your changes
4. Commit using conventional commit messages:
   ```bash
   git commit -m "feat: add new feature"
   git commit -m "fix: resolve bug"
   git commit -m "docs: update documentation"
   ```
5. Push your branch and create a Pull Request

### 3. Pull Request Process
1. Ensure all tests pass
2. Update documentation if needed
3. Request review from maintainers
4. Address review comments
5. Maintainers will merge after approval

## Release Process

Releases follow semantic versioning (MAJOR.MINOR.PATCH):
- MAJOR: Breaking changes
- MINOR: New features (backwards compatible)
- PATCH: Bug fixes

### Creating a Release
1. Ensure all changes are merged to `main`
2. Create and push a new tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. GitHub Actions will automatically:
   - Run tests
   - Build binaries
   - Create a GitHub release
   - Upload artifacts

## Code Standards
- Follow Go best practices and idioms
- Maintain test coverage
- Use meaningful variable and function names
- Document public APIs
- Run linter before committing 