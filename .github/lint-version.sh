#!/usr/bin/env bash

# Lint version should match what is in .github/workflows/build-and-test.yml
LINT_VERSION=v1.54.2

# If command not found then install required version
golangci-lint >/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@$LINT_VERSION

# Show the full version line
golangci-lint version

# Check version
found_version=$(golangci-lint version | head -n 1 | awk '{print $4}')
[ $found_version == $LINT_VERSION ] || echo -e "\033[33mWarning!\033[0m golangci-lint version $found_version does not match version $LINT_VERSION used in github actions"
