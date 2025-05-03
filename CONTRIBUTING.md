# Contributing to DocChain

Thank you for considering contributing to DocChain! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and considerate of others.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with the following information:

1. A clear, descriptive title
2. A detailed description of the issue
3. Steps to reproduce the bug
4. Expected behavior
5. Actual behavior
6. Screenshots (if applicable)
7. Environment information (OS, Go version, etc.)

### Suggesting Enhancements

We welcome suggestions for enhancements! Please create an issue with:

1. A clear, descriptive title
2. A detailed description of the proposed enhancement
3. Any relevant examples or mockups
4. Why this enhancement would be useful

### Pull Requests

We actively welcome pull requests:

1. Fork the repository
2. Create a new branch from `main`
3. Make your changes
4. Run tests and ensure they pass
5. Update documentation if necessary
6. Submit a pull request

#### Pull Request Guidelines

- Follow the Go style guide and code conventions
- Write clear, descriptive commit messages
- Include tests for new features or bug fixes
- Update documentation as needed
- Keep pull requests focused on a single topic

## Development Setup

1. Clone the repository:
   ```
   git clone https://github.com/raykavin/docchain.git
   cd docchain
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Run the development script:
   ```
   ./dev.sh
   ```

## Code Style

- Follow standard Go conventions and the [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` to format your code
- Write clear comments and documentation
- Use meaningful variable and function names

## Testing

- Write tests for all new features and bug fixes
- Ensure all tests pass before submitting a pull request
- Run tests with:
  ```
  make test
  ```

## Documentation

- Update documentation for any changes to the API, features, or behavior
- Use clear, concise language
- Include examples where appropriate

## License

By contributing to DocChain, you agree that your contributions will be licensed under the project's MIT license.
