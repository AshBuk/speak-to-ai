# Contributing to Speak-to-AI

Thank you for your interest to Speak-to-AI! This document provides guidelines for contributing to the project.

## Code of Conduct

- Respect the chosen stack -> Go and Whisper.cpp as AI model
- Be respectful, focus on constructive
- Help others learn and grow
- Maintain a welcoming environment

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR-USERNAME/speak-to-ai.git`
3. Follow the development setup in [DEVELOPMENT.md](DEVELOPMENT.md)

4. **Dev Workflow**
   1. Create a feature branch: `git checkout -b feature/your-feature-name`
   2. Make your changes
   3. Add license headers to new Go files
   ```go
   // Copyright (c) 2025 Asher Buk
   // SPDX-License-Identifier: MIT
   ```
   4. Commit with clear message
   5. Push and create a Pull Request

5. **Code Style**

- **Formatting:** All Go code must be formatted with `gofmt` — run `make fmt`
- **Linting:** Code must pass `golangci-lint` checks — run `make lint`
- **Testing:** New features should include appropriate tests — run `make test`
- **Build:** Changes must not break the build process — run `make build`

**Note:** Our CI automatically validates:
- Lint rules (`golangci-lint`)
- Code formatting (`gofmt`) 
- Build process
- Unit tests
- License headers in all Go files

PRs must pass all checks before merge.

## 🐛 Bug Reports

When reporting bugs, include:

- Create an issue
- Operating system and version
- Desktop environment (GNOME, KDE, etc.)
- Display server (X11/Wayland)
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs

## 💡 Feature Requests

For new features:

- Check existing issues first
- Describe the use case
- Consider backwards compatibility
- Be specific about the desired behavior

## 📜 License

By contributing, you agree that your contributions will be licensed under the MIT License. All contributed code becomes part of the project under the same license terms.
