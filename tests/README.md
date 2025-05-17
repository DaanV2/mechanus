# Integration Tests

- [Integration Tests](#integration-tests)
  - [Overview](#overview)
  - [Setup](#setup)
    - [Quick Start](#quick-start)
    - [Manual Setup](#manual-setup)
  - [Testing](#testing)
    - [Running Tests](#running-tests)
    - [Cleaning Up](#cleaning-up)
  - [Development Workflow](#development-workflow)
  - [CI Integration](#ci-integration)
  - [Troubleshooting](#troubleshooting)
  - [Configuration](#configuration)


## Overview
This directory contains integration tests for the application using Playwright. These tests verify that the application works correctly from an end-user perspective by automating browser interactions.

## Setup

### Quick Start
For a complete setup that prepares everything for development or testing:

```bash
# For development environment
make dev-setup

# For test environment (sets up and runs tests)
make test-setup
```

### Manual Setup
If you prefer to set up components individually:

```bash
# Install dependencies:
make setup

# Build the server Docker image (from project root):
make image

# Start the server:
make server
```

## Testing

### Running Tests
To run all tests:
```bash
# Ensure you have executed something like: make test-setup

make test
```

### Cleaning Up
To clean up Docker containers and test artifacts:
```bash
# Full cleanup (removes dependencies, browser installations, and test results)
make clean
```

## Development Workflow

**initial setup**
```bash
# From project root
make image

# From tests directory
make dev-setup
```

**start local server**:
```bash
# Use Docker container (recommended for consistency with CI)
make server

# Or use local Go server (for development with hot reloading)
make local-server
```

**write and run tests**

Create or modify tests in the `tests` directory
```bash
# Run specific tests with:
npx playwright test <test-file-path>

# Run with UI mode for debugging:
npx playwright test --ui

# View Test Results
npx playwright show-report
```

## CI Integration
The tests are configured to run in GitHub Actions using the Playwright Docker image. The workflow:
1. Builds the application Docker image
2. Sets up the test environment
3. Runs all tests
4. Uploads test reports as artifacts

## Troubleshooting

- **Tests can't connect to server**: Ensure the server is running and accessible at http://127.0.0.1:8080
- **Browser installation issues**: Try running `npx playwright install` manually
- **Docker errors**: Make sure Docker is running and you have permission to create containers
- **Test failures**: Check the Playwright report for detailed error information

## Configuration

The Playwright configuration is in `playwright.config.ts`. Key settings:
- Tests run against http://127.0.0.1:8080 by default (or host.docker.internal in Docker)
- Tests are configured to run in Chromium, Firefox, WebKit, and mobile emulation
- Traces are captured on test retry for debugging

---

For more information on Playwright, visit the [official documentation](https://playwright.dev/docs/intro).