# Contributing to mistral-client

Thank you for your interest in contributing to `mistral-client`! This document will help you get started with local development and explain our standards.

## 🛠️ Technical Requirements

To work on this project, you need:

- **Go**: 1.25 or higher
- **Mockgen**: For generating mocks (`go install go.uber.org/mock/mockgen@latest`)
- **UV**: For running the documentation locally (optional, if you want to preview docs)

## 💻 Local Setup

### 1. Clone the repository

```bash
git clone https://github.com/thomas-marquis/mistral-client.git
cd mistral-client
```

### 2. Install dependencies

```bash
go get .
```

### 3. Generate Mocks

We use `gomock` for testing. Mocks are generated based on the interfaces defined in the project. To generate them, run:

```bash
go generate ./...
```

This will update the files in the `mocks/` directory.

### 4. Run Tests

To ensure everything is working correctly, run the test suite:

```bash
go test ./...
```

## 📁 Project Structure

The project follows a standard Go structure:

- `mistral/`: Core package containing the client implementation and public API.
- `internal/`: Private packages used by the library (e.g., caching engines).
- `mocks/`: Generated mocks for unit testing.
- `examples/`: Runnable examples demonstrating various features.
- `docs/`: Markdown files for the project documentation.
- `tools/`: Development utilities and scripts. This folder contains tools used for development, such as the linting utility.
- `Makefile`: Shortcut commands for documentation management.
- `gen.go`: Configuration for mock generation.

## 🧪 Coding Conventions

### Test Structure

- **Packages**: Use `package mistral_test` for tests in the `mistral/` directory to ensure you are testing the public API as an external user. Use the same package name (e.g., `package cache`) for internal package tests if you need to test unexported symbols.
- **Assertions**: Use [testify/assert](https://github.com/stretchr/testify) for all assertions.
- **Mocking**: Use `gomock` to mock interfaces.
- **Structure**: One function per method/function to test. Use the `// Given`, `// When`, `Then` comments to structure tests.
- **Naming**: Test functions should be named using the following pattern: `Test<FunctionName>>` for a function, `Test<StructName>_<MethodName>` for a method.

### Linting

Before submitting a PR, make sure your code passes the linting check:

```bash
./tools/lint.sh
```

## ✅ Definition of Done

A contribution is considered complete when:

1.  **Implementation**: The code is clean, documented, and follows the existing style.
2.  **Tests**: Unit tests are added for new features or bug fixes. All tests must pass.
3.  **Lint**: The `./tools/lint.sh` script passes without errors.
4.  **Documentation**: Relevant documentation in the `docs/` folder is updated or added.
5.  **CI**: All GitHub Action workflows pass.
6.  **Example**: If you added a new feature, a corresponding example should be added in the `examples/` folder.
7.  **README**: If the change is significant, update the `README.md` to reflect it.
